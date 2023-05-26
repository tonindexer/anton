package fetcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) getTransaction(ctx context.Context, b *ton.BlockIDExt, id ton.TransactionShortInfo) (*core.Transaction, error) {
	addr := address.NewAddress(0, byte(b.Workchain), id.Account)

	tx, err := s.API.GetTransaction(ctx, b, addr, id.LT)
	if err != nil {
		return nil, errors.Wrapf(err, "get transaction (workchain = %d, seq = %d, addr = %s, lt = %d)",
			b.Workchain, b.SeqNo, addr.String(), id.LT)
	}

	return mapTransaction(b, tx)
}

func (s *Service) getTransactions(ctx context.Context, b *ton.BlockIDExt, ids []ton.TransactionShortInfo) ([]*core.Transaction, error) {
	var wg sync.WaitGroup

	type ret struct {
		tx  *core.Transaction
		err error
	}

	defer app.TimeTrack(time.Now(), fmt.Sprintf("getTransactions(%d, %d)", b.Workchain, b.SeqNo))

	ch := make(chan ret, len(ids))

	wg.Add(len(ids))

	for i := range ids {
		go func(id ton.TransactionShortInfo) {
			defer wg.Done()
			tx, err := s.getTransaction(ctx, b, id)
			ch <- ret{tx: tx, err: err}
		}(ids[i])
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	results := make([]*core.Transaction, 0, len(ids))
	for r := range ch {
		if r.err != nil {
			return nil, errors.Wrapf(r.err, "get transaction")
		}
		results = append(results, r.tx)
	}

	return results, nil
}

func (s *Service) BlockTransactions(ctx context.Context, b *ton.BlockIDExt) ([]*core.Transaction, error) {
	var (
		after        *ton.TransactionID3
		fetchedIDs   []ton.TransactionShortInfo
		transactions []*core.Transaction
		more         = true
		err          error
	)

	defer app.TimeTrack(time.Now(), fmt.Sprintf("BlockTransactions(%d, %d)", b.Workchain, b.SeqNo))

	for more {
		fetchedIDs, more, err = s.API.GetBlockTransactionsV2(ctx, b, 100, after)
		if err != nil {
			return nil, errors.Wrapf(err, "get b transactions (workchain = %d, seq = %d)",
				b.Workchain, b.SeqNo)
		}
		if more {
			after = fetchedIDs[len(fetchedIDs)-1].ID3()
		}

		rawTx, err := s.getTransactions(ctx, b, fetchedIDs)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, rawTx...)
	}

	if err := s.getTxAccounts(ctx, b, transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

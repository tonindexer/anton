package fetcher

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) getTransaction(ctx context.Context, master, b *ton.BlockIDExt, id ton.TransactionShortInfo) (*core.Transaction, error) {
	var tx *core.Transaction
	var acc *core.AccountState

	type ret struct {
		res any
		err error
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	a := address.NewAddress(0, byte(b.Workchain), id.Account)

	var wg sync.WaitGroup
	wg.Add(2)

	txCh := make(chan ret)
	go func() {
		rawTx, err := s.API.GetTransaction(ctx, b, a, id.LT)
		if err == nil {
			tx, err := mapTransaction(b, rawTx)
			txCh <- ret{res: tx, err: errors.Wrapf(err, "map transaction (hash = %x)", rawTx.Hash)}
			return
		}
		txCh <- ret{
			err: errors.Wrapf(err, "get transaction (workchain = %d, seq = %d, addr = %s, lt = %d)",
				b.Workchain, b.SeqNo, a.String(), id.LT),
		}
	}()

	accCh := make(chan ret)
	go func() {
		acc, err := s.getAccount(ctx, master, b, *addr.MustFromTonutils(a))
		accCh <- ret{res: acc, err: errors.Wrapf(err, "get account (addr = %s)", a)}
	}()

	go func() {
		wg.Wait()
		close(txCh)
		close(accCh)
	}()

	for i := 0; i < 2; i++ {
		select {
		case accRet := <-accCh:
			if err := accRet.err; err != nil && !errors.Is(err, core.ErrNotFound) {
				return nil, err
			}
			if accRet.err == nil && accRet.res != nil {
				acc = accRet.res.(*core.AccountState) //nolint:forcetypeassert // that's ok
			}

		case txRet := <-txCh:
			if err := txRet.err; err != nil {
				return nil, err
			}
			tx = txRet.res.(*core.Transaction) //nolint:forcetypeassert // that's ok
		}
	}

	tx.Account = acc
	if tx.Account != nil {
		tx.Account.UpdatedAt = tx.CreatedAt
		if tx.InMsg != nil {
			tx.InMsg.DstState = tx.Account
		}
		for _, out := range tx.OutMsg {
			out.SrcState = tx.Account
		}
	}
	return tx, nil
}

func (s *Service) getTransactions(ctx context.Context, master, b *ton.BlockIDExt, ids []ton.TransactionShortInfo) ([]*core.Transaction, error) {
	var wg sync.WaitGroup

	type ret struct {
		tx  *core.Transaction
		err error
	}

	defer core.Timer(time.Now(), "getTransactions(%d, %d)", b.Workchain, b.SeqNo)

	ch := make(chan ret, len(ids))

	wg.Add(len(ids))

	for i := range ids {
		go func(id ton.TransactionShortInfo) {
			defer wg.Done()
			tx, err := s.getTransaction(ctx, master, b, id)
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

func (s *Service) fetchTxIDs(ctx context.Context, b *ton.BlockIDExt, after *ton.TransactionID3) ([]ton.TransactionShortInfo, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return s.API.GetBlockTransactionsV2(ctx, b, 100, after)
}

func (s *Service) BlockTransactions(ctx context.Context, master, b *ton.BlockIDExt) ([]*core.Transaction, error) {
	var (
		after        *ton.TransactionID3
		fetchedIDs   []ton.TransactionShortInfo
		transactions []*core.Transaction
		more         = true
		err          error
	)

	defer core.Timer(time.Now(), "BlockTransactions(%d, %d)", b.Workchain, b.SeqNo)

	for more {
		fetchedIDs, more, err = s.fetchTxIDs(ctx, b, after)
		if err != nil {
			return nil, errors.Wrapf(err, "get block transactions (workchain = %d, seq = %d)", b.Workchain, b.SeqNo)
		}
		if more {
			after = fetchedIDs[len(fetchedIDs)-1].ID3()
		}

		rawTx, err := s.getTransactions(ctx, master, b, fetchedIDs)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, rawTx...)
	}

	return transactions, nil
}

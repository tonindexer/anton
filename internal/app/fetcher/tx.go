package fetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) FetchBlockTransactions(ctx context.Context, b *ton.BlockIDExt) ([]*core.Transaction, error) {
	var (
		after        *ton.TransactionID3
		fetchedIDs   []ton.TransactionShortInfo
		transactions []*tlb.Transaction
		more         = true
		err          error
	)

	defer app.TimeTrack(time.Now(), fmt.Sprintf("FetchBlockTransactions(%d, %d)", b.Workchain, b.SeqNo))

	for more {
		fetchedIDs, more, err = s.API.GetBlockTransactionsV2(ctx, b, 100, after)
		if err != nil {
			return nil, errors.Wrapf(err, "get b transactions (workchain = %d, seq = %d)",
				b.Workchain, b.SeqNo)
		}
		if more {
			after = fetchedIDs[len(fetchedIDs)-1].ID3()
		}

		for _, id := range fetchedIDs {
			addr := address.NewAddress(0, byte(b.Workchain), id.Account)

			tx, err := s.API.GetTransaction(ctx, b, addr, id.LT)
			if err != nil {
				return nil, errors.Wrapf(err, "get transaction (workchain = %d, seq = %d, addr = %s, lt = %d)",
					b.Workchain, b.SeqNo, addr.String(), id.LT)
			}

			transactions = append(transactions, tx)
		}
	}

	ret, err := mapTransactions(b, transactions)
	if err != nil {
		return nil, errors.Wrap(err, "parse block transactions")
	}

	return ret, nil
}

package parser

import (
	"context"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

func (s *Service) GetBlockTransactions(ctx context.Context, b *tlb.BlockInfo) ([]*tlb.Transaction, error) {
	var (
		after        *tlb.TransactionID
		fetchedIDs   []*tlb.TransactionID
		transactions []*tlb.Transaction
		more         = true
		err          error
	)

	for more {
		fetchedIDs, more, err = s.api.GetBlockTransactions(ctx, b, 100, after)
		if err != nil {
			return nil, errors.Wrapf(err, "get b transactions (workchain = %d, seq = %d)",
				b.Workchain, b.SeqNo)
		}
		if more {
			after = fetchedIDs[len(fetchedIDs)-1]
		}

		for _, id := range fetchedIDs {
			addr := address.NewAddress(0, byte(b.Workchain), id.AccountID)

			tx, err := s.api.GetTransaction(ctx, b, addr, id.LT)
			if err != nil {
				return nil, errors.Wrapf(err, "get transaction (workchain = %d, seq = %d, addr = %s, lt = %d)",
					b.Workchain, b.SeqNo, addr.String(), id.LT)
			}

			transactions = append(transactions, tx)
		}
	}

	return transactions, nil
}

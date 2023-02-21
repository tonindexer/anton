package indexer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) fetchBlockTransactions(ctx context.Context, b *tlb.BlockInfo) ([]*tlb.Transaction, error) {
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

func (s *Service) processBlockTransactions(ctx context.Context, tx bun.Tx, shard *tlb.BlockInfo) error {
	var accounts []*core.AccountState
	var accountData []*core.AccountData

	blockTx, err := s.fetchBlockTransactions(ctx, shard)
	if err != nil {
		return errors.Wrap(err, "get block transactions")
	}

	transactions, err := mapTransactions(shard, blockTx)
	if err != nil {
		return errors.Wrap(err, "parse block transactions")
	}

	accountMap, accountDataMap, err := s.processTxAccounts(ctx, shard, transactions)
	if err != nil {
		return errors.Wrap(err, "process tx accounts")
	}

	messages, err := s.processBlockMessages(ctx, tx, shard, blockTx)
	if err != nil {
		return errors.Wrap(err, "parse block messages")
	}

	payloads := s.parseMessagePayloads(ctx, messages, accountMap)

	for _, st := range accountMap {
		accounts = append(accounts, st)
	}
	for _, st := range accountDataMap {
		accountData = append(accountData, st)
	}

	if err := s.accountRepo.AddAccountStates(ctx, tx, accounts); err != nil {
		return errors.Wrap(err, "add account states")
	}
	if err := s.accountRepo.AddAccountData(ctx, tx, accountData); err != nil {
		return errors.Wrap(err, "add account data")
	}
	if err := s.txRepo.AddMessagePayloads(ctx, tx, payloads); err != nil {
		return errors.Wrap(err, "add message payloads")
	}
	if err := s.txRepo.AddMessages(ctx, tx, messages); err != nil {
		return errors.Wrap(err, "add messages")
	}
	if err := s.txRepo.AddTransactions(ctx, tx, transactions); err != nil {
		return errors.Wrap(err, "add transactions")
	}

	return nil
}

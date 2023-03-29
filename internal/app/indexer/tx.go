package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) fetchBlockTransactions(ctx context.Context, b *ton.BlockIDExt) ([]*tlb.Transaction, error) {
	var (
		txID         *ton.TransactionID3
		after        *ton.TransactionShortInfo
		fetchedIDs   []ton.TransactionShortInfo
		transactions []*tlb.Transaction
		more         = true
		err          error
	)

	defer timeTrack(time.Now(), fmt.Sprintf("fetchBlockTransactions(%d, %d)", b.Workchain, b.SeqNo))

	for more {
		if after != nil {
			txID = &ton.TransactionID3{
				Account: after.Account,
				LT:      after.LT,
			}
		}

		fetchedIDs, more, err = s.api.GetBlockTransactionsV2(ctx, b, 100, txID)
		if err != nil {
			return nil, errors.Wrapf(err, "get b transactions (workchain = %d, seq = %d)",
				b.Workchain, b.SeqNo)
		}
		if more {
			after = &fetchedIDs[len(fetchedIDs)-1]
		}

		for _, id := range fetchedIDs {
			addr := address.NewAddress(0, byte(b.Workchain), id.Account)

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

func (s *Service) processBlockTransactions(ctx context.Context, tx bun.Tx, shard *ton.BlockIDExt) error {
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

	payloads := s.parseMessagePayloads(ctx, tx, messages, accountDataMap)

	for _, st := range accountMap {
		accounts = append(accounts, st)
	}
	for _, st := range accountDataMap {
		accountData = append(accountData, st)
	}

	defer timeTrack(time.Now(), fmt.Sprintf("add account and transactions data(%d, %d)", shard.Workchain, shard.SeqNo))

	if err := s.accountRepo.AddAccountStates(ctx, tx, accounts); err != nil {
		return errors.Wrap(err, "add account states")
	}
	if err := s.accountRepo.AddAccountData(ctx, tx, accountData); err != nil {
		return errors.Wrap(err, "add account data")
	}
	if err := s.msgRepo.AddMessagePayloads(ctx, tx, payloads); err != nil {
		return errors.Wrap(err, "add message payloads")
	}
	if err := s.msgRepo.AddMessages(ctx, tx, messages); err != nil {
		return errors.Wrap(err, "add messages")
	}
	if err := s.txRepo.AddTransactions(ctx, tx, transactions); err != nil {
		return errors.Wrap(err, "add transactions")
	}

	return nil
}

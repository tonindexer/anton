package indexer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

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

func (s *Service) processBlockTransactions(ctx context.Context, master, shard *tlb.BlockInfo) error {
	var (
		accounts     []*core.AccountState
		accountMap   = make(map[string]*core.AccountState)
		accountsData []*core.AccountData
	)

	blockTx, err := s.fetchBlockTransactions(ctx, shard)
	if err != nil {
		return errors.Wrap(err, "get block transactions")
	}

	transactions, err := mapTransactions(ctx, shard, blockTx)
	if err != nil {
		return errors.Wrap(err, "parse block transactions")
	}

	for _, tx := range transactions {
		addr := address.MustParseAddr(tx.Address)

		raw, err := s.api.GetAccount(ctx, master, addr)
		if err != nil {
			return errors.Wrapf(err, "get account (%s)", addr.String())
		}
		acc := mapAccount(raw)

		accTypes, err := s.abiRepo.DetermineContractInterfaces(ctx, raw)
		if err != nil {
			return errors.Wrapf(err, "determine contract interfaces")
		}
		for _, t := range accTypes {
			acc.Types = append(acc.Types, string(t))
		}

		accounts = append(accounts, acc)
		if addr.Type() == address.StdAddress {
			accountMap[acc.Address] = acc
		}

		data, err := s.parser.ParseAccountData(ctx, master, raw)
		if err != nil && !errors.Is(err, core.ErrNotAvailable) {
			log.Error().Err(err).Str("addr", tx.Address).Msg("parse account data")
			continue
		}
		if err == nil {
			accountsData = append(accountsData, data)
		}
	}

	messages, err := s.processBlockMessages(ctx, shard, blockTx)
	if err != nil {
		return errors.Wrap(err, "parse block messages")
	}

	payloads := s.parseMessagePayloads(ctx, messages, accountMap)

	// TODO: do not insert duplicated accounts and account data
	if err := s.accountRepo.AddAccountStates(ctx, accounts); err != nil {
		return errors.Wrap(err, "add accounts")
	}
	if err := s.accountRepo.AddAccountData(ctx, accountsData); err != nil {
		return errors.Wrap(err, "add account data")
	}
	if err := s.txRepo.AddTransactions(ctx, transactions); err != nil {
		return errors.Wrap(err, "add transactions")
	}
	if err := s.txRepo.AddMessages(ctx, messages); err != nil {
		return errors.Wrap(err, "add messages")
	}
	if err := s.txRepo.AddMessagePayloads(ctx, payloads); err != nil {
		return errors.Wrap(err, "add message payloads")
	}

	return nil
}

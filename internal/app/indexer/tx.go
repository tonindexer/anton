package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) processBlockTransactions(ctx context.Context, tx bun.Tx, shard *ton.BlockIDExt) error {
	var (
		accounts []*core.AccountState
		messages []*core.Message
	)

	blockTx, err := s.Fetcher.FetchBlockTransactions(ctx, shard)
	if err != nil {
		return errors.Wrap(err, "get block transactions")
	}

	for i := range blockTx {
		accounts = append(accounts, blockTx[i].Account)
	}

	for i := range blockTx {
		messages = append(messages, blockTx[i].InMsg)
		messages = append(messages, blockTx[i].OutMsg...)
	}

	defer app.TimeTrack(time.Now(), fmt.Sprintf("add account and transactions data(%d, %d)", shard.Workchain, shard.SeqNo))

	if err := s.accountRepo.AddAccountStates(ctx, tx, accounts); err != nil {
		return errors.Wrap(err, "add account states")
	}
	if err := s.msgRepo.AddMessages(ctx, tx, messages); err != nil {
		return errors.Wrap(err, "add messages")
	}
	if err := s.txRepo.AddTransactions(ctx, tx, blockTx); err != nil {
		return errors.Wrap(err, "add transactions")
	}

	return nil
}

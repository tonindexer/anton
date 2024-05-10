package indexer

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) insertData(
	ctx context.Context,
	acc []*core.AccountState,
	msg []*core.Message,
	tx []*core.Transaction,
	b []*core.Block,
) error {
	dbTx, err := s.DB.PG.Begin()
	if err != nil {
		return errors.Wrap(err, "cannot begin db tx")
	}
	defer func() {
		_ = dbTx.Rollback()
	}()

	for _, message := range msg {
		err := s.Parser.ParseMessagePayload(ctx, message)
		if errors.Is(err, app.ErrImpossibleParsing) {
			continue
		}
		if err != nil {
			log.Error().Err(err).
				Hex("msg_hash", message.Hash).
				Hex("src_tx_hash", message.SrcTxHash).
				Str("src_addr", message.SrcAddress.String()).
				Hex("dst_tx_hash", message.DstTxHash).
				Str("dst_addr", message.DstAddress.String()).
				Uint32("op_id", message.OperationID).
				Msg("parse message payload")
		}
	}

	if err := func() error {
		defer app.TimeTrack(time.Now(), "AddAccountStates(%d)", len(acc))
		return s.accountRepo.AddAccountStates(ctx, dbTx, acc)
	}(); err != nil {
		return errors.Wrap(err, "add account states")
	}

	if err := func() error {
		defer app.TimeTrack(time.Now(), "AddMessages(%d)", len(msg))
		sort.Slice(msg, func(i, j int) bool { return msg[i].CreatedLT < msg[j].CreatedLT })
		return s.msgRepo.AddMessages(ctx, dbTx, msg)
	}(); err != nil {
		return errors.Wrap(err, "add messages")
	}

	if err := func() error {
		defer app.TimeTrack(time.Now(), "AddTransactions(%d)", len(tx))
		return s.txRepo.AddTransactions(ctx, dbTx, tx)
	}(); err != nil {
		return errors.Wrap(err, "add transactions")
	}

	if err := func() error {
		defer app.TimeTrack(time.Now(), "AddBlocks(%d)", len(b))
		return s.blockRepo.AddBlocks(ctx, dbTx, b)
	}(); err != nil {
		return errors.Wrap(err, "add blocks")
	}

	if err := dbTx.Commit(); err != nil {
		return errors.Wrap(err, "cannot commit db tx")
	}

	return nil
}

func (s *Service) uniqAccounts(transactions []*core.Transaction) []*core.AccountState {
	var ret []*core.AccountState

	uniqAcc := make(map[addr.Address]map[uint64]*core.AccountState)

	for _, tx := range transactions {
		if tx.Account == nil {
			continue
		}
		if uniqAcc[tx.Account.Address] == nil {
			uniqAcc[tx.Account.Address] = map[uint64]*core.AccountState{}
		}
		uniqAcc[tx.Account.Address][tx.Account.LastTxLT] = tx.Account
	}

	for _, accounts := range uniqAcc {
		for _, a := range accounts {
			ret = append(ret, a)
		}
	}

	return ret
}

func (s *Service) addMessage(msg *core.Message, uniqMsg map[string]*core.Message) {
	id := string(msg.Hash)

	if _, ok := uniqMsg[id]; !ok {
		uniqMsg[id] = msg
		return
	}

	switch {
	case msg.SrcTxLT != 0:
		uniqMsg[id].SrcTxLT, uniqMsg[id].SrcTxHash =
			msg.SrcTxLT, msg.SrcTxHash
		uniqMsg[id].SrcWorkchain, uniqMsg[id].SrcShard, uniqMsg[id].SrcBlockSeqNo =
			msg.SrcWorkchain, msg.SrcShard, msg.SrcBlockSeqNo
		uniqMsg[id].SrcState = msg.SrcState

	case msg.DstTxLT != 0:
		uniqMsg[id].DstTxLT, uniqMsg[id].DstTxHash =
			msg.DstTxLT, msg.DstTxHash
		uniqMsg[id].DstWorkchain, uniqMsg[id].DstShard, uniqMsg[id].DstBlockSeqNo =
			msg.DstWorkchain, msg.DstShard, msg.DstBlockSeqNo
		uniqMsg[id].DstState = msg.DstState
	}
}

func (s *Service) getMessageSource(ctx context.Context, msg *core.Message) (skip bool) {
	source, err := s.msgRepo.GetMessage(context.Background(), msg.Hash)
	if err == nil {
		msg.SrcTxLT, msg.SrcShard, msg.SrcBlockSeqNo, msg.SrcState =
			source.SrcTxLT, source.SrcShard, source.SrcBlockSeqNo, source.SrcState
		return false
	}
	if err != nil && !errors.Is(err, core.ErrNotFound) {
		panic(errors.Wrapf(err, "get message with hash %s", msg.Hash))
	}

	// some masterchain messages does not have source
	if msg.SrcAddress.Workchain() == -1 || msg.DstAddress.Workchain() == -1 {
		return false
	}

	blocks, err := s.blockRepo.CountMasterBlocks(ctx)
	if err != nil {
		panic(errors.Wrap(err, "count masterchain blocks"))
	}
	if blocks < 1000 {
		log.Debug().
			Hex("dst_tx_hash", msg.DstTxHash).
			Int32("dst_workchain", msg.DstWorkchain).Int64("dst_shard", msg.DstShard).Uint32("dst_block_seq_no", msg.DstBlockSeqNo).
			Str("src_address", msg.SrcAddress.String()).Str("dst_address", msg.DstAddress.String()).
			Msg("cannot find source message")
		return true
	}

	panic(fmt.Errorf("unknown source of message with dst tx hash %x on block (%d, %d, %d) from %s to %s",
		msg.DstTxHash, msg.DstWorkchain, msg.DstShard, msg.DstBlockSeqNo, msg.SrcAddress.String(), msg.DstAddress.String()))
}

func (s *Service) uniqMessages(ctx context.Context, transactions []*core.Transaction) []*core.Message {
	var ret []*core.Message

	uniqMsg := make(map[string]*core.Message)

	for j := range transactions {
		tx := transactions[j]

		if tx.InMsg != nil {
			s.addMessage(tx.InMsg, uniqMsg)
		}
		for _, out := range tx.OutMsg {
			s.addMessage(out, uniqMsg)
		}
	}

	for _, msg := range uniqMsg {
		if msg.Type == core.Internal && (msg.SrcTxLT == 0 && msg.DstTxLT != 0) {
			if s.getMessageSource(ctx, msg) {
				continue
			}
		}

		ret = append(ret, msg)
	}

	return ret
}

var lastLog = time.Now()

func (s *Service) saveBlock(ctx context.Context, master *core.Block) {
	newBlocks := append([]*core.Block{master}, master.Shards...)

	var newTransactions []*core.Transaction
	for i := range newBlocks {
		newTransactions = append(newTransactions, newBlocks[i].Transactions...)
	}

	if err := s.insertData(ctx, s.uniqAccounts(newTransactions), s.uniqMessages(ctx, newTransactions), newTransactions, newBlocks); err != nil {
		panic(err)
	}

	lvl := log.Debug()
	if time.Since(lastLog) > 10*time.Minute {
		lvl = log.Info()
		lastLog = time.Now()
	}
	lvl.Uint32("last_inserted_seq", master.SeqNo).Msg("inserted new block")
}

func (s *Service) saveBlocksLoop(results <-chan *core.Block) {
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

	for s.running() {
		var b *core.Block

		select {
		case b = <-results:
		case <-t.C:
			continue
		}

		log.Debug().
			Uint32("master_seq_no", b.SeqNo).
			Int("master_tx", len(b.Transactions)).
			Int("shards", len(b.Shards)).
			Msg("new master")

		s.saveBlock(context.Background(), b)
	}
}

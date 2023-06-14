package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func (s *Service) insertData(
	acc []*core.AccountState,
	msg []*core.Message,
	tx []*core.Transaction,
	b []*core.Block,
) error {
	ctx := context.Background()

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

var lastLog = time.Now()

func (s *Service) dumpMatchedData() {
	var (
		seq          uint32
		minMasterSeq uint32 = 1e9
		maxMasterSeq uint32 = 0
		dumpMasters  []pendingMaster
		insertBlocks []*core.Block
		insertTx     []*core.Transaction
		insertAcc    []*core.AccountState
		insertMsg    []*core.Message
	)

	if len(s.pendingMasters) < s.InsertBlockBatch {
		return
	}

	for _, msg := range s.unknownDstMsg {
		if msg.SrcWorkchain == -1 {
			if msg.SrcBlockSeqNo < minMasterSeq {
				minMasterSeq = msg.SrcBlockSeqNo
			}
			continue
		}
		blockID := core.BlockID{
			Workchain: msg.SrcWorkchain,
			Shard:     msg.SrcShard,
			SeqNo:     msg.SrcBlockSeqNo,
		}
		master := s.shardsMasterMap[blockID]
		if master.SeqNo < minMasterSeq {
			minMasterSeq = master.SeqNo
			continue
		}
	}

	for seq = range s.pendingMasters {
		if seq >= minMasterSeq {
			continue
		}
		if seq > maxMasterSeq {
			maxMasterSeq = seq
		}
		dumpMasters = append(dumpMasters, s.pendingMasters[seq])
		delete(s.pendingMasters, seq)
	}
	for master, shards := range s.masterShardsCache {
		if master.SeqNo >= minMasterSeq {
			continue
		}
		for _, shard := range shards {
			delete(s.shardsMasterMap, shard)
		}
		delete(s.masterShardsCache, master)
	}

	if len(dumpMasters) == 0 {
		return
	}

	for it := range dumpMasters {
		insertBlocks = append(insertBlocks, dumpMasters[it].Info...)
		insertTx = append(insertTx, dumpMasters[it].Tx...)
		insertMsg = append(insertMsg, dumpMasters[it].Msg...)
		insertAcc = append(insertAcc, dumpMasters[it].Acc...)
	}

	if err := s.insertData(insertAcc, insertMsg, insertTx, insertBlocks); err != nil {
		panic(err)
	}

	lvl := log.Debug()
	if time.Since(lastLog) > 10*time.Minute {
		lvl = log.Info()
		lastLog = time.Now()
	}
	lvl.Uint32("last_inserted_seq", maxMasterSeq).
		Int("pending_masters", len(s.pendingMasters)).
		Int("unknown_dst_msg", len(s.unknownDstMsg)).
		Uint32("min_master_seq", minMasterSeq).
		Msg("inserted new block")
}

func (s *Service) uniqAccounts(transactions []*core.Transaction) []*core.AccountState {
	var ret []*core.AccountState

	uniqAcc := make(map[addr.Address]*core.AccountState)

	for j := range transactions {
		tx := transactions[j]
		if tx.Account != nil {
			uniqAcc[tx.Account.Address] = tx.Account
		}
	}

	for _, a := range uniqAcc {
		ret = append(ret, a)
	}

	return ret
}

func (s *Service) addMessage(msg *core.Message, uniqMsg map[string]*core.Message) {
	id := string(msg.Hash)

	srcMsg, ok := s.unknownDstMsg[id]
	if ok {
		uniqMsg[id] = srcMsg
		delete(s.unknownDstMsg, id)
	}

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

func (s *Service) uniqMessages(transactions []*core.Transaction) []*core.Message {
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
		if (msg.Type != core.Internal) || (msg.SrcTxLT != 0 && msg.DstTxLT != 0) {
			ret = append(ret, msg)
			continue
		}

		if msg.SrcTxLT == 0 && msg.DstTxLT != 0 {
			if msg.SrcAddress.Workchain() == -1 && msg.DstAddress.Workchain() == -1 {
				ret = append(ret, msg)
				continue
			}
			_, err := s.msgRepo.FilterMessages(context.Background(), &filter.MessagesReq{Hash: msg.Hash})
			if err == nil {
				continue // message is already in a database
			}
			panic(fmt.Errorf("unknown source message with dst tx hash %x on block (%d, %x, %d) from %s to %s",
				msg.DstTxHash, msg.DstWorkchain, msg.DstShard, msg.DstBlockSeqNo, msg.SrcAddress.String(), msg.DstAddress.String()))
		}

		// unknown destination, waiting for next transactions
		s.unknownDstMsg[string(msg.Hash)] = msg
	}

	return ret
}

func (s *Service) addPendingBlocks(master *core.Block) {
	var (
		newBlocks       = []*core.Block{master}
		newTransactions []*core.Transaction
	)

	for i := range master.Shards {
		newBlocks = append(newBlocks, master.Shards[i])
		s.masterShardsCache[master.ID()] = append(s.masterShardsCache[master.ID()], master.Shards[i].ID())
		s.shardsMasterMap[master.Shards[i].ID()] = master.ID()
	}

	for i := range newBlocks {
		newTransactions = append(newTransactions, newBlocks[i].Transactions...)
	}

	s.pendingMasters[master.SeqNo] = pendingMaster{
		Info: newBlocks,
		Tx:   newTransactions,
		Acc:  s.uniqAccounts(newTransactions),
		Msg:  s.uniqMessages(newTransactions),
	}
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

		s.addPendingBlocks(b)

		s.dumpMatchedData()
	}
}

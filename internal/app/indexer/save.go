package indexer

import (
	"bytes"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

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

	for seq, m := range s.unknownDstMsg {
		for id, srcMsg := range m {
			if !bytes.Equal(msg.Hash, srcMsg.Hash) {
				continue
			}
			uniqMsg[id] = srcMsg
			delete(m, id)
		}
		if len(m) == 0 {
			delete(s.unknownDstMsg, seq)
		}
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
				continue
			}
			panic(fmt.Errorf("unknown source message with hash %x on block (%d, %x, %d) from %s to %s",
				msg.Hash, msg.DstWorkchain, msg.DstShard, msg.DstBlockSeqNo, msg.SrcAddress.String(), msg.DstAddress.String()))
		}

		// unknown destination, waiting for next transactions
		if _, ok := s.unknownDstMsg[msg.SrcBlockSeqNo]; !ok {
			s.unknownDstMsg[msg.SrcBlockSeqNo] = make(map[string]*core.Message)
		}
		s.unknownDstMsg[msg.SrcBlockSeqNo][string(msg.Hash)] = msg
	}

	return ret
}

func (s *Service) saveBlocksLoop(results <-chan processedMasterBlock) {
	var (
		insertBlocks []*core.Block
		insertTx     []*core.Transaction
		insertAcc    []*core.AccountState
		insertMsg    []*core.Message
		lastLog      = time.Now()
	)

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

	for s.running() {
		var b processedMasterBlock

		select {
		case b = <-results:
		case <-t.C:
			continue
		}

		newMaster := b.master.block

		log.Debug().
			Uint32("master_seq_no", newMaster.SeqNo).
			Int("master_tx", len(newMaster.Transactions)).
			Int("shards", len(b.shards)).
			Msg("new master")

		newBlocks := b.shards
		newBlocks = append(newBlocks, b.master)

		for i := range newBlocks {
			newBlock := newBlocks[i].block
			insertBlocks = append(insertBlocks, newBlock)
			insertTx = append(insertTx, newBlock.Transactions...)
			insertAcc = append(insertAcc, s.uniqAccounts(newBlock.Transactions)...)
			insertMsg = append(insertMsg, s.uniqMessages(newBlock.Transactions)...)
		}

		if len(s.unknownDstMsg) != 0 {
			continue
		}

		if err := s.insertData(insertAcc, insertMsg, insertTx, insertBlocks); err != nil {
			panic(err)
		}
		insertAcc, insertMsg, insertTx, insertBlocks = nil, nil, nil, nil

		lvl := log.Debug()
		if time.Since(lastLog) > time.Minute {
			lvl = log.Info()
			lastLog = time.Now()
		}
		lvl.Uint32("master_seq_no", newMaster.SeqNo).Msg("inserted new block")
	}
}

package indexer

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

type processedBlock struct {
	block *core.Block
	err   error
}

type processedMasterBlock struct {
	master processedBlock
	shards []processedBlock
}

func (s *Service) processMasterSeqNo(seq uint32) (ret processedMasterBlock) {
	defer app.TimeTrack(time.Now(), fmt.Sprintf("processMasterSeqNo(%d)", seq))

	for {
		ctx := context.Background()

		master, shards, err := s.Fetcher.UnseenBlocks(ctx, seq)
		if err != nil {
			lvl := log.Error()
			if errors.Is(err, ton.ErrBlockNotFound) || (err != nil && strings.Contains(err.Error(), "block is not applied")) {
				lvl = log.Debug()
			}
			lvl.Err(err).Uint32("master_seq", seq).Msg("cannot fetch unseen blocks")
			time.Sleep(300 * time.Millisecond)
			continue
		}
		ret.master, ret.shards = processedBlock{}, nil

		var wg sync.WaitGroup
		wg.Add(len(shards) + 1)

		ch := make(chan processedBlock, len(shards)+1)

		go func() {
			defer wg.Done()

			tx, err := s.Fetcher.BlockTransactions(ctx, master)

			ch <- processedBlock{
				block: &core.Block{
					Workchain:    master.Workchain,
					Shard:        master.Shard,
					SeqNo:        master.SeqNo,
					FileHash:     master.FileHash,
					RootHash:     master.RootHash,
					Transactions: tx,
				},
				err: err,
			}
		}()

		for i := range shards {
			go func(shard *ton.BlockIDExt) {
				defer wg.Done()

				tx, err := s.Fetcher.BlockTransactions(ctx, shard)

				ch <- processedBlock{
					block: &core.Block{
						Workchain: shard.Workchain,
						Shard:     shard.Shard,
						SeqNo:     shard.SeqNo,
						RootHash:  shard.RootHash,
						FileHash:  shard.FileHash,
						MasterID: &core.BlockID{
							Workchain: master.Workchain,
							Shard:     master.Shard,
							SeqNo:     master.SeqNo,
						},
						Transactions: tx,
					},
					err: err,
				}
			}(shards[i])
		}

		wg.Wait()
		close(ch)

		var errBlock processedBlock
		for i := range ch {
			if i.err != nil {
				errBlock = i
			}
			if i.block.Workchain == master.Workchain {
				ret.master = i
			} else {
				ret.shards = append(ret.shards, i)
			}
		}
		if errBlock.err != nil {
			log.Error().
				Err(errBlock.err).
				Int32("workchain", errBlock.block.Workchain).
				Uint64("shard", uint64(errBlock.block.Shard)).
				Uint32("seq", seq).
				Msg("cannot process block")
		} else {
			return ret
		}
	}
}

func (s *Service) fetchBlocksLoop(fromBlock uint32, results chan<- processedMasterBlock) {
	defer s.wg.Done()

	for s.running() {
		var wg sync.WaitGroup
		wg.Add(s.Workers)

		ch := make(chan processedMasterBlock, s.Workers)

		for i := 0; i < s.Workers; i++ {
			go func(seq uint32) {
				defer wg.Done()
				r := s.processMasterSeqNo(seq)
				ch <- r
			}(fromBlock + uint32(i))
		}

		wg.Wait()
		close(ch)

		var blocks []processedMasterBlock
		for b := range ch {
			blocks = append(blocks, b)
		}

		sort.Slice(blocks, func(i, j int) bool {
			return blocks[i].master.block.SeqNo < blocks[j].master.block.SeqNo
		})

		for i := range blocks {
			results <- blocks[i]
		}
		fromBlock = blocks[len(blocks)-1].master.block.SeqNo + 1
	}
}

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

	if err := s.accountRepo.AddAccountStates(ctx, dbTx, acc); err != nil {
		return errors.Wrap(err, "add account states")
	}
	if err := s.msgRepo.AddMessages(ctx, dbTx, msg); err != nil {
		return errors.Wrap(err, "add messages")
	}
	if err := s.txRepo.AddTransactions(ctx, dbTx, tx); err != nil {
		return errors.Wrap(err, "add transactions")
	}
	if err := s.blockRepo.AddBlocks(ctx, dbTx, b); err != nil {
		return errors.Wrap(err, "add shard block")
	}

	if err := dbTx.Commit(); err != nil {
		return errors.Wrap(err, "cannot commit db tx")
	}

	return nil
}

func (s *Service) saveBlocksLoop(results <-chan processedMasterBlock) {
	var (
		insertBlocks  []*core.Block
		insertTx      []*core.Transaction
		insertAcc     []*core.AccountState
		insertMsg     []*core.Message
		unknownDstMsg = make(map[uint32]map[string]*core.Message)
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
			var (
				newBlock     = newBlocks[i].block
				uniqMessages = make(map[string]*core.Message)
				uniqAccounts = make(map[addr.Address]*core.AccountState)
			)

			addMessage := func(msg *core.Message) {
				id := string(msg.Hash)

				for seq, m := range unknownDstMsg {
					for id, srcMsg := range m {
						if !bytes.Equal(msg.Hash, srcMsg.Hash) {
							continue
						}
						uniqMessages[id] = srcMsg
						delete(m, id)
					}
					if len(m) == 0 {
						delete(unknownDstMsg, seq)
					}
				}

				if _, ok := uniqMessages[id]; !ok {
					uniqMessages[id] = msg
					return
				}

				switch {
				case msg.SrcTxLT != 0:
					uniqMessages[id].SrcTxLT, uniqMessages[id].SrcTxHash =
						msg.SrcTxLT, msg.SrcTxHash
					uniqMessages[id].SrcWorkchain, uniqMessages[id].SrcShard, uniqMessages[id].SrcBlockSeqNo =
						msg.SrcWorkchain, msg.SrcShard, msg.SrcBlockSeqNo

				case msg.DstTxLT != 0:
					uniqMessages[id].DstTxLT, uniqMessages[id].DstTxHash =
						msg.DstTxLT, msg.DstTxHash
					uniqMessages[id].DstWorkchain, uniqMessages[id].DstShard, uniqMessages[id].DstBlockSeqNo =
						msg.DstWorkchain, msg.DstShard, msg.DstBlockSeqNo
				}
			}

			insertBlocks = append(insertBlocks, newBlock)

			insertTx = append(insertTx, newBlock.Transactions...)

			for j := range newBlock.Transactions {
				tx := newBlock.Transactions[j]

				if tx.Account != nil {
					uniqAccounts[tx.Account.Address] = tx.Account
				}

				if tx.InMsg != nil {
					addMessage(tx.InMsg)
				}
				for _, out := range tx.OutMsg {
					addMessage(out)
				}
			}

			for _, msg := range uniqMessages {
				if (msg.Type != core.Internal) || msg.SrcTxLT != 0 && msg.DstTxLT != 0 {
					insertMsg = append(insertMsg, msg)
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
				if _, ok := unknownDstMsg[msg.SrcBlockSeqNo]; !ok {
					unknownDstMsg[msg.SrcBlockSeqNo] = make(map[string]*core.Message)
				}
				unknownDstMsg[msg.SrcBlockSeqNo][string(msg.Hash)] = msg
			}

			for _, a := range uniqAccounts {
				insertAcc = append(insertAcc, a)
			}
		}

		if len(unknownDstMsg) != 0 {
			continue
		}

		if err := s.insertData(insertAcc, insertMsg, insertTx, insertBlocks); err != nil {
			panic(err)
		}
		insertAcc, insertMsg, insertTx, insertBlocks = nil, nil, nil, nil

		log.Debug().Uint32("master_seq_no", newMaster.SeqNo).Msg("inserted new block")
	}
}

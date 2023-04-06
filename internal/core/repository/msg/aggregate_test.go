package msg_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core/aggregate"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestRepository_AggregateMessages(t *testing.T) {
	var (
		recvCount, sentCount   = 50, 100
		recvAmount, sentAmount = new(bunbig.Int), new(bunbig.Int)
		recvFromAddress        = make(map[addr.Address]*bunbig.Int)
		sentToAddress          = make(map[addr.Address]*bunbig.Int)
	)

	initdb(t)

	address := rndm.Address()

	messagesTo := rndm.MessagesTo(address, recvCount)
	messagesFrom := rndm.MessagesFrom(address, sentCount)
	for _, msg := range messagesTo {
		recvAmount = recvAmount.Add(msg.Amount)
		recvFromAddress[msg.SrcAddress] = msg.Amount
	}
	for _, msg := range messagesFrom {
		sentAmount = sentAmount.Add(msg.Amount)
		sentToAddress[msg.DstAddress] = msg.Amount
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("insert test data", func(t *testing.T) {
		tx, err := pg.Begin()
		assert.Nil(t, err)

		err = repo.AddMessages(ctx, tx, rndm.Messages(100))
		assert.Nil(t, err)
		err = repo.AddMessages(ctx, tx, messagesTo)
		assert.Nil(t, err)
		err = repo.AddMessages(ctx, tx, messagesFrom)
		assert.Nil(t, err)

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("aggregate by address", func(t *testing.T) {
		res, err := repo.AggregateMessages(ctx, &aggregate.MessagesReq{
			Address: address,
			OrderBy: "amount",
			Limit:   150,
		})
		assert.Nil(t, err)
		assert.Equal(t, recvCount, res.RecvCount)
		assert.Equal(t, recvAmount, res.RecvAmount)
		for _, r := range res.RecvByAddress {
			assert.True(t, r.Sender != nil)
			assert.True(t, r.Amount != nil && r.Amount.ToUInt64() > 0)
			assert.Equal(t, recvFromAddress[*r.Sender], r.Amount)
			assert.Equal(t, 1, r.Count)
		}
		assert.Equal(t, sentCount, res.SentCount)
		assert.Equal(t, sentAmount, res.SentAmount)
		for _, r := range res.SentByAddress {
			assert.True(t, r.Receiver != nil)
			assert.True(t, r.Amount != nil && r.Amount.ToUInt64() > 0)
			assert.Equal(t, sentToAddress[*r.Receiver], r.Amount)
			assert.Equal(t, 1, r.Count)
		}
	})
}

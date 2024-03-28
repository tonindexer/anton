package msg_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/addr"
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
		require.Nil(t, err)

		err = repo.AddMessages(ctx, tx, rndm.Messages(100))
		require.Nil(t, err)
		err = repo.AddMessages(ctx, tx, messagesTo)
		require.Nil(t, err)
		err = repo.AddMessages(ctx, tx, messagesFrom)
		require.Nil(t, err)

		err = tx.Commit()
		require.Nil(t, err)
	})

	t.Run("aggregate by address", func(t *testing.T) {
		res, err := repo.AggregateMessages(ctx, &aggregate.MessagesReq{
			Address: address,
			OrderBy: "amount",
			Limit:   150,
		})
		require.Nil(t, err)
		require.Equal(t, recvCount, res.RecvCount)
		require.Equal(t, recvAmount, res.RecvAmount)
		for _, r := range res.RecvByAddress {
			require.True(t, r.Sender != nil)
			require.True(t, r.Amount != nil && r.Amount.ToUInt64() > 0)
			require.Equal(t, recvFromAddress[*r.Sender], r.Amount)
			require.Equal(t, 1, r.Count)
		}
		require.Equal(t, sentCount, res.SentCount)
		require.Equal(t, sentAmount, res.SentAmount)
		for _, r := range res.SentByAddress {
			require.True(t, r.Receiver != nil)
			require.True(t, r.Amount != nil && r.Amount.ToUInt64() > 0)
			require.Equal(t, sentToAddress[*r.Receiver], r.Amount)
			require.Equal(t, 1, r.Count)
		}
	})
}

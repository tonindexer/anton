package msg_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/internal/core/aggregate/history"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestRepository_AggregateMessagesHistory(t *testing.T) {
	var (
		amountSum = new(bunbig.Int)
	)

	initdb(t)

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

		messages := rndm.Messages(10)
		messagesToSpecial := rndm.Messages(50)

		payloads := append(
			rndm.MessagePayloads(messages),
			rndm.MessagesToContract(messagesToSpecial, "special")...)

		for _, m := range messagesToSpecial {
			amountSum = amountSum.Add(m.Amount)
		}

		err = repo.AddMessagePayloads(ctx, tx, payloads)
		assert.Nil(t, err)
		err = repo.AddMessages(ctx, tx, messages)
		assert.Nil(t, err)

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("count messages to special contract", func(t *testing.T) {
		res, err := repo.AggregateMessagesHistory(ctx, &history.MessagesReq{
			Metric:       history.MessageCount,
			DstContracts: []string{"special"},
			ReqParams: history.ReqParams{
				From:     time.Now().Add(-time.Minute),
				Interval: 24 * time.Hour,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(res.CountRes))
		assert.Equal(t, 50, res.CountRes[0].Value)
	})

	t.Run("sum messages amount to special contract", func(t *testing.T) {
		res, err := repo.AggregateMessagesHistory(ctx, &history.MessagesReq{
			Metric:       history.MessageAmountSum,
			DstContracts: []string{"special"},
			ReqParams: history.ReqParams{
				From:     time.Now().Add(-time.Minute),
				Interval: 24 * time.Hour,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(res.BigIntRes))
		assert.Equal(t, amountSum, res.BigIntRes[0].Value)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}

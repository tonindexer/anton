package msg_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestFilterRepository(t *testing.T) {
	initdb(t)

	messages := rndm.Messages(100)
	payloads := rndm.MessagePayloads(messages)

	specialOperation := messages[len(messages)-1]
	specialOperation.Payload = payloads[len(payloads)-1]
	specialOperation.Payload.OperationName = "special_op"

	specialDestination := messages[len(messages)-2]
	specialDestination.Payload = payloads[len(payloads)-2]
	specialDestination.Payload.DstContract = "special"

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

		err = repo.AddMessagePayloads(ctx, tx, payloads)
		assert.Nil(t, err)
		err = repo.AddMessages(ctx, tx, messages)
		assert.Nil(t, err)

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("filter by hash", func(t *testing.T) {
		expected := *messages[0]
		expected.Payload = payloads[0]

		res, err := repo.FilterMessages(ctx, &filter.MessagesReq{
			Hash:        messages[0].Hash,
			WithPayload: true,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, 1, len(res.Rows))
		assert.JSONEq(t, string(expected.Payload.DataJSON), string(res.Rows[0].Payload.DataJSON))
		res.Rows[0].Payload.DataJSON = expected.Payload.DataJSON
		assert.Equal(t, []*core.Message{&expected}, res.Rows)
	})

	t.Run("filter by address", func(t *testing.T) {
		res, err := repo.FilterMessages(ctx, &filter.MessagesReq{
			DstAddresses: []*addr.Address{&messages[0].DstAddress},
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, []*core.Message{messages[0]}, res.Rows)
	})

	t.Run("filter by contract", func(t *testing.T) {
		res, err := repo.FilterMessages(ctx, &filter.MessagesReq{
			DstContracts: []string{"special"},
			WithPayload:  true,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, 1, len(res.Rows))
		assert.JSONEq(t, string(specialDestination.Payload.DataJSON), string(res.Rows[0].Payload.DataJSON))
		res.Rows[0].Payload.DataJSON = specialDestination.Payload.DataJSON
		assert.Equal(t, []*core.Message{specialDestination}, res.Rows)
	})

	t.Run("filter by operation name", func(t *testing.T) {
		res, err := repo.FilterMessages(ctx, &filter.MessagesReq{
			OperationNames: []string{"special_op"},
			WithPayload:    true,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, 1, len(res.Rows))
		assert.JSONEq(t, string(specialOperation.Payload.DataJSON), string(res.Rows[0].Payload.DataJSON))
		res.Rows[0].Payload.DataJSON = specialOperation.Payload.DataJSON
		assert.Equal(t, []*core.Message{specialOperation}, res.Rows)
	})

	t.Run("filter by minter address", func(t *testing.T) {
		res, err := repo.FilterMessages(ctx, &filter.MessagesReq{
			MinterAddress: specialOperation.Payload.MinterAddress,
			WithPayload:   true,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, 1, len(res.Rows))
		assert.JSONEq(t, string(specialOperation.Payload.DataJSON), string(res.Rows[0].Payload.DataJSON))
		res.Rows[0].Payload.DataJSON = specialOperation.Payload.DataJSON
		assert.Equal(t, []*core.Message{specialOperation}, res.Rows)
	})
}

// TODO: benchmarks on filtering by msg payload

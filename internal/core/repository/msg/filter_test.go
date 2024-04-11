package msg_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestRepository_FilterMessages(t *testing.T) {
	initdb(t)

	messages := rndm.Messages(100)

	specialOperation := messages[len(messages)-1]
	specialOperation.OperationName = "special_op"

	specialDestination := messages[len(messages)-2]
	specialDestination.DstContract = "special"

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

		err = repo.AddMessages(ctx, tx, messages)
		require.Nil(t, err)

		err = tx.Commit()
		require.Nil(t, err)
	})

	t.Run("filter by hash", func(t *testing.T) {
		expected := *messages[0]

		res, err := repo.FilterMessages(ctx, &filter.MessagesReq{
			Hash: messages[0].Hash,
		})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, 1, len(res.Rows))
		require.JSONEq(t, string(expected.DataJSON), string(res.Rows[0].DataJSON))
		res.Rows[0].DataJSON = expected.DataJSON
		require.Equal(t, []*core.Message{&expected}, res.Rows)
	})

	t.Run("filter by address", func(t *testing.T) {
		res, err := repo.FilterMessages(ctx, &filter.MessagesReq{
			DstAddresses: []*addr.Address{&messages[0].DstAddress},
		})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, []*core.Message{messages[0]}, res.Rows)
	})

	t.Run("filter by contract", func(t *testing.T) {
		res, err := repo.FilterMessages(ctx, &filter.MessagesReq{
			DstContracts: []string{"special"},
		})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, 1, len(res.Rows))
		require.JSONEq(t, string(specialDestination.DataJSON), string(res.Rows[0].DataJSON))
		res.Rows[0].DataJSON = specialDestination.DataJSON
		require.Equal(t, []*core.Message{specialDestination}, res.Rows)
	})

	t.Run("filter by operation name", func(t *testing.T) {
		res, err := repo.FilterMessages(ctx, &filter.MessagesReq{
			OperationNames: []string{"special_op"},
		})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, 1, len(res.Rows))
		require.JSONEq(t, string(specialOperation.DataJSON), string(res.Rows[0].DataJSON))
		res.Rows[0].DataJSON = specialOperation.DataJSON
		require.Equal(t, []*core.Message{specialOperation}, res.Rows)
	})
}

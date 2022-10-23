package db

import (
	"context"
	"time"

	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
)

func Connect(ctx context.Context, dsn string, opts ...ch.Option) (*ch.DB, error) {
	opts = append(opts, ch.WithDSN(dsn), ch.WithAutoCreateDatabase(true), ch.WithPoolSize(16))

	db := ch.Connect(opts...)

	var err error
	for i := 0; i < 8; i++ { // wait for ch start
		err = db.Ping(ctx)
		if err == nil {
			return db, nil
		}
		time.Sleep(2 * time.Second)
	}

	return db, err
}

func CreateTables(ctx context.Context, db *ch.DB) error {
	_, err := db.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.MasterBlockInfo{}).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.ShardBlockInfo{}).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.Transaction{}).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.Message{}).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.MessagePayload{}).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.Account{}).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.AccountData{}).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.ContractInterface{}).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.ContractOperation{}).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

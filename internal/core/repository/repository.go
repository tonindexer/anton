package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core/repository/abi"
	"github.com/iam047801/tonidx/internal/core/repository/account"
	"github.com/iam047801/tonidx/internal/core/repository/block"
	"github.com/iam047801/tonidx/internal/core/repository/tx"
)

type DB struct {
	CH *ch.DB // TODO: do not insert duplicates to ch
	PG *bun.DB
}

func (db *DB) Close() {
	_ = db.CH.Close()
	_ = db.PG.Close()
}

func ConnectDB(ctx context.Context, dsnCH, dsnPG string, opts ...ch.Option) (*DB, error) {
	var err error

	opts = append(opts, ch.WithDSN(dsnCH), ch.WithAutoCreateDatabase(true), ch.WithPoolSize(16))
	chDB := ch.Connect(opts...)

	for i := 0; i < 8; i++ { // wait for ch start
		err = chDB.Ping(ctx)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, errors.Wrap(err, "cannot ping ch")
	}

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsnPG)))
	pgDB := bun.NewDB(sqlDB, pgdialect.New())

	for i := 0; i < 8; i++ { // wait for pg start
		err = pgDB.Ping()
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, errors.Wrap(err, "cannot ping pg")
	}

	return &DB{CH: chDB, PG: pgDB}, nil
}

func CreateTablesDB(ctx context.Context, db *DB) error {
	err := block.CreateTables(ctx, db.CH, db.PG)
	if err != nil {
		return err
	}
	err = account.CreateTables(ctx, db.CH, db.PG)
	if err != nil {
		return err
	}
	err = tx.CreateTables(ctx, db.CH, db.PG)
	if err != nil {
		return err
	}
	err = abi.CreateTables(ctx, db.CH, db.PG)
	if err != nil {
		return err
	}
	return nil
}

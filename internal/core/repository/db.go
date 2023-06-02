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
)

type DB struct {
	CH *ch.DB
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

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsnPG), pgdriver.WithWriteTimeout(time.Minute)))
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

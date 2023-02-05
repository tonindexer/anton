package tx

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
)

var _ core.TxRepository = (*Repository)(nil)

type Repository struct {
	ch *ch.DB
	pg *bun.DB
}

func NewRepository(_ch *ch.DB, _pg *bun.DB) *Repository {
	return &Repository{ch: _ch, pg: _pg}
}

func CreateTables(ctx context.Context, chDB *ch.DB, pgDB *bun.DB) error {
	_, err := pgDB.ExecContext(ctx, "CREATE TYPE message_type AS ENUM (?, ?, ?)",
		core.ExternalIn, core.ExternalOut, core.Internal)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return errors.Wrap(err, "account status pg create enum")
	}

	_, err = chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.MessagePayload{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload ch create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.MessagePayload{}).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create table")
	}

	_, err = chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.Message{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message ch create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.Message{}).
		IfNotExists().
		// WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message pg create table")
	}

	_, err = chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.Transaction{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "transaction ch create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.Transaction{}).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "transaction pg create table")
	}

	return nil
}

func (r *Repository) AddTransactions(ctx context.Context, transactions []*core.Transaction) error {
	_, err := r.ch.NewInsert().Model(&transactions).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.pg.NewInsert().Model(&transactions).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) AddMessages(ctx context.Context, messages []*core.Message) error {
	_, err := r.ch.NewInsert().Model(&messages).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.pg.NewInsert().Model(&messages).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) AddMessagePayloads(ctx context.Context, payloads []*core.MessagePayload) error {
	_, err := r.ch.NewInsert().Model(&payloads).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.pg.NewInsert().Model(&payloads).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetSourceMessageHash(ctx context.Context, from, to string, bodyHash []byte, lt uint64) (ret []byte, err error) {
	err = r.pg.NewSelect().Model(&core.Message{}).
		Column("hash").
		Where("src_address = ?", from).
		Where("dst_address = ?", to).
		Where("body_hash = ?", bodyHash).
		Where("created_lt = ?", lt).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *Repository) GetTransactions(ctx context.Context, filter *core.TransactionFilter, offset, limit int) ([]*core.Transaction, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *Repository) GetMessages(ctx context.Context, filter *core.MessageFilter, offset, limit int) ([]*core.MessageFilter, error) {
	panic(fmt.Errorf("not implemented"))
}

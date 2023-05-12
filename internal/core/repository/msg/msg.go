package msg

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
)

var _ repository.Message = (*Repository)(nil)

type Repository struct {
	ch *ch.DB
	pg *bun.DB
}

func NewRepository(ck *ch.DB, pg *bun.DB) *Repository {
	return &Repository{ch: ck, pg: pg}
}

func createIndexes(ctx context.Context, pgDB *bun.DB) error {
	var err error

	// messages

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Column("src_address", "source_tx_lt").
		Where("length(src_address) > 0").
		Where("source_tx_lt > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message src_address source_tx_lt pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("HASH").
		Column("src_address").
		Where("length(src_address) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message src addr pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Unique().
		Column("src_address", "created_lt").
		Where("length(src_address) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message src addr lt pg create unique index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("HASH").
		Column("dst_address").
		Where("length(dst_address) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message dst addr pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("BTREE").
		Column("created_lt").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message created_lt pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Column("operation_id").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message operation id pg create index")
	}

	// message payloads

	_, err = pgDB.NewCreateIndex().
		Model(&core.MessagePayload{}).
		Using("HASH").
		Column("src_contract").
		Where("length(src_contract) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create src_contract index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.MessagePayload{}).
		Using("HASH").
		Column("dst_contract").
		Where("length(dst_contract) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create dst_contract index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.MessagePayload{}).
		Using("HASH").
		Column("operation_name").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create operation name index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.MessagePayload{}).
		Using("HASH").
		Column("minter_address").
		Where("length(minter_address) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "address state pg create unique index")
	}

	return nil
}

func CreateTables(ctx context.Context, chDB *ch.DB, pgDB *bun.DB) error {
	_, err := pgDB.ExecContext(ctx, "CREATE TYPE message_type AS ENUM (?, ?, ?)",
		core.ExternalIn, core.ExternalOut, core.Internal)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return errors.Wrap(err, "messages pg create enum")
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

	_, err = pgDB.ExecContext(ctx, "ALTER TABLE messages ADD CONSTRAINT messages_source_tx_hash_notnull "+
		"CHECK (NOT (source_tx_hash IS NULL AND src_address != decode('11ff0000000000000000000000000000000000000000000000000000000000000000', 'hex')));")
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return errors.Wrap(err, "messages pg create source tx hash check")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.Message{}).
		IfNotExists().
		// WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message pg create table")
	}

	if err := createIndexes(ctx, pgDB); err != nil {
		return err
	}

	return nil
}

func (r *Repository) AddMessages(ctx context.Context, tx bun.Tx, messages []*core.Message) error {
	var unknown []*core.Message

	for _, msg := range messages {
		if msg.Known {
			continue
		}
		unknown = append(unknown, msg)
	}

	if len(unknown) == 0 {
		return nil
	}

	_, err := tx.NewInsert().Model(&unknown).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.ch.NewInsert().Model(&unknown).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) AddMessagePayloads(ctx context.Context, tx bun.Tx, payloads []*core.MessagePayload) error {
	if len(payloads) == 0 {
		return nil
	}
	_, err := tx.NewInsert().Model(&payloads).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.ch.NewInsert().Model(&payloads).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

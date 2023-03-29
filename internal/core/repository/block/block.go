package block

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
)

var _ repository.Block = (*Repository)(nil)

type Repository struct {
	ch *ch.DB
	pg *bun.DB
}

func NewRepository(ck *ch.DB, pg *bun.DB) *Repository {
	return &Repository{ch: ck, pg: pg}
}

func createIndexes(ctx context.Context, pgDB *bun.DB) error {
	_, err := pgDB.NewCreateIndex().
		Model(&core.Block{}).
		Using("HASH").
		Column("workchain").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "block workchain pg create index")
	}

	return nil
}

func CreateTables(ctx context.Context, chDB *ch.DB, pgDB *bun.DB) error {
	_, err := chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.Block{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "block ch create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.Block{}).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "block pg create table")
	}

	return createIndexes(ctx, pgDB)
}

func (r *Repository) AddBlocks(ctx context.Context, tx bun.Tx, info []*core.Block) error {
	if len(info) == 0 {
		return nil
	}
	_, err := tx.NewInsert().Model(&info).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.ch.NewInsert().Model(&info).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetLastMasterBlock(ctx context.Context) (*core.Block, error) {
	ret := new(core.Block)

	err := r.ch.NewSelect().Model(ret).
		Where("workchain = ?", -1).
		Order("seq_no DESC").
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return ret, nil
}

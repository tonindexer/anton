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
		Using("BTREE").
		Column("workchain", "seq_no").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "block workchain pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Block{}).
		Using("BTREE").
		Column("master_workchain", "master_shard", "master_seq_no").
		Where("workchain != -1").
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
	for _, b := range info {
		_, err := tx.NewInsert().Model(b).Exec(ctx)
		if err != nil {
			return err
		}
	}
	_, err := r.ch.NewInsert().Model(&info).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetLastMasterBlock(ctx context.Context) (*core.Block, error) {
	ret := new(core.Block)

	err := r.pg.NewSelect().Model(ret).
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

func (r *Repository) CountMasterBlocks(ctx context.Context) (int, error) {
	ret, err := r.ch.NewSelect().Model((*core.Block)(nil)).
		Where("workchain = -1").
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (r *Repository) GetMissedMasterBlocks(ctx context.Context) (res []uint32, err error) {
	var ret []struct {
		SeqNo     uint32
		NextSeqNo uint32
	}

	err = r.ch.NewSelect().
		TableExpr("(?) as sq",
			r.ch.NewSelect().Model((*core.Block)(nil)).
				ColumnExpr("seq_no").
				ColumnExpr("any(seq_no) over (order by seq_no asc rows between 1 following and 1 following) as next_seq_no").
				Where("workchain = -1").
				Order("seq_no asc"),
		).
		Where("seq_no != next_seq_no - 1").
		Where("next_seq_no != 0").
		Order("seq_no asc").
		Scan(ctx, &ret)
	if err != nil {
		return nil, err
	}

	var lastMissedBlock uint32
	for _, r := range ret {
		for i := r.SeqNo - 10; i < r.NextSeqNo+10; i++ {
			if i <= lastMissedBlock {
				continue
			}
			lastMissedBlock = i
			res = append(res, i)
		}
	}

	return res, nil
}

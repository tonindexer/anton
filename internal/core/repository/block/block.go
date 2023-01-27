package block

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
)

var _ core.BlockRepository = (*Repository)(nil)

type Repository struct {
	db *ch.DB
}

func NewRepository(db *ch.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetLastMasterBlockInfo(ctx context.Context) (*core.BlockInfo, error) {
	ret := new(core.BlockInfo)

	err := r.db.NewSelect().Model(ret).
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

func (r *Repository) AddBlocksInfo(ctx context.Context, info []*core.BlockInfo) error {
	for _, b := range info {
		_, err := r.db.NewInsert().Model(b).Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

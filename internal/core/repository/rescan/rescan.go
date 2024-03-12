package rescan

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"

	"github.com/tonindexer/anton/internal/core"
)

var _ core.RescanRepository = (*Repository)(nil)

type Repository struct {
	pg *bun.DB
}

func NewRepository(db *bun.DB) *Repository {
	return &Repository{pg: db}
}

func (r *Repository) AddRescanTask(ctx context.Context, task *core.RescanTask) error {
	_, err := r.pg.NewInsert().Model(task).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetUnfinishedRescanTask(ctx context.Context) (bun.Tx, *core.RescanTask, error) {
	var task core.RescanTask

	tx, err := r.pg.Begin()
	if err != nil {
		return bun.Tx{}, nil, err
	}

	err = tx.NewSelect().Model(&task).
		For("UPDATE").
		Where("finished = ?", false).
		Order("id").
		Limit(1).
		Scan(ctx)
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return bun.Tx{}, nil, errors.Wrap(core.ErrNotFound, "no unfinished tasks")
		}
		return bun.Tx{}, nil, err
	}

	if task.Type != core.DelInterface {
		task.Contract = new(core.ContractInterface)
		err := r.pg.NewSelect().Model(task.Contract).
			Where("name = ?", task.ContractName).
			Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = core.ErrNotFound
			}
			return bun.Tx{}, nil, errors.Wrapf(core.ErrNotFound, "no %s contract interface for %s task", task.ContractName, task.Type)
		}
	}

	if task.Type == core.UpdOperation {
		task.Operation = new(core.ContractOperation)
		err := r.pg.NewSelect().Model(task.Operation).
			Where("contract_name = ?", task.ContractName).
			Where("outgoing IS ?", task.Operation).
			Where("message_type = ?", task.MessageType).
			Where("operation_id = ?", task.OperationID).
			Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = core.ErrNotFound
			}
			return bun.Tx{}, nil, errors.Wrapf(err, "get 0x%x operation of %s contract for %s task", task.OperationID, task.ContractName, task.Type)
		}
	}

	return tx, &task, nil
}

func (r *Repository) SetRescanTask(ctx context.Context, tx bun.Tx, task *core.RescanTask) error {
	_, err := tx.NewUpdate().Model(task).
		Set("finished = ?finished").
		Set("last_address = ?last_address").
		Set("last_tx_lt = ?last_tx_lt").
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

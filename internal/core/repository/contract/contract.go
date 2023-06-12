package contract

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
)

var _ repository.Contract = (*Repository)(nil)

type Repository struct {
	pg    *bun.DB
	cache *cache
}

func NewRepository(db *bun.DB) *Repository {
	return &Repository{pg: db, cache: newCache()}
}

func CreateTables(ctx context.Context, pgDB *bun.DB) error {
	_, err := pgDB.NewCreateTable().
		Model(&core.ContractInterface{}).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "contract interface pg create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.ContractOperation{}).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "contract operations pg create table")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.ContractInterface{}).
		Unique().
		Column("get_method_hashes").
		Where("addresses IS NULL and code IS NULL").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "contract interface get_method_hashes create unique index")
	}

	_, err = pgDB.ExecContext(ctx, `ALTER TABLE contract_operations ADD CONSTRAINT contract_interfaces_uniq_name UNIQUE (operation_name, contract_name)`)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return errors.Wrap(err, "messages pg create source tx hash check")
	}

	return nil
}

func (r *Repository) AddInterface(ctx context.Context, i *core.ContractInterface) error {
	_, err := r.pg.NewInsert().Model(i).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) AddOperation(ctx context.Context, op *core.ContractOperation) error {
	_, err := r.pg.NewInsert().Model(op).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) DelInterface(ctx context.Context, name string) error {
	_, err := r.pg.NewDelete().
		Model((*core.ContractOperation)(nil)).
		Where("contract_name = ?", name).
		Exec(ctx)
	if err != nil {
		return err
	}

	ret, err := r.pg.NewDelete().
		Model((*core.ContractInterface)(nil)).
		Where("name = ?", name).
		Exec(ctx)
	if err != nil {
		return err
	}

	rows, err := ret.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "rows affected")
	}
	if rows == 0 {
		return errors.Wrap(core.ErrNotFound, "no such interface")
	}

	return nil
}

func (r *Repository) GetInterfaces(ctx context.Context) ([]*core.ContractInterface, error) {
	var ret []*core.ContractInterface

	if i := r.cache.getInterfaces(); i != nil {
		return i, nil
	}

	err := r.pg.NewSelect().Model(&ret).Scan(ctx)
	if err != nil {
		return nil, err
	}

	r.cache.setInterfaces(ret)

	return ret, nil
}

func (r *Repository) GetOperations(ctx context.Context) ([]*core.ContractOperation, error) {
	var ret []*core.ContractOperation

	if op := r.cache.getOperations(); op != nil {
		return op, nil
	}

	err := r.pg.NewSelect().Model(&ret).Scan(ctx)
	if err != nil {
		return nil, err
	}

	r.cache.setOperations(ret)

	return ret, nil
}

func (r *Repository) GetOperationByID(ctx context.Context, types []abi.ContractName, outgoing bool, id uint32) (*core.ContractOperation, error) {
	var ret []*core.ContractOperation

	if len(types) == 0 {
		return nil, errors.Wrap(core.ErrNotFound, "no contract types")
	}

	if op := r.cache.getOperationByID(types, outgoing, id); op != nil {
		return op, nil
	}

	err := r.pg.NewSelect().Model(&ret).
		Where("contract_name IN (?)", bun.In(types)).
		Where("outgoing IS ?", outgoing).
		Where("operation_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	if len(ret) < 1 {
		return nil, errors.Wrap(core.ErrNotFound, "unknown operation")
	}

	op := ret[0]

	return op, nil
}

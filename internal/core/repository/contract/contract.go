package contract

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
)

var _ repository.Contract = (*Repository)(nil)

type Repository struct {
	pg         *bun.DB
	interfaces []*core.ContractInterface
	operations []*core.ContractOperation
}

func NewRepository(db *bun.DB) *Repository {
	return &Repository{pg: db}
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

	return nil
}

func (r *Repository) GetInterfaces(ctx context.Context) ([]*core.ContractInterface, error) {
	var ret []*core.ContractInterface

	// TODO: clear cache

	if len(r.interfaces) > 0 {
		return r.interfaces, nil
	}

	err := r.pg.NewSelect().Model(&ret).Scan(ctx)
	if err != nil {
		return nil, err
	}

	if len(ret) > 0 {
		r.interfaces = ret
	}

	return ret, nil
}

func (r *Repository) GetOperations(ctx context.Context) ([]*core.ContractOperation, error) {
	var ret []*core.ContractOperation

	// TODO: clear cache

	if len(r.operations) > 0 {
		return r.operations, nil
	}

	err := r.pg.NewSelect().Model(&ret).Scan(ctx)
	if err != nil {
		return nil, err
	}

	if len(ret) > 0 {
		r.operations = ret
	}

	return ret, nil
}

func (r *Repository) GetOperationByID(ctx context.Context, types []abi.ContractName, outgoing bool, id uint32) (*core.ContractOperation, error) {
	var ret []*core.ContractOperation

	if len(types) == 0 {
		return nil, errors.Wrap(core.ErrNotFound, "no contract types")
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

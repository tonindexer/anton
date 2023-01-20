package account

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
)

var _ core.AccountRepository = (*Repository)(nil)

type Repository struct {
	db         *ch.DB
	interfaces []*core.ContractInterface
	operations []*core.ContractOperation
}

func NewRepository(db *ch.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetContractInterfaces(ctx context.Context) ([]*core.ContractInterface, error) {
	var ret []*core.ContractInterface

	// TODO: clear cache

	if len(r.interfaces) > 0 {
		return r.interfaces, nil
	}

	err := r.db.NewSelect().Model(&ret).Scan(ctx)
	if err != nil {
		return nil, err
	}

	if len(ret) > 0 {
		r.interfaces = ret
	}

	return ret, nil
}

func (r *Repository) InsertContractOperations(ctx context.Context, operations []*core.ContractOperation) error {
	var err error

	for _, op := range operations {
		op.Schema, err = marshalStructSchema(op.StructSchema)
		if err != nil {
			return errors.Wrap(err, "marshal struct schema")
		}

		_, err = r.db.NewInsert().Model(op).Exec(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) GetContractOperations(ctx context.Context, types []core.ContractType) ([]*core.ContractOperation, error) {
	var ret []*core.ContractOperation

	// TODO: clear cache

	if len(r.operations) > 0 {
		return r.operations, nil
	}

	err := r.db.NewSelect().Model(&ret).Where("type in (?)", ch.In(types)).Scan(ctx)
	if err != nil {
		return nil, err
	}

	for _, op := range ret {
		op.StructSchema, err = unmarshalStructSchema(op.Schema)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal struct schema")
		}
	}

	if len(ret) > 0 {
		r.operations = ret
	}

	return ret, nil
}

func (r *Repository) GetContractOperationByID(ctx context.Context, a *core.Account, outgoing bool, id uint32) (*core.ContractOperation, error) {
	var ret []*core.ContractOperation

	if len(a.Types) == 0 {
		return nil, errors.Wrap(core.ErrNotFound, "no contract types")
	}

	var out uint16 // TODO: fix this, go-ch bug
	if outgoing {
		out = 1
	}

	err := r.db.NewSelect().Model(&ret).
		Where("contract_name in (?)", ch.In(a.Types)).
		Where("outgoing = ?", out).
		Where("operation_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	if len(ret) < 1 {
		return nil, errors.Wrap(core.ErrNotFound, "unknown operation")
	}

	op := ret[0]
	op.StructSchema, err = unmarshalStructSchema(op.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal struct schema")
	}

	return op, nil
}

func (r *Repository) AddAccounts(ctx context.Context, accounts []*core.Account) error {
	for _, acc := range accounts {
		_, err := r.db.NewInsert().Model(acc).Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) AddAccountData(ctx context.Context, data []*core.AccountData) error {
	for _, d := range data {
		// TODO: cache data hashes, do not duplicate account data

		_, err := r.db.NewInsert().Model(d).Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

package abi

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/iam047801/tonidx/internal/core"
)

var _ core.ContractRepository = (*Repository)(nil)

type Repository struct {
	db         *ch.DB
	interfaces []*core.ContractInterface
	operations []*core.ContractOperation
}

func NewRepository(db *ch.DB) *Repository {
	return &Repository{db: db}
}

func CreateTables(ctx context.Context, chDB *ch.DB, pgDB *bun.DB) error {
	_, err := chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.ContractInterface{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "contract interface ch create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.ContractInterface{}).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "contract interface pg create table")
	}

	_, err = chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.ContractOperation{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "contract operations ch create table")
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

func matchByAddress(acc *tlb.Account, addr string) bool {
	if addr == "" {
		return false
	}
	return acc.State != nil && addr == acc.State.Address.String()
}

func matchByCode(acc *tlb.Account, code []byte) bool {
	if len(code) == 0 {
		return false
	}

	codeCell, err := cell.FromBOC(code)
	if err != nil {
		log.Error().Err(err).Msg("parse contract interface code")
		return false
	}

	return acc.Code != nil && bytes.Equal(acc.Code.Hash(), codeCell.Hash())
}

func (s *Repository) DetermineContractInterfaces(ctx context.Context, acc *tlb.Account) ([]core.ContractType, error) {
	var ret []core.ContractType

	version := wallet.GetWalletVersion(acc)
	if version != wallet.Unknown {
		ret = append(ret,
			core.ContractType(fmt.Sprintf("wallet_%s", version.String())))
	}

	ifaces, err := s.GetContractInterfaces(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get contract interfaces")
	}

	for _, iface := range ifaces {
		if matchByAddress(acc, iface.Address) {
			ret = append(ret, iface.Name)
			continue
		}

		if matchByCode(acc, iface.Code) {
			ret = append(ret, iface.Name)
			continue
		}

		if len(iface.GetMethods) == 0 {
			continue
		}

		var hasMethods = true
		for _, get := range iface.GetMethods {
			if !acc.HasGetMethod(get) {
				hasMethods = false
				break
			}
		}
		if hasMethods {
			ret = append(ret, iface.Name)
		}
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

func ParseOperationID(msg *core.Message) error {
	payload, err := cell.FromBOC(msg.Body)
	if err != nil {
		return errors.Wrap(err, "msg body from boc")
	}

	slice := payload.BeginParse()

	op, _ := slice.LoadUInt(32)
	msg.OperationID = uint32(op)

	if msg.OperationID == 0 {
		// simple transfer with comment
		msg.TransferComment, _ = slice.LoadStringSnake()
	}

	return nil
}

func (r *Repository) GetContractOperationByID(ctx context.Context, a *core.AccountState, outgoing bool, id uint32) (*core.ContractOperation, error) {
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

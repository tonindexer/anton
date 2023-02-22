package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/abi"
)

type ContractInterface struct {
	ch.CHModel    `ch:"contract_interfaces"`
	bun.BaseModel `bun:"table:contract_interfaces"`

	Name       abi.ContractName `ch:",pk" bun:",pk"`
	Address    string           `ch:",pk" bun:",pk"`
	Code       []byte           `bun:",unique"`
	GetMethods []string         `bun:",unique,array"`
}

type ContractOperation struct {
	ch.CHModel    `ch:"contract_operations"`
	bun.BaseModel `bun:"table:contract_operations"`

	Name         string           `bun:",unique"`
	ContractName abi.ContractName `ch:",pk" bun:",pk"`
	Outgoing     bool             `ch:",pk" bun:",pk"` // if operation is going from contract
	OperationID  uint32           `ch:",pk" bun:",pk"`
	Schema       []byte           //
}

type ContractRepository interface {
	GetInterfaces(context.Context) ([]*ContractInterface, error)
	GetOperationByID(context.Context, []abi.ContractName, bool, uint32) (*ContractOperation, error)
}

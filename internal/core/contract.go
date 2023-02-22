package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/abi"
)

type ContractInterface struct {
	ch.CHModel    `ch:"contract_interfaces" json:"-"`
	bun.BaseModel `bun:"table:contract_interfaces" json:"-"`

	Name       abi.ContractName `ch:",pk" bun:",pk" json:"name"`
	Address    string           `ch:",pk" bun:",pk" json:"address"`
	Code       []byte           `bun:",unique" json:"code"`
	GetMethods []string         `bun:",unique,array" json:"get_methods"`
}

type ContractOperation struct {
	ch.CHModel    `ch:"contract_operations" json:"-"`
	bun.BaseModel `bun:"table:contract_operations" json:"-"`

	Name         string           `bun:",unique" json:"name"`
	ContractName abi.ContractName `ch:",pk" bun:",pk" json:"contract_name"`
	Outgoing     bool             `ch:",pk" bun:",pk" json:"outgoing"` // if operation is going from contract
	OperationID  uint32           `ch:",pk" bun:",pk" json:"operation_id"`
	Schema       []byte           `json:"schema"`
}

type ContractRepository interface {
	GetInterfaces(context.Context) ([]*ContractInterface, error)
	GetOperations(ctx context.Context) ([]*ContractOperation, error)
	GetOperationByID(context.Context, []abi.ContractName, bool, uint32) (*ContractOperation, error)
}

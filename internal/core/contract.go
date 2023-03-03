package core

import (
	"context"
	"encoding/json"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/addr"
)

// TODO: contract addresses labels

type ContractInterface struct {
	ch.CHModel    `ch:"contract_interfaces" json:"-"`
	bun.BaseModel `bun:"table:contract_interfaces" json:"-"`

	Name            abi.ContractName `ch:",pk" bun:",pk" json:"name"`
	Addresses       []*addr.Address  `ch:"type:Array(String),pk" bun:"type:bytea[]" json:"addresses"`
	Code            []byte           `ch:"type:String" bun:"type:bytea,unique" json:"code"`
	GetMethods      []string         `ch:"type:Array(String)" bun:",unique,array" json:"get_methods"`
	GetMethodHashes []uint32         `ch:"type:Array(UInt32)" bun:",unique,array" json:"get_method_hashes"`
}

type ContractOperation struct {
	ch.CHModel    `ch:"contract_operations" json:"-"`
	bun.BaseModel `bun:"table:contract_operations" json:"-"`

	Name         string           `bun:",unique" json:"name"`
	ContractName abi.ContractName `ch:",pk" bun:",pk" json:"contract_name"`
	Outgoing     bool             `ch:",pk" bun:",pk" json:"outgoing"` // if operation is going from contract
	OperationID  uint32           `ch:",pk" bun:",pk" json:"operation_id"`
	Schema       json.RawMessage  `ch:"type:String" bun:"type:jsonb" json:"schema"`
}

type ContractRepository interface {
	GetInterfaces(context.Context) ([]*ContractInterface, error)
	GetOperations(ctx context.Context) ([]*ContractOperation, error)
	GetOperationByID(context.Context, []abi.ContractName, bool, uint32) (*ContractOperation, error)
}

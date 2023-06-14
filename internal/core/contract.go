package core

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
)

type ContractInterface struct {
	bun.BaseModel `bun:"table:contract_interfaces" json:"-"`

	Name            abi.ContractName     `bun:",pk" json:"name"`
	Addresses       []*addr.Address      `bun:"type:bytea[],unique" json:"addresses,omitempty"`
	Code            []byte               `bun:"type:bytea,unique" json:"code,omitempty"`
	GetMethodsDesc  []abi.GetMethodDesc  `bun:"type:text" json:"get_methods_descriptors,omitempty"`
	GetMethodHashes []int32              `bun:"type:integer[]" json:"get_method_hashes,omitempty"`
	Operations      []*ContractOperation `ch:"-" bun:"rel:has-many,join:name=contract_name" json:"operations,omitempty"`
}

type ContractOperation struct {
	bun.BaseModel `bun:"table:contract_operations" json:"-"`

	OperationName string            `json:"operation_name"`
	ContractName  abi.ContractName  `bun:",pk" json:"contract_name"`
	MessageType   MessageType       `bun:"type:message_type,notnull" json:"message_type"` // only internal is supported now
	Outgoing      bool              `bun:",pk" json:"outgoing"`                           // if operation is going from contract
	OperationID   uint32            `bun:",pk" json:"operation_id"`
	Schema        abi.OperationDesc `bun:"type:jsonb" json:"schema"`
}

type ContractRepository interface {
	AddInterface(context.Context, *ContractInterface) error
	AddOperation(context.Context, *ContractOperation) error

	DelInterface(ctx context.Context, name string) error

	GetInterfaces(context.Context) ([]*ContractInterface, error)
	GetOperations(context.Context) ([]*ContractOperation, error)
	GetOperationByID(context.Context, []abi.ContractName, bool, uint32) (*ContractOperation, error)
}

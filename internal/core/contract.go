package core

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
)

type ContractDefinition struct {
	bun.BaseModel `bun:"table:contract_definitions" json:"-"`

	Name   abi.TLBType       `bun:",pk,notnull" json:"name"`
	Schema abi.TLBFieldsDesc `bun:"type:jsonb,notnull" json:"schema"`
}

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
	AddDefinition(context.Context, abi.TLBType, abi.TLBFieldsDesc) error
	UpdateDefinition(ctx context.Context, dn abi.TLBType, d abi.TLBFieldsDesc) error
	DeleteDefinition(ctx context.Context, dn abi.TLBType) error
	GetDefinitions(context.Context) (map[abi.TLBType]abi.TLBFieldsDesc, error)

	AddInterface(context.Context, *ContractInterface) error
	UpdateInterface(ctx context.Context, i *ContractInterface) error
	DeleteInterface(ctx context.Context, name abi.ContractName) error
	GetInterface(ctx context.Context, name abi.ContractName) (*ContractInterface, error)
	GetInterfaces(context.Context) ([]*ContractInterface, error)
	GetMethodDescription(ctx context.Context, name abi.ContractName, method string) (abi.GetMethodDesc, error)

	AddOperation(context.Context, *ContractOperation) error
	UpdateOperation(ctx context.Context, op *ContractOperation) error
	DeleteOperation(ctx context.Context, opName string) error
	GetOperations(context.Context) ([]*ContractOperation, error)
	GetOperationsByID(ctx context.Context, t MessageType, interfaces []abi.ContractName, outgoing bool, id uint32) ([]*ContractOperation, error)
}

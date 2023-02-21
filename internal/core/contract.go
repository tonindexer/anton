package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/abi"
)

type NFTCollectionData struct {
	NextItemIndex uint64
	// OwnerAddress  string
}

type NFTRoyaltyData struct {
	RoyaltyAddress string
	RoyaltyFactor  uint16
	RoyaltyBase    uint16
}

type NFTContentData struct {
	ContentURI         string
	ContentName        string
	ContentDescription string
	ContentImage       string
	ContentImageData   []byte
}

type NFTItemData struct {
	Initialized       bool
	ItemIndex         uint64
	CollectionAddress string
	EditorAddress     string
	// OwnerAddress      string
}

type ContractInterface struct {
	ch.CHModel    `ch:"contract_interfaces"`
	bun.BaseModel `bun:"table:contract_interfaces"`

	Name       abi.ContractName `ch:",pk" bun:",pk"`
	Address    string           //
	Code       []byte           //
	GetMethods []string         //
}

type ContractOperation struct {
	ch.CHModel    `ch:"contract_operations"`
	bun.BaseModel `bun:"table:contract_operations"`

	Name         string           //
	ContractName abi.ContractName `ch:",pk" bun:",pk"`
	Outgoing     bool             // if operation is going from contract
	OperationID  uint32           `ch:",pk" bun:",pk"`
	Schema       []byte           //
}

type ContractRepository interface {
	GetInterfaces(context.Context) ([]*ContractInterface, error)
	GetOperationByID(context.Context, []abi.ContractName, bool, uint32) (*ContractOperation, error)
}

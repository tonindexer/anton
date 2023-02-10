package core

import (
	"context"
	"reflect"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"
)

type ContractType string

const (
	NFTCollection = "nft_collection"
	NFTItem       = "nft_item"
	NFTItemSBT    = "nft_item_sbt"
	NFTRoyalty    = "nft_royalty"
	NFTEditable   = "nft_editable"
	NFTSale       = "nft_sale"
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

	Name       ContractType `ch:",pk" bun:",pk"`
	Address    string       //
	Code       []byte       //
	GetMethods []string     //
}

type ContractOperation struct {
	ch.CHModel    `ch:"contract_operations"`
	bun.BaseModel `bun:"table:contract_operations"`

	Name         string                //
	ContractName ContractType          `ch:",pk" bun:",pk"`
	Outgoing     bool                  // if operation is going from contract
	OperationID  uint32                `ch:",pk" bun:",pk"`
	Schema       string                //
	StructSchema []reflect.StructField `ch:"-"`
}

type ContractRepository interface {
	GetContractInterfaces(context.Context) ([]*ContractInterface, error)
	DetermineContractInterfaces(context.Context, *tlb.Account) ([]ContractType, error)

	InsertContractOperations(context.Context, []*ContractOperation) error
	GetContractOperationByID(context.Context, *AccountState, bool, uint32) (*ContractOperation, error)
}

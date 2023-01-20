package core

import (
	"context"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"
)

type AccountStatus string

const (
	Uninit   = AccountStatus(tlb.AccountStatusUninit)
	Active   = AccountStatus(tlb.AccountStatusActive)
	Frozen   = AccountStatus(tlb.AccountStatusFrozen)
	NonExist = AccountStatus(tlb.AccountStatusNonExist)
)

type Account struct {
	ch.CHModel `ch:"accounts,partition:status,address,types,round(balance,-9)"`

	Types []string `ch:",lc"` // TODO: ContractType here, go-ch bug

	Address   string        `ch:",pk"`
	IsActive  bool          //
	Status    AccountStatus `ch:",lc"` // TODO: enum
	Balance   uint64        // TODO: uint256
	StateHash []byte        //

	Data     []byte //
	DataHash []byte `ch:",pk"`
	Code     []byte //
	CodeHash []byte `ch:",pk"`

	// TODO: do we need it?
	Depth uint64 //
	Tick  bool   //
	Tock  bool   //
	Lib   []byte

	LastTxLT   uint64 //
	LastTxHash []byte //
}

type ContractType string

const (
	NFTCollection = "nft_collection"
	NFTItem       = "nft_item"
	NFTItemSBT    = "nft_item_sbt"
	NFTRoyalty    = "nft_royalty"
	NFTEditable   = "nft_editable"
	NFTSale       = "nft_sale"
)

type ContractInterface struct {
	ch.CHModel `ch:"contract_interfaces"`

	Name       ContractType `ch:",pk"`
	Address    string       //
	Code       []byte       //
	GetMethods []string     //
}

type AccountData struct {
	ch.CHModel `ch:"account_data,partition:types"`

	Address  string   `ch:",pk"`
	DataHash []byte   `ch:",pk"`
	Types    []string `ch:",lc"` // TODO: ContractType here, ch bug
	// GetMethodError string

	// nft collection
	NextItemIndex      uint64
	OwnerAddress       string
	ContentURI         string
	ContentName        string
	ContentDescription string
	ContentImage       string
	ContentImageData   []byte
	RoyaltyFactor      uint16
	RoyaltyBase        uint16
	RoyaltyAddress     string

	// nft item
	Initialized       bool
	ItemIndex         uint64
	CollectionAddress string
	EditorAddress     string
	// OwnerAddress
}

type AccountRepository interface {
	GetContractInterfaces(context.Context) ([]*ContractInterface, error)

	InsertContractOperations(context.Context, []*ContractOperation) error
	GetContractOperationByID(context.Context, *Account, bool, uint32) (*ContractOperation, error)

	AddAccounts(context.Context, []*Account) error
	AddAccountData(context.Context, []*AccountData) error
}

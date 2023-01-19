package core

import (
	"context"
	"reflect"

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

type ContractType string

const (
	NFTCollection = "nft_collection"
	NFTItem       = "nft_item"
	NFTItemSBT    = "nft_item_sbt"
	NFTRoyalty    = "nft_royalty"
	NFTEditable   = "nft_editable"
	NFTSale       = "nft_sale"
	NFTSwap       = "nft_swap"

	JettonWallet = "jetton_wallet"
	JettonMinter = "jetton_minter"

	DNSRoot       = "dns_root"
	DNSCollection = "dns_collect"
	DNSItem       = "dns_item"
)

type ContractOperation struct {
	ch.CHModel `ch:"contract_operations"`

	Name         string                //
	ContractName ContractType          `ch:",pk"`
	OperationID  uint32                `ch:",pk"`
	Schema       string                //
	StructSchema []reflect.StructField `ch:"-"`
}

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

type Account struct {
	ch.CHModel `ch:"accounts,partition:status,address,types,round(balance,-9),last_tx_lt"`

	Address    string        `ch:",pk"`
	IsActive   bool          //
	Status     AccountStatus `ch:",lc"` // TODO: enum
	Types      []string      `ch:",lc"` // TODO: ContractType here, go-ch bug
	Data       []byte        //
	DataHash   []byte        `ch:",pk"`
	Code       []byte        //
	CodeHash   []byte        `ch:",pk"`
	LastTxLT   uint64        //
	LastTxHash []byte        //
	Balance    uint64        // TODO: uint256
}

type AccountRepository interface {
	GetContractInterfaces(context.Context) ([]*ContractInterface, error)

	InsertContractOperations(context.Context, []*ContractOperation) error
	GetContractOperationByID(context.Context, []ContractType, uint32) (*ContractOperation, error)

	AddAccounts(context.Context, []*Account) error
	AddAccountData(context.Context, []*AccountData) error
}

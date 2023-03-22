package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/addr"
)

type AccountStatus string

const (
	Uninit   = AccountStatus(tlb.AccountStatusUninit)
	Active   = AccountStatus(tlb.AccountStatusActive)
	Frozen   = AccountStatus(tlb.AccountStatusFrozen)
	NonExist = AccountStatus(tlb.AccountStatusNonExist)
)

type LatestAccountState struct {
	bun.BaseModel `bun:"table:latest_account_states" json:"-"`

	Address      addr.Address  `bun:"type:bytea,pk,notnull" json:"address"`
	LastTxLT     uint64        `bun:"type:bigint,notnull" json:"last_tx_lt"`
	AccountState *AccountState `bun:"rel:has-one,join:address=address,join:last_tx_lt=last_tx_lt" json:"account"`
}

type AccountState struct {
	ch.CHModel    `ch:"account_states,partition:status" json:"-"`
	bun.BaseModel `bun:"table:account_states" json:"-"`

	Address  addr.Address  `ch:"type:String,pk" bun:"type:bytea,pk,notnull" json:"address"`
	IsActive bool          `json:"is_active"`
	Status   AccountStatus `ch:",lc" bun:"type:account_status" json:"status"` // TODO: ch enum
	Balance  *bunbig.Int   `ch:"type:UInt256" bun:"type:numeric" json:"balance"`

	LastTxLT   uint64 `ch:",pk" bun:"type:bigint,pk,notnull" json:"last_tx_lt"`
	LastTxHash []byte `ch:",pk" bun:"type:bytea,unique,notnull" json:"last_tx_hash"`

	StateData *AccountData `ch:"-" bun:"rel:belongs-to,join:address=address,join:last_tx_lt=last_tx_lt" json:"state_data,omitempty"`

	StateHash []byte `bun:"type:bytea" json:"state_hash,omitempty"` // only if account is frozen
	Code      []byte `bun:"type:bytea" json:"code,omitempty"`
	CodeHash  []byte `bun:"type:bytea" json:"code_hash,omitempty"`
	Data      []byte `bun:"type:bytea" json:"data,omitempty"`
	DataHash  []byte `bun:"type:bytea" json:"data_hash,omitempty"`

	GetMethodHashes []uint32 `ch:"type:Array(UInt32)" bun:",array" json:"get_method_hashes"`
}

type NFTCollectionData struct {
	NextItemIndex *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"next_item_index,omitempty"`
	// OwnerAddress Address
}

type NFTRoyaltyData struct {
	RoyaltyAddress *addr.Address `ch:"type:String" bun:"type:bytea" json:"royalty_address,omitempty"`
	RoyaltyFactor  uint16        `ch:"type:UInt16" bun:"type:integer" json:"royalty_factor,omitempty"`
	RoyaltyBase    uint16        `ch:"type:UInt16" bun:"type:integer" json:"royalty_base,omitempty"`
}

type NFTContentData struct {
	ContentURI         string `ch:"type:String" json:"content_uri,omitempty"`
	ContentName        string `ch:"type:String" json:"content_name,omitempty"`
	ContentDescription string `ch:"type:String" json:"content_description,omitempty"`
	ContentImage       string `ch:"type:String" json:"content_image,omitempty"`
	ContentImageData   []byte `ch:"type:String" json:"content_image_data,omitempty"`
}

type NFTItemData struct {
	Initialized bool        `ch:"type:Bool" json:"initialized,omitempty"`
	ItemIndex   *bunbig.Int `ch:"type:UInt256" json:"item_index,omitempty"`
	// CollectionAddress *addr.Address `ch:"type:String" bun:"type:bytea" json:"collection_address,omitempty"`
	EditorAddress *addr.Address `ch:"type:String" bun:"type:bytea" json:"editor_address,omitempty"`
	// OwnerAddress      *addr.Address
}

type FTMasterData struct {
	TotalSupply  *bunbig.Int   `ch:"type:UInt256" bun:"type:numeric" json:"total_supply,omitempty" swaggertype:"string"`
	Mintable     bool          `json:"mintable,omitempty"`
	AdminAddress *addr.Address `ch:"type:String" bun:"type:bytea" json:"admin_addr,omitempty"`
	// Content     nft.ContentAny
	// WalletCode  *cell.Cell
}

type FTWalletData struct {
	JettonBalance *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"balance,omitempty" swaggertype:"string"`
	// OwnerAddress  Address
	// MasterAddress *addr.Address `ch:"type:String" bun:"type:bytea" json:"master_address,omitempty"`
	// WalletCode    *cell.Cell
}

type AccountData struct {
	ch.CHModel    `ch:"account_data,partition:types" json:"-"`
	bun.BaseModel `bun:"table:account_data" json:"-"`

	Address    addr.Address `ch:"type:String,pk" bun:"type:bytea,pk,notnull" json:"address"`
	LastTxLT   uint64       `ch:",pk" bun:",pk,notnull" json:"last_tx_lt"`
	LastTxHash []byte       `ch:",pk" bun:"type:bytea,notnull,unique" json:"last_tx_hash"`
	Balance    *bunbig.Int  `ch:"type:UInt256" bun:"type:numeric" json:"balance"`

	Types []abi.ContractName `ch:"type:Array(String)" bun:"type:text[],array" json:"types"`

	// common fields for FT and NFT
	OwnerAddress  *addr.Address `ch:"type:String" bun:"type:bytea" json:"owner_address,omitempty"` // universal column for many contracts
	MinterAddress *addr.Address `ch:"type:String" bun:"type:bytea" json:"minter_address,omitempty"`

	NFTCollectionData
	NFTRoyaltyData
	NFTContentData
	NFTItemData

	FTMasterData
	FTWalletData

	Errors []string `json:"error,omitempty"`
}

type AccountStateFilter struct {
	Addresses   []*addr.Address // `form:"addresses"`
	LatestState bool            `form:"latest"`

	// contract data filter
	WithData      bool
	ContractTypes []abi.ContractName `form:"interface"`
	OwnerAddress  *addr.Address      // `form:"owner_address"`
	MinterAddress *addr.Address      // `form:"minter_address"`

	Order string `form:"order"` // ASC, DESC

	AfterTxLT *uint64 `form:"after"`
	Limit     int     `form:"limit"`
}

type AccountRepository interface {
	AddAccountStates(ctx context.Context, tx bun.Tx, states []*AccountState) error
	AddAccountData(ctx context.Context, tx bun.Tx, data []*AccountData) error
	GetAccountStates(ctx context.Context, filter *AccountStateFilter) ([]*AccountState, error)
}

package core

import (
	"context"

	"github.com/uptrace/bun"
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

type AccountState struct {
	ch.CHModel    `ch:"account_states,partition:types,is_active,status"`
	bun.BaseModel `bun:"table:account_states"`

	Latest bool

	Address  string        `ch:",pk"`
	IsActive bool          //
	Status   AccountStatus `ch:",lc" bun:"type:account_status"` // TODO: enum
	Balance  uint64        // TODO: uint256

	LastTxLT   uint64 `ch:",pk"`
	LastTxHash []byte `ch:",pk" bun:"type:bytea,unique"`

	StateHash []byte       `ch:",pk" bun:"type:bytea,pk"`
	StateData *AccountData `ch:"-" bun:"rel:belongs-to,join:state_hash=state_hash"`

	Code     []byte `bun:"type:bytea"`
	CodeHash []byte `bun:"type:bytea"`
	Data     []byte `bun:"type:bytea"`
	DataHash []byte `bun:"type:bytea"`

	// TODO: do we need it?
	Depth uint64 //
	Tick  bool   //
	Tock  bool   //

	Types []string `ch:",lc"` // TODO: ContractType here, go-ch bug
}

type AccountData struct {
	ch.CHModel    `ch:"account_data,partition:types"`
	bun.BaseModel `bun:"table:account_data"`

	Address    string `ch:",pk" bun:",notnull"`
	LastTxLT   uint64 `ch:",pk" bun:",notnull"`
	LastTxHash []byte `ch:",pk" bun:"type:bytea,notnull,unique"`
	StateHash  []byte `ch:",pk" bun:"type:bytea,pk"`

	Types []string `ch:",lc"` // TODO: ContractType here, ch bug

	OwnerAddress string // universal column for many contracts

	NFTCollectionData
	NFTRoyaltyData
	NFTContentData
	NFTItemData
}

type AccountStateFilter struct {
	Address     string
	LatestState bool

	// contract data filter
	WithData          bool
	ContractTypes     []ContractType
	OwnerAddress      string
	CollectionAddress string
}

type AccountRepository interface {
	AddAccountStates(ctx context.Context, states []*AccountState) error
	AddAccountData(ctx context.Context, data []*AccountData) error
	GetAccountStates(ctx context.Context, filter *AccountStateFilter, offset, limit int) ([]*AccountState, error)
}

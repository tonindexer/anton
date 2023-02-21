package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/abi"
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

	Address  string        `ch:",pk" bun:",pk"`
	IsActive bool          //
	Status   AccountStatus `ch:",lc" bun:"type:account_status"` // TODO: enum
	Balance  uint64        // TODO: uint256

	LastTxLT   uint64 `ch:",pk" bun:",pk"`
	LastTxHash []byte `ch:",pk" bun:"type:bytea,unique"`

	StateData *AccountData `ch:"-" bun:"rel:belongs-to,join:address=address,join:last_tx_lt=last_tx_lt"`

	StateHash []byte `bun:"type:bytea"`
	Code      []byte `bun:"type:bytea"`
	CodeHash  []byte `bun:"type:bytea"`
	Data      []byte `bun:"type:bytea"`
	DataHash  []byte `bun:"type:bytea"`

	// TODO: do we need it?
	Depth uint64 //
	Tick  bool   //
	Tock  bool   //

	// TODO: list all get method hashes
	Types []string `ch:",lc"` // TODO: ContractType here, go-ch bug
}

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

type AccountData struct {
	ch.CHModel    `ch:"account_data,partition:types"`
	bun.BaseModel `bun:"table:account_data"`

	Address    string `ch:",pk" bun:",pk,notnull"`
	LastTxLT   uint64 `ch:",pk" bun:",pk,notnull"`
	LastTxHash []byte `ch:",pk" bun:"type:bytea,notnull,unique"`

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
	ContractTypes     []abi.ContractName
	OwnerAddress      string
	CollectionAddress string
}

type AccountRepository interface {
	AddAccountStates(ctx context.Context, tx bun.Tx, states []*AccountState) error
	AddAccountData(ctx context.Context, tx bun.Tx, data []*AccountData) error
	GetAccountStates(ctx context.Context, filter *AccountStateFilter, offset, limit int) ([]*AccountState, error)
}

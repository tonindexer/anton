package core

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
)

type AccountStatus string

const (
	Uninit   = AccountStatus(tlb.AccountStatusUninit)
	Active   = AccountStatus(tlb.AccountStatusActive)
	Frozen   = AccountStatus(tlb.AccountStatusFrozen)
	NonExist = AccountStatus(tlb.AccountStatusNonExist)
)

type LabelCategory string

var (
	CentralizedExchange LabelCategory = "centralized_exchange"
	Scam                LabelCategory = "scam"
)

type AddressLabel struct {
	bun.BaseModel `bun:"table:address_labels" json:"-"`

	Address  addr.Address  `bun:"type:bytea,pk,notnull" json:"-"`
	Name     string        `bun:"type:text,notnull" json:"name"`
	Category LabelCategory `bun:"type:label_category" json:"category,omitempty"`
}

type NFTContentData struct {
	ContentURI         string `ch:"type:String" bun:",nullzero" json:"content_uri,omitempty"`
	ContentName        string `ch:"type:String" bun:",nullzero" json:"content_name,omitempty"`
	ContentDescription string `ch:"type:String" bun:",nullzero" json:"content_description,omitempty"`
	ContentImage       string `ch:"type:String" bun:",nullzero" json:"content_image,omitempty"`
	ContentImageData   []byte `ch:"type:String" bun:",nullzero" json:"content_image_data,omitempty"`
}

type FTWalletData struct {
	JettonBalance *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"balance,omitempty" swaggertype:"string"`
}

type AccountState struct {
	ch.CHModel    `ch:"account_states,partition:toYYYYMM(updated_at)" json:"-"`
	bun.BaseModel `bun:"table:account_states" json:"-"`

	Address addr.Address  `ch:"type:String,pk" bun:"type:bytea,pk,notnull" json:"address"`
	Label   *AddressLabel `ch:"-" bun:"rel:has-one,join:address=address" json:"label,omitempty"`

	Workchain  int32  `bun:"type:integer,notnull" json:"workchain"`
	Shard      int64  `bun:"type:bigint,notnull" json:"shard"`
	BlockSeqNo uint32 `bun:"type:integer,notnull" json:"block_seq_no"`

	IsActive bool          `json:"is_active"`
	Status   AccountStatus `ch:",lc" bun:"type:account_status" json:"status"` // TODO: ch enum

	Balance *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"balance"`

	LastTxLT   uint64 `ch:",pk" bun:"type:bigint,pk,notnull" json:"last_tx_lt"`
	LastTxHash []byte `bun:"type:bytea,unique,notnull" json:"last_tx_hash"`

	StateHash []byte `bun:"type:bytea" json:"state_hash,omitempty"` // only if account is frozen
	Code      []byte `bun:"type:bytea" json:"code,omitempty"`
	CodeHash  []byte `bun:"type:bytea" json:"code_hash,omitempty"`
	Data      []byte `bun:"type:bytea" json:"data,omitempty"`
	DataHash  []byte `bun:"type:bytea" json:"data_hash,omitempty"`

	GetMethodHashes []int32 `ch:"type:Array(UInt32)" bun:"type:integer[]" json:"get_method_hashes"`

	Types []abi.ContractName `ch:"type:Array(String)" bun:"type:text[],array" json:"types"`

	// common fields for FT and NFT
	OwnerAddress  *addr.Address `ch:"type:String" bun:"type:bytea" json:"owner_address,omitempty"` // universal column for many contracts
	MinterAddress *addr.Address `ch:"type:String" bun:"type:bytea" json:"minter_address,omitempty"`

	ExecutedGetMethods map[abi.ContractName][]abi.GetMethodExecution `ch:"-" bun:"type:jsonb" json:"executed_get_methods,omitempty"`

	// TODO: remove this
	NFTContentData
	FTWalletData

	UpdatedAt time.Time `bun:"type:timestamp without time zone,notnull" json:"updated_at"`
}

type LatestAccountState struct {
	bun.BaseModel `bun:"table:latest_account_states" json:"-"`

	Address      addr.Address  `bun:"type:bytea,pk,notnull" json:"address"`
	LastTxLT     uint64        `bun:"type:bigint,notnull" json:"last_tx_lt"`
	AccountState *AccountState `bun:"rel:has-one,join:address=address,join:last_tx_lt=last_tx_lt" json:"account"`
}

type AccountRepository interface {
	AddAccountStates(ctx context.Context, tx bun.Tx, states []*AccountState) error
}

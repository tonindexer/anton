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

	ExecutedGetMethods map[string]abi.GetMethodExecution `ch:"type:json" bun:"type:jsonb" json:"executed_get_methods,omitempty"`

	UpdatedAt time.Time `bun:"type:timestamp without time zone,notnull" json:"updated_at"`
}

type AccountRepository interface {
	AddAccountStates(ctx context.Context, tx bun.Tx, states []*AccountState) error
}

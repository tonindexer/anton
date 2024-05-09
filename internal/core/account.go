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
	ch.CHModel    `ch:"address_labels" json:"-"`
	bun.BaseModel `bun:"table:address_labels" json:"-"`

	Address    addr.Address    `ch:"type:String,pk" bun:"type:bytea,pk,notnull" json:"address"`
	Name       string          `bun:"type:text" json:"name"`
	Categories []LabelCategory `ch:",lc" bun:"type:label_category[]" json:"categories,omitempty"`
}

type NFTContentData struct {
	ContentURI         string `ch:"type:String" bun:",nullzero" json:"content_uri,omitempty"`
	ContentName        string `ch:"type:String" bun:",nullzero" json:"content_name,omitempty"`
	ContentDescription string `ch:"type:String" bun:",nullzero" json:"content_description,omitempty"`
	ContentImage       string `ch:"type:String" bun:",nullzero" json:"content_image,omitempty"`
	ContentImageData   []byte `ch:"type:String" bun:",nullzero" json:"content_image_data,omitempty"`
}

type FTWalletData struct {
	JettonBalance *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"jetton_balance,omitempty" swaggertype:"string"`
}

type AccountStateID struct {
	Address  addr.Address `ch:"type:String"`
	LastTxLT uint64
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
	Code      []byte `ch:"-" bun:"type:bytea" json:"code,omitempty"`
	CodeHash  []byte `bun:"type:bytea" json:"code_hash,omitempty"`
	Data      []byte `ch:"-" bun:"type:bytea" json:"data,omitempty"`
	DataHash  []byte `bun:"type:bytea" json:"data_hash,omitempty"`
	Libraries []byte `bun:"type:bytea" json:"libraries,omitempty"`

	GetMethodHashes []int32 `ch:"type:Array(UInt32)" bun:"type:integer[]" json:"get_method_hashes,omitempty"`

	Types []abi.ContractName `ch:"type:Array(String)" bun:"type:text[],array" json:"types,omitempty"`

	// common fields for FT and NFT
	OwnerAddress  *addr.Address `ch:"type:String" bun:"type:bytea" json:"owner_address,omitempty"` // universal column for many contracts
	MinterAddress *addr.Address `ch:"type:String" bun:"type:bytea" json:"minter_address,omitempty"`

	Fake bool `ch:"type:Bool" bun:"type:boolean" json:"fake"`

	ExecutedGetMethods map[abi.ContractName][]abi.GetMethodExecution `ch:"type:String" bun:"type:jsonb" json:"executed_get_methods,omitempty"`

	// TODO: remove this
	NFTContentData
	FTWalletData

	UpdatedAt time.Time `bun:"type:timestamp without time zone,notnull" json:"updated_at"`
}

type AccountStateCode struct {
	ch.CHModel `ch:"account_states_code" json:"-"`

	CodeHash []byte `ch:"type:String"`
	Code     []byte `ch:"type:String"`
}

type AccountStateData struct {
	ch.CHModel `ch:"account_states_data" json:"-"`

	DataHash []byte `ch:"type:String"`
	Data     []byte `ch:"type:String"`
}

func (a *AccountState) BlockID() BlockID {
	return BlockID{
		Workchain: a.Workchain,
		Shard:     a.Shard,
		SeqNo:     a.BlockSeqNo,
	}
}

type LatestAccountState struct {
	bun.BaseModel `bun:"table:latest_account_states" json:"-"`

	Address      addr.Address  `bun:"type:bytea,pk,notnull" json:"address"`
	LastTxLT     uint64        `bun:"type:bigint,notnull" json:"last_tx_lt"`
	AccountState *AccountState `bun:"rel:has-one,join:address=address,join:last_tx_lt=last_tx_lt" json:"account"`
}

func SkipAddress(a addr.Address) bool {
	switch a.Base64() {
	case "EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM9c": // burn address
		return true
	case "Ef8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAU": // system contract
		return true
	case "Ef8zMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzM0vF": // elector contract
		return true
	case "Ef80UXx731GHxVr0-LYf3DIViMerdo3uJLAG3ykQZFjXz2kW": // log tests contract
		return true
	case "Ef9VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVbxn": // config contract
		return true
	case "EQAHI1vGuw7d4WG-CtfDrWqEPNtmUuKjKFEFeJmZaqqfWTvW": // BSC Bridge Collector
		return true
	case "EQCuzvIOXLjH2tv35gY4tzhIvXCqZWDuK9kUhFGXKLImgxT5": // ETH Bridge Collector
		return true
	case "EQA2u5Z5Fn59EUvTI-TIrX8PIGKQzNj3qLixdCPPujfJleXC",
		"EQA2Pnxp0rMB9L6SU2z1VqfMIFIfutiTjQWFEXnwa_zPh0P3",
		"EQDhIloDu1FWY9WFAgQDgw0RjuT5bLkf15Rmd5LCG3-0hyoe": // strange heavy testnet address
		return true
	case "EQAWBIxrfQDExJSfFmE5UL1r9drse0dQx_eaV8w9S77VK32F": // tongo emulator segmentation fault
		return true
	default:
		return false
	}
}

type AccountRepository interface {
	AddAddressLabel(context.Context, *AddressLabel) error
	GetAddressLabel(context.Context, addr.Address) (*AddressLabel, error)

	AddAccountStates(ctx context.Context, tx bun.Tx, states []*AccountState) error
	UpdateAccountStates(ctx context.Context, states []*AccountState) error

	// MatchStatesByInterfaceDesc returns (address, last_tx_lt) pairs for suitable account states.
	MatchStatesByInterfaceDesc(ctx context.Context,
		contractName abi.ContractName,
		addresses []*addr.Address,
		codeHash []byte,
		getMethodHashes []int32,
		afterAddress *addr.Address,
		afterTxLt uint64,
		limit int) ([]*AccountStateID, error)

	// GetAllAccountInterfaces returns transaction LT, on which contract interface was updated.
	// It also considers, that contract can be both upgraded and downgraded.
	GetAllAccountInterfaces(context.Context, addr.Address) (map[uint64][]abi.ContractName, error)

	// GetAllAccountStates is pretty much similar to GetAllAccountInterfaces, but it returns updates of code or data.
	GetAllAccountStates(ctx context.Context, a addr.Address, beforeTxLT uint64, limit int) ([]*AccountState, error)
}

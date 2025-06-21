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

type AccountBlockStateID struct {
	Address    addr.Address `ch:"type:String"`
	Workchain  int32
	Shard      int64
	BlockSeqNo uint32
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

	Address  addr.Address `bun:"type:bytea,pk,notnull" json:"address"`
	LastTxLT uint64       `bun:"type:bigint,notnull" json:"last_tx_lt"`

	Types []abi.ContractName `bun:"type:text[],array" json:"types,omitempty"`

	OwnerAddress  *addr.Address `bun:"type:bytea" json:"owner_address,omitempty"` // universal column for many contracts
	MinterAddress *addr.Address `bun:"type:bytea" json:"minter_address,omitempty"`

	CreatedLT uint64 `bun:"type:bigint,notnull" json:"created_lt"`

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
	case "EQCnBscEi-KGfqJ5Wk6R83yrqtmUum94SXnSDz3AOQfHGjDw",
		"EQA9xJgsYbsTjWxEcaxv8DLW3iRJtHzjwFzFAEWVxup0WH0R": // quackquack (?)
		return true
	case "EQCqNjAPkigLdS5gxHiHitWuzF3ZN-gX7MlX4Qfy2cGS3FWx": // ton-squid
		return true
	case "EQCp6qUScSUYB66ExDIlla8kfnUpP5cLZ_zhy4nlOPC-fqFo": // highload wallet v2 with heavy data
		return true
	case "EQC1Bq1GJY9ON_2WpSroVlXpejzfLNA8XoL2MYxtN50ZbJfN": // TryTON
		return true
	case "EQCTsnUmD2wvN-SBaa7CMF1sgTfC-YNywqbdPepKw34VBglS": // TryTON NFT collection
		return true
	case "EQCatS3EvWAhYaFEmLK_rOWViVgzN9RrHYh_PpNQ01X_WTPh": // TON lama jetton distribution
		return true
	case "EQBvc1QLuqTMx0NNTZ4DD__UzfTvkEOJMs67XoZhHVihWtMN": // POO jetton distribution
		return true
	case "EQDF6fj6ydJJX_ArwxINjP-0H8zx982W4XgbkKzGvceUWvXl": // ETH Token Bridge Collector
		return true
	case "EQC_0ScHnb7bVoyInXLkZ2G4XRHg97S9XrPKCUDaO1ZRyFhZ": // Gemz Checkin
		return true
	case "EQD_QUnVTBzwG-8GCkqnQ4xiWxU0oPZn9Pon_rq0MZVdIBuf",
		"EQB2MfIcTbwtshE8VOv0YA6ZWpb9bbj79D_SUXHZYv04X47c": // Wonton (?)
		return true
	case "EQAqk4SStGaodBsjW0zc8H4psrsx258cCdqw4Nm3ScnMYpLf": // some service (?)
		return true
	case "EQDlHrYvmV9R91wNbqvpzo-_pXu4Q6vQZo0-t2CplC6Zgh4y": // RBT trader
		return true
	case "EQD5iFPj0zk1mA-GatG_3QtBNWVzuRKatszH1MUYAw6aVeK2": // some service claims (?)
		return true
	case "EQCfrctTcgYp6cd2iqgAVKiLKauJvBNC4sc84xYBvspyw3q7",
		"EQAlMRLTYOoG6kM0d3dLHqgK30ol3qIYwMNtEelktzXP_pD5",
		"EQDa5wUCdTj1tqYV-LyIcefBHd3IGacvzhcBrSjmlKY2xnaK",
		"EQAU35_2hAbymisgUrhGa4bIJUtEJjVNVS7zBrqfKaENd67N",
		"EQCxr1o-x7cEFb3vALiYMOW7QPuAoGHMtw1Yab5m6HrnuIuZ",
		"EQDCR0XQ0qNQJNjITRpo59mFsP0pjx81ImtXx92mJBnIc7m4",
		"EQAYNJOQTA9FqZF4QGxzcPEvvMWkP76snfI7gATCur_86psC",
		"EQD-r3joXyZ2kWRxraqze6ypKoVtSx1qlKlJsNEjyLM7ujs7",
		"EQDTCD85dI5Cu8O1eDecuARaagwaOPMacnXwqn8KB0-1DN8P": // unknown
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

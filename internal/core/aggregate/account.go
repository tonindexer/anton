package aggregate

import (
	"context"

	"github.com/uptrace/bun/extra/bunbig"

	"github.com/iam047801/tonidx/internal/addr"
)

type AccountStatesReq struct {
	MinterAddress *addr.Address // NFT or FT minter

	Limit int `form:"limit"`
}

type AccountStatesRes struct {
	// NFT minter
	Items       int `json:"items,omitempty"`
	OwnersCount int `json:"owners_count,omitempty"`
	OwnedItems  []*struct {
		OwnerAddress *addr.Address `ch:"type:String" json:"owner_address"`
		ItemsCount   int           `json:"items_count"`
	} `json:"owned_items,omitempty"`
	UniqueOwners []*struct {
		ItemAddress *addr.Address `ch:"type:String" json:"item_address"`
		OwnersCount int           `json:"owners_count"`
	} `json:"unique_owners,omitempty"`

	// FT minter
	Wallets      int         `json:"wallets,omitempty"`
	TotalSupply  *bunbig.Int `json:"total_supply,omitempty"`
	OwnedBalance []*struct {
		WalletAddress *addr.Address `ch:"item_address,type:String" json:"wallet_address"`
		OwnerAddress  *addr.Address `ch:"type:String" json:"owner_address"`
		Balance       *bunbig.Int   `ch:"type:UInt256" json:"balance"`
	} `json:"owned_balance,omitempty"`
}

type AccountRepository interface {
	AggregateAccountStates(ctx context.Context, req *AccountStatesReq) (*AccountStatesRes, error)
}

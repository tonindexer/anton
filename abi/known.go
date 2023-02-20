package abi

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type ContractName string

const (
	Wallet     = "wallet"
	WalletV4R2 = "wallet_v4r2"

	NFTCollection = "nft_collection"
	NFTItem       = "nft_item"
	NFTRoyalty    = "nft_royalty"
	NFTEditable   = "nft_editable"
	NFTItemSBT    = "nft_item_sbt"
	NFTSale       = "nft_sale"

	JettonMinter = "jetton_minter"
	JettonWallet = "jetton_wallet"

	TelemintNFTCollection = "telemint_nft_collection"
	TelemintNFTItem       = "telemint_nft_item"
	TelemintNFTDNS        = "telemint_nft_dns"

	DNSResolver = "dns_resolver"
)

type (
	NFTCollectionItemMint      nft.ItemMintPayload
	NFTCollectionItemMintBatch struct {
		_       tlb.Magic `tlb:"#00000002"`
		QueryID uint64    `tlb:"## 64"`
		// TODO: content dictionary
	}
	NFTCollectionChangeOwner   nft.CollectionChangeOwner
	NFTCollectionChangeContent struct {
		_       tlb.Magic `tlb:"#00000004"`
		QueryID uint64    `tlb:"## 64"`
		// TODO: content
	}
	NFTCollectionGetRoyaltyParams struct {
		_ tlb.Magic `tlb:"#693d3950"`
	}

	NFTCollectionReportRoyaltyParams struct {
		_ tlb.Magic `tlb:"#a8cb00ad"`
	}

	NFTItemTransfer      nft.TransferPayload
	NFTItemEdit          nft.ItemEditPayload
	NFTItemGetStaticData struct {
		_ tlb.Magic `tlb:"#2fcb26a2"`
	}
	NFTItemTransferEditorship struct {
		_                   tlb.Magic        `tlb:"#1c04412a"`
		QueryID             uint64           `tlb:"## 64"`
		NewOwner            *address.Address `tlb:"addr"`
		ResponseDestination *address.Address `tlb:"addr"`
		CustomPayload       *cell.Cell       `tlb:"maybe ^"`
		ForwardAmount       tlb.Coins        `tlb:"."`
		ForwardPayload      *cell.Cell       `tlb:"either . ^"`
	}

	NFTItemReportStaticData struct {
		_       tlb.Magic `tlb:"#8b771735"`
		QueryID uint64    `tlb:"## 64"`
	}
	NFTItemOwnershipAssigned struct {
		_       tlb.Magic `tlb:"#05138d91"`
		QueryID uint64    `tlb:"## 64"`
	}
	NFTItemEditorshipAssigned struct {
		_       tlb.Magic `tlb:"#511a4463"`
		QueryID uint64    `tlb:"## 64"`
	}

	JettonMint             jetton.MintPayload
	JettonTransfer         jetton.TransferPayload
	JettonInternalTransfer struct {
		_                tlb.Magic        `tlb:"#178d4519"`
		QueryID          uint64           `tlb:"## 64"`
		Amount           tlb.Coins        `tlb:"."`
		From             *address.Address `tlb:"addr"`
		ResponseAddress  *address.Address `tlb:"addr"`
		ForwardTONAmount tlb.Coins        `tlb:"."`
		ForwardPayload   *cell.Cell       `tlb:"either . ^"`
	}
	JettonTransferNotification struct {
		_              tlb.Magic        `tlb:"#7362d09c"`
		QueryID        uint64           `tlb:"## 64"`
		Amount         tlb.Coins        `tlb:"."`
		Sender         *address.Address `tlb:"addr"`
		ForwardPayload *cell.Cell       `tlb:"either . ^"`
	}
	JettonBurn jetton.BurnPayload

	TeleitemAuctionConfig struct {
		BeneficiaryAddress *address.Address `tlb:"addr"`
		InitialMinBid      tlb.Coins        `tlb:"."`
		MaxBid             tlb.Coins        `tlb:"."`
		MinBidStep         uint8            `tlb:"## 8"`
		MinExtendTime      uint32           `tlb:"## 32"`
		Duration           uint32           `tlb:"## 32"`
	}
	TelemintRoyaltyParams struct {
		Numerator   uint16           `tlb:"## 16"`
		Denominator uint16           `tlb:"## 16"`
		Destination *address.Address `tlb:"addr"`
	}

	TelemintMsgDeploy struct {
		_             tlb.Magic              `tlb:"#4637289b"`
		Sig           []byte                 `tlb:"bits 512"`
		SubwalletID   uint32                 `tlb:"## 32"`
		ValidSince    uint32                 `tlb:"## 32"`
		ValidTill     uint32                 `tlb:"## 32"`
		TokenName     *TelemintText          `tlb:"."`
		Content       *cell.Cell             `tlb:"^"`
		AuctionConfig *TeleitemAuctionConfig `tlb:"^"`
		RoyaltyParams *TelemintRoyaltyParams `tlb:"maybe ^"`
	}
	TeleitemMsgDeploy struct {
		_             tlb.Magic              `tlb:"#299a3e15"`
		SenderAddress *address.Address       `tlb:"addr"`
		Bid           tlb.Coins              `tlb:"."`
		Info          *cell.Cell             `tlb:"^"`
		Content       *cell.Cell             `tlb:"^"`
		AuctionConfig *TeleitemAuctionConfig `tlb:"^"`
		RoyaltyParams *TelemintRoyaltyParams `tlb:"^"`
	}
	TeleitemStartAuction struct {
		_             tlb.Magic              `tlb:"#487a8e81"`
		QueryID       uint64                 `tlb:"## 64"`
		AuctionConfig *TeleitemAuctionConfig `tlb:"^"`
	}
	TeleitemCancelAuction struct {
		_       tlb.Magic `tlb:"#371638ae"`
		QueryID uint64    `tlb:"## 64"`
	}
	TeleitemOK struct {
		_ tlb.Magic `tlb:"#a37a0983"`
	}
	TeleitemOutbidNotification struct {
		_ tlb.Magic `tlb:"#557cea20"`
	}

	Excesses struct {
		_       tlb.Magic `tlb:"#d53276db"`
		QueryID uint64    `tlb:"## 64"`
	}
)

var (
	// KnownContractOperations is a map[contract] -> map[is outgoing message] -> message schema
	KnownContractOperations = map[ContractName]map[bool][]any{
		NFTCollection: {
			false: []any{
				(*NFTCollectionItemMint)(nil), (*NFTCollectionItemMintBatch)(nil),
			},
			true: []any{},
		},

		NFTRoyalty: {
			false: []any{
				(*NFTCollectionGetRoyaltyParams)(nil),
			},
			true: []any{
				(*NFTCollectionReportRoyaltyParams)(nil),
			},
		},

		NFTEditable: {
			false: []any{
				(*NFTCollectionChangeOwner)(nil), (*NFTCollectionChangeContent)(nil),
				(*NFTItemEdit)(nil), (*NFTItemTransferEditorship)(nil),
			},
			true: []any{
				(*NFTItemEditorshipAssigned)(nil),
			},
		},

		NFTItem: {
			false: []any{
				(*NFTItemTransfer)(nil),
				(*NFTItemGetStaticData)(nil),
			},
			true: []any{
				(*NFTItemOwnershipAssigned)(nil),
				(*Excesses)(nil),
				(*NFTItemReportStaticData)(nil),
			},
		},

		JettonMinter: {
			false: []any{
				(*JettonMint)(nil),
			},
		},

		JettonWallet: {
			false: {
				(*JettonTransfer)(nil), (*JettonInternalTransfer)(nil), (*JettonTransferNotification)(nil),
				(*JettonBurn)(nil),
			},
		},

		TelemintNFTCollection: {
			false: {
				(*TelemintMsgDeploy)(nil),
			},
			true: {
				(*TeleitemMsgDeploy)(nil),
			},
		},
		TelemintNFTItem: {
			false: {
				(*TeleitemMsgDeploy)(nil),
				(*TeleitemStartAuction)(nil), (*TeleitemCancelAuction)(nil),
			},
			true: {
				(*TeleitemOK)(nil), (*TeleitemOutbidNotification)(nil),
			},
		},
	}

	KnownContractMethods = map[ContractName][]string{
		Wallet:     {"seqno", "get_public_key"},
		WalletV4R2: {"get_subwallet_id", "is_plugin_installed"},

		NFTCollection: {"get_collection_data", "get_nft_address_by_index", "get_nft_content"},
		NFTRoyalty:    {"royalty_params"},
		NFTEditable:   {"get_editor"},
		NFTItem:       {"get_nft_data"},
		NFTItemSBT:    {"get_nonce", "get_public_key", "get_authority_address"},
		NFTSale:       {"get_sale_data"},

		JettonMinter: {"get_jetton_data", "get_wallet_address"},
		JettonWallet: {"get_wallet_data"},

		DNSResolver: {"dnsresolve"},

		TelemintNFTCollection: {"get_collection_data", "get_nft_address_by_index", "get_nft_content"},
		TelemintNFTItem:       {"get_nft_data", "get_telemint_token_name", "get_telemint_auction_state", "get_telemint_auction_config"},
		TelemintNFTDNS:        {"get_full_domain"},
	}
)

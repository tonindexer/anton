package abi

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

var (
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

		// TelemintNFTCollection: {"get_collection_data", "get_nft_address_by_index", "get_nft_content"},
		TelemintNFTItem: {"get_telemint_token_name", "get_telemint_auction_state", "get_telemint_auction_config"},
		TelemintNFTDNS:  {"get_full_domain"},
	}

	KnownAddresses = map[string]ContractName{
		"EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N": TelemintNFTCollection,
		"EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi": TelemintNFTCollection,
	}

	// KnownContractOperations is a map[contract] -> map[is outgoing message] -> message schema
	KnownContractOperations = map[ContractName]map[bool][]any{
		NFTCollection: {
			false: []any{
				(*NFTCollectionItemMint)(nil), (*NFTCollectionItemMintBatch)(nil),
				(*NFTCollectionChangeOwner)(nil), (*NFTCollectionChangeContent)(nil),
			},
		},

		NFTRoyalty: {
			false: []any{
				(*NFTGetRoyaltyParams)(nil),
			},
			true: []any{
				(*NFTReportRoyaltyParams)(nil),
			},
		},

		NFTEditable: {
			false: []any{
				(*NFTEdit)(nil), (*NFTTransferEditorship)(nil),
			},
			true: []any{
				(*NFTEditorshipAssigned)(nil),
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
				(*JettonTransfer)(nil), (*JettonInternalTransfer)(nil),
				(*JettonBurn)(nil),
			},
			true: {
				(*JettonTransferNotification)(nil),
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
				// (*TeleitemMsgDeploy)(nil),
				(*TeleitemStartAuction)(nil), (*TeleitemCancelAuction)(nil),
			},
			true: {
				(*TeleitemOK)(nil), (*TeleitemOutbidNotification)(nil),
			},
		},
	}
)

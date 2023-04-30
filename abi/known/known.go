package known

import "github.com/tonindexer/anton/abi"

var (
	NFTCollection abi.ContractName = "nft_collection"
	NFTItem       abi.ContractName = "nft_item"
	NFTEditable   abi.ContractName = "nft_editable"
	NFTRoyalty    abi.ContractName = "nft_royalty"

	JettonMinter abi.ContractName = "jetton_minter"
	JettonWallet abi.ContractName = "jetton_wallet"
)

func GetAllWalletNames() []abi.ContractName {
	return []abi.ContractName{
		"wallet_v1r1",
		"wallet_v1r2",
		"wallet_v1r3",
		"wallet_v2r1",
		"wallet_v2r2",
		"wallet_v3r1",
		"wallet_v3r2",
		"wallet_v4r1",
		"wallet_v4r2",
		"wallet_lockup",
		"wallet_highload_v1r1",
		"wallet_highload_v1r2",
		"wallet_highload_v2r1",
		"wallet_highload_v2r2",
	}
}

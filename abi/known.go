package abi

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/tvm/cell"
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
	NFTItemExcesses struct {
		_       tlb.Magic `tlb:"#d53276db"`
		QueryID uint64    `tlb:"## 64"`
	}

	JettonMint     jetton.MintPayload
	JettonTransfer jetton.TransferPayload
	JettonBurn     jetton.BurnPayload
)

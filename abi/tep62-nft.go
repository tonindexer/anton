package abi

import (
	"context"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// https://github.com/ton-blockchain/TEPs/blob/master/text/0062-nft-standard.md
// https://github.com/ton-blockchain/token-contract/tree/main/nft)

// NFTCollection
type (
	NFTCollectionData nft.CollectionData

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
)

// NFTRoyalty
type (
	NFTRoyaltyData nft.CollectionRoyaltyParams

	NFTGetRoyaltyParams struct {
		_ tlb.Magic `tlb:"#693d3950"`
	}
	NFTReportRoyaltyParams struct {
		_ tlb.Magic `tlb:"#a8cb00ad"`
	}
)

// NFTItem
type (
	NFTItemData nft.ItemData

	NFTItemTransfer      nft.TransferPayload
	NFTItemGetStaticData struct {
		_ tlb.Magic `tlb:"#2fcb26a2"`
	}
	NFTItemOwnershipAssigned struct {
		_       tlb.Magic `tlb:"#05138d91"`
		QueryID uint64    `tlb:"## 64"`
	}
	NFTItemReportStaticData struct {
		_       tlb.Magic `tlb:"#8b771735"`
		QueryID uint64    `tlb:"## 64"`
	}
)

// NFTEditable
type (
	NFTEditableData struct {
		Editor *address.Address
	}

	NFTEdit               nft.ItemEditPayload
	NFTTransferEditorship struct {
		_                   tlb.Magic        `tlb:"#1c04412a"`
		QueryID             uint64           `tlb:"## 64"`
		NewOwner            *address.Address `tlb:"addr"`
		ResponseDestination *address.Address `tlb:"addr"`
		CustomPayload       *cell.Cell       `tlb:"maybe ^"`
		ForwardAmount       tlb.Coins        `tlb:"."`
		ForwardPayload      *cell.Cell       `tlb:"either . ^"`
	}
	NFTEditorshipAssigned struct {
		_       tlb.Magic `tlb:"#511a4463"`
		QueryID uint64    `tlb:"## 64"`
	}
)

type Excesses struct {
	_       tlb.Magic `tlb:"#d53276db"`
	QueryID uint64    `tlb:"## 64"`
}

func GetNFTCollectionData(ctx context.Context, api *ton.APIClient, b *ton.BlockIDExt, addr *address.Address) (*NFTCollectionData, error) {
	c := nft.NewCollectionClient(api, addr)

	data, err := c.GetCollectionDataAtBlock(ctx, b)
	if err != nil {
		return nil, err
	}

	return (*NFTCollectionData)(data), nil
}

func GetNFTRoyaltyData(ctx context.Context, api *ton.APIClient, b *ton.BlockIDExt, addr *address.Address) (*NFTRoyaltyData, error) {
	c := nft.NewCollectionClient(api, addr)

	data, err := c.RoyaltyParamsAtBlock(ctx, b)
	if err != nil {
		return nil, err
	}

	return (*NFTRoyaltyData)(data), nil
}

func GetNFTItemData(ctx context.Context, api *ton.APIClient, b *ton.BlockIDExt, addr *address.Address) (*NFTItemData, error) {
	c := nft.NewItemClient(api, addr)

	data, err := c.GetNFTDataAtBlock(ctx, b)
	if err != nil {
		return nil, errors.Wrap(err, "get nft item data")
	}

	if data.CollectionAddress != nil && data.Content != nil {
		collect := nft.NewCollectionClient(api, data.CollectionAddress)

		data.Content, err = collect.GetNFTContentAtBlock(ctx, data.Index, data.Content, b)
		if err != nil {
			return nil, errors.Wrap(err, "get nft content")
		}
	}

	return (*NFTItemData)(data), nil
}

func GetNFTEditableData(ctx context.Context, api *ton.APIClient, b *ton.BlockIDExt, addr *address.Address) (*NFTEditableData, error) {
	c := nft.NewItemEditableClient(api, addr)

	editor, err := c.GetEditorAtBlock(ctx, b)
	if err != nil {
		return nil, errors.Wrap(err, "get editor")
	}

	return &NFTEditableData{Editor: editor}, nil
}

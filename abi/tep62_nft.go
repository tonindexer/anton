package abi

import (
	"context"

	"github.com/pkg/errors"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/nft"
)

type NFTCollectionData nft.CollectionData

type NFTRoyaltyData nft.CollectionRoyaltyParams

type NFTItemData nft.ItemData

type NFTEditableData struct {
	Editor *address.Address
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

	if data.CollectionAddress != nil && !data.CollectionAddress.IsAddrNone() && data.Content != nil {
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

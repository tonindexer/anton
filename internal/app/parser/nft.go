package parser

import (
	"context"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/nft"

	"github.com/iam047801/tonidx/internal/core"
)

func getContentDataNFT(ret *core.AccountData, c nft.ContentAny) {
	switch content := c.(type) {
	case *nft.ContentSemichain: // TODO: remove this (?)
		ret.ContentURI = content.URI
		ret.ContentName = content.Name
		ret.ContentDescription = content.Description
		ret.ContentImage = content.Image
		ret.ContentImageData = content.ImageData

	case *nft.ContentOnchain:
		ret.ContentName = content.Name
		ret.ContentDescription = content.Description
		ret.ContentImage = content.Image
		ret.ContentImageData = content.ImageData

	case *nft.ContentOffchain:
		ret.ContentURI = content.URI
	}
}

func (s *Service) getAccountDataNFT(ctx context.Context, master *tlb.BlockInfo, acc *core.Account, ret *core.AccountData) error {
	ret.Address = acc.Address
	ret.DataHash = acc.DataHash

	addr, err := address.ParseAddr(acc.Address)
	if err != nil {
		return errors.Wrap(err, "parse address")
	}

	var collection, item, editable, royalty bool

	for _, t := range acc.Types {
		switch t {
		case core.NFTCollection:
			collection = true
		case core.NFTItem, core.NFTItemSBT:
			item = true
		case core.NFTEditable:
			editable = true
		case core.NFTRoyalty:
			royalty = true
		}
	}

	switch {
	case collection:
		c := nft.NewCollectionClient(s.api, addr)

		data, err := c.GetCollectionDataAtBlock(ctx, master)
		if err != nil {
			return errors.Wrap(err, "get collection data")
		}

		ret.Types = append(ret.Types, core.NFTCollection)

		ret.NextItemIndex = data.NextItemIndex.Uint64()
		ret.OwnerAddress = data.OwnerAddress.String()
		getContentDataNFT(ret, data.Content)

		if royalty {
			params, err := c.RoyaltyParamsAtBlock(ctx, master)
			if err != nil {
				return errors.Wrap(err, "get royalty params")
			}
			ret.RoyaltyAddress = params.Address.String()
			ret.RoyaltyBase = params.Base
			ret.RoyaltyFactor = params.Factor
		}

	case item:
		c := nft.NewItemClient(s.api, addr)

		data, err := c.GetNFTDataAtBlock(ctx, master)
		if err != nil {
			return errors.Wrap(err, "get nft item data")
		}

		ret.Types = append(ret.Types, core.NFTItem)

		ret.Initialized = data.Initialized
		ret.ItemIndex = data.Index.Uint64()
		ret.CollectionAddress = data.CollectionAddress.String()
		ret.OwnerAddress = data.OwnerAddress.String()

		if data.Content != nil {
			collect := nft.NewCollectionClient(s.api, data.CollectionAddress)
			con, err := collect.GetNFTContentAtBlock(ctx, data.Index, data.Content, master)
			if err != nil {
				return errors.Wrap(err, "get nft content")
			}
			getContentDataNFT(ret, con)
		}

		if editable {
			c := nft.NewItemEditableClient(s.api, addr)

			ret.Types = append(ret.Types, core.NFTEditable)

			editor, err := c.GetEditorAtBlock(ctx, master)
			if err != nil {
				return errors.Wrap(err, "get editor")
			}

			ret.EditorAddress = editor.String()
		}

	default:
		return errors.Wrap(core.ErrNotAvailable, "get account nft data")
	}

	return nil
}

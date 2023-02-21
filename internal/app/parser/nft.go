package parser

import (
	"context"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/nft"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/core"
)

func mapContentDataNFT(ret *core.AccountData, c nft.ContentAny) {
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

func mapCollectionDataNFT(ret *core.AccountData, data *abi.NFTCollectionData) {
	ret.NextItemIndex = data.NextItemIndex.Uint64()
	ret.OwnerAddress = data.OwnerAddress.String()
	mapContentDataNFT(ret, data.Content)
}

func mapRoyaltyDataNFT(ret *core.AccountData, params *abi.NFTRoyaltyData) {
	ret.RoyaltyAddress = params.Address.String()
	ret.RoyaltyBase = params.Base
	ret.RoyaltyFactor = params.Factor
}

func mapItemDataNFT(ret *core.AccountData, data *abi.NFTItemData) {
	ret.Initialized = data.Initialized
	ret.ItemIndex = data.Index.Uint64()
	ret.CollectionAddress = data.CollectionAddress.String()
	ret.OwnerAddress = data.OwnerAddress.String()
	mapContentDataNFT(ret, data.Content)
}

func mapEditorDataNFT(ret *core.AccountData, data *abi.NFTEditableData) {
	ret.EditorAddress = data.Editor.String()
}

func (s *Service) getAccountDataNFT(ctx context.Context, b *tlb.BlockInfo, acc *tlb.Account, types []abi.ContractName, ret *core.AccountData) error {
	var unknown int

	addr := acc.State.Address

	for _, t := range types {
		switch t {
		case abi.NFTCollection:
			data, err := abi.GetNFTCollectionData(ctx, s.api, b, addr)
			if err != nil {
				return errors.Wrap(err, "get nft collection data")
			}
			mapCollectionDataNFT(ret, data)

		case abi.NFTRoyalty:
			data, err := abi.GetNFTRoyaltyData(ctx, s.api, b, addr)
			if err != nil {
				return errors.Wrap(err, "get nft royalty data")
			}
			mapRoyaltyDataNFT(ret, data)

		case abi.NFTItem:
			data, err := abi.GetNFTItemData(ctx, s.api, b, addr)
			if err != nil {
				return errors.Wrap(err, "get nft item data")
			}
			mapItemDataNFT(ret, data)

		case abi.NFTEditable:
			data, err := abi.GetNFTEditableData(ctx, s.api, b, addr)
			if err != nil {
				return errors.Wrap(err, "get nft editable data")
			}
			mapEditorDataNFT(ret, data)

		default:
			unknown++
		}

		ret.Types = append(ret.Types, string(t))
	}
	if unknown == len(types) {
		return errors.Wrap(core.ErrNotAvailable, "unknown contract")
	}

	return nil
}

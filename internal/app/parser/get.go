package parser

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun/extra/bunbig"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/abi/known"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func getMethodByName(i *core.ContractInterface, n string) *abi.GetMethodDesc {
	for it := range i.GetMethodsDesc {
		if i.GetMethodsDesc[it].Name == n {
			return &i.GetMethodsDesc[it]
		}
	}
	return nil
}

func (s *Service) getMethodCall(ctx context.Context, d *abi.GetMethodDesc, acc *core.AccountState, args []any) (abi.VmStack, error) {
	var argsStack abi.VmStack

	if len(acc.Code) == 0 || len(acc.Data) == 0 {
		return nil, errors.Wrap(app.ErrImpossibleParsing, "no account code or data")
	}

	if len(d.Arguments) != len(args) {
		return nil, errors.New("length of passed and described arguments does not match")
	}
	for it := range args {
		argsStack = append(argsStack, abi.VmValue{
			VmValueDesc: d.Arguments[it],
			Payload:     args[it],
		})
	}

	code, err := cell.FromBOC(acc.Code)
	if err != nil {
		return nil, errors.Wrap(err, "account code from boc")
	}
	data, err := cell.FromBOC(acc.Data)
	if err != nil {
		return nil, errors.Wrap(err, "account data from boc")
	}

	e, err := abi.NewEmulator(acc.Address.MustToTonutils(), code, data, s.bcConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new emulator")
	}

	ret, err := e.RunGetMethod(ctx, d.Name, argsStack, d.ReturnValues)
	if err != nil {
		return nil, errors.Wrap(err, "run get method")
	}

	return ret, nil
}

// TODO: map automatically by field names with reflect
// TODO: check return values descriptors, do not panic on wrong type assertions

func (s *Service) getMethodCallNoArgs(ctx context.Context, i *core.ContractInterface, gmName string, acc *core.AccountState) (abi.VmStack, error) {
	gm := getMethodByName(i, gmName)
	if gm == nil {
		// we panic as contract interface was defined, but there are no standard get-method
		panic(fmt.Errorf("%s `%s` get-method was not found", i.Name, gmName))
	}
	if len(gm.Arguments) != 0 {
		// we panic as get-method has the wrong description and dev must fix this bug
		panic(fmt.Errorf("%s `%s` get-method has arguments", i.Name, gmName))
	}

	stack, err := s.getMethodCall(ctx, gm, acc, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "%s `%s`", i.Name, gmName)
	}

	return stack, nil
}

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

func (s *Service) getAccountDataNFT(ctx context.Context, acc *core.AccountState, interfaces []*core.ContractInterface, ret *core.AccountData) {
	for _, i := range interfaces {
		switch i.Name {
		case known.NFTCollection:
			stack, err := s.getMethodCallNoArgs(ctx, i, "get_collection_data", acc)
			if err != nil {
				ret.Errors = append(ret.Errors, err.Error())
				continue
			}

			ret.NFTCollectionData.NextItemIndex = bunbig.FromMathBig(stack[0].Payload.(*big.Int)) //nolint:forcetypeassert // panic on wrong interface
			mapContentDataNFT(ret, stack[1].Payload.(nft.ContentAny))                             //nolint:forcetypeassert // panic on wrong interface
			ret.OwnerAddress = addr.MustFromTonutils(stack[2].Payload.(*address.Address))         //nolint:forcetypeassert // panic on wrong interface

		case known.NFTRoyalty:
			stack, err := s.getMethodCallNoArgs(ctx, i, "royalty_params", acc)
			if err != nil {
				ret.Errors = append(ret.Errors, err.Error())
				continue
			}

			ret.NFTRoyaltyData.RoyaltyBase = stack[0].Payload.(uint16)                                     //nolint:forcetypeassert // panic on wrong interface
			ret.NFTRoyaltyData.RoyaltyFactor = stack[1].Payload.(uint16)                                   //nolint:forcetypeassert // panic on wrong interface
			ret.NFTRoyaltyData.RoyaltyAddress = addr.MustFromTonutils(stack[2].Payload.(*address.Address)) //nolint:forcetypeassert // panic on wrong interface

		case known.NFTItem:
			stack, err := s.getMethodCallNoArgs(ctx, i, "get_nft_data", acc)
			if err != nil {
				ret.Errors = append(ret.Errors, err.Error())
				continue
			}

			ret.NFTItemData.Initialized = stack[0].Payload.(bool)                          //nolint:forcetypeassert // panic on wrong interface
			ret.NFTItemData.ItemIndex = bunbig.FromMathBig(stack[1].Payload.(*big.Int))    //nolint:forcetypeassert // panic on wrong interface
			ret.MinterAddress = addr.MustFromTonutils(stack[2].Payload.(*address.Address)) //nolint:forcetypeassert // panic on wrong interface
			ret.OwnerAddress = addr.MustFromTonutils(stack[3].Payload.(*address.Address))  //nolint:forcetypeassert // panic on wrong interface

			// TODO: get nft collection account state and full nft content
			// individualContent := stack[4].Payload.(*cell.Cell)

		case known.NFTEditable:
			stack, err := s.getMethodCallNoArgs(ctx, i, "get_editor", acc)
			if err != nil {
				ret.Errors = append(ret.Errors, err.Error())
				continue
			}

			ret.NFTEditable.EditorAddress = addr.MustFromTonutils(stack[0].Payload.(*address.Address)) //nolint:forcetypeassert // panic on wrong interface
		}
	}
}

func (s *Service) getAccountDataFT(ctx context.Context, acc *core.AccountState, interfaces []*core.ContractInterface, ret *core.AccountData) {
	for _, i := range interfaces {
		switch i.Name {
		case known.JettonMinter:
			stack, err := s.getMethodCallNoArgs(ctx, i, "get_jetton_data", acc)
			if err != nil {
				ret.Errors = append(ret.Errors, err.Error())
				continue
			}

			ret.FTMasterData.TotalSupply = bunbig.FromMathBig(stack[0].Payload.(*big.Int))             //nolint:forcetypeassert // panic on wrong interface
			ret.FTMasterData.Mintable = stack[1].Payload.(bool)                                        //nolint:forcetypeassert // panic on wrong interface
			ret.FTMasterData.AdminAddress = addr.MustFromTonutils(stack[2].Payload.(*address.Address)) //nolint:forcetypeassert // panic on wrong interface
			mapContentDataNFT(ret, stack[3].Payload.(nft.ContentAny))                                  //nolint:forcetypeassert // panic on wrong interface

		case known.JettonWallet:
			stack, err := s.getMethodCallNoArgs(ctx, i, "get_wallet_data", acc)
			if err != nil {
				ret.Errors = append(ret.Errors, err.Error())
				continue
			}

			ret.FTWalletData.JettonBalance = bunbig.FromMathBig(stack[0].Payload.(*big.Int)) //nolint:forcetypeassert // panic on wrong interface
			ret.OwnerAddress = addr.MustFromTonutils(stack[1].Payload.(*address.Address))    //nolint:forcetypeassert // panic on wrong interface
			ret.MinterAddress = addr.MustFromTonutils(stack[2].Payload.(*address.Address))   //nolint:forcetypeassert // panic on wrong interface
		}
	}
}

func (s *Service) getAccountDataWallet(ctx context.Context, acc *core.AccountState, interfaces []*core.ContractInterface, ret *core.AccountData) {
	for _, i := range interfaces {
		if !strings.HasPrefix(string(i.Name), "wallet") || strings.HasPrefix(string(i.Name), "wallet_highload") {
			continue
		}

		stack, err := s.getMethodCallNoArgs(ctx, i, "seqno", acc)
		if err != nil {
			ret.Errors = append(ret.Errors, err.Error())
			continue
		}
		if len(stack) != 1 || stack[0].Format != "uint64" {
			// we panic as standard contract interface has the wrong description
			panic(fmt.Errorf("wrong wallet `seqno` get-method description"))
		}

		ret.WalletSeqNo = stack[0].Payload.(uint64) //nolint:forcetypeassert // panic on wrong interface
	}
}

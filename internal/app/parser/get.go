package parser

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

func (s *Service) getMethodCall(ctx context.Context, d *abi.GetMethodDesc, acc *core.AccountState, args []any) (ret abi.GetMethodExecution, err error) {
	var argsStack abi.VmStack

	if len(acc.Code) == 0 || len(acc.Data) == 0 {
		return ret, errors.Wrap(app.ErrImpossibleParsing, "no account code or data")
	}

	if len(d.Arguments) != len(args) {
		return ret, errors.New("length of passed and described arguments does not match")
	}
	for it := range args {
		argsStack = append(argsStack, abi.VmValue{
			VmValueDesc: d.Arguments[it],
			Payload:     args[it],
		})
	}

	code, err := cell.FromBOC(acc.Code)
	if err != nil {
		return ret, errors.Wrap(err, "account code from boc")
	}
	data, err := cell.FromBOC(acc.Data)
	if err != nil {
		return ret, errors.Wrap(err, "account data from boc")
	}

	e, err := abi.NewEmulator(acc.Address.MustToTonutils(), code, data, s.bcConfig)
	if err != nil {
		return ret, errors.Wrap(err, "new emulator")
	}

	retStack, err := e.RunGetMethod(ctx, d.Name, argsStack, d.ReturnValues)

	ret = abi.GetMethodExecution{
		Name:      d.Name,
		Arguments: argsStack,
		Returns:   retStack,
	}
	if err != nil {
		ret.Error = errors.Wrap(err, "run get method").Error()
		log.Warn().Err(err).Str("get_method", d.Name).Str("address", acc.Address.Base64()).Msg("run get method")
	}
	return ret, nil
}

// TODO: map automatically by field names with reflect
// TODO: check return values descriptors, do not panic on wrong type assertions

func (s *Service) getMethodCallNoArgs(ctx context.Context, i *core.ContractInterface, gmName string, acc *core.AccountState) (ret abi.GetMethodExecution, err error) {
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
		return ret, errors.Wrapf(err, "%s `%s`", i.Name, gmName)
	}

	return stack, nil
}

func mapContentDataNFT(ret *core.AccountState, c any) {
	if c == nil {
		return
	}
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

func (s *Service) getNFTItemContent(ctx context.Context, collection *core.AccountState, idx *big.Int, itemContent *cell.Cell, acc *core.AccountState) {
	exec, err := s.getMethodCall(ctx,
		&abi.GetMethodDesc{
			Name: "get_nft_content",
			Arguments: []abi.VmValueDesc{{
				Name:      "index",
				StackType: "int",
			}, {
				Name:      "individual_content",
				StackType: "cell",
			}},
			ReturnValues: []abi.VmValueDesc{{
				Name:      "full_content",
				StackType: "cell",
				Format:    "content",
			}},
		},
		collection, []any{idx, itemContent},
	)
	if err != nil {
		log.Error().Err(err).Msg("execute get_nft_content nft_collection get-method")
		return
	}
	acc.ExecutedGetMethods[known.NFTCollection] = append(acc.ExecutedGetMethods[known.NFTCollection], exec)

	mapContentDataNFT(acc, exec.Returns[0].Payload)
}

func (s *Service) getAccountDataNFT(
	ctx context.Context,
	acc *core.AccountState,
	others func(context.Context, *addr.Address) (*core.AccountState, error),
	interfaces []*core.ContractInterface,
) {
	for _, i := range interfaces {
		switch i.Name {
		case known.NFTCollection:
			exec, err := s.getMethodCallNoArgs(ctx, i, "get_collection_data", acc)
			if err != nil {
				log.Error().Err(err).Msg("execute get_collection_data nft_collection get-method")
				return
			}
			acc.ExecutedGetMethods[i.Name] = append(acc.ExecutedGetMethods[i.Name], exec)

			acc.OwnerAddress = addr.MustFromTonutils(exec.Returns[2].Payload.(*address.Address)) //nolint:forcetypeassert // panic on wrong interface

			mapContentDataNFT(acc, exec.Returns[1].Payload)

		case known.NFTRoyalty:
			exec, err := s.getMethodCallNoArgs(ctx, i, "royalty_params", acc)
			if err != nil {
				log.Error().Err(err).Msg("execute royalty_params nft_royalty get-method")
				return
			}
			acc.ExecutedGetMethods[i.Name] = append(acc.ExecutedGetMethods[i.Name], exec)

		case known.NFTItem:
			exec, err := s.getMethodCallNoArgs(ctx, i, "get_nft_data", acc)
			if err != nil {
				log.Error().Err(err).Msg("execute get_nft_data nft_item get-method")
				return
			}
			acc.ExecutedGetMethods[i.Name] = append(acc.ExecutedGetMethods[i.Name], exec)

			acc.MinterAddress = addr.MustFromTonutils(exec.Returns[2].Payload.(*address.Address)) //nolint:forcetypeassert // panic on wrong interface
			acc.OwnerAddress = addr.MustFromTonutils(exec.Returns[3].Payload.(*address.Address))  //nolint:forcetypeassert // panic on wrong interface

			if acc.MinterAddress == nil {
				continue
			}
			collection, err := others(ctx, acc.MinterAddress)
			if err != nil {
				log.Error().Err(err).Msg("get nft collection state")
				return
			}
			s.getNFTItemContent(ctx, collection, exec.Returns[1].Payload.(*big.Int), exec.Returns[4].Payload.(*cell.Cell), acc) //nolint:forcetypeassert // panic on wrong interface

		case known.NFTEditable:
			exec, err := s.getMethodCallNoArgs(ctx, i, "get_editor", acc)
			if err != nil {
				log.Error().Err(err).Msg("execute get_editor nft_editable get-method")
				return
			}
			acc.ExecutedGetMethods[i.Name] = append(acc.ExecutedGetMethods[i.Name], exec)
		}
	}
}

func (s *Service) getAccountDataFT(
	ctx context.Context,
	acc *core.AccountState,
	_ func(context.Context, *addr.Address) (*core.AccountState, error),
	interfaces []*core.ContractInterface,
) {
	for _, i := range interfaces {
		switch i.Name {
		case known.JettonMinter:
			exec, err := s.getMethodCallNoArgs(ctx, i, "get_jetton_data", acc)
			if err != nil {
				log.Error().Err(err).Msg("call get_jetton_data jetton_minter get-method")
				continue
			}
			acc.ExecutedGetMethods[i.Name] = append(acc.ExecutedGetMethods[i.Name], exec)

			mapContentDataNFT(acc, exec.Returns[3].Payload)

		case known.JettonWallet:
			exec, err := s.getMethodCallNoArgs(ctx, i, "get_wallet_data", acc)
			if err != nil {
				log.Error().Err(err).Msg("call get_wallet_data jetton_wallet get-method")
				continue
			}
			acc.ExecutedGetMethods[i.Name] = append(acc.ExecutedGetMethods[i.Name], exec)

			acc.JettonBalance = bunbig.FromMathBig(exec.Returns[0].Payload.(*big.Int))            //nolint:forcetypeassert // panic on wrong interface
			acc.OwnerAddress = addr.MustFromTonutils(exec.Returns[1].Payload.(*address.Address))  //nolint:forcetypeassert // panic on wrong interface
			acc.MinterAddress = addr.MustFromTonutils(exec.Returns[2].Payload.(*address.Address)) //nolint:forcetypeassert // panic on wrong interface
		}
	}
}

func (s *Service) getAccountDataWallet(
	ctx context.Context,
	acc *core.AccountState,
	_ func(context.Context, *addr.Address) (*core.AccountState, error),
	interfaces []*core.ContractInterface,
) {
	for _, i := range interfaces {
		if !strings.HasPrefix(string(i.Name), "wallet") {
			continue
		}
		if len(i.GetMethodsDesc) == 0 {
			continue
		}

		exec, err := s.getMethodCallNoArgs(ctx, i, "seqno", acc)
		if err != nil {
			log.Error().Err(err).Msg("call seqno wallet get-method")
			continue
		}
		acc.ExecutedGetMethods[i.Name] = append(acc.ExecutedGetMethods[i.Name], exec)
	}
}

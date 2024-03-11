package parser

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"

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

func (s *Service) emulateGetMethod(ctx context.Context, d *abi.GetMethodDesc, acc *core.AccountState, args []any) (ret abi.GetMethodExecution, err error) {
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

	codeBase64, dataBase64, librariesBase64 :=
		base64.StdEncoding.EncodeToString(acc.Code),
		base64.StdEncoding.EncodeToString(acc.Data),
		base64.StdEncoding.EncodeToString(acc.Libraries)

	e, err := abi.NewEmulatorBase64(acc.Address.MustToTonutils(), codeBase64, dataBase64, s.bcConfigBase64, librariesBase64)
	if err != nil {
		return ret, errors.Wrap(err, "new emulator")
	}

	retStack, err := e.RunGetMethod(ctx, d.Name, argsStack, d.ReturnValues)

	ret = abi.GetMethodExecution{
		Name: d.Name,
	}
	for i := range argsStack {
		ret.Receives = append(ret.Receives, argsStack[i].Payload)
	}
	for i := range retStack {
		ret.Returns = append(ret.Returns, retStack[i].Payload)
	}
	if err != nil {
		ret.Error = err.Error()

		lvl := log.Warn()
		if d.Name == "get_telemint_auction_state" && ret.Error == "tvm execution failed with code 219" {
			lvl = log.Debug() // err::no_auction
		}
		lvl.Err(err).
			Str("get_method", d.Name).
			Str("address", acc.Address.Base64()).
			Int32("workchain", acc.Workchain).
			Int64("shard", acc.Shard).
			Uint32("block_seq_no", acc.BlockSeqNo).
			Msg("run get method")
	}
	return ret, nil
}

func (s *Service) emulateGetMethodNoArgs(ctx context.Context, i *core.ContractInterface, gmName string, acc *core.AccountState) (ret abi.GetMethodExecution, err error) {
	gm := getMethodByName(i, gmName)
	if gm == nil {
		// we panic as contract interface was defined, but there are no standard get-method
		panic(fmt.Errorf("%s `%s` get-method was not found", i.Name, gmName))
	}
	if len(gm.Arguments) != 0 {
		// we panic as get-method has the wrong description and dev must fix this bug
		panic(fmt.Errorf("%s `%s` get-method has arguments", i.Name, gmName))
	}

	stack, err := s.emulateGetMethod(ctx, gm, acc, nil)
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
	desc := &abi.GetMethodDesc{
		Name: "get_nft_content",
		Arguments: []abi.VmValueDesc{{
			Name:      "index",
			StackType: "int",
			Format:    "bytes",
		}, {
			Name:      "individual_content",
			StackType: "cell",
		}},
		ReturnValues: []abi.VmValueDesc{{
			Name:      "full_content",
			StackType: "cell",
			Format:    "content",
		}},
	}

	args := []any{idx.Bytes(), itemContent}

	exec, err := s.emulateGetMethod(ctx, desc, collection, args)
	if err != nil {
		log.Error().Err(err).Msg("execute get_nft_content nft_collection get-method")
		return
	}

	exec.Address = &collection.Address

	acc.ExecutedGetMethods[known.NFTCollection] = append(acc.ExecutedGetMethods[known.NFTCollection], exec)
	if exec.Error != "" {
		return
	}

	mapContentDataNFT(acc, exec.Returns[0])
}

func (s *Service) checkMinter(ctx context.Context, minter, item *core.AccountState, i abi.ContractName, desc *abi.GetMethodDesc, args []any) {
	item.Fake = true

	exec, err := s.emulateGetMethod(ctx, desc, minter, args)
	if err != nil {
		log.Error().Err(err).Msgf("execute %s %s get-method", desc.Name, i)
		return
	}

	exec.Address = &minter.Address

	item.ExecutedGetMethods[i] = append(item.ExecutedGetMethods[i], exec)
	if exec.Error != "" {
		log.Error().Err(err).Msgf("execute %s %s get-method", desc.Name, i)
		return
	}

	itemAddr := addr.MustFromTonutils(exec.Returns[0].(*address.Address)) //nolint:forcetypeassert // panic on wrong interface
	if addr.Equal(itemAddr, &item.Address) {
		item.Fake = false
	}
}

func (s *Service) checkNFTMinter(ctx context.Context, minter *core.AccountState, idx *big.Int, item *core.AccountState) {
	desc := &abi.GetMethodDesc{
		Name: "get_nft_address_by_index",
		Arguments: []abi.VmValueDesc{{
			Name:      "index",
			StackType: "int",
			Format:    "bytes",
		}},
		ReturnValues: []abi.VmValueDesc{{
			Name:      "address",
			StackType: "slice",
			Format:    "addr",
		}},
	}

	args := []any{idx.Bytes()}

	s.checkMinter(ctx, minter, item, known.NFTCollection, desc, args)
}

func (s *Service) checkJettonMinter(ctx context.Context, minter *core.AccountState, ownerAddr *addr.Address, walletAcc *core.AccountState) {
	desc := &abi.GetMethodDesc{
		Name: "get_wallet_address",
		Arguments: []abi.VmValueDesc{{
			Name:      "owner_address",
			StackType: "slice",
			Format:    "addr",
		}},
		ReturnValues: []abi.VmValueDesc{{
			Name:      "wallet_address",
			StackType: "slice",
			Format:    "addr",
		}},
	}

	args := []any{ownerAddr.MustToTonutils()}

	s.checkMinter(ctx, minter, walletAcc, known.JettonMinter, desc, args)
}

func (s *Service) callGetMethod(
	ctx context.Context,
	acc *core.AccountState,
	i *core.ContractInterface,
	getMethodDesc *abi.GetMethodDesc,
	others func(context.Context, addr.Address) (*core.AccountState, error),
) error {
	exec, err := s.emulateGetMethodNoArgs(ctx, i, getMethodDesc.Name, acc)
	if err != nil {
		return errors.Wrapf(err, "execute get-method")
	}

	acc.ExecutedGetMethods[i.Name] = append(acc.ExecutedGetMethods[i.Name], exec)
	if exec.Error != "" {
		return nil
	}

	switch getMethodDesc.Name {
	case "get_collection_data":
		acc.OwnerAddress = addr.MustFromTonutils(exec.Returns[2].(*address.Address)) //nolint:forcetypeassert // panic on wrong interface
		mapContentDataNFT(acc, exec.Returns[1])

	case "get_nft_data":
		acc.MinterAddress = addr.MustFromTonutils(exec.Returns[2].(*address.Address)) //nolint:forcetypeassert // panic on wrong interface
		acc.OwnerAddress = addr.MustFromTonutils(exec.Returns[3].(*address.Address))  //nolint:forcetypeassert // panic on wrong interface

		if acc.MinterAddress == nil {
			return nil
		}

		collection, err := others(ctx, *acc.MinterAddress)
		if err != nil {
			log.Error().Str("minter_address", acc.MinterAddress.Base64()).Err(err).Msg("get nft collection state")
			return nil
		}

		index, individualContent := new(big.Int).SetBytes(exec.Returns[1].([]byte)), exec.Returns[4].(*cell.Cell) //nolint:forcetypeassert // panic on wrong interface

		s.getNFTItemContent(ctx, collection, index, individualContent, acc)
		s.checkNFTMinter(ctx, collection, index, acc)

	case "get_jetton_data":
		mapContentDataNFT(acc, exec.Returns[3])

	case "get_wallet_data":
		acc.JettonBalance = bunbig.FromMathBig(exec.Returns[0].(*big.Int))            //nolint:forcetypeassert // panic on wrong interface
		acc.OwnerAddress = addr.MustFromTonutils(exec.Returns[1].(*address.Address))  //nolint:forcetypeassert // panic on wrong interface
		acc.MinterAddress = addr.MustFromTonutils(exec.Returns[2].(*address.Address)) //nolint:forcetypeassert // panic on wrong interface

		if acc.MinterAddress == nil || acc.OwnerAddress == nil {
			return nil
		}

		minter, err := others(ctx, *acc.MinterAddress)
		if err != nil {
			log.Error().Str("minter_address", acc.MinterAddress.Base64()).Err(err).Msg("get jetton minter state")
			return nil
		}

		s.checkJettonMinter(ctx, minter, acc.OwnerAddress, acc)
	}

	return nil
}

func (s *Service) callPossibleGetMethods(
	ctx context.Context,
	acc *core.AccountState,
	others func(context.Context, addr.Address) (*core.AccountState, error),
	interfaces []*core.ContractInterface,
) {
	for _, i := range interfaces {
		for it := range i.GetMethodsDesc {
			d := &i.GetMethodsDesc[it]

			if len(d.Arguments) != 0 {
				continue
			}

			if err := s.callGetMethod(ctx, acc, i, d, others); err != nil {
				log.Error().Err(err).Str("contract_name", string(i.Name)).Str("get_method", d.Name).Msg("execute get-method")
			}
		}
	}
}

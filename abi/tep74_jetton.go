package abi

import (
	"context"
	"fmt"
	"math/big"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// https://github.com/ton-blockchain/TEPs/blob/master/text/0074-jettons-standard.md
// https://github.com/ton-blockchain/token-contract/tree/main/ft)

type (
	JettonData       jetton.Data
	JettonWalletData struct {
		Balance       *big.Int
		OwnerAddress  *address.Address
		MasterAddress *address.Address
		WalletCode    *cell.Cell
	}

	JettonMint             jetton.MintPayload
	JettonTransfer         jetton.TransferPayload
	JettonInternalTransfer struct {
		_                tlb.Magic        `tlb:"#178d4519"`
		QueryID          uint64           `tlb:"## 64"`
		Amount           tlb.Coins        `tlb:"."`
		From             *address.Address `tlb:"addr"`
		ResponseAddress  *address.Address `tlb:"addr"`
		ForwardTONAmount tlb.Coins        `tlb:"."`
		ForwardPayload   *cell.Cell       `tlb:"either . ^"`
	}
	JettonTransferNotification struct {
		_              tlb.Magic        `tlb:"#7362d09c"`
		QueryID        uint64           `tlb:"## 64"`
		Amount         tlb.Coins        `tlb:"."`
		Sender         *address.Address `tlb:"addr"`
		ForwardPayload *cell.Cell       `tlb:"either . ^"`
	}
	JettonBurn jetton.BurnPayload
)

func GetJettonData(ctx context.Context, api *ton.APIClient, b *ton.BlockIDExt, addr *address.Address) (*JettonData, error) {
	c := jetton.NewJettonMasterClient(api, addr)

	data, err := c.GetJettonDataAtBlock(ctx, b)
	if err != nil {
		return nil, err
	}

	return (*JettonData)(data), nil
}

func GetJettonWalletData(ctx context.Context, api *ton.APIClient, b *ton.BlockIDExt, addr *address.Address) (*JettonWalletData, error) {
	res, err := api.RunGetMethod(ctx, b, addr, "get_wallet_data")
	if err != nil {
		return nil, fmt.Errorf("failed to run get_wallet_data method: %w", err)
	}

	balance, err := res.Int(0)
	if err != nil {
		return nil, fmt.Errorf("balance get err: %w", err)
	}

	ownerAddrS, err := res.Slice(1)
	if err != nil {
		return nil, fmt.Errorf("owner addr get err: %w", err)
	}
	ownerAddr, err := ownerAddrS.LoadAddr()
	if err != nil {
		return nil, fmt.Errorf("failed to load address from ownerAddr slice: %w", err)
	}

	masterAddrS, err := res.Slice(2)
	if err != nil {
		return nil, fmt.Errorf("master addr get err: %w", err)
	}
	masterAddr, err := masterAddrS.LoadAddr()
	if err != nil {
		return nil, fmt.Errorf("failed to load address from masterAddr slice: %w", err)
	}

	code, err := res.Cell(3)
	if err != nil {
		return nil, fmt.Errorf("wallet code cell get err: %w", err)
	}

	return &JettonWalletData{
		Balance:       balance,
		OwnerAddress:  ownerAddr,
		MasterAddress: masterAddr,
		WalletCode:    code,
	}, nil
}

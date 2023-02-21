package abi

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// https://github.com/ton-blockchain/TEPs/blob/master/text/0074-jettons-standard.md
// https://github.com/ton-blockchain/token-contract/tree/main/ft)

type (
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

package rndm

import (
	"math/rand"
	"time"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/abi/known"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

var (
	contractNames = []abi.ContractName{known.NFTCollection, known.NFTItem, known.JettonMinter, known.JettonWallet, "wallet_v3r1", "wallet_v4r2"}
	lastTxLT      uint64
	timestamp     = time.Now().UTC()
)

func GetMethodHashes() (ret []int32) {
	for i := 0; i < 1+rand.Int()%16; i++ {
		ret = append(ret, int32(rand.Uint32()))
	}
	return
}

func ContractNames(a *addr.Address) (ret []abi.ContractName) {
	seed := int(a[30])<<8 + int(a[31])
	for i := 0; i < 1+seed%16; i++ {
		ret = append(ret, contractNames[(seed+i)%len(contractNames)])
	}
	return
}

func AddressState(a *addr.Address, t []abi.ContractName, minter *addr.Address) *core.AccountState {
	lastTxLT++
	timestamp = timestamp.Add(time.Minute)

	b := Block(0)

	s := &core.AccountState{
		Address:         *a,
		Workchain:       b.Workchain,
		Shard:           b.Shard,
		BlockSeqNo:      b.SeqNo,
		IsActive:        true,
		Status:          core.Active,
		Balance:         BigInt(),
		LastTxLT:        lastTxLT,
		LastTxHash:      Bytes(32),
		StateHash:       Bytes(32),
		Code:            Bytes(32),
		CodeHash:        Bytes(32),
		Data:            Bytes(32),
		DataHash:        Bytes(32),
		GetMethodHashes: GetMethodHashes(),
		Types:           t,
		OwnerAddress:    Address(),
		MinterAddress:   minter,
		ExecutedGetMethods: map[abi.ContractName][]abi.GetMethodExecution{
			"nft_item": {{
				Name: "get_nft_data",
				Returns: abi.VmStack{{
					VmValueDesc: abi.VmValueDesc{
						Name:      "init",
						StackType: "int",
						Format:    "bool",
					},
					Payload: true,
				}},
			}},
		},
		FTWalletData:   core.FTWalletData{JettonBalance: BigInt()},
		NFTContentData: core.NFTContentData{ContentImageData: []byte{}}, // TODO: i dunno why ",nullzero" tag does not work in pg
		UpdatedAt:      timestamp,
	}

	return s
}

func AddressStateContract(a *addr.Address, t abi.ContractName, minter *addr.Address) *core.AccountState {
	if minter == nil {
		minter = new(addr.Address)
		copy((*minter)[:], a[:])
		minter[16] = '\xde'
	}

	var types []abi.ContractName
	if t != "" {
		types = append(types, t)
	} else {
		types = append(types, ContractNames(a)...)
	}

	return AddressState(a, types, minter)
}

func AddressStates(a *addr.Address, n int) (ret []*core.AccountState) {
	for i := 0; i < n; i++ {
		ret = append(ret, AddressState(a, nil, nil))
	}
	return ret
}

func AccountStates(n int) (ret []*core.AccountState) {
	return AddressStates(Address(), n)
}

func AccountStatesContract(n int, t abi.ContractName, minter *addr.Address) (ret []*core.AccountState) {
	a := Address()
	for i := 0; i < n; i++ {
		ret = append(ret, AddressStateContract(a, t, minter))
	}
	return ret
}

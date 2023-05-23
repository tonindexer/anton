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

	s := &core.AccountState{
		Address:         *a,
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
		UpdatedAt:       timestamp,
	}

	return s
}

func AddressStateContract(a *addr.Address, t abi.ContractName, minter *addr.Address) *core.AccountState {
	var types []abi.ContractName

	if minter == nil {
		minter = new(addr.Address)
		copy((*minter)[:], a[:])
		minter[16] = '\xde'
	}

	if t != "" {
		types = append(types, t)
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

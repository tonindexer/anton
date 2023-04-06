package rndm

import (
	"math/rand"
	"time"

	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
)

var (
	contractNames []abi.ContractName
	lastTxLT      uint64
	timestamp     = time.Now().UTC()
)

func initContractNames() {
	for n := range abi.KnownContractMethods {
		contractNames = append(contractNames, n)
	}
	for v := range abi.WalletCode {
		contractNames = append(contractNames, v.Name())
	}
}

func GetMethodHashes() (ret []int32) {
	for i := 0; i < 1+rand.Int()%16; i++ {
		ret = append(ret, int32(rand.Uint32()))
	}
	return
}

func ContractNames(a *addr.Address) (ret []abi.ContractName) {
	seed := int(a[30])<<8 + int(a[31])

	if contractNames == nil {
		initContractNames()
	}
	for i := 0; i < 1+seed%16; i++ {
		ret = append(ret, contractNames[(seed+i)%len(contractNames)])
	}
	return
}

func AddressStates(a *addr.Address, n int) (ret []*core.AccountState) {
	for i := 0; i < n; i++ {
		lastTxLT++
		timestamp = timestamp.Add(time.Minute)

		s := &core.AccountState{
			Address:         *a,
			IsActive:        true,
			Status:          core.Active,
			Balance:         bunbig.FromUInt64(rand.Uint64()),
			LastTxLT:        lastTxLT,
			LastTxHash:      Bytes(32),
			StateHash:       Bytes(32),
			Code:            Bytes(32),
			CodeHash:        Bytes(32),
			Data:            Bytes(32),
			DataHash:        Bytes(32),
			GetMethodHashes: GetMethodHashes(),
			UpdatedAt:       timestamp,
		}

		ret = append(ret, s)
	}

	return ret
}

func AccountStates(n int) (ret []*core.AccountState) {
	return AddressStates(Address(), n)
}

func ContractsData(states []*core.AccountState, t abi.ContractName, minter *addr.Address) (ret []*core.AccountData) {
	for _, s := range states {
		if minter == nil {
			minter = new(addr.Address)
			copy((*minter)[:], s.Address[:])
			minter[16] = '\xde'
		}
		data := &core.AccountData{
			Address:           s.Address,
			LastTxLT:          s.LastTxLT,
			LastTxHash:        s.LastTxHash,
			Balance:           s.Balance,
			Types:             ContractNames(&s.Address),
			OwnerAddress:      Address(),
			MinterAddress:     minter,
			NFTCollectionData: core.NFTCollectionData{NextItemIndex: bunbig.FromUInt64(rand.Uint64())},
			NFTRoyaltyData:    core.NFTRoyaltyData{RoyaltyAddress: Address()},
			NFTContentData:    core.NFTContentData{ContentURI: String(16), ContentImageData: Bytes(128)},
			NFTItemData:       core.NFTItemData{ItemIndex: bunbig.FromUInt64(rand.Uint64())},
			FTMasterData:      core.FTMasterData{TotalSupply: bunbig.FromUInt64(rand.Uint64())},
			FTWalletData:      core.FTWalletData{JettonBalance: bunbig.FromUInt64(uint64(rand.Uint32()))},
			Errors:            []string{String(16)},
			UpdatedAt:         s.UpdatedAt,
		}
		if t != "" {
			data.Types = append(data.Types, t)
		}
		ret = append(ret, data)
	}
	return ret
}

func AccountData(states []*core.AccountState) (ret []*core.AccountData) {
	return ContractsData(states, "", nil)
}

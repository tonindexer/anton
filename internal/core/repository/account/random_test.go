package account

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
)

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randBytes(l int) []byte {
	token := make([]byte, l)
	rand.Read(token) // nolint
	return token
}

func randAddr() *addr.Address {
	a, err := new(addr.Address).FromString(fmt.Sprintf("0:%x", randBytes(32)))
	if err != nil {
		panic(err)
	}
	return a
}

func randGetMethodHashes() (ret []int32) {
	for i := 0; i < 1+rand.Int()%16; i++ {
		ret = append(ret, int32(rand.Uint32()))
	}
	return
}

var (
	contractNames []abi.ContractName
	lastTxLT      uint64
)

func initContractNames() {
	for n := range abi.KnownContractMethods {
		contractNames = append(contractNames, n)
	}
	for v := range abi.WalletCode {
		contractNames = append(contractNames, v.Name())
	}
}

func randInterfaces(a *addr.Address) (ret []abi.ContractName) {
	seed := int(a[30])<<8 + int(a[31])

	if contractNames == nil {
		initContractNames()
	}
	for i := 0; i < 1+seed%16; i++ {
		ret = append(ret, contractNames[(seed+i)%len(contractNames)])
	}
	return
}

func randAddressStates(a *addr.Address, n int) (ret []*core.AccountState) {
	for i := 0; i < n; i++ {
		lastTxLT++

		s := &core.AccountState{
			Address:         *a,
			IsActive:        true,
			Status:          core.Active,
			Balance:         bunbig.FromUInt64(rand.Uint64()),
			LastTxLT:        lastTxLT,
			LastTxHash:      randBytes(32),
			StateHash:       randBytes(32),
			Code:            randBytes(32),
			CodeHash:        randBytes(32),
			Data:            randBytes(32),
			DataHash:        randBytes(32),
			GetMethodHashes: randGetMethodHashes(),
			UpdatedAt:       time.Now().UTC(),
		}

		ret = append(ret, s)
	}

	return ret
}

func randAccountStates(n int) (ret []*core.AccountState) {
	return randAddressStates(randAddr(), n)
}

func randContractData(states []*core.AccountState, t abi.ContractName) (ret []*core.AccountData) {
	for _, s := range states {
		minter := s.Address
		minter[16] = '\xde'

		data := &core.AccountData{
			Address:           s.Address,
			LastTxLT:          s.LastTxLT,
			LastTxHash:        s.LastTxHash,
			Balance:           s.Balance,
			Types:             randInterfaces(&s.Address),
			OwnerAddress:      randAddr(),
			MinterAddress:     &minter,
			NFTCollectionData: core.NFTCollectionData{NextItemIndex: bunbig.FromUInt64(rand.Uint64())},
			NFTRoyaltyData:    core.NFTRoyaltyData{RoyaltyAddress: randAddr()},
			NFTContentData:    core.NFTContentData{ContentURI: randString(16), ContentImageData: randBytes(128)},
			NFTItemData:       core.NFTItemData{ItemIndex: bunbig.FromUInt64(rand.Uint64())},
			FTMasterData:      core.FTMasterData{TotalSupply: bunbig.FromUInt64(rand.Uint64())},
			FTWalletData:      core.FTWalletData{JettonBalance: bunbig.FromUInt64(uint64(rand.Uint32()))},
			Errors:            []string{randString(16)},
			UpdatedAt:         s.UpdatedAt,
		}
		if t != "" {
			data.Types = append(data.Types, t)
		}
		ret = append(ret, data)
	}
	return ret
}

func randAccountData(states []*core.AccountState) (ret []*core.AccountData) {
	return randContractData(states, "")
}

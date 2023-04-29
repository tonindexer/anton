package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

func TestService_ParseAccount(t *testing.T) {
	s := testService(t)
	master := getCurrentMaster(t)

	type testCase struct {
		addr           *address.Address
		contract       abi.ContractName
		status         tlb.AccountStatus
		contentURI     string
		collectionAddr string
	}

	var cases = []*testCase{
		{
			addr:     address.MustParseAddr("EQA-IU8sn_aSCxCpufZtjTm1uxOyCe3LAYEJlH09e8nElCnp"),
			contract: "wallet_v3r1",
			status:   tlb.AccountStatusActive,
		},
		{
			addr:     address.MustParseAddr("EQC6KV4zs8TJtSZapOrRFmqSkxzpq-oSCoxekQRKElf4nC1I"),
			contract: abi.NFTItem,
			status:   tlb.AccountStatusActive,
		},
		{
			addr:       address.MustParseAddr("EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N"),
			contract:   abi.NFTCollection,
			status:     tlb.AccountStatusActive,
			contentURI: "https://nft.fragment.com/numbers.json",
		},
		{
			addr:           address.MustParseAddr("EQBu6eCK84PxTdjEKyY7z8TQGhN3dbzx-935nj-Lx4FCKPaF"),
			contract:       abi.NFTItem,
			status:         tlb.AccountStatusActive,
			contentURI:     "https://nft.fragment.com/number/88809696960.json",
			collectionAddr: "EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N",
		},
		{
			addr:       address.MustParseAddr("EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi"),
			contract:   abi.NFTCollection,
			status:     tlb.AccountStatusActive,
			contentURI: "https://nft.fragment.com/usernames.json",
		},
		{
			addr:           address.MustParseAddr("EQDOZIib-2DZPCKPir1tT5KtOYWzwoDGM404m9NxXeKVEDpC"),
			contract:       abi.NFTItem, // username
			status:         tlb.AccountStatusActive,
			contentURI:     "https://nft.fragment.com/username/datboi420.json",
			collectionAddr: "EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi",
		},
		{
			addr:     address.MustParseAddr("EQB2NJFK0H5OxJTgyQbej0fy5zuicZAXk2vFZEDrqbQ_n5YW"),
			contract: abi.NFTItem,
			status:   tlb.AccountStatusActive,
		},
	}

	for _, c := range cases {
		acc, err := s.api.GetAccount(ctx, master, c.addr)
		assert.Nil(t, err)

		st := &core.AccountState{
			Address:    *addr.MustFromBase64(c.addr.String()),
			IsActive:   true,
			Status:     core.Active,
			LastTxLT:   acc.LastTxLT,
			LastTxHash: acc.LastTxHash,
			Code:       acc.Code.ToBOC(),
		}
		st.GetMethodHashes, err = abi.GetMethodHashes(acc.Code)
		assert.Nil(t, err)

		types, err := s.DetermineInterfaces(ctx, st)
		assert.Nil(t, err)

		found := false
		for _, t := range types {
			if t == c.contract {
				found = true
				break
			}
		}
		assert.True(t, found)
		assert.Equal(t, c.status, acc.State.Status)

		if c.contract != abi.NFTCollection && c.contract != abi.NFTItem {
			continue
		}

		data, err := s.ParseAccountData(ctx, master, st, types)
		assert.Nil(t, err)
		assert.Equal(t, c.contentURI, data.ContentURI)
		if c.collectionAddr != "" {
			assert.Equal(t, c.collectionAddr, data.MinterAddress.Base64())
		}
	}
}

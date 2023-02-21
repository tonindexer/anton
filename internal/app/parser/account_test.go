package parser

import (
	"testing"

	"github.com/xssnick/tonutils-go/address"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/core"
)

func TestService_ParseAccount(t *testing.T) {
	s := testService(t)
	master := getCurrentMaster(t)

	type testCase struct {
		addr           *address.Address
		contract       abi.ContractName
		status         core.AccountStatus
		contentURI     string
		collectionAddr string
	}

	var cases = []*testCase{
		{
			addr:     address.MustParseAddr("EQA-IU8sn_aSCxCpufZtjTm1uxOyCe3LAYEJlH09e8nElCnp"),
			contract: "wallet_V3R1",
			status:   core.Active,
		},
		{
			addr:     address.MustParseAddr("EQC6KV4zs8TJtSZapOrRFmqSkxzpq-oSCoxekQRKElf4nC1I"),
			contract: abi.NFTItem,
			status:   core.Active,
		},
		{
			addr:       address.MustParseAddr("EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N"),
			contract:   abi.NFTCollection,
			status:     core.Active,
			contentURI: "https://nft.fragment.com/numbers.json",
		},
		{
			addr:           address.MustParseAddr("EQBu6eCK84PxTdjEKyY7z8TQGhN3dbzx-935nj-Lx4FCKPaF"),
			contract:       abi.NFTItem,
			status:         core.Active,
			contentURI:     "https://nft.fragment.com/number/88809696960.json",
			collectionAddr: "EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N",
		},
		{
			addr:       address.MustParseAddr("EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi"),
			contract:   abi.NFTCollection,
			status:     core.Active,
			contentURI: "https://nft.fragment.com/usernames.json",
		},
		{
			addr:           address.MustParseAddr("EQDOZIib-2DZPCKPir1tT5KtOYWzwoDGM404m9NxXeKVEDpC"),
			contract:       abi.NFTItem, // username
			status:         core.Active,
			contentURI:     "https://nft.fragment.com/username/datboi420.json",
			collectionAddr: "EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi",
		},
	}

	for _, c := range cases {
		acc, err := s.api.GetAccount(ctx, master, c.addr)
		if err != nil {
			t.Fatal(c.addr.String(), err)
		}

		types, err := s.DetermineInterfaces(ctx, acc)
		if err != nil {
			t.Fatal(c.addr.String(), err)
		}

		if len(types) < 1 || types[0] != c.contract {
			t.Fatalf("[%s] expected: %s, got: %v", c.addr, c.contract, types)
		}
		if core.AccountStatus(acc.State.Status) != c.status {
			t.Fatalf("[%s] expected: %s, got: %s", c.addr, c.status, acc.State.Status)
		}

		if c.contract != abi.NFTCollection && c.contract != abi.NFTItem {
			continue
		}

		data, err := s.ParseAccountData(ctx, master, acc)
		if err != nil {
			t.Fatal(c.addr.String(), err)
		}
		if c.contentURI != "" && c.contentURI != data.ContentURI {
			t.Fatalf("[%s] expected: %s, got: %s", c.addr, c.contentURI, data.ContentURI)
		}
		if c.collectionAddr != "" && c.collectionAddr != data.CollectionAddress {
			t.Fatalf("[%s] expected: %s, got: %s", c.addr, c.collectionAddr, data.CollectionAddress)
		}
	}
}

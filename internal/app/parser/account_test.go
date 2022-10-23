package parser

import (
	"context"
	"testing"

	"github.com/xssnick/tonutils-go/address"

	"github.com/iam047801/tonidx/internal/core"
)

func TestService_ParseAccount(t *testing.T) {
	s := testService(t)
	master := getCurrentMaster(t)

	type testCase struct {
		addr     *address.Address
		contract core.ContractType
		status   core.AccountStatus
	}

	var cases = []*testCase{
		{
			addr:     address.MustParseAddr("EQA-IU8sn_aSCxCpufZtjTm1uxOyCe3LAYEJlH09e8nElCnp"),
			contract: "V3R1",
			status:   core.Active,
		},
		{
			addr:     address.MustParseAddr("EQAo92DYMokxghKcq-CkCGSk_MgXY5Fo1SPW20gkvZl75iCN"),
			contract: core.NFTCollection,
			status:   core.Active,
		},
		{
			addr:     address.MustParseAddr("EQC6KV4zs8TJtSZapOrRFmqSkxzpq-oSCoxekQRKElf4nC1I"),
			contract: core.NFTItem,
			status:   core.Active,
		},
		{
			addr:     address.MustParseAddr("EQCVRJ-RqeZWcDqgTzzcxUIrChFYs0SyKGUvye9kGOuEWndQ"),
			contract: core.NFTSale,
			status:   core.Active,
		},
	}

	for _, c := range cases {
		acc, err := s.ParseAccount(context.Background(), master, c.addr)
		if err != nil {
			t.Fatal(err)
		}
		if len(acc.Types) < 1 || core.ContractType(acc.Types[0]) != c.contract {
			t.Fatalf("expected: %s, got: %v", c.contract, acc.Types)
		}
		if acc.Status != c.status {
			t.Fatalf("expected: %s, got: %s", c.status, acc.Status)
		}
		// t.Logf("acc: %+v", acc)
		// t.Logf("data: %+v", data)
	}
}

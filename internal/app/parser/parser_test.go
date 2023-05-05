package parser

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
)

var ctx = context.Background()

var _ core.ContractRepository = (*mockContractRepo)(nil)

type mockContractRepo struct {
	interfaces []*core.ContractInterface
}

func (m *mockContractRepo) AddInterface(_ context.Context, _ *core.ContractInterface) error {
	panic("implement me")
}

func (m *mockContractRepo) AddOperation(_ context.Context, _ *core.ContractOperation) error {
	panic("implement me")
}

func (m *mockContractRepo) GetInterfaces(_ context.Context) ([]*core.ContractInterface, error) {
	return m.interfaces, nil
}

func (m *mockContractRepo) GetOperations(_ context.Context) ([]*core.ContractOperation, error) {
	panic("implement me")
}

func (m *mockContractRepo) GetOperationByID(_ context.Context, _ []abi.ContractName, _ bool, _ uint32) (*core.ContractOperation, error) {
	panic("implement me")
}

func newService(t *testing.T) *Service {
	s, err := NewService(ctx, &app.ParserConfig{
		DB: &repository.DB{},
		Servers: []*app.ServerAddr{
			{
				IPPort:    "135.181.177.59:53312",
				PubKeyB64: "aF91CuUHuuOv9rm2W5+O/4h38M3sRm40DtSdRxQhmtQ=",
			},
		},
	})
	require.Nil(t, err)

	walletV3R2Code, err := base64.StdEncoding.DecodeString("te6cckEBAQEAcQAA3v8AIN0gggFMl7ohggEznLqxn3Gw7UTQ0x/THzHXC//jBOCk8mCDCNcYINMf0x/TH/gjE7vyY+1E0NMf0x/T/9FRMrryoVFEuvKiBPkBVBBV+RDyo/gAkyDXSpbTB9QC+wDo0QGkyMsfyx/L/8ntVBC9ba0=")
	require.Nil(t, err)

	walletV3R2 := core.ContractInterface{
		Name: "wallet_v3r2",
		Code: walletV3R2Code,
		GetMethodsDesc: []abi.GetMethodDesc{{
			Name: "seqno",
			ReturnValues: []abi.VmValueDesc{{
				Name:      "seqno",
				StackType: "int",
				Format:    "uint64",
			}},
		}},
	}

	walletV4R2Code, err := base64.StdEncoding.DecodeString("te6cckECFAEAAtQAART/APSkE/S88sgLAQIBIAcCBPjygwjXGCDTH9Mf0x8C+CO78mTtRNDTH9Mf0//0BNFRQ7ryoVFRuvKiBfkBVBBk+RDyo/gAJKTIyx9SQMsfUjDL/1IQ9ADJ7VT4DwHTByHAAJ9sUZMg10qW0wfUAvsA6DDgIcAB4wAhwALjAAHAA5Ew4w0DpMjLHxLLH8v/BgUEAwAK9ADJ7VQAbIEBCNcY+gDTPzBSJIEBCPRZ8qeCEGRzdHJwdIAYyMsFywJQBc8WUAP6AhPLassfEss/yXP7AABwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwICAUgRCAIBIAoJAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCASAMCwARuMl+1E0NcLH4AgFYEA0CASAPDgAZrx32omhAEGuQ64WPwAAZrc52omhAIGuQ64X/wAA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYALm0AHQ0wMhcbCSXwTgItdJwSCSXwTgAtMfIYIQcGx1Z70ighBkc3RyvbCSXwXgA/pAMCD6RAHIygfL/8nQ7UTQgQFA1yH0BDBcgQEI9ApvoTGzkl8H4AXTP8glghBwbHVnupI4MOMNA4IQZHN0crqSXwbjDRMSAIpQBIEBCPRZMO1E0IEBQNcgyAHPFvQAye1UAXKwjiOCEGRzdHKDHrFwgBhQBcsFUAPPFiP6AhPLassfyz/JgED7AJJfA+IAeAH6APQEMPgnbyIwUAqhIb7y4FCCEHBsdWeDHrFwgBhQBMsFJs8WWPoCGfQAy2kXyx9SYMs/IMmAQPsABqZCg7I=")
	require.Nil(t, err)

	walletV4R2 := core.ContractInterface{
		Name: "wallet_v4r2",
		Code: walletV4R2Code,
		GetMethodsDesc: []abi.GetMethodDesc{{
			Name: "seqno",
			ReturnValues: []abi.VmValueDesc{{
				Name:      "seqno",
				StackType: "int",
				Format:    "uint64",
			}},
		}},
	}

	nftItem := core.ContractInterface{
		Name: "nft_item",
		GetMethodsDesc: []abi.GetMethodDesc{{
			Name:      "get_nft_data",
			Arguments: []abi.VmValueDesc{},
			ReturnValues: []abi.VmValueDesc{{
				Name:      "init",
				StackType: "int",
				Format:    "bool",
			}, {
				Name:      "index",
				StackType: "int",
			}, {
				Name:      "collection_address",
				StackType: "slice",
				Format:    "addr",
			}, {
				Name:      "owner_address",
				StackType: "slice",
				Format:    "addr",
			}, {
				Name:      "individual_content",
				StackType: "cell",
			}},
		}},
	}
	for it := range nftItem.GetMethodsDesc {
		nftItem.GetMethodHashes = append(nftItem.GetMethodHashes, abi.MethodNameHash(nftItem.GetMethodsDesc[it].Name))
	}

	s.contractRepo = &mockContractRepo{
		interfaces: []*core.ContractInterface{&walletV3R2, &walletV4R2, &nftItem},
	}

	return s
}

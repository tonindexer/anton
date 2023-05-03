package parser

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

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
		CHModel:   ch.CHModel{},
		BaseModel: bun.BaseModel{},
		Name:      "wallet_v3r2",
		Code:      walletV3R2Code,
		GetMethodsDesc: []abi.GetMethodDesc{
			{
				Name: "seqno",
				ReturnValues: []abi.VmValueDesc{
					{
						Name:      "seqno",
						StackType: "int",
						Format:    "uint64",
					},
				},
			},
		},
	}

	s.contractRepo = &mockContractRepo{
		interfaces: []*core.ContractInterface{&walletV3R2},
	}

	return s
}

package parser

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

func TestService_ParseAccountData_WalletV3R2(t *testing.T) {
	s := newService(t)

	code, err := base64.StdEncoding.DecodeString("te6cckEBAQEAcQAA3v8AIN0gggFMl7ohggEznLqxn3Gw7UTQ0x/THzHXC//jBOCk8mCDCNcYINMf0x/TH/gjE7vyY+1E0NMf0x/T/9FRMrryoVFEuvKiBPkBVBBV+RDyo/gAkyDXSpbTB9QC+wDo0QGkyMsfyx/L/8ntVBC9ba0=")
	require.Nil(t, err)
	data, err := base64.StdEncoding.DecodeString("te6cckEBAQEAKgAAUAAAAAEGQZj7UhMYn0DGJKa8VAJx2X9dF+VkfoJrgOKgW7MinX6Pqkvc3Pev")
	require.Nil(t, err)

	ret, err := s.ParseAccountData(ctx, &core.AccountState{
		Address:  *addr.MustFromBase64("EQDj5AA8mQvM5wJEQsFFFof79y3ZsuX6wowktWQFhz_Anton"),
		IsActive: true, Status: core.Active,
		Balance: bunbig.FromInt64(1e9),
		Code:    code,
		Data:    data,
	})
	require.Nil(t, err)
	require.Equal(t, []abi.ContractName{"wallet_v3r2"}, ret.Types)
	require.Equal(t, uint64(1), ret.WalletSeqNo)
}

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

func TestService_ParseAccountData_WalletV4R2(t *testing.T) {
	s := newService(t)

	code, err := base64.StdEncoding.DecodeString("te6cckECFAEAAtQAART/APSkE/S88sgLAQIBIAcCBPjygwjXGCDTH9Mf0x8C+CO78mTtRNDTH9Mf0//0BNFRQ7ryoVFRuvKiBfkBVBBk+RDyo/gAJKTIyx9SQMsfUjDL/1IQ9ADJ7VT4DwHTByHAAJ9sUZMg10qW0wfUAvsA6DDgIcAB4wAhwALjAAHAA5Ew4w0DpMjLHxLLH8v/BgUEAwAK9ADJ7VQAbIEBCNcY+gDTPzBSJIEBCPRZ8qeCEGRzdHJwdIAYyMsFywJQBc8WUAP6AhPLassfEss/yXP7AABwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwICAUgRCAIBIAoJAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCASAMCwARuMl+1E0NcLH4AgFYEA0CASAPDgAZrx32omhAEGuQ64WPwAAZrc52omhAIGuQ64X/wAA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYALm0AHQ0wMhcbCSXwTgItdJwSCSXwTgAtMfIYIQcGx1Z70ighBkc3RyvbCSXwXgA/pAMCD6RAHIygfL/8nQ7UTQgQFA1yH0BDBcgQEI9ApvoTGzkl8H4AXTP8glghBwbHVnupI4MOMNA4IQZHN0crqSXwbjDRMSAIpQBIEBCPRZMO1E0IEBQNcgyAHPFvQAye1UAXKwjiOCEGRzdHKDHrFwgBhQBcsFUAPPFiP6AhPLassfyz/JgED7AJJfA+IAeAH6APQEMPgnbyIwUAqhIb7y4FCCEHBsdWeDHrFwgBhQBMsFJs8WWPoCGfQAy2kXyx9SYMs/IMmAQPsABqZCg7I=")
	require.Nil(t, err)
	data, err := base64.StdEncoding.DecodeString("te6cckEBAQEAKwAAUQAAACIpqaMXt5/GUJUGuDtk+HdlAcW91x/58gRLxYvfD26hyGLEcWxAm7pXnQ==")
	require.Nil(t, err)

	ret, err := s.ParseAccountData(ctx, &core.AccountState{
		Address:  *addr.MustFromBase64("EQBCPrKazoIMW0CBYbHitNdrh2Lf_s70EtqdSqp0Y4k9Ul6N"),
		IsActive: true, Status: core.Active,
		Balance: bunbig.FromInt64(1e9),
		Code:    code,
		Data:    data,
	})
	require.Nil(t, err)
	require.Equal(t, []abi.ContractName{"wallet_v4r2"}, ret.Types)
	require.Equal(t, uint64(0x22), ret.WalletSeqNo)
}

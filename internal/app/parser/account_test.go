package parser

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/extra/bunbig"
	"github.com/xssnick/tonutils-go/tvm/cell"

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
	}, nil)
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
	}, nil)
	require.Nil(t, err)
	require.Equal(t, []abi.ContractName{"wallet_v4r2"}, ret.Types)
	require.Equal(t, uint64(0x22), ret.WalletSeqNo)
}

func TestService_ParseAccountData_NFTItem(t *testing.T) {
	s := newService(t)

	code, err := base64.StdEncoding.DecodeString("te6cckECDQEAAdAAART/APSkE/S88sgLAQIBYgIDAgLOBAUACaEfn+AFAgEgBgcCASALDALXDIhxwCSXwPg0NMDAXGwkl8D4PpA+kAx+gAxcdch+gAx+gAw8AIEs44UMGwiNFIyxwXy4ZUB+kDUMBAj8APgBtMf0z+CEF/MPRRSMLqOhzIQN14yQBPgMDQ0NTWCEC/LJqISuuMCXwSED/LwgCAkAET6RDBwuvLhTYAH2UTXHBfLhkfpAIfAB+kDSADH6AIIK+vCAG6EhlFMVoKHeItcLAcMAIJIGoZE24iDC//LhkiGOPoIQBRONkchQCc8WUAvPFnEkSRRURqBwgBDIywVQB88WUAX6AhXLahLLH8s/Im6zlFjPFwGRMuIByQH7ABBHlBAqN1viCgBycIIQi3cXNQXIy/9QBM8WECSAQHCAEMjLBVAHzxZQBfoCFctqEssfyz8ibrOUWM8XAZEy4gHJAfsAAIICjjUm8AGCENUydtsQN0QAbXFwgBDIywVQB88WUAX6AhXLahLLH8s/Im6zlFjPFwGRMuIByQH7AJMwMjTiVQLwAwA7O1E0NM/+kAg10nCAJp/AfpA1DAQJBAj4DBwWW1tgAB0A8jLP1jPFgHPFszJ7VSC/dQQb")
	require.Nil(t, err)
	data, err := base64.StdEncoding.DecodeString("te6cckEBAgEAWAABlQAAAAAAAABkgAmZdBGwAyeH1p8lmxniF4hL/lrgtKpVWt5op0KDyjb28AIihaT5me2lhAhFtxowTSuLb3JY8S1sv5rLvgAnLsoWVgEAEDEwMC5qc29u7rJBww==")
	require.Nil(t, err)

	others := func(_ context.Context, a *addr.Address) (*core.AccountState, error) {
		aStr := "EQBMy6CNgBk8PrT5LNjPELxCX_LXBaVSqtbzRToUHlG3t-fg"
		if a.Base64() != aStr {
			return nil, errors.Wrapf(core.ErrNotFound, "unexpected address %s", a.Base64())
		}

		code, err := base64.StdEncoding.DecodeString("te6cckECFAEAAh8AART/APSkE/S88sgLAQIBYgIDAgLNBAUCASAODwTn0QY4BIrfAA6GmBgLjYSK3wfSAYAOmP6Z/2omh9IGmf6mpqGEEINJ6cqClAXUcUG6+CgOhBCFRlgFa4QAhkZYKoAueLEn0BCmW1CeWP5Z+A54tkwCB9gHAbKLnjgvlwyJLgAPGBEuABcYES4AHxgRgZgeACQGBwgJAgEgCgsAYDUC0z9TE7vy4ZJTE7oB+gDUMCgQNFnwBo4SAaRDQ8hQBc8WE8s/zMzMye1Ukl8F4gCmNXAD1DCON4BA9JZvpSCOKQakIIEA+r6T8sGP3oEBkyGgUyW78vQC+gDUMCJUSzDwBiO6kwKkAt4Ekmwh4rPmMDJQREMTyFAFzxYTyz/MzMzJ7VQALDI0AfpAMEFEyFAFzxYTyz/MzMzJ7VQAPI4V1NQwEDRBMMhQBc8WE8s/zMzMye1U4F8EhA/y8AIBIAwNAD1FrwBHAh8AV3gBjIywVYzxZQBPoCE8trEszMyXH7AIAC0AcjLP/gozxbJcCDIywET9AD0AMsAyYAAbPkAdMjLAhLKB8v/ydCACASAQEQAlvILfaiaH0gaZ/qamoYLehqGCxABDuLXTHtRND6QNM/1NTUMBAkXwTQ1DHUMNBxyMsHAc8WzMmAIBIBITAC+12v2omh9IGmf6mpqGDYg6GmH6Yf9IBhAALbT0faiaH0gaZ/qamoYCi+CeAI4APgCwGlAMbg==")
		require.Nil(t, err)
		data, err := base64.StdEncoding.DecodeString("te6cckECEgEAAmcAA1OAH+KPIWfXRAHhzc8BIGKAZ7CGFDhMB09Wc+npbBemPgcgAAAAAAAAaBABAgMCAAQFART/APSkE/S88sgLBgBLAGQD6IAf4o8hZ9dEAeHNzwEgYoBnsIYUOEwHT1Zz6elsF6Y+BzAARAFodHRwczovL2xvdG9uLmZ1bi9jb2xsZWN0aW9uLmpzb24ALGh0dHBzOi8vbG90b24uZnVuL25mdC8CAWIHCAICzgkKAAmhH5/gBQIBIAsMAgEgEBEC1wyIccAkl8D4NDTAwFxsJJfA+D6QPpAMfoAMXHXIfoAMfoAMPACBLOOFDBsIjRSMscF8uGVAfpA1DAQI/AD4AbTH9M/ghBfzD0UUjC6jocyEDdeMkAT4DA0NDU1ghAvyyaiErrjAl8EhA/y8IA0OABE+kQwcLry4U2AB9lE1xwXy4ZH6QCHwAfpA0gAx+gCCCvrwgBuhIZRTFaCh3iLXCwHDACCSBqGRNuIgwv/y4ZIhjj6CEAUTjZHIUAnPFlALzxZxJEkUVEagcIAQyMsFUAfPFlAF+gIVy2oSyx/LPyJus5RYzxcBkTLiAckB+wAQR5QQKjdb4g8AcnCCEIt3FzUFyMv/UATPFhAkgEBwgBDIywVQB88WUAX6AhXLahLLH8s/Im6zlFjPFwGRMuIByQH7AACCAo41JvABghDVMnbbEDdEAG1xcIAQyMsFUAfPFlAF+gIVy2oSyx/LPyJus5RYzxcBkTLiAckB+wCTMDI04lUC8AMAOztRNDTP/pAINdJwgCafwH6QNQwECQQI+AwcFltbYAAdAPIyz9YzxYBzxbMye1Ugb+s9wA==")
		require.Nil(t, err)

		return &core.AccountState{
			Address:  *addr.MustFromBase64(aStr),
			IsActive: true, Status: core.Active,
			Balance: bunbig.FromInt64(1e9),
			Code:    code,
			Data:    data,
		}, nil
	}

	codeCell, err := cell.FromBOC(code)
	require.Nil(t, err)
	getMethodHashes, err := abi.GetMethodHashes(codeCell)
	require.Nil(t, err)

	ret, err := s.ParseAccountData(ctx, &core.AccountState{
		Address:  *addr.MustFromBase64("EQAQKmY9GTsEb6lREv-vxjT5sVHJyli40xGEYP3tKZSDuTBj"),
		IsActive: true, Status: core.Active,
		Balance:         bunbig.FromInt64(1e9),
		Code:            code,
		Data:            data,
		GetMethodHashes: getMethodHashes,
	}, others)
	require.Nil(t, err)
	require.Equal(t, []abi.ContractName{"nft_item"}, ret.Types)
	require.Equal(t, "https://loton.fun/nft/100.json", ret.NFTContentData.ContentURI)
}

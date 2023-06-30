package parser

import (
	"context"
	"encoding/base64"
	"encoding/json"
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

	ret := &core.AccountState{
		Address:  *addr.MustFromBase64("EQDj5AA8mQvM5wJEQsFFFof79y3ZsuX6wowktWQFhz_Anton"),
		IsActive: true, Status: core.Active,
		Balance: bunbig.FromInt64(1e9),
		Code:    code,
		Data:    data,
	}
	err = s.ParseAccountData(ctx, ret, nil)
	require.Nil(t, err)
	require.Equal(t, []abi.ContractName{"wallet_v3r2"}, ret.Types)
	j, err := json.Marshal(ret.ExecutedGetMethods)
	require.Nil(t, err)
	require.Equal(t, `{"wallet_v3r2":[{"name":"seqno","returns":[1]}]}`, string(j))
}

func TestService_ParseAccountData_WalletV4R2(t *testing.T) {
	s := newService(t)

	code, err := base64.StdEncoding.DecodeString("te6cckECFAEAAtQAART/APSkE/S88sgLAQIBIAcCBPjygwjXGCDTH9Mf0x8C+CO78mTtRNDTH9Mf0//0BNFRQ7ryoVFRuvKiBfkBVBBk+RDyo/gAJKTIyx9SQMsfUjDL/1IQ9ADJ7VT4DwHTByHAAJ9sUZMg10qW0wfUAvsA6DDgIcAB4wAhwALjAAHAA5Ew4w0DpMjLHxLLH8v/BgUEAwAK9ADJ7VQAbIEBCNcY+gDTPzBSJIEBCPRZ8qeCEGRzdHJwdIAYyMsFywJQBc8WUAP6AhPLassfEss/yXP7AABwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwICAUgRCAIBIAoJAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCASAMCwARuMl+1E0NcLH4AgFYEA0CASAPDgAZrx32omhAEGuQ64WPwAAZrc52omhAIGuQ64X/wAA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYALm0AHQ0wMhcbCSXwTgItdJwSCSXwTgAtMfIYIQcGx1Z70ighBkc3RyvbCSXwXgA/pAMCD6RAHIygfL/8nQ7UTQgQFA1yH0BDBcgQEI9ApvoTGzkl8H4AXTP8glghBwbHVnupI4MOMNA4IQZHN0crqSXwbjDRMSAIpQBIEBCPRZMO1E0IEBQNcgyAHPFvQAye1UAXKwjiOCEGRzdHKDHrFwgBhQBcsFUAPPFiP6AhPLassfyz/JgED7AJJfA+IAeAH6APQEMPgnbyIwUAqhIb7y4FCCEHBsdWeDHrFwgBhQBMsFJs8WWPoCGfQAy2kXyx9SYMs/IMmAQPsABqZCg7I=")
	require.Nil(t, err)
	data, err := base64.StdEncoding.DecodeString("te6cckEBAQEAKwAAUQAAACIpqaMXt5/GUJUGuDtk+HdlAcW91x/58gRLxYvfD26hyGLEcWxAm7pXnQ==")
	require.Nil(t, err)

	ret := &core.AccountState{
		Address:  *addr.MustFromBase64("EQBCPrKazoIMW0CBYbHitNdrh2Lf_s70EtqdSqp0Y4k9Ul6N"),
		IsActive: true, Status: core.Active,
		Balance: bunbig.FromInt64(1e9),
		Code:    code,
		Data:    data,
	}
	err = s.ParseAccountData(ctx, ret, nil)
	require.Nil(t, err)
	require.Equal(t, []abi.ContractName{"wallet_v4r2"}, ret.Types)
	j, err := json.Marshal(ret.ExecutedGetMethods)
	require.Nil(t, err)
	require.Equal(t, `{"wallet_v4r2":[{"name":"seqno","returns":[34]}]}`, string(j))
}

func TestService_ParseAccountData_NFTItem(t *testing.T) {
	s := newService(t)

	code, err := base64.StdEncoding.DecodeString("te6cckECDQEAAdAAART/APSkE/S88sgLAQIBYgIDAgLOBAUACaEfn+AFAgEgBgcCASALDALXDIhxwCSXwPg0NMDAXGwkl8D4PpA+kAx+gAxcdch+gAx+gAw8AIEs44UMGwiNFIyxwXy4ZUB+kDUMBAj8APgBtMf0z+CEF/MPRRSMLqOhzIQN14yQBPgMDQ0NTWCEC/LJqISuuMCXwSED/LwgCAkAET6RDBwuvLhTYAH2UTXHBfLhkfpAIfAB+kDSADH6AIIK+vCAG6EhlFMVoKHeItcLAcMAIJIGoZE24iDC//LhkiGOPoIQBRONkchQCc8WUAvPFnEkSRRURqBwgBDIywVQB88WUAX6AhXLahLLH8s/Im6zlFjPFwGRMuIByQH7ABBHlBAqN1viCgBycIIQi3cXNQXIy/9QBM8WECSAQHCAEMjLBVAHzxZQBfoCFctqEssfyz8ibrOUWM8XAZEy4gHJAfsAAIICjjUm8AGCENUydtsQN0QAbXFwgBDIywVQB88WUAX6AhXLahLLH8s/Im6zlFjPFwGRMuIByQH7AJMwMjTiVQLwAwA7O1E0NM/+kAg10nCAJp/AfpA1DAQJBAj4DBwWW1tgAB0A8jLP1jPFgHPFszJ7VSC/dQQb")
	require.Nil(t, err)
	data, err := base64.StdEncoding.DecodeString("te6cckEBAgEAWAABlQAAAAAAAABkgAmZdBGwAyeH1p8lmxniF4hL/lrgtKpVWt5op0KDyjb28AIihaT5me2lhAhFtxowTSuLb3JY8S1sv5rLvgAnLsoWVgEAEDEwMC5qc29u7rJBww==")
	require.Nil(t, err)

	others := func(_ context.Context, a addr.Address) (*core.AccountState, error) {
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

	ret := &core.AccountState{
		Address:  *addr.MustFromBase64("EQAQKmY9GTsEb6lREv-vxjT5sVHJyli40xGEYP3tKZSDuTBj"),
		IsActive: true, Status: core.Active,
		Balance:         bunbig.FromInt64(1e9),
		Code:            code,
		Data:            data,
		GetMethodHashes: getMethodHashes,
	}
	err = s.ParseAccountData(ctx, ret, others)
	require.Nil(t, err)
	require.Equal(t, []abi.ContractName{"nft_item"}, ret.Types)
	require.Equal(t, "https://loton.fun/nft/100.json", ret.NFTContentData.ContentURI)
	j, err := json.Marshal(ret.ExecutedGetMethods)
	require.Nil(t, err)
	require.Equal(t, `{"nft_collection":[{"name":"get_nft_content","receives":[100,"te6cckEBAQEACgAAEDEwMC5qc29ue9bV9g=="],"returns":[{"URI":"https://loton.fun/nft/100.json"}]},{"name":"get_nft_address_by_index","receives":[100],"returns":["EQAQKmY9GTsEb6lREv-vxjT5sVHJyli40xGEYP3tKZSDuTBj"]}],"nft_item":[{"name":"get_nft_data","returns":[true,100,"EQBMy6CNgBk8PrT5LNjPELxCX_LXBaVSqtbzRToUHlG3t-fg","EQCIoWk-ZntpYQIRbcaME0ri29yWPEtbL-ay74AJy7KFlcfj","te6cckEBAQEACgAAEDEwMC5qc29ue9bV9g=="]}]}`, string(j))
	require.Equal(t, false, ret.Fake)
}

func TestService_ParseAccountData_JettonWallet_Fake(t *testing.T) {
	s := newService(t)

	code, err := base64.StdEncoding.DecodeString("te6cckECEwEAA8oAART/APSkE/S88sgLAQIBYgIDAgLMBAUAYaD2BdqJofQB9IEaEMAGWqybXjgOrpKNs8jiONm8DWGpMg/cK8ei9sh9b+35IIgDqGECAdQGBwIBSAgJAMMIMcAkl8E4AHQ0wMBcbCVE18D8Avg+kD6QDH6ADFx1yH6ADH6ADBzqbQAAtMfghAPin6lUiC6lTE0WfAI4IIQF41FGVIgupYxREQD8AngNYIQWV8HvLqTWfAK4F8EhA/y8IAARPpEMHC68uFNgAgEgCgsCASAQEQH1APTP/oA+kAh8AHtRND6APpAjQhgAy1WTa8cB1dJRtnkcRxs3gaw1JkH7hXj0XtkPrf2/JBEAdQwUTahUirHBfLiwSjC//LiwlQ0QnBUIBNUFAPIUAT6AljPFgHPFszJIsjLARL0APQAywDJIPkAcHTIywLKB8v/ydAEgDAPlDPtRND6APpAjQhgAy1WTa8cB1dJRtnkcRxs3gaw1JkH7hXj0XtkPrf2/JBEAdQwB9M/+gBRUaAF+kD6QPoAUauhggiYloCCCJiWgBK2CKGCCOThwKAboSqWEEpQmF8F4w0k1wsBwwAkwgCwkmwz4w1VAoA0ODwDw+kD0BDH6ACDXScIA8uLEd4AYyMsFUAjPFnD6AhfLaxPMghAXjUUZyMsfGcs/UAf6AiLPFlAGzxYl+gJQA88WyVAFzCORcpFx4lAIqBOgggjk4cCqAIIImJaAoKAUvPLixQTJgED7ABAjyFAE+gJYzxYBzxbMye1UALJSqaAYoYIQc2LQnMjLH1JAyz9QA/oCAc8WUAfPFslxgBDIywWNCGAFtaFwmOlCb1HItK/Cw0QAL5ThHake2AtrQI1iKsjqliTPFlAJ+gIYy2oXzMlx+wAQNABEghDVMnbbcIAQyMsFUAfPFlAF+gIVy2oTyx8Uyz/JcvsAAQAeyFAE+gJYzxYBzxbMye1UAfc7UTQ+gD6QI0IYAMtVk2vHAdXSUbZ5HEcbN4GsNSZB+4V49F7ZD639vyQRAHUMAfTP/oA+kAwUVGhUknHBfLiwSfC//LiwoII5OHAqgAWoBa88uLDghB73ZfeyMsfFcs/UAP6AiLPFgHPFslxgBjIywUkzxZw+gLLaszJgEgDJIAg1yHtRND6APpAjQhgAy1WTa8cB1dJRtnkcRxs3gaw1JkH7hXj0XtkPrf2/JBEAdQwBNMfghAXjUUZUiC6ghB73ZfeE7oSsfLixdM/MfoAMBOgUCPIUAT6AljPFgHPFszJ7VSAAKoBA+wBAE8hQBPoCWM8WAc8WzMntVOMAf0I=")
	require.Nil(t, err)
	data, err := base64.StdEncoding.DecodeString("te6cckECFAEABBgAAZVx+LHi0aLACAHm2P3/xfKWJgfnQi1JlJXfQO0vEjAqDxsMyYMWC9YLUwAZarJteOA6uko2zyOI42bwNYakyD9wrx6L2yH1v7fkgiABART/APSkE/S88sgLAgIBYgMEAgLMBQYAYaD2BdqJofQB9IEaEMAGWqybXjgOrpKNs8jiONm8DWGpMg/cK8ei9sh9b+35IIgDqGECAdQHCAIBSAkKAMMIMcAkl8E4AHQ0wMBcbCVE18D8Avg+kD6QDH6ADFx1yH6ADH6ADBzqbQAAtMfghAPin6lUiC6lTE0WfAI4IIQF41FGVIgupYxREQD8AngNYIQWV8HvLqTWfAK4F8EhA/y8IAARPpEMHC68uFNgAgEgCwwCASAREgH1APTP/oA+kAh8AHtRND6APpAjQhgAy1WTa8cB1dJRtnkcRxs3gaw1JkH7hXj0XtkPrf2/JBEAdQwUTahUirHBfLiwSjC//LiwlQ0QnBUIBNUFAPIUAT6AljPFgHPFszJIsjLARL0APQAywDJIPkAcHTIywLKB8v/ydAEgDQPlDPtRND6APpAjQhgAy1WTa8cB1dJRtnkcRxs3gaw1JkH7hXj0XtkPrf2/JBEAdQwB9M/+gBRUaAF+kD6QPoAUauhggiYloCCCJiWgBK2CKGCCOThwKAboSqWEEpQmF8F4w0k1wsBwwAkwgCwkmwz4w1VAoA4PEADw+kD0BDH6ACDXScIA8uLEd4AYyMsFUAjPFnD6AhfLaxPMghAXjUUZyMsfGcs/UAf6AiLPFlAGzxYl+gJQA88WyVAFzCORcpFx4lAIqBOgggjk4cCqAIIImJaAoKAUvPLixQTJgED7ABAjyFAE+gJYzxYBzxbMye1UALJSqaAYoYIQc2LQnMjLH1JAyz9QA/oCAc8WUAfPFslxgBDIywWNCGAFtaFwmOlCb1HItK/Cw0QAL5ThHake2AtrQI1iKsjqliTPFlAJ+gIYy2oXzMlx+wAQNABEghDVMnbbcIAQyMsFUAfPFlAF+gIVy2oTyx8Uyz/JcvsAAQAeyFAE+gJYzxYBzxbMye1UAfc7UTQ+gD6QI0IYAMtVk2vHAdXSUbZ5HEcbN4GsNSZB+4V49F7ZD639vyQRAHUMAfTP/oA+kAwUVGhUknHBfLiwSfC//LiwoII5OHAqgAWoBa88uLDghB73ZfeyMsfFcs/UAP6AiLPFgHPFslxgBjIywUkzxZw+gLLaszJgEwDJIAg1yHtRND6APpAjQhgAy1WTa8cB1dJRtnkcRxs3gaw1JkH7hXj0XtkPrf2/JBEAdQwBNMfghAXjUUZUiC6ghB73ZfeE7oSsfLixdM/MfoAMBOgUCPIUAT6AljPFgHPFszJ7VSAAKoBA+wBAE8hQBPoCWM8WAc8WzMntVGuroJQ=")
	require.Nil(t, err)

	others := func(_ context.Context, a addr.Address) (*core.AccountState, error) {
		aStr := "EQBlqsm144Dq6SjbPI4jjZvA1hqTIP3CvHovbIfW_t-SCALE"
		if a.Base64() != aStr {
			return nil, errors.Wrapf(core.ErrNotFound, "unexpected address %s", a.Base64())
		}

		code, err := base64.StdEncoding.DecodeString("te6cckECCQEAAbAAART/APSkE/S88sgLAQIBYgIDAgLMBAUCA3pgBwgB3dmRDjgEit8GhpgYC42Eit8H0gGADpj+mf9qJofQB9IGpqGEAKqThdRxoamoDgAHlwJCiiY4L5cCT9IH0AahgQaEAwa5D9ABgSCBooIXgEKpBkKAJ9ASxni2ZmZPaqcEEIPe7L7wvdcYEvg8IH+XhAYAk9/BQiAbgqEAmqCgHkKAJ9ASxniwDni2ZkkWRlgIl6AHoAZYBkkHyAODpkZYFlA+X/5Og7wAxkZYKsZ4soAn0BCeW1iWZmZLj9gEAPwD+gD6QPgoVBIIcFQgE1QUA8hQBPoCWM8WAc8WzMkiyMsBEvQA9ADLAMn5AHB0yMsCygfL/8nQUAjHBfLgShKhA1AkyFAE+gJYzxbMzMntVAH6QDAg1wsBwwCOH4IQ1TJ223CAEMjLBVADzxYi+gISy2rLH8s/yYBC+wCRW+IAfa289qJofQB9IGpqGDYY/BQAuCoQCaoKAeQoAn0BLGeLAOeLZmSRZGWAiXoAegBlgGT8gDg6ZGWBZQPl/+ToQAAjrxb2omh9AH0gamoYEeAAKpBAQ6gA1A==")
		require.Nil(t, err)
		data, err := base64.StdEncoding.DecodeString("te6cckECFAEAA40AAlFztIY1u7U7eAGRDZCVSHaHb9cm1hPKMRV84bFGDAD3HkxTW5nQAcumsQECAGwBaXBmczovL1FtZWViWm00c0NtWEdkTTlpaUQ4VHRZRHlRd3lGYTNEbTI3aHhvblZReUZUNFABFP8A9KQT9LzyyAsDAgFiBAUCAswGBwAboPYF2omh9AH0gfSBqGECAdQICQIBSAoLALsIMcAkl8E4AHQ0wMBcbCVE18D8Avg+kD6QDH6ADFx1yH6ADH6ADAC0x+CEA+KfqVSILqVMTRZ8AjgghAXjUUZUiC6ljFERAPwCeA1ghBZXwe8upNZ8ArgXwSED/LwgABE+kQwcLry4U2ACASAMDQIBIBITAfUA9M/+gD6QCHwAe1E0PoA+kD6QNQwUTahUirHBfLiwSjC//LiwlQ0QnBUIBNUFAPIUAT6AljPFgHPFszJIsjLARL0APQAywDJIPkAcHTIywLKB8v/ydAE+kD0BDH6AHeAGMjLBVAIzxZw+gIXy2sTzIIQF41FGcjLHxmAOA/c7UTQ+gD6QPpA1DAI0z/6AFFRoAX6QPpAU1vHBVRzbXBUIBNUFAPIUAT6AljPFgHPFszJIsjLARL0APQAywDJ+QBwdMjLAsoHy//J0FANxwUcsfLiwwr6AFGooYIImJaAggiYloAStgihggiYloCgGKEn4w8l1wsBwwAjgDxARAJrLP1AH+gIizxZQBs8WJfoCUAPPFslQBcwjkXKRceJQCKgToIIImJaAqgCCCJiWgKCgFLzy4sUEyYBA+wAQI8hQBPoCWM8WAc8WzMntVABwUnmgGKGCEHNi0JzIyx9SMMs/WPoCUAfPFlAHzxbJcYAYyMsFJM8WUAb6AhXLahTMyXH7ABAkECMADhBJEDg3XwQAdsIAsI4hghDVMnbbcIAQyMsFUAjPFlAE+gIWy2oSyx8Syz/JcvsAkzVsIeIDyFAE+gJYzxYBzxbMye1UANs7UTQ+gD6QPpA1DAH0z/6APpAMFFRoVJJxwXy4sEnwv/y4sKCCJiWgKoAFqAWvPLiw4IQe92X3sjLHxXLP1AD+gIizxYBzxbJcYAYyMsFJM8WcPoCy2rMyYBA+wBAE8hQBPoCWM8WAc8WzMntVIACDIAg1yHtRND6APpA+kDUMATTH4IQF41FGVIguoIQe92X3hO6ErHy4sXTPzH6ADAToFAjyFAE+gJYzxYBzxbMye1UgClVyqA==")
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

	ret := &core.AccountState{
		Address:  *addr.MustFromBase64("EQD9vvnQn6Y9pC7cHnHTYw00yvEsL0Pob5zDMCV9iYMNgY2x"),
		IsActive: true, Status: core.Active,
		Balance:         bunbig.FromInt64(1e9),
		Code:            code,
		Data:            data,
		GetMethodHashes: getMethodHashes,
	}
	err = s.ParseAccountData(ctx, ret, others)
	require.Nil(t, err)
	require.Equal(t, []abi.ContractName{"jetton_wallet"}, ret.Types)
	require.Equal(t, true, ret.Fake)
	// j, err := json.Marshal(ret.ExecutedGetMethods)
	// require.Nil(t, err)
	// require.Equal(t, `{"jetton_minter":[{"name":"get_wallet_address","receives":["EQDzbH7_4vlLEwPzoRakykrvoHaXiRgVB42GZMGLBesFqemt"],"returns":["EQBWICDwlBzfMdyM56TAMikgKVNfssQzvqKK964A1SIlC8jb"]}],"jetton_wallet":[{"name":"get_wallet_data","returns":[8878686000000000,"EQDzbH7_4vlLEwPzoRakykrvoHaXiRgVB42GZMGLBesFqemt","EQBlqsm144Dq6SjbPI4jjZvA1hqTIP3CvHovbIfW_t-SCALE","te6cckECEwEAA8oAART/APSkE/S88sgLAQIBYgMCAGGg9gXaiaH0AfSBGhDABlqsm144Dq6SjbPI4jjZvA1hqTIP3CvHovbIfW/t+SCIA6hhAgLMBgQCAUgFDAIBIBEJAgHUCAcAET6RDBwuvLhTYADDCDHAJJfBOAB0NMDAXGwlRNfA/AL4PpA+kAx+gAxcdch+gAx+gAwc6m0AALTH4IQD4p+pVIgupUxNFnwCOCCEBeNRRlSILqWMUREA/AJ4DWCEFlfB7y6k1nwCuBfBIQP8vCAD5Qz7UTQ+gD6QI0IYAMtVk2vHAdXSUbZ5HEcbN4GsNSZB+4V49F7ZD639vyQRAHUMAfTP/oAUVGgBfpA+kD6AFGroYIImJaAggiYloAStgihggjk4cCgG6EqlhBKUJhfBeMNJNcLAcMAJMIAsJJsM+MNVQKAQCwoAHshQBPoCWM8WAc8WzMntVABEghDVMnbbcIAQyMsFUAfPFlAF+gIVy2oTyx8Uyz/JcvsAAQIBIA4NAMkgCDXIe1E0PoA+kCNCGADLVZNrxwHV0lG2eRxHGzeBrDUmQfuFePRe2Q+t/b8kEQB1DAE0x+CEBeNRRlSILqCEHvdl94TuhKx8uLF0z8x+gAwE6BQI8hQBPoCWM8WAc8WzMntVIAH3O1E0PoA+kCNCGADLVZNrxwHV0lG2eRxHGzeBrDUmQfuFePRe2Q+t/b8kEQB1DAH0z/6APpAMFFRoVJJxwXy4sEnwv/y4sKCCOThwKoAFqAWvPLiw4IQe92X3sjLHxXLP1AD+gIizxYBzxbJcYAYyMsFJM8WcPoCy2rMyYA8AKoBA+wBAE8hQBPoCWM8WAc8WzMntVACyUqmgGKGCEHNi0JzIyx9SQMs/UAP6AgHPFlAHzxbJcYAQyMsFjQhgBbWhcJjpQm9RyLSvwsNEAC+U4R2pHtgLa0CNYirI6pYkzxZQCfoCGMtqF8zJcfsAEDQB9QD0z/6APpAIfAB7UTQ+gD6QI0IYAMtVk2vHAdXSUbZ5HEcbN4GsNSZB+4V49F7ZD639vyQRAHUMFE2oVIqxwXy4sEowv/y4sJUNEJwVCATVBQDyFAE+gJYzxYBzxbMySLIywES9AD0AMsAySD5AHB0yMsCygfL/8nQBIBIA8PpA9AQx+gAg10nCAPLixHeAGMjLBVAIzxZw+gIXy2sTzIIQF41FGcjLHxnLP1AH+gIizxZQBs8WJfoCUAPPFslQBcwjkXKRceJQCKgToIII5OHAqgCCCJiWgKCgFLzy4sUEyYBA+wAQI8hQBPoCWM8WAc8WzMntVB9hzdY="]}]}`, string(j))
}

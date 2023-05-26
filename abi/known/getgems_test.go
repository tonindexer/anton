package known_test

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
)

func TestGetMethodDesc_GetGemsNFTAuction(t *testing.T) {
	var (
		interfaces   []*abi.InterfaceDesc
		interfaceMap = make(map[string]*abi.InterfaceDesc)
	)

	j, err := os.ReadFile("getgems.json")
	require.Nil(t, err)

	err = json.Unmarshal(j, &interfaces)
	require.Nil(t, err)

	for _, i := range interfaces {
		err := i.RegisterDefinitions()
		require.Nil(t, err)

		b64, err := base64.StdEncoding.DecodeString(i.CodeBoc)
		require.Nil(t, err)

		c, err := cell.FromBOC(b64)
		require.Nil(t, err)

		interfaceMap[string(c.Hash())] = i
	}

	getInterfaceByCode := func(h []byte) *abi.InterfaceDesc {
		i, ok := interfaceMap[string(h)]
		require.True(t, ok)
		return i
	}

	var testCases = []*struct {
		name     string
		addr     *address.Address
		code     string
		data     string
		expected []any
	}{
		{
			name: "get_sale_data",
			addr: addr.MustFromBase64("EQCGCDRk4nYZDIH4Z-ZlEp3bpOvzQ6u96BnE8yw3r3pzdW9I").MustToTonutils(),
			code: `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`,
			data: `te6cckEBBAEA5AADS8AFhO5hst/wg3EW0Py1B42TlkvL6cBf1qFBsb/KXWpD4YsYES46AQIDAKWAFHJrDD7ztes0J62jBcLIBCGAXzHHvjH7TpcetWKDV+MAAAAAoAAADJAAXXlQ6yZE0u0tK+YE2MpwfL1QhR6PGxiFRAY74fT0ALgAAAAoAAABkgA5ZXT73mAAAEO5rKAAAAAAABjFJrHAAABLAAAASyAAhYABfAsRKKg4e+3J8Gn3aWhdKFw/Pas5LOYwbqntU7CMcXAAx+LkNPvaNIajlzitRgaFrD/drrrKzKCkFN2TE0ld0H4YGxQ8`,
			expected: []any{
				big.NewInt(4281667),
				true,
				big.NewInt(1662294727),
				address.MustParseAddr("EQBYTuYbLf8INxFtD8tQeNk5ZLy-nAX9ahQbG_yl1qQ-GEMS"),
				address.MustParseAddr("EQAx-LkNPvaNIajlzitRgaFrD_drrrKzKCkFN2TE0ld0H-D6"),
				address.MustParseAddr("EQAL4FiJRUHD325Pg0-7S0LpQuH57VnJZzGDdU9qnYRji-e_"),
				big.NewInt(0),
				address.NewAddressNone(),
				big.NewInt(1000000000),
				address.MustParseAddr("EQCjk1hh952vWaE9bRguFkAhDAL5jj3xj9p0uPWrFBq_GEMS"),
				big.NewInt(5),
				big.NewInt(100),
				address.MustParseAddr("EQAXXlQ6yZE0u0tK-YE2MpwfL1QhR6PGxiFRAY74fT0ALgNw"),
				big.NewInt(10),
				big.NewInt(100),
				big.NewInt(0),
				big.NewInt(96000000000000),
				big.NewInt(1661085127),
				big.NewInt(0),
				false,
			},
		},
		{
			name: "get_sale_data",
			addr: addr.MustFromBase64("EQBBhkR9sVrASMQRlTC8TZR2RHP8bkFIkEF3wcFe2phJhp22").MustToTonutils(),
			code: `te6cckECHQEABZMAART/APSkE/S88sgLAQIBIAIDAgFIBAUCKPIw2zyBA+74RMD/8vL4AH/4ZNs8GxwCAs4GBwKLoDhZtnm2eQQQgqqH8IXwofCH8KfwpfCd8JvwmfCX8JXwi/Cf8IwaIiYaGCIkGBYiIhYUIiAUIT4hHCD6INggtiD0INIgsRsaAgEgCAkCASAYGQT1AHQ0wMBcbDyQPpAMNs8+ELA//hDUiDHBbCO0DMx0x8hwACNBJyZXBlYXRfZW5kX2F1Y3Rpb26BSIMcFsI6DW9s84DLAAI0EWVtZXJnZW5jeV9tZXNzYWdlgUiDHBbCa1DDQ0wfUMAH7AOAw4PhTUhDHBY6EMzHbPOABgGxIKCwATIIQO5rKAAGphIAFcMYED6fhS10nCAvLygQPqAdMfghAFE42REroS8vSAQNch+kAw+HJw+GJ/+GTbPBwEhts8IMABjzgwgQPt+CP4UL7y8oED7fhCwP/y8oED8AKCEDuaygC5EvLy+FJSEMcF+ENSIMcFsfLhkwF/2zzbPOAgwAIMFQ0OAIwgxwDA/5IwcODTHzGLZjYW5jZWyCHHBZIwceCLRzdG9wghxwWSMHLgi2ZmluaXNoghxwWSMHLgi2ZGVwbG95gBxwWRc+BwAYpwIPglghBfzD0UyMsfyz/4Us8WUAPPFhLLACH6AssAyXGAGMjLBfhTzxZw+gLLasyCCA9CQHD7AsmDBvsAf/hif/hm2zwcBPyOwzAygQPt+ELA//LygQPwAYIQO5rKALny8vgj+FC+jhf4UlIQxwX4Q1IgxwWx+E1SIMcFsfLhk5n4UlIQxwXy4ZPi2zzgwAOSXwPg+ELA//gj+FC+sZdfA4ED7fLw4PhLghA7msoAoFIgvvhLwgCw4wL4UPhRofgjueMA+E4SDxARAiwCcNs8IfhtghA7msoAofhu+CP4b9s8FRIADvhQ+FGg+HADcI6VMoED6PhKUiC58vL4bvht+CP4b9s84fhO+EygUiC5l18DgQPo8vDgAnDbPAH4bfhu+CP4b9s8HBUcApT4TsAAjj1wIPglghBfzD0UyMsfyz/4Us8WUAPPFhLLACH6AssAyXGAGMjLBfhTzxZw+gLLasyCCA9CQHD7AsmDBvsA4w5/+GLbPBMcAvrbPPhOQFTwAyDCAI4rcCCAEMjLBVAHzxYi+gIWy2oVyx+L9NYXJrZXRwbGFjZSBmZWWM8WyXL7AJE04vhOQAPwAyDCAI4jcCCAEMjLBVAEzxYi+gITy2oSyx+LdSb3lhbHR5jPFsly+wCRMeKCCA9CQHD7AvhOWKEBoSDCABoUAMCOInAggBDIywX4Us8WUAP6AhLLassfi2UHJvZml0jPFsly+wCRMOJwIPglghBfzD0UyMsfyz/4Tc8WUAPPFhLLAIIImJaA+gLLAMlxgBjIywX4U88WcPoCy2rMyYMG+wAC8vhOwQGRW+D4TvhHoSKCCJiWgKFSELyZMAGCCJiWgKEBkTLijQpWW91ciBiaWQgaGFzIGJlZW4gb3V0YmlkIGJ5IGFub3RoZXIgdXNlci6ABwP+OHzCNBtBdWN0aW9uIGhhcyBiZWVuIGNhbmNlbGxlZC6DeIcIA4w8WFwA4cCCAGMjLBfhNzxZQBPoCE8tqEssfAc8WyXL7AAACWwARIIQO5rKAKmEgAB0IMAAk18DcOBZ8AIB8AGAAIPhI0PpA0x/TH/pA0x/THzAAyvhBbt3tRNDSAAH4YtIAAfhk0gAB+Gb6QAH4bfoAAfhu0x8B+G/THwH4cPpAAfhy1AH4aNQw+Gn4SdDSHwH4Z/pAAfhj+gAB+Gr6AAH4a/oAAfhs0x8B+HH6QAH4c9MfMPhlf/hhAFT4SfhI+FD4T/hG+ET4QsjKAMoAygD4Tc8W+E76Assfyx/4Us8WzMzJ7VQBqlR8`,
			data: `te6cckEBAwEA5wACVcAAAAAAMh4aEMAIK9sXWs6Gja7sTGW8RMnRJOd2eVHm3glGc8M6RTfAq6gBAgClgBRyaww+87XrNCetowXCyAQhgF8xx74x+06XHrVig1fjAAAAAKAAAAyQADKf5sPbNoDE1wWr8GnWGDIsANPb8eaaRz6r29bf/HE0AAAAKAAAAZIAwQCA6+iACwncw2W/4QbiLaH5ag8bJyyXl9OAv61Cg2N/lLrUh8MKF0h26ADAV0+95gAKBKgXyAAAAAJZADntpmm2XS5oYrslQmTOYkrPY2vOEb5tWs4sqzdS+bQcWQpvy2Cqii/C`,
			expected: []any{
				big.NewInt(4281667),
				true,
				big.NewInt(1681667105),
				address.MustParseAddr("EQBYTuYbLf8INxFtD8tQeNk5ZLy-nAX9ahQbG_yl1qQ-GEMS"),
				address.MustParseAddr("EQDntpmm2XS5oYrslQmTOYkrPY2vOEb5tWs4sqzdS-bQcfHF"),
				address.MustParseAddr("EQCCvbF1rOho2u7ExlvETJ0STndnlR5t4JRnPDOkU3wKujeJ"),
				big.NewInt(0),
				address.NewAddressNone(),
				big.NewInt(10000000000),
				address.MustParseAddr("EQCjk1hh952vWaE9bRguFkAhDAL5jj3xj9p0uPWrFBq_GEMS"),
				big.NewInt(5),
				big.NewInt(100),
				address.MustParseAddr("EQAMp_mw9s2gMTXBavwadYYMiwA09vx5ppHPqvb1t_8cTbZo"),
				big.NewInt(10),
				big.NewInt(100),
				big.NewInt(3000000000000),
				big.NewInt(50000000000),
				big.NewInt(1680457517),
				big.NewInt(0),
				false,
			},
		},
		{
			name: "get_sale_data",
			addr: addr.MustFromBase64("EQB3iXx-uvTsBsmU7exvGe97Q0JvHfgJdzli0qhZQ5vIUcfo").MustToTonutils(),
			code: `te6cckECHAEABZcAART/APSkE/S88sgLAQIBIAIDAgFIBAUELPLbPPhEwACOiDD4AH/4ZNs84Ns8wAIUGxUWAgLOBgcCi6A4WbZ5tnkEEIKqh/CF8KHwh/Cn8KXwnfCb8Jnwl/CV8Ivwn/CMGiImGhgiJBgWIiIWFCIgFCE+IRwg+iDYILYg9CDSILEUGQIBIAgJAgEgEhME9QB0NMDAXGw8kD6QDDbPPhCwP/4Q1IgxwWwjtAzMdMfIcAAjQScmVwZWF0X2VuZF9hdWN0aW9ugUiDHBbCOg1vbPOAywACNBFlbWVyZ2VuY3lfbWVzc2FnZYFIgxwWwmtQw0NMH1DAB+wDgMOD4U1IQxwWOhDMx2zzgAYBQXCgsAEyCEDuaygABqYSABXDGBA+n4UtdJwgLy8oED6gHTH4IQBRONkRK6EvL0gEDXIfpAMPhycPhif/hk2zwbBOjbPCDAAY69MDKBA+34I/hQvvLygQPt+ELA//LygQPwAYIQO5rKALny8oED8fhOwgDy8vhSUhDHBfhDUiDHBbHy4ZPbPOAgwALjAsADkl8D4PhCwP/4I/hQvrGXXwOBA+3y8OD4S4IQO5rKAKBSIL74S8IAsBUYDA0BdjAygQPt+ELA//LygQPwAYIQO5rKALny8oED8vgj+FC58vL4UlIQxwX4Q1IgxwWx+E1SIMcFsfLhk9s8FwTOjxYCcNs8IfhtghA7msoAofhu+CP4b9s84PhQ+FGh+CO5l/hQ+FGg+HDe+E6OlTKBA+j4SlIgufLy+G74bfgj+G/bPOH4ToIQBfXhAKD4TvhMpmSAZPADtglSILmXXwOBA+jy8OACcA8XGw4CGts8Afht+G74I/hv2zwPGwLy+E7BAZFb4PhO+EehIoIImJaAoVIQvJkwAYIImJaAoQGRMuKNClZb3VyIGJpZCBoYXMgYmVlbiBvdXRiaWQgYnkgYW5vdGhlciB1c2VyLoAHA/44fMI0G0F1Y3Rpb24gaGFzIGJlZW4gY2FuY2VsbGVkLoN4hwgDjDxARADhwIIAYyMsF+E3PFlAE+gITy2oSyx8BzxbJcvsAAAJbABEghA7msoAqYSAAHQgwACTXwNw4FnwAgHwAYADK+EFu3e1E0NIAAfhi0gAB+GTSAAH4ZvpAAfht+gAB+G7THwH4b9MfAfhw+kAB+HLUAfho1DD4afhJ0NIfAfhn+kAB+GP6AAH4avoAAfhr+gAB+GzTHwH4cfpAAfhz0x8w+GV/+GEAjCDHAMD/kjBw4NMfMYtmNhbmNlbIIccFkjBx4ItHN0b3CCHHBZIwcuCLZmaW5pc2iCHHBZIwcuCLZkZXBsb3mAHHBZFz4HABZI6rgQPt+ELA//LygQPy+CP4ULny8vgnbyIwgQPwAYIQO5rKALny8vgA+FLbPOCED/LwFwP2+E7AAI6C2zzg2zz4TkBU8AMgwgCOK3AggBDIywVQB88WIvoCFstqFcsfi/TWFya2V0cGxhY2UgZmVljPFsly+wCRNOL4TkAD8AMgwgCOI3AggBDIywVQBM8WIvoCE8tqEssfi3Um95YWx0eYzxbJcvsAkTHigggPQkBwGBkaAYpwIPglghBfzD0UyMsfyz/4Us8WUAPPFhLLACH6AssAyXGAGMjLBfhTzxZw+gLLasyCCA9CQHD7AsmDBvsAf/hif/hm2zwbACD4SND6QNMf0x/6QNMf0x8wAeD7AvhOWKEBoSDCAI4icCCAEMjLBfhSzxZQA/oCEstqyx+LZQcm9maXSM8WyXL7AJEw4nAg+CWCEF/MPRTIyx/LP/hNzxZQA88WEssAggiYloD6AssAyXGAGMjLBfhTzxZw+gLLaszJgwb7AH/4Yts8GwBU+En4SPhQ+E/4RvhE+ELIygDKAMoA+E3PFvhO+gLLH8sf+FLPFszMye1UZp3BZw==`,
			data: `te6cckECAwEAAQcAAqFQAh2XUDm6vnWfzFYn9c7GS1FD6f3iXkkybnLneiVyTQKpQFeea4AZGXl2GRmm1eACF48fnE8fKyIRyYAKHIFXNAoX760tZqKXZXiVlhu/QNQBAgClgBRyaww+87XrNCetowXCyAQhgF8xx74x+06XHrVig1fjAAAAAKAAAAyQAexixhd154NpDCiyJ46u1VpM/KonvKULWXdWfBGqsgiEAAAAFAAAAZIAtQCA6+iACwncw2W/4QbiLaH5ag8bJyyXl9OAv61Cg2N/lLrUh8MIdzWUAKF0h26AAgoAAAJZACh8R/3IghYujp318YlvxG2AjDOL81AWc3mjVjkJp7Ge2RdYNiDRCRg9`,
			expected: []any{
				big.NewInt(4281667),
				false,
				big.NewInt(1684446039),
				address.MustParseAddr("EQBYTuYbLf8INxFtD8tQeNk5ZLy-nAX9ahQbG_yl1qQ-GEMS"),
				address.MustParseAddr("EQCh8R_3IghYujp318YlvxG2AjDOL81AWc3mjVjkJp7Ge639"),
				address.MustParseAddr("EQBC8ePziePlZEI5MAFDkCrmgUL99aWs1FLsrxKyw3foGiBt"),
				big.NewInt(5880000000),
				address.MustParseAddr("EQCHZdQObq-dZ_MVif1zsZLUUPp_eJeSTJucud6JXJNAqp8v"),
				big.NewInt(5),
				address.MustParseAddr("EQCjk1hh952vWaE9bRguFkAhDAL5jj3xj9p0uPWrFBq_GEMS"),
				big.NewInt(5),
				big.NewInt(100),
				address.MustParseAddr("EQB7GLGF3Xng2kMKLInjq7VWkz8qie8pQtZd1Z8EaqyCIYTw"),
				big.NewInt(5),
				big.NewInt(100),
				big.NewInt(50000000000),
				big.NewInt(1000000000),
				big.NewInt(1683841240),
				big.NewInt(1684399576),
				false,
			},
		},
		{
			name: "get_sale_data",
			addr: addr.MustFromBase64("EQAZIWzvPG68dMpw06nURIXxi76P-2naW4Wvhptz26KgXAHz").MustToTonutils(),
			code: `te6cckECDAEAAikAART/APSkE/S88sgLAQIBIAIDAgFIBAUABPIwAgLNBgcAUaA4WdqJoaYBpj/0gfSB9IH0AahhofSB9AH0gfQAYQQgjJKwoBWAAqsBAvfQDoaYGAuNhJL4JwfSAYdqJoaYBpj/0gfSB9IH0AahgTYAD5aMoRa6ThAVnHIBkcHCmg44LJL4RwKKJjgvlw+gJpj8EIAonGyIldeXD66Z+Y/SAYICsDZGWACuWPqAHniwDniwDniwD9AWZk9qpwGxPjgHGBA+mP6Z+YEMCAkB92YIQO5rKAFJgoFIwvvLhwiTQ+kD6APpA+gAwU5KhIaFQh6EWoFKQcIAQyMsFUAPPFgH6AstqyXH7ACXCACXXScICsI4XUEVwgBDIywVQA88WAfoCy2rJcfsAECOSNDTiWnCAEMjLBVADzxYB+gLLaslx+wBwIIIQX8w9FILABY3EDhHZRRDMHDwBQFKwAGSXwvgIcACnzEQSRA4R2AQJRAkECPwBeA6wAPjAl8JhA/y8AoAyoIQO5rKABi+8uHJU0bHBVFSxwUVsfLhynAgghBfzD0UIYAQyMsFKM8WIfoCy2rLHxnLPyfPFifPFhjKACf6AhfKAMmAQPsAcQZQREUVBsjLABXLH1ADzxYBzxYBzxYB+gLMye1UAIIhgBjIywUqzxYh+gLLassfE8s/I88WUAPPFsoAIfoCygDJgwb7AHFVUAbIywAVyx9QA88WAc8WAc8WAfoCzMntVBhUCL4=`,
			data: `te6cckEBAgEAvQAB3bIy7JbAAezbdnLVsLSq8tnVVzaHxxkKpoSYBNnn1673GXWsA+LoAAdFEGvkv2fjwjz7qzfdiXdqPR5AK5SnHJS8bwaMarvrACusxA6As841+hvM/J45oW8+8nSsukfHR4LShggiGp6o2AV0+95gAgEAkYARKckLuq22GdsiptP7thQNN+yRFCi1BGWZHcyFbaVcf4EAIlOSF3VbbDO2RU2n92woGm/ZIihRagjLMjuZCttKuP8UHrvQKAJwGsGO`,
			expected: []any{
				big.NewInt(1179211856),
				true,
				big.NewInt(1684396333),
				address.MustParseAddr("EQAezbdnLVsLSq8tnVVzaHxxkKpoSYBNnn1673GXWsA-Lu_w"),
				address.MustParseAddr("EQADoog18l-z8eEefdWb7sS7tR6PIBXKU45KXjeDRjVd9XIW"),
				address.MustParseAddr("EQCusxA6As841-hvM_J45oW8-8nSsukfHR4LShggiGp6o3hj"),
				big.NewInt(1500000000000),
				address.MustParseAddr("EQCJTkhd1W2wztkVNp_dsKBpv2SIoUWoIyzI7mQrbSrj_NSh"),
				big.NewInt(0),
				address.MustParseAddr("EQCJTkhd1W2wztkVNp_dsKBpv2SIoUWoIyzI7mQrbSrj_NSh"),
				big.NewInt(33000000000),
			},
		},
		{
			name: "get_sale_data",
			addr: addr.MustFromBase64("EQC8kT8r4ZaHEb9NKcuXW9bGBpU3ente9hoB7NKELlE3rF8r").MustToTonutils(),
			code: `te6cckECDAEAAqAAART/APSkE/S88sgLAQIBIAIDAgFIBAUAfvIw7UTQ0wDTH/pA+kD6QPoA1NMAMMABjh34AHAHyMsAFssfUATPFljPFgHPFgH6AszLAMntVOBfB4IA//7y8AICzQYHAFegOFnaiaGmAaY/9IH0gfSB9AGppgBgYaH0gfQB9IH0AGEEIIySsKAVgAKrAQP10A6GmBgLjYSS+CcH0gGHaiaGmAaY/9IH0gfSB9AGppgBgYOCmE44BgAEwthGmP6Z+lVW8Q4AHxgRDAgRXdFOAA2CnT44LYTwhWL4ZqGGhpg+oYAP2AcBRgAPloyhJrpOEBWfGBHByUYABOGxuIHCOyiiGYOHgC8BRgAMCAkKAfdmCEDuaygBSYKBSML7y4cIk0PpA+gD6QPoAMFOSoSGhUIehFqBSkHCAEMjLBVADzxYB+gLLaslx+wAlwgAl10nCArCOF1BFcIAQyMsFUAPPFgH6AstqyXH7ABAjkjQ04lpwgBDIywVQA88WAfoCy2rJcfsAcCCCEF/MPRSCwDYMTc4OYIQO5rKABi+8uHJU0bHBVFSxwUVsfLhynAgghBfzD0UIYAQyMsFKM8WIfoCy2rLHxXLPyfPFifPFhTKACP6AhPKAMmAQPsAcVBmRRUEcAfIywAWyx9QBM8WWM8WAc8WAfoCzMsAye1UAIAwMzk5U1LHBZJfCeBRUccF8uH0ghAFE42RFbry4fUE+kAwQGYFcAfIywAWyx9QBM8WWM8WAc8WAfoCzMsAye1UAC6SXwvgCMACmFVEECQQI/AF4F8KhA/y8ACWyMsfE8s/I88WUAPPFsoAggnJw4D6AsoAyXGAGMjLBSbPFnD6AstqzMmDBvsAcVVQcAfIywAWyx9QBM8WWM8WAc8WAfoCzMsAye1UbaV+xg==`,
			data: `te6cckEBAgEAugAB2zIy9FpACgRvXbOJeFSRKnEg1D+i0SqDMlaNVGvpSSKCzDQU/wDIATaCjXjFji61m1/GffT4Syla7jz4B2kJBYN4/yJSozbLADg7c9NYvuOnxzYB8tCZyF6lHDwu54T9DmJP1rezeX7VlGZyCzABAQCNgBQI3rtnEvCpIlTiQah/RaJVBmStGqjX0pJFBZhoKf4BhzEtAQAMLA26BZf4hrmn5TZ+fefRzmpnp5L0BEA8GpDueQmNkgJZ4N7+`,
			expected: []any{
				big.NewInt(1179211856),
				false,
				big.NewInt(1684400308),
				address.MustParseAddr("EQCgRvXbOJeFSRKnEg1D-i0SqDMlaNVGvpSSKCzDQU_wDAR4"),
				address.MustParseAddr("EQCbQUa8YscXWs2v4z76fCWUrXcefAO0hILBvH-RKVGbZQ3q"),
				address.MustParseAddr("EQDg7c9NYvuOnxzYB8tCZyF6lHDwu54T9DmJP1rezeX7Vpsw"),
				big.NewInt(110000000000),
				address.MustParseAddr("EQCgRvXbOJeFSRKnEg1D-i0SqDMlaNVGvpSSKCzDQU_wDAR4"),
				big.NewInt(10000000),
				address.MustParseAddr("EQAwsDboFl_iGuaflNn5959HOamenkvQEQDwakO55CY2SGB8"),
				big.NewInt(0),
			},
		},
		{
			name: "get_sale_data",
			addr: addr.MustFromBase64("EQCKhCl6uCNSSOZdBPODUvdGeLl-30XBW4IGV83JDvyg4A0Y").MustToTonutils(),
			code: `te6cckECCwEAArkAART/APSkE/S88sgLAQIBIAIDAgFIBAUAfvIw7UTQ0wDTH/pA+kD6QPoA1NMAMMABjh34AHAHyMsAFssfUATPFljPFgHPFgH6AszLAMntVOBfB4IA//7y8AICzQYHAFegOFnaiaGmAaY/9IH0gfSB9AGppgBgYaH0gfQB9IH0AGEEIIySsKAVgAKrAQH30A6GmBgLjYSS+CcH0gGHaiaGmAaY/9IH0gfSB9AGppgBgYOCmE44BgAEqYhOmPhW8Q4YBKGATpn8cIxbMbC3MbK2QV44LJOZlvKAVxFWAAyS+G8BJrpOEBFcCBFd0VYACRWdjYKdxjgthOjq+G6hhoaYPqGAD9gHAU4ADAgB92YIQO5rKAFJgoFIwvvLhwiTQ+kD6APpA+gAwU5KhIaFQh6EWoFKQcIAQyMsFUAPPFgH6AstqyXH7ACXCACXXScICsI4XUEVwgBDIywVQA88WAfoCy2rJcfsAECOSNDTiWnCAEMjLBVADzxYB+gLLaslx+wBwIIIQX8w9FIKAejy0ZSzjkIxMzk5U1LHBZJfCeBRUccF8uH0ghAFE42RFrry4fUD+kAwRlAQNFlwB8jLABbLH1AEzxZYzxYBzxYB+gLMywDJ7VTgMDcowAPjAijAAJw2NxA4R2UUQzBw8AXgCMACmFVEECQQI/AF4F8KhA/y8AkA1Dg5ghA7msoAGL7y4clTRscFUVLHBRWx8uHKcCCCEF/MPRQhgBDIywUozxYh+gLLassfFcs/J88WJ88WFMoAI/oCE8oAyYMG+wBxUGZFFQRwB8jLABbLH1AEzxZYzxYBzxYB+gLMywDJ7VQAlsjLHxPLPyPPFlADzxbKAIIJycOA+gLKAMlxgBjIywUmzxZw+gLLaszJgwb7AHFVUHAHyMsAFssfUATPFljPFgHPFgH6AszLAMntVNZeZYk=`,
			data: `te6cckEBAgEAvAAB2zIy9xlABYTuYbLf8INxFtD8tQeNk5ZLy+nAX9ahQbG/yl1qQ+GIAZ6IgjGL9UMyHT9NmBeqYGaOTDb+HM7VffFFqPh/f6/DABjRMWY6Jubyhnlx0X4hXy3aYvHBJalWB+eu/TW2OcVM1F0h26ABAQCRgBRyaww+87XrNCetowXCyAQhgF8xx74x+06XHrVig1fjCgJUC+QBAC3dNlesgVD8YbAazcauIrXBPfiVhMMr5YYk2in0Mtszwu8PdIc=`,
			expected: []any{
				big.NewInt(1179211856),
				false,
				big.NewInt(1684401714),
				address.MustParseAddr("EQBYTuYbLf8INxFtD8tQeNk5ZLy-nAX9ahQbG_yl1qQ-GEMS"),
				address.MustParseAddr("EQDPREEYxfqhmQ6fpswL1TAzRyYbfw5nar74otR8P7_X4Uad"),
				address.MustParseAddr("EQBjRMWY6Jubyhnlx0X4hXy3aYvHBJalWB-eu_TW2OcVM-IA"),
				big.NewInt(100000000000),
				address.MustParseAddr("EQCjk1hh952vWaE9bRguFkAhDAL5jj3xj9p0uPWrFBq_GEMS"),
				big.NewInt(5000000000),
				address.MustParseAddr("EQC3dNlesgVD8YbAazcauIrXBPfiVhMMr5YYk2in0Mtsz0Bz"),
				big.NewInt(0),
			},
		},
		{
			name: "get_offer_data",
			addr: addr.MustFromBase64("EQCvouWpaYLxd3BVlZfZaH2xwUNY9blfLJSPprC6yu2zuE4c").MustToTonutils(),
			code: `te6cckECFgEABEkAART/APSkE/S88sgLAQIBIAIDAgFIBAUBVvLtRNDTANMf0x/6QPpA+kD6ANTTADAwB8AB8tGU+CMlvuMCXwiCAP/+8vAVAgLMBgcAg6FGH9qJoaYBpj+mP/SB9IH0gfQBqaYAYGGh9IGmP6Y/9IGmP6Y+YKjsIeAGqO6p4AalBUIDQwQwnoyMiqQdgAKrgQIBSAgJA/fbYREWh9IGmP6Y/9IGmP6Y+YFKz4AaokIfgBqbhQkdDFsoOTezNLpBOsuBBACGRlgqgC54soAf0BCeW1ZY+A54tkuP2AEWEAEWuk4QFYSTYQ8YaQYQBIrfGGuBBBCC/mHopkZY+J5Z+TZ4soAeeLZQBBBExLQH0BZQBkuMEhMUAgEgCgsCASAQEQH3AHQ0wMBcbCSXwTg+kAw7UTQ0wDTH9Mf+kD6QPpA+gDU0wAwwAGOJTE3NzhVMxJwCMjLABfLHxXLH1ADzxYBzxYBzxYB+gLMywDJ7VTgfyrHAcAAjhowCdMfIcAAi2Y2FuY2VshSIMcFsJJzMt5Qqt4ggQIruinAAbBTpoAwAEyCEDuaygABqYSAC0McFsJ4QrF8M1DDQ0wfUMAH7AOCCEAUTjZFSELrjAjwnwAHy0ZQrwABTk8cFsI4rODg5UHagEDdGUEQDAnAIyMsAF8sfFcsfUAPPFgHPFgHPFgH6AszLAMntVOA3OQnAA+MCXwmED/LwDQ4AxjAJ0z/6QDBTlMcFCcAAGbArghA7msoAvrCeOBBaEEkQOEcVA0Rk8AiOODlfBjMzcCCCEF/MPRTIyx8Tyz8jzxZQA88WygAh+gLKAMlxgBjIywVQA88WcPoCEstqzMmAQPsA4gGsU1jHBVNixwWx8uHKgggPQkBw+wJRUccFjhQ1cIAQyMsFKM8WIfoCy2rJgwb7AOMNcUcXUGYFBANwCMjLABfLHxXLH1ADzxYBzxYBzxYB+gLMywDJ7VQPALYF+gAhghAdzWUAvJeCEB3NZQAy3o0EE9mZmVyIGNhbmNlbCBmZWWBURzNwIIAQyMsFUAXPFlAD+gITy2rLHwHPFslx+wDUMHGAEMjLBSnPFnD6AstqzMmDBvsAABEghA7msoAqYSAAHQgwACTXwNw4FnwAgHwAYABYi+T2ZmZXIgcm95YWxpZXOBNwIIAQyMsFUAXPFlAD+gITy2rLHwHPFslx+wAATIuU9mZmVyIGZlZYcCCAEMjLBVAFzxZQA/oCE8tqyx8BzxbJcfsAAHiAGMjLBSbPFnD6AstqzIIID0JAcPsCyYMG+wBxVWBwCMjLABfLHxXLH1ADzxYBzxYBzxYB+gLMywDJ7VQAzgfTH4EPowLDABLy8oEPpCHXSsMA8vKBD6Uh10mBAfS88vL4AIIID0JAcPsCcCCAEMjLBSTPFiH6Astqyx8BzxbJgwb7AHEHVQVwCMjLABfLHxXLH1ADzxYBzxYBzxYB+gLMywDJ7VRFze+v`,
			data: `te6cckEBAgEAyQAB4bIuXFSyMvmUwAWE7mGy3/CDcRbQ/LUHjZOWS8vpwF/WoUGxv8pdakPhiAHMXjRhWPFwvnSLPMmcCSNoVRn5DNXikXGrB+TEmvo5XwA4O3PTWL7jp8c2AfLQmchepRw8LueE/Q5iT9a3s3l+1ZLLQXgBAQClgBRyaww+87XrNCetowXCyAQhgF8xx74x+06XHrVig1fjAAAAAKAAAAyQAmZLXdvN/Q6HNm6q9CthOC37te3pwMe9BOWp3Qgd1gloAAAACAAAAZKD0yeg`,
			expected: []any{
				big.NewInt(340481426770),
				true,
				big.NewInt(1683798185),
				big.NewInt(1684402985),
				address.MustParseAddr("EQBYTuYbLf8INxFtD8tQeNk5ZLy-nAX9ahQbG_yl1qQ-GEMS"),
				address.MustParseAddr("EQDmLxowrHi4XzpFnmTOBJG0Koz8hmrxSLjVg_JiTX0cr_Ba"),
				address.MustParseAddr("EQDg7c9NYvuOnxzYB8tCZyF6lHDwu54T9DmJP1rezeX7Vpsw"),
				big.NewInt(3000000000),
				address.MustParseAddr("EQCjk1hh952vWaE9bRguFkAhDAL5jj3xj9p0uPWrFBq_GEMS"),
				big.NewInt(5),
				big.NewInt(100),
				address.MustParseAddr("EQCZktd2839Doc2bqr0K2E4Lfu17enAx70E5andCB3WCWmjB"),
				big.NewInt(2),
				big.NewInt(100),
				big.NewInt(2790000000),
			},
		},
	}

	for _, test := range testCases {
		ret := execGetMethod(t, getInterfaceByCode(getCodeHash(t, test.code)), test.addr, test.name, test.code, test.data)
		if !assert.Equal(t, test.expected, ret) {
			for _, r := range ret {
				t.Logf("%v", r)
			}
			t.Logf("")
		}
	}
}

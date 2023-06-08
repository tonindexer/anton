package known_test

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
)

type (
	JettonMint struct {
		_         tlb.Magic        `tlb:"#00000015"`
		QueryID   uint64           `tlb:"## 64"`
		ToAddress *address.Address `tlb:"addr"`
		Amount    tlb.Coins        `tlb:"."`
		MasterMsg struct {
			OpCode       uint32    `tlb:"## 32"`
			QueryID      uint64    `tlb:"## 64"`
			JettonAmount tlb.Coins `tlb:"."`
		} `tlb:"^"`
	}
	JettonBurnNotification struct {
		_               tlb.Magic        `tlb:"#7bdd97de"`
		QueryID         uint64           `tlb:"## 64"`
		JettonAmount    tlb.Coins        `tlb:"."`
		FromAddress     *address.Address `tlb:"addr"`
		ResponseAddress *address.Address `tlb:"addr"`
	}
	JettonChangeAdmin struct {
		_               tlb.Magic        `tlb:"#00000003"`
		QueryID         uint64           `tlb:"## 64"`
		NewAdminAddress *address.Address `tlb:"addr"`
	}
	JettonChangeContent struct {
		_       tlb.Magic  `tlb:"#00000004"`
		QueryID uint64     `tlb:"## 64"`
		Content *cell.Cell `tlb:"^"`
	}
)

func TestNewOperationDesc_JettonMinter(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
		{
			structType: (*JettonMint)(nil),
			expected:   `{"op_name":"jetton_mint","op_code":"0x15","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"to_address","tlb_type":"addr","format":"addr"},{"name":"amount","tlb_type":".","format":"coins"},{"name":"master_msg","tlb_type":"^","format":"struct","struct_fields":[{"name":"op_code","tlb_type":"## 32","format":"uint32"},{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"jetton_amount","tlb_type":".","format":"coins"}]}]}`,
		}, {
			structType: (*JettonBurnNotification)(nil),
			expected:   `{"op_name":"jetton_burn_notification","op_code":"0x7bdd97de","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"jetton_amount","tlb_type":".","format":"coins"},{"name":"from_address","tlb_type":"addr","format":"addr"},{"name":"response_address","tlb_type":"addr","format":"addr"}]}`,
		}, {
			structType: (*JettonChangeAdmin)(nil),
			expected:   `{"op_name":"jetton_change_admin","op_code":"0x3","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"new_admin_address","tlb_type":"addr","format":"addr"}]}`,
		}, {
			structType: (*JettonChangeContent)(nil),
			expected:   `{"op_name":"jetton_change_content","op_code":"0x4","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"content","tlb_type":"^","format":"cell"}]}`,
		},
	}

	for _, test := range testCases {
		d, err := abi.NewOperationDesc(test.structType)
		require.Nil(t, err)
		got, err := json.Marshal(d)
		require.Nil(t, err)
		assert.Equal(t, test.expected, string(got))
	}
}

type (
	JettonTransfer struct {
		_                   tlb.Magic        `tlb:"#0f8a7ea5"`
		QueryID             uint64           `tlb:"## 64"`
		Amount              tlb.Coins        `tlb:"."`
		Destination         *address.Address `tlb:"addr"`
		ResponseDestination *address.Address `tlb:"addr"`
		CustomPayload       *cell.Cell       `tlb:"maybe ^"`
		ForwardTONAmount    tlb.Coins        `tlb:"."`
		ForwardPayload      *cell.Cell       `tlb:"either . ^"`
	}
	JettonInternalTransfer struct {
		_                tlb.Magic        `tlb:"#178d4519"`
		QueryID          uint64           `tlb:"## 64"`
		Amount           tlb.Coins        `tlb:"."`
		From             *address.Address `tlb:"addr"`
		ResponseAddress  *address.Address `tlb:"addr"`
		ForwardTONAmount tlb.Coins        `tlb:"."`
		ForwardPayload   *cell.Cell       `tlb:"either . ^"`
	}
	JettonTransferNotification struct {
		_              tlb.Magic        `tlb:"#7362d09c"`
		QueryID        uint64           `tlb:"## 64"`
		Amount         tlb.Coins        `tlb:"."`
		Sender         *address.Address `tlb:"addr"`
		ForwardPayload *cell.Cell       `tlb:"either . ^"`
	}
	JettonBurn struct {
		_                   tlb.Magic        `tlb:"#595f07bc"`
		QueryID             uint64           `tlb:"## 64"`
		Amount              tlb.Coins        `tlb:"."`
		ResponseDestination *address.Address `tlb:"addr"`
		CustomPayload       *cell.Cell       `tlb:"maybe ^"`
	}
)

func TestNewOperationDesc_JettonWallet(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
		{
			structType: (*JettonTransfer)(nil),
			expected:   `{"op_name":"jetton_transfer","op_code":"0xf8a7ea5","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"amount","tlb_type":".","format":"coins"},{"name":"destination","tlb_type":"addr","format":"addr"},{"name":"response_destination","tlb_type":"addr","format":"addr"},{"name":"custom_payload","tlb_type":"maybe ^","format":"cell"},{"name":"forward_ton_amount","tlb_type":".","format":"coins"},{"name":"forward_payload","tlb_type":"either . ^","format":"cell"}]}`,
		}, {
			structType: (*JettonInternalTransfer)(nil),
			expected:   `{"op_name":"jetton_internal_transfer","op_code":"0x178d4519","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"amount","tlb_type":".","format":"coins"},{"name":"from","tlb_type":"addr","format":"addr"},{"name":"response_address","tlb_type":"addr","format":"addr"},{"name":"forward_ton_amount","tlb_type":".","format":"coins"},{"name":"forward_payload","tlb_type":"either . ^","format":"cell"}]}`,
		}, {
			structType: (*JettonTransferNotification)(nil),
			expected:   `{"op_name":"jetton_transfer_notification","op_code":"0x7362d09c","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"amount","tlb_type":".","format":"coins"},{"name":"sender","tlb_type":"addr","format":"addr"},{"name":"forward_payload","tlb_type":"either . ^","format":"cell"}]}`,
		}, {
			structType: (*JettonBurn)(nil),
			expected:   `{"op_name":"jetton_burn","op_code":"0x595f07bc","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"amount","tlb_type":".","format":"coins"},{"name":"response_destination","tlb_type":"addr","format":"addr"},{"name":"custom_payload","tlb_type":"maybe ^","format":"cell"}]}`,
		},
	}

	for _, test := range testCases {
		d, err := abi.NewOperationDesc(test.structType)
		require.Nil(t, err)
		got, err := json.Marshal(d)
		require.Nil(t, err)
		assert.Equal(t, test.expected, string(got))
	}
}

func TestOperationDesc_JettonMinter(t *testing.T) {
	var (
		interfaces []*abi.InterfaceDesc
		i          *abi.InterfaceDesc
	)

	j, err := os.ReadFile("tep74_jetton.json")
	require.Nil(t, err)

	err = json.Unmarshal(j, &interfaces)
	require.Nil(t, err)

	for _, i = range interfaces {
		if i.Name == "jetton_minter" {
			err := i.RegisterDefinitions()
			require.Nil(t, err)
			break
		}
	}

	var testCases = []*struct {
		name     string
		boc      string
		expected string
	}{
		{
			// tx hash e5782dd2b1e2186038c1f92db2cdb709bd12eba25a295ec4db9561aa3928c317
			name:     "jetton_mint",
			boc:      `te6cckECBgEAAY4AAWMAAAAVpRNS/gQ80YGAFSIq9XSS6um704WS4suGgdULW5b13fha77/GRjN77mgoZVPxAQEBbReNRRmlE1L+BDzRgUHc1lACADvjYKkD7+YYy/VvGEAr2NDd0ROzABirjBhcbeEorNtYoX14QAYCAZdJKpgbgBB56R97rZGAXZwg4u1afeJ934d0DPUZtObWsRZvvxY00AHfGwVIH38wxl+reMIBXsaG7oidmADFXGDC428JRWbaxQF9eEAgAwJf0zuweeADvjYKkD7+YYy/VvGEAr2NDd0ROzABirjBhcbeEorNtYqCVrO8SiAvrwgEBQQAl6iXCtCADvjYKkD7+YYy/VvGEAr2NDd0ROzABirjBhcbeEorNtYwAd8bBUgffzDGX6t4wgFexobuiJ2YAMVcYMLjbwlFZtrFAJiWgCAAl+kWu++ADvjYKkD7+YYy/VvGEAr2NDd0ROzABirjBhcbeEorNtYwAd8bBUgffzDGX6t4wgFexobuiJ2YAMVcYMLjbwlFZtrFAJiWgCAnWbE8`,
			expected: `{"query_id":11894942291761877377,"to_address":"EQCpEVerpJdXTd6cLJcWXDQOqFrct67vwtd9_jIxm99zQZV6","amount":"850000000","master_msg":{"op_code":395134233,"query_id":11894942291761877377,"jetton_amount":"500000000"}}`,
		},
	}

	for _, test := range testCases {
		j := loadOperation(t, i, test.name, test.boc)
		assert.Equal(t, test.expected, j)
	}
}

func TestOperationDesc_JettonWallet_JettonBurn_Optional(t *testing.T) {
	// issue #13
	// tx hash 8ad6febf40eed096d8d911466c069ebc693e4d5e641a4e806a19d5831233141b

	var (
		interfaces []*abi.InterfaceDesc
		i          *abi.InterfaceDesc
	)

	j, err := os.ReadFile("tep74_jetton.json")
	require.Nil(t, err)

	err = json.Unmarshal(j, &interfaces)
	require.Nil(t, err)

	for _, i = range interfaces {
		if i.Name == "jetton_wallet" {
			err := i.RegisterDefinitions()
			require.Nil(t, err)
			break
		}
	}

	var testCases = []*struct {
		name     string
		boc      string
		expected string
	}{
		{
			// tx hash e5782dd2b1e2186038c1f92db2cdb709bd12eba25a295ec4db9561aa3928c317
			name:     "jetton_burn",
			boc:      `te6ccuEBAQEAMwBmAGFZXwe8AAAAAACFYI8walQ4AJvi30153Ex53ULaIU/S0hqruxuRQNfFygS/4vFtl92PwofAGw==`, // out msg
			expected: `{"query_id":8741007,"amount":"435523","response_destination":"EQBN8W-mvO4mPO6hbRCn6WkNVd2NyKBr4uUCX_F4tsvux5oO"}`,
		},
	}

	for _, test := range testCases {
		dp := getOperationDescByName(i, test.name)
		require.NotNilf(t, dp, "operation name %s", test.name)

		boc, err := base64.StdEncoding.DecodeString(test.boc)
		require.Nil(t, err)

		c, err := cell.FromBOC(boc)
		require.Nil(t, err)

		op, err := dp.New()
		require.Nil(t, err)

		err = tlb.LoadFromCell(op, c.BeginParse())
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "not enough data in reader"))

		op, err = dp.New(true)
		require.Nil(t, err)

		err = tlb.LoadFromCell(op, c.BeginParse())
		require.Nil(t, err)

		j, err := json.Marshal(op)
		require.Nil(t, err)

		assert.Equal(t, test.expected, string(j))
	}
}

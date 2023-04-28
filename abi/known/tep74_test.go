package known

import (
	"encoding/json"
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
		assert.Equal(t, test.expected, string(got))
	}
}

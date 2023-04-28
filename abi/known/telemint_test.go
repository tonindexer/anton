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
	TeleitemAuctionConfig struct {
		BeneficiaryAddress *address.Address `tlb:"addr"`
		InitialMinBid      tlb.Coins        `tlb:"."`
		MaxBid             tlb.Coins        `tlb:"."`
		MinBidStep         uint8            `tlb:"## 8"`
		MinExtendTime      uint32           `tlb:"## 32"`
		Duration           uint32           `tlb:"## 32"`
	}
	TelemintRoyaltyParams struct {
		Numerator   uint16           `tlb:"## 16"`
		Denominator uint16           `tlb:"## 16"`
		Destination *address.Address `tlb:"addr"`
	}
	TelemintTokenInfo struct {
		Name   *abi.TelemintText `tlb:"."`
		Domain *abi.TelemintText `tlb:"."`
	}
)

type (
	TelemintMsgDeploy struct {
		_             tlb.Magic              `tlb:"#4637289a"`
		Sig           []byte                 `tlb:"bits 512"`
		SubwalletID   uint32                 `tlb:"## 32"`
		ValidSince    uint32                 `tlb:"## 32"`
		ValidTill     uint32                 `tlb:"## 32"`
		TokenName     *abi.TelemintText      `tlb:"."`
		Content       *cell.Cell             `tlb:"^"`
		AuctionConfig *TeleitemAuctionConfig `tlb:"^"`
		RoyaltyParams *TelemintRoyaltyParams `tlb:"maybe ^"`
	}
	TelemintMsgDeployV2 struct {
		_             tlb.Magic              `tlb:"#4637289b"`
		Sig           []byte                 `tlb:"bits 512"`
		SubwalletID   uint32                 `tlb:"## 32"`
		ValidSince    uint32                 `tlb:"## 32"`
		ValidTill     uint32                 `tlb:"## 32"`
		TokenName     *abi.TelemintText      `tlb:"."`
		Content       *cell.Cell             `tlb:"^"`
		AuctionConfig *TeleitemAuctionConfig `tlb:"^"`
		RoyaltyParams *TelemintRoyaltyParams `tlb:"maybe ^"`
		// Restrictions
	}
	TeleitemMsgDeploy struct {
		_             tlb.Magic              `tlb:"#299a3e15"`
		SenderAddress *address.Address       `tlb:"addr"`
		Bid           tlb.Coins              `tlb:"."`
		Info          *TelemintTokenInfo     `tlb:"^"`
		Content       *cell.Cell             `tlb:"^"`
		AuctionConfig *TeleitemAuctionConfig `tlb:"^"`
		RoyaltyParams *TelemintRoyaltyParams `tlb:"^"`
	}
)

func TestNewOperationDesc_TelemintNFTCollection(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
		{
			structType: (*TelemintMsgDeploy)(nil),
			expected:   `{"op_name":"telemint_msg_deploy","op_code":"0x4637289a","body":[{"name":"sig","tlb_type":"bits 512","format":"bytes"},{"name":"subwallet_id","tlb_type":"## 32","format":"uint32"},{"name":"valid_since","tlb_type":"## 32","format":"uint32"},{"name":"valid_till","tlb_type":"## 32","format":"uint32"},{"name":"token_name","tlb_type":".","format":"telemintText"},{"name":"content","tlb_type":"^","format":"cell"},{"name":"auction_config","tlb_type":"^","format":"struct","struct_fields":[{"name":"beneficiary_address","tlb_type":"addr","format":"addr"},{"name":"initial_min_bid","tlb_type":".","format":"coins"},{"name":"max_bid","tlb_type":".","format":"coins"},{"name":"min_bid_step","tlb_type":"## 8","format":"uint8"},{"name":"min_extend_time","tlb_type":"## 32","format":"uint32"},{"name":"duration","tlb_type":"## 32","format":"uint32"}]},{"name":"royalty_params","tlb_type":"maybe ^","format":"struct","struct_fields":[{"name":"numerator","tlb_type":"## 16","format":"uint16"},{"name":"denominator","tlb_type":"## 16","format":"uint16"},{"name":"destination","tlb_type":"addr","format":"addr"}]}]}`,
		}, {
			structType: (*TelemintMsgDeployV2)(nil),
			expected:   `{"op_name":"telemint_msg_deploy_v_2","op_code":"0x4637289b","body":[{"name":"sig","tlb_type":"bits 512","format":"bytes"},{"name":"subwallet_id","tlb_type":"## 32","format":"uint32"},{"name":"valid_since","tlb_type":"## 32","format":"uint32"},{"name":"valid_till","tlb_type":"## 32","format":"uint32"},{"name":"token_name","tlb_type":".","format":"telemintText"},{"name":"content","tlb_type":"^","format":"cell"},{"name":"auction_config","tlb_type":"^","format":"struct","struct_fields":[{"name":"beneficiary_address","tlb_type":"addr","format":"addr"},{"name":"initial_min_bid","tlb_type":".","format":"coins"},{"name":"max_bid","tlb_type":".","format":"coins"},{"name":"min_bid_step","tlb_type":"## 8","format":"uint8"},{"name":"min_extend_time","tlb_type":"## 32","format":"uint32"},{"name":"duration","tlb_type":"## 32","format":"uint32"}]},{"name":"royalty_params","tlb_type":"maybe ^","format":"struct","struct_fields":[{"name":"numerator","tlb_type":"## 16","format":"uint16"},{"name":"denominator","tlb_type":"## 16","format":"uint16"},{"name":"destination","tlb_type":"addr","format":"addr"}]}]}`,
		}, {
			structType: (*TeleitemMsgDeploy)(nil),
			expected:   `{"op_name":"teleitem_msg_deploy","op_code":"0x299a3e15","body":[{"name":"sender_address","tlb_type":"addr","format":"addr"},{"name":"bid","tlb_type":".","format":"coins"},{"name":"info","tlb_type":"^","format":"struct","struct_fields":[{"name":"name","tlb_type":".","format":"telemintText"},{"name":"domain","tlb_type":".","format":"telemintText"}]},{"name":"content","tlb_type":"^","format":"cell"},{"name":"auction_config","tlb_type":"^","format":"struct","struct_fields":[{"name":"beneficiary_address","tlb_type":"addr","format":"addr"},{"name":"initial_min_bid","tlb_type":".","format":"coins"},{"name":"max_bid","tlb_type":".","format":"coins"},{"name":"min_bid_step","tlb_type":"## 8","format":"uint8"},{"name":"min_extend_time","tlb_type":"## 32","format":"uint32"},{"name":"duration","tlb_type":"## 32","format":"uint32"}]},{"name":"royalty_params","tlb_type":"^","format":"struct","struct_fields":[{"name":"numerator","tlb_type":"## 16","format":"uint16"},{"name":"denominator","tlb_type":"## 16","format":"uint16"},{"name":"destination","tlb_type":"addr","format":"addr"}]}]}`,
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
	TeleitemStartAuction struct {
		_             tlb.Magic              `tlb:"#487a8e81"`
		QueryID       uint64                 `tlb:"## 64"`
		AuctionConfig *TeleitemAuctionConfig `tlb:"^"`
	}
	TeleitemCancelAuction struct {
		_       tlb.Magic `tlb:"#371638ae"`
		QueryID uint64    `tlb:"## 64"`
	}

	TeleitemOK struct {
		_       tlb.Magic `tlb:"#a37a0983"`
		QueryID uint64    `tlb:"## 64"`
	}
	TeleitemOutbidNotification struct {
		_ tlb.Magic `tlb:"#557cea20"`
	}
)

func TestNewOperationDesc_TelemintNFTItem(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
		{
			structType: (*TeleitemStartAuction)(nil),
			expected:   `{"op_name":"teleitem_start_auction","op_code":"0x487a8e81","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"auction_config","tlb_type":"^","format":"struct","struct_fields":[{"name":"beneficiary_address","tlb_type":"addr","format":"addr"},{"name":"initial_min_bid","tlb_type":".","format":"coins"},{"name":"max_bid","tlb_type":".","format":"coins"},{"name":"min_bid_step","tlb_type":"## 8","format":"uint8"},{"name":"min_extend_time","tlb_type":"## 32","format":"uint32"},{"name":"duration","tlb_type":"## 32","format":"uint32"}]}]}`,
		}, {
			structType: (*TeleitemCancelAuction)(nil),
			expected:   `{"op_name":"teleitem_cancel_auction","op_code":"0x371638ae","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"}]}`,
		}, {
			structType: (*TeleitemOK)(nil),
			expected:   `{"op_name":"teleitem_ok","op_code":"0xa37a0983","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"}]}`,
		}, {
			structType: (*TeleitemOutbidNotification)(nil),
			expected:   `{"op_name":"teleitem_outbid_notification","op_code":"0x557cea20","body":null}`,
		},
	}

	for _, test := range testCases {
		d, err := abi.NewOperationDesc(test.structType)
		require.Nil(t, err)
		got, err := json.Marshal(d)
		assert.Equal(t, test.expected, string(got))
	}
}

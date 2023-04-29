package known_test

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
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

func TestOperationDesc_TelemintNFTCollection(t *testing.T) {
	var (
		interfaces []*abi.InterfaceDesc
		i          *abi.InterfaceDesc
	)

	j, err := os.ReadFile("telemint.json")
	require.Nil(t, err)

	err = json.Unmarshal(j, &interfaces)
	require.Nil(t, err)

	for _, i = range interfaces {
		if i.Name == "telemint_nft_collection" {
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
			// tx hash 0ff74e54e6bff335658c3dac65045632ca11d840b99d1650788ceefa7c37b616
			name:     "telemint_msg_deploy",
			boc:      `te6cckEBBAEA6gADsUY3KJphrCuza/Lb313RaBRyRsTRVoTNp6ceHrgVNS0nFgxrIl1BQpKB8HyFHCXrHrM4ma3Cq8fpX1ls4eXsrBHk3OkIAAAAA2PDFEtjwxThB25vbWVyNTjAAwIBAEsABQBkgAgRtHZRbYDLSyfCByJNN1VTi/GV14GNOz4VV3VlUyStcABhgAwcNTYmtuVtY9ClO4/AGIimMJuro731Y7+KIrFPpsxTqgSoF8gAAKAAAcIAASdQEABeAWh0dHBzOi8vbmZ0LmZyYWdtZW50LmNvbS91c2VybmFtZS9ub21lcjU4Lmpzb26nqG30`,
			expected: `{"sig":"Yawrs2vy299d0WgUckbE0VaEzaenHh64FTUtJxYMayJdQUKSgfB8hRwl6x6zOJmtwqvH6V9ZbOHl7KwR5NzpCA==","subwallet_id":3,"valid_since":1673729099,"valid_till":1673729249,"token_name":{"Len":7,"Text":"nomer58"},"content":{},"auction_config":{"beneficiary_address":"EQBg4amxNbcrax6FKdx-AMRFMYTdXR3vqx38URWKfTZinejc","initial_min_bid":"10000000000","max_bid":"0","min_bid_step":5,"min_extend_time":3600,"duration":604800},"royalty_params":{"numerator":5,"denominator":100,"destination":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE"}}`,
		}, {
			// tx hash 726b637d8127068310350fcacfe45918a641b163759074e2ceb23ac4a901cb1d
			name:     "telemint_msg_deploy_v2",
			boc:      `te6cckEBBAEA8AADuUY3KJscZHOp8v4ACiYXmiUzxlVQuojEbXs1wST89fxdPptM8bTC/X7cQ+lPEl6mbhLnHetMET9ptfXXk5PGoxdv7cEHAAAADGOQcQhjkHGeCzg4ODA5ODYwNjA2oAMCAQBLAAUAZIAIEbR2UW2Ay0snwgciTTdVU4vxldeBjTs+FVd1ZVMkrXAAYYAIEbR2UW2Ay0snwgciTTdVU4vxldeBjTs+FVd1ZVMkrWoJUC+QAACgAAHCAAEnUBAAYgFodHRwczovL25mdC5mcmFnbWVudC5jb20vbnVtYmVyLzg4ODA5ODYwNjA2Lmpzb27chHEq`,
			expected: `{"sig":"HGRzqfL+AAomF5olM8ZVULqIxG17NcEk/PX8XT6bTPG0wv1+3EPpTxJepm4S5x3rTBE/abX115OTxqMXb+3BBw==","subwallet_id":12,"valid_since":1670410504,"valid_till":1670410654,"token_name":{"Len":11,"Text":"88809860606"},"content":{},"auction_config":{"beneficiary_address":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE","initial_min_bid":"20000000000","max_bid":"0","min_bid_step":5,"min_extend_time":3600,"duration":604800},"royalty_params":{"numerator":5,"denominator":100,"destination":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE"}}`,
		}, {
			// tx hash 0b73df250c943a08bee93eff7edd3e10d85b4d3700f1858adf8547cc0f4dba4d
			name:     "teleitem_msg_deploy",
			boc:      `te6cckEBBQEA0wAEVSmaPhWAAKkpQssPUR7gVnU4J4Usgy79aEsghesXqNV3ZiG0a3zqBKgXyAEDAgEEAGGACcFWZ9FV+Xo7WsatuSn9hb+IIgkeSBJao/47XuB+LiIKBKgXyAAAoAABwgABJ1AQAGQBaHR0cHM6Ly9uZnQuZnJhZ21lbnQuY29tL3VzZXJuYW1lL29ubHlfc2FsZXMuanNvbgAiCm9ubHlfc2FsZXMFbWUAdAAASwAFAGSACBG0dlFtgMtLJ8IHIk03VVOL8ZXXgY07PhVXdWVTJK1wCtjbZA==`,
			expected: `{"sender_address":"EQAFSUoWWHqI9wKzqcE8KWQZd-tCWQQvWL1Gq7sxDaNb51vm","bid":"10000000000","info":{"name":{"Len":10,"Text":"only_sales"},"domain":{"Len":5,"Text":"me\u0000t\u0000"}},"content":{},"auction_config":{"beneficiary_address":"EQBOCrM-iq_L0drWNW3JT-wt_EEQSPJAktUf8dr3A_FxEFt1","initial_min_bid":"10000000000","max_bid":"0","min_bid_step":5,"min_extend_time":3600,"duration":604800},"royalty_params":{"numerator":5,"denominator":100,"destination":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE"}}`,
		},
	}

	for _, test := range testCases {
		j := loadOperation(t, i, test.name, test.boc)
		assert.Equal(t, test.expected, j)
	}
}

func TestGetMethodDesc_TelemintNFTItem(t *testing.T) {
	var (
		interfaces []*abi.InterfaceDesc
		i          *abi.InterfaceDesc
	)

	j, err := os.ReadFile("telemint.json")
	require.Nil(t, err)

	err = json.Unmarshal(j, &interfaces)
	require.Nil(t, err)

	for _, i = range interfaces {
		if i.Name == "telemint_nft_item" {
			err := i.RegisterDefinitions()
			require.Nil(t, err)
			break
		}
	}

	var testCases = []*struct {
		name     string
		addr     *address.Address
		code     string
		data     string
		expected []any
	}{
		{
			name: "get_telemint_auction_config",
			addr: addr.MustFromBase64("EQC2yZuyN-Aph6upzg8hyZB6PvpzMlCOZd4YgkrM2aWBqVDq").MustToTU(),
			code: `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`,
			data: `te6cckECCwEAAUAAAgHAAQIAg+LkF1bB5GhSvfy7vspAVh0PCJ9/NfLh4OCKCn1cp4mkgBAa8Ua/KrQpbPP1UQ/+mat/geh2lJ3UVPzST+bgNk54UAMBMAMEBQIBQAYHAgAICQBLAAUAZIAIEbR2UW2Ay0snwgciTTdVU4vxldeBjTs+FVd1ZVMkrXAAbgFodHRwczovL25mdC5mcmFnbWVudC5jb20vdXNlcm5hbWUvY3J5cHRvY29tbXVuaXR5Lmpzb24ALA9jcnlwdG9jb21tdW5pdHkFbWUAdAABFbAvXdHmYAMikBoECgBhgBbfihY3txYCaIx7CkZDmR9iJdtkWjMBj/V+YJg42dJd69GpSiAAAKAAAcIAASdQEABXgBQS0HLD6O9gcaNhqgvbubqPCuATvbFKENj1v2auCHcCzAtHGZhgAMiWsnEd9Shp`,
			expected: []any{
				addr.MustFromBase64("EQC2_FCxvbiwE0Rj2FIyHMj7ES7bItGYDH-r8wTBxs6S78qs").MustToTU(),
				big.NewInt(1000000000000),
				big.NewInt(0),
				big.NewInt(5),
				big.NewInt(3600),
				big.NewInt(604800),
			},
		}, {
			name: "get_telemint_auction_config",
			addr: addr.MustFromBase64("EQDOZIib-2DZPCKPir1tT5KtOYWzwoDGM404m9NxXeKVEDpC").MustToTU(),
			code: `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`,
			data: `te6cckEBBwEA4QACAcABAgCDbtUkW4oqei5fQe23Q1RxuuJVw/B+29hX3ypBCY13dHuAEBrxRr8qtCls8/VRD/6Zq3+B6HaUndRU/NJP5uA2TnhQAkOAE4XScCtkdWJ/MX6FslZwUyrjfgkkRGJ1Gc2116e8KORIAwQCAUAFBgBLAAUAZIAIEbR2UW2Ay0snwgciTTdVU4vxldeBjTs+FVd1ZVMkrXAAYgFodHRwczovL25mdC5mcmFnbWVudC5jb20vdXNlcm5hbWUvZGF0Ym9pNDIwLmpzb24AIAlkYXRib2k0MjAFbWUAdAAfhCkV`,
			expected: []any{
				(*cell.Slice)(nil),
				big.NewInt(0),
				big.NewInt(0),
				big.NewInt(0),
				big.NewInt(0),
				big.NewInt(0),
			},
		},
	}

	for _, test := range testCases {
		ret := execGetMethod(t, i, test.addr, test.name, test.code, test.data)
		assert.Equal(t, test.expected, ret)
	}
}

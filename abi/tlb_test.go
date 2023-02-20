package abi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func mustBase64(t *testing.T, str string) []byte {
	ret, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		t.Fatal(err)
	}
	return ret
}

func TestMarshalSchema(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
		// nft collection
		{
			structType: (*NFTCollectionItemMint)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#00000001\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"Index","type":"bigInt","tag":"tlb:\"## 64\""},{"name":"TonAmount","type":"coins","tag":"tlb:\".\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""}]`,
		}, {
			structType: (*NFTCollectionChangeOwner)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#00000003\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"NewOwner","type":"address","tag":"tlb:\"addr\""}]`,
		},
		// nft item
		{
			structType: (*NFTItemTransfer)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#5fcc3d14\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"NewOwner","type":"address","tag":"tlb:\"addr\""},{"name":"ResponseDestination","type":"address","tag":"tlb:\"addr\""},{"name":"CustomPayload","type":"cell","tag":"tlb:\"maybe ^\""},{"name":"ForwardAmount","type":"coins","tag":"tlb:\".\""},{"name":"ForwardPayload","type":"cell","tag":"tlb:\"either . ^\""}]`,
		}, {
			structType: (*NFTItemEdit)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#1a0b9d51\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""}]`,
		},
		// jetton minter
		{
			structType: (*JettonMint)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#00000001\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"Index","type":"uint64","tag":"tlb:\"## 64\""},{"name":"TonAmount","type":"coins","tag":"tlb:\".\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""}]`,
		},
		// jetton wallet
		{
			structType: (*JettonTransfer)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#0f8a7ea5\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"Amount","type":"coins","tag":"tlb:\".\""},{"name":"Destination","type":"address","tag":"tlb:\"addr\""},{"name":"ResponseDestination","type":"address","tag":"tlb:\"addr\""},{"name":"CustomPayload","type":"cell","tag":"tlb:\"maybe ^\""},{"name":"ForwardTONAmount","type":"coins","tag":"tlb:\".\""},{"name":"ForwardPayload","type":"cell","tag":"tlb:\"either . ^\""}]`,
		}, {
			structType: (*JettonBurn)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#595f07bc\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"Amount","type":"coins","tag":"tlb:\".\""},{"name":"ResponseDestination","type":"address","tag":"tlb:\"addr\""},{"name":"CustomPayload","type":"cell","tag":"tlb:\"maybe ^\""}]`,
		},
		// telemint
		{
			structType: (*TelemintMsgDeploy)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#4637289b\""},{"name":"Sig","type":"bytes","tag":"tlb:\"bits 512\""},{"name":"SubwalletID","type":"uint32","tag":"tlb:\"## 32\""},{"name":"ValidSince","type":"uint32","tag":"tlb:\"## 32\""},{"name":"ValidTill","type":"uint32","tag":"tlb:\"## 32\""},{"name":"TokenName","type":"telemintText","tag":"tlb:\".\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""},{"name":"AuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]},{"name":"RoyaltyParams","type":"struct","tag":"tlb:\"maybe ^\"","struct_fields":[{"name":"Numerator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Denominator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Destination","type":"address","tag":"tlb:\"addr\""}]}]`,
		}, {
			structType: (*TeleitemMsgDeploy)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#299a3e15\""},{"name":"SenderAddress","type":"address","tag":"tlb:\"addr\""},{"name":"Bid","type":"coins","tag":"tlb:\".\""},{"name":"Info","type":"cell","tag":"tlb:\"^\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""},{"name":"AuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]},{"name":"RoyaltyParams","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"Numerator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Denominator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Destination","type":"address","tag":"tlb:\"addr\""}]}]`,
		}, {
			structType: (*TeleitemStartAuction)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#487a8e81\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"AuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]}]`,
		}, {
			structType: (*TeleitemCancelAuction)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#371638ae\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""}]`,
		},
	}

	for i, test := range testCases {
		raw, err := MarshalSchema(test.structType)
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != test.expected {
			t.Errorf("[%d] (%s)\nexpected: %s\n     got: %s", i, reflect.TypeOf(test.structType), test.expected, string(raw))
		}

		got, err := UnmarshalSchema(raw)
		if err != nil {
			t.Fatal(err)
		}

		gotRaw, err := MarshalSchema(got)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(raw, gotRaw) {
			t.Errorf("[%d] (%s)\nexpected: %s\n     got: %s", i, reflect.TypeOf(test.structType), raw, gotRaw)
		}
	}
}

func TestUnmarshalSchema(t *testing.T) {
	var testCases = []*struct {
		schema     []byte
		payloadBOC []byte
		expected   []byte
	}{
		{
			// nft item transfer
			// https://ton.cx/tx/35447977000003:JH9pr5my6TDD5q4YivuMTQaNmXWAAyfxEb04iCkdz84=:EQAiZupbLhdE7UWQgnTirCbIJRg6yxfmkvTDjxsFh33Cu5rM
			schema:     []byte(`[{"name":"Op","type":"magic","tag":"tlb:\"#5fcc3d14\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"NewOwner","type":"address","tag":"tlb:\"addr\""},{"name":"ResponseDestination","type":"address","tag":"tlb:\"addr\""},{"name":"CustomPayload","type":"cell","tag":"tlb:\"maybe ^\""},{"name":"ForwardAmount","type":"coins","tag":"tlb:\".\""},{"name":"ForwardPayload","type":"cell","tag":"tlb:\"either . ^\""}]`),
			payloadBOC: mustBase64(t, "te6cckEBAQEAVgAAp1/MPRQAAAAAAAAAAIAcBFOrOsuVgqBQaiEMi9EgdejAvDwScckxQey2WToQIDADqCJitJw/6R85JCMxuBEjulzVWrYz5gWeqFkMS5xCV6yAJiWgCPhXYXQ="),
			expected:   []byte(`{"Op":{},"QueryID":0,"NewOwner":"EQDgIp1Z1lysFQKDUQhkXokDr0YF4eCTjkmKD2WyydCBAcnZ","ResponseDestination":"EQDqCJitJw_6R85JCMxuBEjulzVWrYz5gWeqFkMS5xCV6w3N","CustomPayload":null,"ForwardAmount":"20000000","ForwardPayload":{}}`),
		}, {
			// telemint deploy
			// https://tonwhales.com/explorer/address/EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N/33913089000003_e926079f4bfccde9b90b1edcd7e2d06e6b719997f9345c692ccf46ebf1ff43a6
			schema:     []byte(`[{"name":"Op","type":"magic","tag":"tlb:\"#4637289b\""},{"name":"Sig","type":"bytes","tag":"tlb:\"bits 512\""},{"name":"SubwalletID","type":"uint32","tag":"tlb:\"## 32\""},{"name":"ValidSince","type":"uint32","tag":"tlb:\"## 32\""},{"name":"ValidTill","type":"uint32","tag":"tlb:\"## 32\""},{"name":"TokenName","type":"telemintText","tag":"tlb:\".\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""},{"name":"TeleitemAuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]},{"name":"RoyaltyParams","type":"struct","tag":"tlb:\"maybe ^\"","struct_fields":[{"name":"Numerator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Denominator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Destination","type":"address","tag":"tlb:\"addr\""}]}]`),
			payloadBOC: mustBase64(t, "te6cckEBBAEA8AADuUY3KJtYJs44aLts46UNlLN1ToWlVd+hM15PiiFyyjO2RCNZj0Sb2wOnHfCCd8IRF354X2+1ESquboKybfW7MBPaGMEKAAAADGOm6lhjpuruCzg4ODAwMDkwMDIxoAMCAQBLAAUAZIAIEbR2UW2Ay0snwgciTTdVU4vxldeBjTs+FVd1ZVMkrXAAYYAIEbR2UW2Ay0snwgciTTdVU4vxldeBjTs+FVd1ZVMkrWpYAo5EAACgAAHCAAEnUBAAYgFodHRwczovL25mdC5mcmFnbWVudC5jb20vbnVtYmVyLzg4ODAwMDkwMDIxLmpzb25dwVfs"),
			expected:   []byte(`{"Op":{},"Sig":"WCbOOGi7bOOlDZSzdU6FpVXfoTNeT4ohcsoztkQjWY9Em9sDpx3wgnfCERd+eF9vtREqrm6Csm31uzAT2hjBCg==","SubwalletID":12,"ValidSince":1671883352,"ValidTill":1671883502,"TokenName":{"Len":11,"Text":"ODg4MDAwOTAwMjE="},"Content":{},"TeleitemAuctionConfig":{"BeneficiaryAddress":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE","InitialMinBid":"189000000000","MaxBid":"0","MinBidStep":5,"MinExtendTime":3600,"Duration":604800},"RoyaltyParams":{"Numerator":5,"Denominator":100,"Destination":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE"}}`),
		}, {
			// teleitem deploy
			// https://tonwhales.com/explorer/address/EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N/33913089000003_e926079f4bfccde9b90b1edcd7e2d06e6b719997f9345c692ccf46ebf1ff43a6
			schema:     []byte(`[{"name":"Op","type":"magic","tag":"tlb:\"#299a3e15\""},{"name":"SenderAddress","type":"address","tag":"tlb:\"addr\""},{"name":"Bid","type":"coins","tag":"tlb:\".\""},{"name":"Info","type":"cell","tag":"tlb:\"^\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""},{"name":"AuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]},{"name":"RoyaltyParams","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"Numerator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Denominator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Destination","type":"address","tag":"tlb:\"addr\""}]}]`),
			payloadBOC: mustBase64(t, "te6cckEBBQEAzwAEVSmaPhWAETu49GWAKN5gWTwjrtUtPWdevOKLzRrYyOi/fgW+2WuKWAKORAEEAwIBAEsABQBkgAgRtHZRbYDLSyfCByJNN1VTi/GV14GNOz4VV3VlUyStcABhgAgRtHZRbYDLSyfCByJNN1VTi/GV14GNOz4VV3VlUyStalgCjkQAAKAAAcIAASdQEABiAWh0dHBzOi8vbmZ0LmZyYWdtZW50LmNvbS9udW1iZXIvODg4MDAwOTAwMjEuanNvbgAcCzg4ODAwMDkwMDIxAQAaXJNe"),
			expected:   []byte(`{"Op":{},"SenderAddress":"EQCJ3cejLAFG8wLJ4R12qWnrOvXnFF5o1sZHRfvwLfbLXKzI","Bid":"189000000000","Info":{},"Content":{},"AuctionConfig":{"BeneficiaryAddress":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE","InitialMinBid":"189000000000","MaxBid":"0","MinBidStep":5,"MinExtendTime":3600,"Duration":604800},"RoyaltyParams":{"Numerator":5,"Denominator":100,"Destination":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE"}}`),
		}, {
			// teleitem start auction
			// https://tonwhales.com/explorer/address/EQBu6eCK84PxTdjEKyY7z8TQGhN3dbzx-935nj-Lx4FCKPaF/33643655000003_bbbb0c3fc14918aa07ef91e5a4f2c5d256bbca46a6622a0930e0e69b3cce4fe3
			schema:     []byte(`[{"name":"Op","type":"magic","tag":"tlb:\"#487a8e81\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"AuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]}]`),
			payloadBOC: mustBase64(t, "te6cckEBAgEARwABGEh6joFflHKrV6ZFAgEAa4AThdJwK2R1Yn8xfoWyVnBTKuN+CSREYnUZzbXXp7wo5EtEGapgALwznNNAAKAAAcIAASdQEO0NbQ0="),
			expected:   []byte(`{"Op":{},"QueryID":6887255810391819522,"AuctionConfig":{"BeneficiaryAddress":"EQCcLpOBWyOrE_mL9C2Ss4KZVxvwSSIjE6jOba69PeFHIgt1","InitialMinBid":"696000000000","MaxBid":"969000000000","MinBidStep":5,"MinExtendTime":3600,"Duration":604800}}`),
		},
	}

	for i, test := range testCases {
		payloadCell, err := cell.FromBOC(test.payloadBOC)
		if err != nil {
			t.Fatal(err)
		}
		payloadSlice := payloadCell.BeginParse()

		s, err := UnmarshalSchema(test.schema)
		if err != nil {
			t.Fatal(err)
		}

		schema, err := MarshalSchema(s)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(schema, test.schema) {
			t.Fatalf("[%d]unmarshalled and marshalled schema is different\nexpected: %s\ngot: %s", i, test.schema, schema)
		}

		if err = tlb.LoadFromCell(s, payloadSlice); err != nil {
			t.Fatal(err)
		}

		raw, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(raw, test.expected) {
			t.Fatalf("[%d]\nexpected: %s\ngot: %s", i, test.expected, raw)
		}
	}
}

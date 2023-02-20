package abi

import (
	"bytes"
	"encoding/base64"
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
	}

	for i, test := range testCases {
		raw, err := MarshalSchema(test.structType)
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != test.expected {
			t.Errorf("[%d] (%s)\nexpected: %s\ngot: %s", i, reflect.TypeOf(test.structType), test.expected, string(raw))
		}
	}
}

func TestUnmarshalSchema(t *testing.T) {
	var testCases = []*struct {
		schema     []byte
		payloadBOC []byte
		// expected   any
	}{
		{
			// nft item transfer
			// https://ton.cx/tx/35447977000003:JH9pr5my6TDD5q4YivuMTQaNmXWAAyfxEb04iCkdz84=:EQAiZupbLhdE7UWQgnTirCbIJRg6yxfmkvTDjxsFh33Cu5rM
			schema:     []byte(`[{"name":"Op","type":"magic","tag":"tlb:\"#5fcc3d14\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"NewOwner","type":"address","tag":"tlb:\"addr\""},{"name":"ResponseDestination","type":"address","tag":"tlb:\"addr\""},{"name":"CustomPayload","type":"cell","tag":"tlb:\"maybe ^\""},{"name":"ForwardAmount","type":"coins","tag":"tlb:\".\""},{"name":"ForwardPayload","type":"cell","tag":"tlb:\"either . ^\""}]`),
			payloadBOC: mustBase64(t, "te6cckEBAQEAVgAAp1/MPRQAAAAAAAAAAIAcBFOrOsuVgqBQaiEMi9EgdejAvDwScckxQey2WToQIDADqCJitJw/6R85JCMxuBEjulzVWrYz5gWeqFkMS5xCV6yAJiWgCPhXYXQ="),
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
			t.Fatalf("unmarshalled and marshalled schema is different\nexpected: %s\ngot: %s", test.schema, schema)
		}

		if err = tlb.LoadFromCell(s, payloadSlice); err != nil {
			t.Fatal(err)
		}

		t.Logf("[%d] %+v", i, s)
		// if !reflect.DeepEqual(test.expected, s) {
		// 	t.Errorf("[%d]\nexpected: %+v\ngot: %+v", i, test.expected, s)
		// }
	}
}

package abi_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	NFTCollectionItemMint struct {
		_         tlb.Magic `tlb:"#00000001"`
		QueryID   uint64    `tlb:"## 64"`
		Index     *big.Int  `tlb:"## 64"`
		TonAmount tlb.Coins `tlb:"."`
		Content   struct {
			Owner   *address.Address `tlb:"addr"`
			Content nft.ContentAny   `tlb:"^"`
			// Editor   *address.Address `tlb:"addr"` // TODO: optional
		} `tlb:"^"`
	}
	NFTCollectionItemMintBatch struct {
		_          tlb.Magic        `tlb:"#00000002"`
		QueryID    uint64           `tlb:"## 64"`
		DeployList *cell.Dictionary `tlb:"dict 64"`
	}
	NFTCollectionChangeOwner struct {
		_        tlb.Magic        `tlb:"#00000003"`
		QueryID  uint64           `tlb:"## 64"`
		NewOwner *address.Address `tlb:"addr"`
	}
	NFTCollectionChangeContent struct {
		_       tlb.Magic  `tlb:"#00000004"`
		QueryID uint64     `tlb:"## 64"`
		Content *cell.Cell `tlb:"^"`
	}
)

func TestNewOperationDesc_NFTCollection(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
		{
			structType: (*NFTCollectionItemMint)(nil),
			expected:   ``,
		}, {
			structType: (*NFTCollectionItemMintBatch)(nil),
			expected:   ``,
		}, {
			structType: (*NFTCollectionChangeOwner)(nil),
			expected:   ``,
		}, {
			structType: (*NFTCollectionChangeContent)(nil),
			expected:   ``,
		},
	}

	for _, test := range testCases {
		got := makeOperationDesc(t, test.structType)
		assert.Equal(t, test.expected, got)
	}
}

// func TestNFTMarshalSchema(t *testing.T) {
// 	var testCases = []*struct {
// 		structType any
// 		expected   string
// 	}{
// 		// nft collection
// 		{
// 			structType: (*NFTCollectionItemMint)(nil),
// 			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#00000001\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"Index","type":"bigInt","tag":"tlb:\"## 64\""},{"name":"TonAmount","type":"coins","tag":"tlb:\".\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""}]`,
// 		}, {
// 			structType: (*NFTCollectionChangeOwner)(nil),
// 			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#00000003\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"NewOwner","type":"address","tag":"tlb:\"addr\""}]`,
// 		},
// 		// nft item
// 		{
// 			structType: (*NFTItemTransfer)(nil),
// 			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#5fcc3d14\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"NewOwner","type":"address","tag":"tlb:\"addr\""},{"name":"ResponseDestination","type":"address","tag":"tlb:\"addr\""},{"name":"CustomPayload","type":"cell","tag":"tlb:\"maybe ^\""},{"name":"ForwardAmount","type":"coins","tag":"tlb:\".\""},{"name":"ForwardPayload","type":"cell","tag":"tlb:\"either . ^\""}]`,
// 		}, {
// 			structType: (*NFTEdit)(nil),
// 			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#1a0b9d51\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""}]`,
// 		},
// 		// jetton minter
// 		{
// 			structType: (*JettonMint)(nil),
// 			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#00000001\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"Index","type":"uint64","tag":"tlb:\"## 64\""},{"name":"TonAmount","type":"coins","tag":"tlb:\".\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""}]`,
// 		},
// 		// jetton wallet
// 		{
// 			structType: (*JettonTransfer)(nil),
// 			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#0f8a7ea5\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"Amount","type":"coins","tag":"tlb:\".\""},{"name":"Destination","type":"address","tag":"tlb:\"addr\""},{"name":"ResponseDestination","type":"address","tag":"tlb:\"addr\""},{"name":"CustomPayload","type":"cell","tag":"tlb:\"maybe ^\""},{"name":"ForwardTONAmount","type":"coins","tag":"tlb:\".\""},{"name":"ForwardPayload","type":"cell","tag":"tlb:\"either . ^\""}]`,
// 		}, {
// 			structType: (*JettonBurn)(nil),
// 			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#595f07bc\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"Amount","type":"coins","tag":"tlb:\".\""},{"name":"ResponseDestination","type":"address","tag":"tlb:\"addr\""},{"name":"CustomPayload","type":"cell","tag":"tlb:\"maybe ^\""}]`,
// 		},
// 	}
//
// 	for _, test := range testCases {
// 		testMarshalSchema(t, test.structType, test.expected)
// 	}
// }
//
// func TestNFTUnmarshalSchema(t *testing.T) {
// 	var testCases = []*struct {
// 		schema     []byte
// 		payloadBOC []byte
// 		expected   []byte
// 	}{
// 		{
// 			// nft item transfer
// 			// https://ton.cx/tx/35447977000003:JH9pr5my6TDD5q4YivuMTQaNmXWAAyfxEb04iCkdz84=:EQAiZupbLhdE7UWQgnTirCbIJRg6yxfmkvTDjxsFh33Cu5rM
// 			schema:     []byte(`[{"name":"Op","type":"magic","tag":"tlb:\"#5fcc3d14\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"NewOwner","type":"address","tag":"tlb:\"addr\""},{"name":"ResponseDestination","type":"address","tag":"tlb:\"addr\""},{"name":"CustomPayload","type":"cell","tag":"tlb:\"maybe ^\""},{"name":"ForwardAmount","type":"coins","tag":"tlb:\".\""},{"name":"ForwardPayload","type":"cell","tag":"tlb:\"either . ^\""}]`),
// 			payloadBOC: mustBase64(t, "te6cckEBAQEAVgAAp1/MPRQAAAAAAAAAAIAcBFOrOsuVgqBQaiEMi9EgdejAvDwScckxQey2WToQIDADqCJitJw/6R85JCMxuBEjulzVWrYz5gWeqFkMS5xCV6yAJiWgCPhXYXQ="),
// 			expected:   []byte(`{"Op":{},"QueryID":0,"NewOwner":"EQDgIp1Z1lysFQKDUQhkXokDr0YF4eCTjkmKD2WyydCBAcnZ","ResponseDestination":"EQDqCJitJw_6R85JCMxuBEjulzVWrYz5gWeqFkMS5xCV6w3N","CustomPayload":null,"ForwardAmount":"20000000","ForwardPayload":{}}`),
// 		},
// 	}
//
// 	for _, test := range testCases {
// 		testUnmarshalSchema(t, test.payloadBOC, test.schema, test.expected)
// 	}
// }

package abi

import "testing"

func TestNFTMarshalSchema(t *testing.T) {
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
			structType: (*NFTEdit)(nil),
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

	for _, test := range testCases {
		testMarshalSchema(t, test.structType, test.expected)
	}
}

func TestNFTUnmarshalSchema(t *testing.T) {
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
		},
	}

	for _, test := range testCases {
		testUnmarshalSchema(t, test.payloadBOC, test.schema, test.expected)
	}
}

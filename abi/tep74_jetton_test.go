package abi

import "testing"

func TestJettonMarshalSchema(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
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

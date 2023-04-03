package abi

import (
	"testing"
)

func TestTelemintMarshalSchema(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
		// telemint
		{
			structType: (*TelemintMsgDeploy)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#4637289a\""},{"name":"Sig","type":"bytes","tag":"tlb:\"bits 512\""},{"name":"SubwalletID","type":"uint32","tag":"tlb:\"## 32\""},{"name":"ValidSince","type":"uint32","tag":"tlb:\"## 32\""},{"name":"ValidTill","type":"uint32","tag":"tlb:\"## 32\""},{"name":"TokenName","type":"telemintText","tag":"tlb:\".\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""},{"name":"AuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]},{"name":"RoyaltyParams","type":"struct","tag":"tlb:\"maybe ^\"","struct_fields":[{"name":"Numerator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Denominator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Destination","type":"address","tag":"tlb:\"addr\""}]}]`,
		}, {
			structType: (*TeleitemMsgDeploy)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#299a3e15\""},{"name":"SenderAddress","type":"address","tag":"tlb:\"addr\""},{"name":"Bid","type":"coins","tag":"tlb:\".\""},{"name":"Info","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"Name","type":"telemintText","tag":"tlb:\".\""},{"name":"Domain","type":"telemintText","tag":"tlb:\".\""}]},{"name":"Content","type":"cell","tag":"tlb:\"^\""},{"name":"AuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]},{"name":"RoyaltyParams","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"Numerator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Denominator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Destination","type":"address","tag":"tlb:\"addr\""}]}]`,
		}, {
			structType: (*TeleitemStartAuction)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#487a8e81\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"AuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]}]`,
		}, {
			structType: (*TeleitemCancelAuction)(nil),
			expected:   `[{"name":"Op","type":"magic","tag":"tlb:\"#371638ae\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""}]`,
		},
	}

	for _, test := range testCases {
		testMarshalSchema(t, test.structType, test.expected)
	}
}

func TestTelemintUnmarshalSchema(t *testing.T) {
	var testCases = []*struct {
		schema     []byte
		payloadBOC []byte
		expected   []byte
	}{
		{
			// telemint deploy
			// https://tonwhales.com/explorer/address/EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N/33913089000003_e926079f4bfccde9b90b1edcd7e2d06e6b719997f9345c692ccf46ebf1ff43a6
			schema:     []byte(`[{"name":"Op","type":"magic","tag":"tlb:\"#4637289b\""},{"name":"Sig","type":"bytes","tag":"tlb:\"bits 512\""},{"name":"SubwalletID","type":"uint32","tag":"tlb:\"## 32\""},{"name":"ValidSince","type":"uint32","tag":"tlb:\"## 32\""},{"name":"ValidTill","type":"uint32","tag":"tlb:\"## 32\""},{"name":"TokenName","type":"telemintText","tag":"tlb:\".\""},{"name":"Content","type":"cell","tag":"tlb:\"^\""},{"name":"TeleitemAuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]},{"name":"RoyaltyParams","type":"struct","tag":"tlb:\"maybe ^\"","struct_fields":[{"name":"Numerator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Denominator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Destination","type":"address","tag":"tlb:\"addr\""}]}]`),
			payloadBOC: mustBase64(t, "te6cckEBBAEA8AADuUY3KJtYJs44aLts46UNlLN1ToWlVd+hM15PiiFyyjO2RCNZj0Sb2wOnHfCCd8IRF354X2+1ESquboKybfW7MBPaGMEKAAAADGOm6lhjpuruCzg4ODAwMDkwMDIxoAMCAQBLAAUAZIAIEbR2UW2Ay0snwgciTTdVU4vxldeBjTs+FVd1ZVMkrXAAYYAIEbR2UW2Ay0snwgciTTdVU4vxldeBjTs+FVd1ZVMkrWpYAo5EAACgAAHCAAEnUBAAYgFodHRwczovL25mdC5mcmFnbWVudC5jb20vbnVtYmVyLzg4ODAwMDkwMDIxLmpzb25dwVfs"),
			expected:   []byte(`{"Op":{},"Sig":"WCbOOGi7bOOlDZSzdU6FpVXfoTNeT4ohcsoztkQjWY9Em9sDpx3wgnfCERd+eF9vtREqrm6Csm31uzAT2hjBCg==","SubwalletID":12,"ValidSince":1671883352,"ValidTill":1671883502,"TokenName":{"Len":11,"Text":"88800090021"},"Content":{},"TeleitemAuctionConfig":{"BeneficiaryAddress":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE","InitialMinBid":"189000000000","MaxBid":"0","MinBidStep":5,"MinExtendTime":3600,"Duration":604800},"RoyaltyParams":{"Numerator":5,"Denominator":100,"Destination":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE"}}`),
		}, {
			// teleitem deploy
			// https://tonwhales.com/explorer/address/EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N/33913089000003_e926079f4bfccde9b90b1edcd7e2d06e6b719997f9345c692ccf46ebf1ff43a6
			schema:     []byte(`[{"name":"Op","type":"magic","tag":"tlb:\"#299a3e15\""},{"name":"SenderAddress","type":"address","tag":"tlb:\"addr\""},{"name":"Bid","type":"coins","tag":"tlb:\".\""},{"name":"Info","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"Name","type":"telemintText","tag":"tlb:\".\""},{"name":"Domain","type":"telemintText","tag":"tlb:\".\""}]},{"name":"Content","type":"cell","tag":"tlb:\"^\""},{"name":"AuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]},{"name":"RoyaltyParams","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"Numerator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Denominator","type":"uint16","tag":"tlb:\"## 16\""},{"name":"Destination","type":"address","tag":"tlb:\"addr\""}]}]`),
			payloadBOC: mustBase64(t, "te6cckEBBQEAzwAEVSmaPhWAETu49GWAKN5gWTwjrtUtPWdevOKLzRrYyOi/fgW+2WuKWAKORAEEAwIBAEsABQBkgAgRtHZRbYDLSyfCByJNN1VTi/GV14GNOz4VV3VlUyStcABhgAgRtHZRbYDLSyfCByJNN1VTi/GV14GNOz4VV3VlUyStalgCjkQAAKAAAcIAASdQEABiAWh0dHBzOi8vbmZ0LmZyYWdtZW50LmNvbS9udW1iZXIvODg4MDAwOTAwMjEuanNvbgAcCzg4ODAwMDkwMDIxAQAaXJNe"),
			expected:   []byte(`{"Op":{},"SenderAddress":"EQCJ3cejLAFG8wLJ4R12qWnrOvXnFF5o1sZHRfvwLfbLXKzI","Bid":"189000000000","Info":{"Name":{"Len":11,"Text":"88800090021"},"Domain":{"Len":1,"Text":"\u0000"}},"Content":{},"AuctionConfig":{"BeneficiaryAddress":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE","InitialMinBid":"189000000000","MaxBid":"0","MinBidStep":5,"MinExtendTime":3600,"Duration":604800},"RoyaltyParams":{"Numerator":5,"Denominator":100,"Destination":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE"}}`),
		}, {
			// teleitem start auction
			// https://tonwhales.com/explorer/address/EQBu6eCK84PxTdjEKyY7z8TQGhN3dbzx-935nj-Lx4FCKPaF/33643655000003_bbbb0c3fc14918aa07ef91e5a4f2c5d256bbca46a6622a0930e0e69b3cce4fe3
			schema:     []byte(`[{"name":"Op","type":"magic","tag":"tlb:\"#487a8e81\""},{"name":"QueryID","type":"uint64","tag":"tlb:\"## 64\""},{"name":"AuctionConfig","type":"struct","tag":"tlb:\"^\"","struct_fields":[{"name":"BeneficiaryAddress","type":"address","tag":"tlb:\"addr\""},{"name":"InitialMinBid","type":"coins","tag":"tlb:\".\""},{"name":"MaxBid","type":"coins","tag":"tlb:\".\""},{"name":"MinBidStep","type":"uint8","tag":"tlb:\"## 8\""},{"name":"MinExtendTime","type":"uint32","tag":"tlb:\"## 32\""},{"name":"Duration","type":"uint32","tag":"tlb:\"## 32\""}]}]`),
			payloadBOC: mustBase64(t, "te6cckEBAgEARwABGEh6joFflHKrV6ZFAgEAa4AThdJwK2R1Yn8xfoWyVnBTKuN+CSREYnUZzbXXp7wo5EtEGapgALwznNNAAKAAAcIAASdQEO0NbQ0="),
			expected:   []byte(`{"Op":{},"QueryID":6887255810391819522,"AuctionConfig":{"BeneficiaryAddress":"EQCcLpOBWyOrE_mL9C2Ss4KZVxvwSSIjE6jOba69PeFHIgt1","InitialMinBid":"696000000000","MaxBid":"969000000000","MinBidStep":5,"MinExtendTime":3600,"Duration":604800}}`),
		},
	}

	for _, test := range testCases {
		testUnmarshalSchema(t, test.payloadBOC, test.schema, test.expected)
	}
}

package known_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xssnick/tonutils-go/address"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
)

func TestGetMethodDesc_Wallets(t *testing.T) {
	var (
		interfaces []*abi.InterfaceDesc
		i          *abi.InterfaceDesc
	)

	j, err := os.ReadFile("wallets.json")
	require.Nil(t, err)

	err = json.Unmarshal(j, &interfaces)
	require.Nil(t, err)

	for _, i = range interfaces {
		if i.Name == "wallet_v1r3" {
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
			name: "seqno",
			addr: addr.MustFromBase64("EQDj5AA8mQvM5wJEQsFFFof79y3ZsuX6wowktWQFhz_Anton").MustToTonutils(),
			code: `te6cckEBAQEAcQAA3v8AIN0gggFMl7ohggEznLqxn3Gw7UTQ0x/THzHXC//jBOCk8mCDCNcYINMf0x/TH/gjE7vyY+1E0NMf0x/T/9FRMrryoVFEuvKiBPkBVBBV+RDyo/gAkyDXSpbTB9QC+wDo0QGkyMsfyx/L/8ntVBC9ba0=`,
			data: `te6cckEBAQEAKgAAUAAAAAEGQZj7UhMYn0DGJKa8VAJx2X9dF+VkfoJrgOKgW7MinX6Pqkvc3Pev`,
			expected: []any{
				uint32(1),
			},
		}, {
			name: "get_public_key",
			addr: addr.MustFromBase64("EQDj5AA8mQvM5wJEQsFFFof79y3ZsuX6wowktWQFhz_Anton").MustToTonutils(),
			code: `te6cckEBAQEAcQAA3v8AIN0gggFMl7ohggEznLqxn3Gw7UTQ0x/THzHXC//jBOCk8mCDCNcYINMf0x/TH/gjE7vyY+1E0NMf0x/T/9FRMrryoVFEuvKiBPkBVBBV+RDyo/gAkyDXSpbTB9QC+wDo0QGkyMsfyx/L/8ntVBC9ba0=`,
			data: `te6cckEBAQEAKgAAUAAAAAEGQZj7UhMYn0DGJKa8VAJx2X9dF+VkfoJrgOKgW7MinX6Pqkvc3Pev`,
			expected: []any{
				[]uint8{0x52, 0x13, 0x18, 0x9f, 0x40, 0xc6, 0x24, 0xa6, 0xbc, 0x54, 0x2, 0x71, 0xd9, 0x7f, 0x5d, 0x17, 0xe5, 0x64, 0x7e, 0x82, 0x6b, 0x80, 0xe2, 0xa0, 0x5b, 0xb3, 0x22, 0x9d, 0x7e, 0x8f, 0xaa, 0x4b},
			},
		}, {
			name: "get_public_key",
			addr: addr.MustFromBase64("EQAy4fglSdK3G5YYd2aiWqh_d7dT_2NDmxVrnq9Ed4uEvNGe").MustToTonutils(),
			code: `te6cckEBCQEA6QABFP8A9KQT9LzyyAsBAgEgAgMCAUgEBQHu8oMI1xgg0x/TP/gjqh9TILnyY+1E0NMf0z/T//QE0VNggED0Dm+hMfJgUXO68qIH+QFUEIf5EPKjAvQE0fgAf44YIYAQ9HhvoW+hIJgC0wfUMAH7AJEy4gGz5luDJaHIQDSAQPRDiuYxyBLLHxPLP8v/9ADJ7VQIAATQMAIBIAYHABe9nOdqJoaa+Y64X/wAQb5fl2omhpj5jpn+n/mPoCaKkQQCB6BzfQmMktv8ld0fFAA4IIBA9JZvoW+hMlEQlDBTA7neIJMzNgGSMjDis/eNfE4=`,
			data: `te6cckEBAgEAPAABWSmpoxdkZS/GN/MXxZShySjGzcYn+KyymF5QkFD4R0wbXtC4FhePRjHtd2HgwAEAE6AyMp9JSQvSaMCoCmRt`,
			expected: []any{
				[]uint8{0x94, 0xa1, 0xc9, 0x28, 0xc6, 0xcd, 0xc6, 0x27, 0xf8, 0xac, 0xb2, 0x98, 0x5e, 0x50, 0x90, 0x50, 0xf8, 0x47, 0x4c, 0x1b, 0x5e, 0xd0, 0xb8, 0x16, 0x17, 0x8f, 0x46, 0x31, 0xed, 0x77, 0x61, 0xe0},
			},
		}, {
			name: "get_public_key",
			addr: addr.MustFromBase64("EQAkcrRxpQZQS2bfhJ69Pd9GEUMQbOSB7-D2biP0pG6Qbbva").MustToTonutils(),
			code: `te6cckECHgEAAmEAART/APSkE/S88sgLAQIBIAIDAgFIBAUB8vKDCNcYINMf0x/TH4AkA/gjuxPy8vADgCJRqboa8vSAI1G3uhvy9IAfC/kBVBDF+RAa8vT4AFBX+CPwBlCY+CPwBiBxKJMg10qOi9MHMdRRG9s8ErAB6DCSKaDfcvsCBpMg10qW0wfUAvsA6NEDpEdoFBVDMPAE7VQdAgLNBgcCASATFAIBIAgJAgEgDxACASAKCwAtXtRNDTH9Mf0//T//QE+gD0BPoA9ATRgD9wB0NMDAXGwkl8D4PpAMCHHAJJfA+AB0x8hwQKSXwTg8ANRtPABghCC6vnEUrC9sJJfDOCAKIIQgur5xBu6GvL0gCErghA7msoAvvL0B4MI1xiAICH5AVQQNvkQEvL00x+AKYIQNzqp9BO6EvL00wDTHzAB4w8QSBA3XjKAMDQ4AEwh10n0qG+lbDGAADBA5SArwBQAWEDdBCvAFCBBXUFYAEBAkQwDwBO1UAgEgERIARUjh4igCD0lm+lIJMwI7uRMeIgmDX6ANEToUATkmwh4rPmMIADUCMjKHxfKHxXL/xPL//QAAfoC9AAB+gL0AMmAAQxRIqBTE4Ag9A5voZb6ANEToAKRMOLIUAP6AkATgCD0QwGACASAVFgAVven3gBiCQvhHgAwCASAXGAIBSBscAC21GH4AbYiGioJgngDGIH4Axj8E7eILMAIBWBkaABetznaiaGmfmOuF/8AAF6x49qJoaY+Y64WPwAARsyX7UTQ1wsfgABex0b4I4IBCMPtQ9iAAKAHQ0wMBeLCSW3/g+kAx+kAwAfABqA7apA==`,
			data: `te6cckEBCAEA7wADqQAAAAwpqaMXGb4cMG6ojLNpNk3djQ+1hRNnkQpBh4xgOYVrD8jUBSCV++yZT7+kMrngsA7CHoC5xLGOupqaDEHiclgod0th26Hc1lAFwLXmL9sygDABAgMCA3TABAUAE6BndFsUQ7msoAgAGaBliz6UcC15i/bMoAgARaDgHQlFJs6x5OJhjcQnrIaeHowB4o13LmlQQ0YQzYMnJmDQAgV/v7AGBwBDv43zy3jusChgjyZZkoVOtBMzhJFoxaUnZLl+IAY6hi1EQABDv69TGho07kLUXYIufYfD+CPKMZveRM9ghkREWR9nyv3hQMngwgs=`,
			expected: []any{
				[]uint8{0x19, 0xbe, 0x1c, 0x30, 0x6e, 0xa8, 0x8c, 0xb3, 0x69, 0x36, 0x4d, 0xdd, 0x8d, 0xf, 0xb5, 0x85, 0x13, 0x67, 0x91, 0xa, 0x41, 0x87, 0x8c, 0x60, 0x39, 0x85, 0x6b, 0xf, 0xc8, 0xd4, 0x5, 0x20},
			},
		},
	}

	for _, test := range testCases {
		ret := execGetMethod(t, i, test.addr, test.name, test.code, test.data)
		assert.Equal(t, test.expected, ret)
	}
}

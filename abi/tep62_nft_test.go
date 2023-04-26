package abi_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
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
			Content *cell.Cell       `tlb:"^"`
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
			expected:   `{"op_name":"nft_collection_item_mint","op_code":"0x1","body":[{"name":"query_id","tlb_type":"## 64","map_to":"uint64"},{"name":"index","tlb_type":"## 64","map_to":"bigInt"},{"name":"ton_amount","tlb_type":".","map_to":"coins"},{"name":"content","tlb_type":"^","map_to":"struct","struct_fields":[{"name":"owner","tlb_type":"addr","map_to":"addr"},{"name":"content","tlb_type":"^","map_to":"cell"}]}]}`,
		}, {
			structType: (*NFTCollectionItemMintBatch)(nil),
			expected:   `{"op_name":"nft_collection_item_mint_batch","op_code":"0x2","body":[{"name":"query_id","tlb_type":"## 64","map_to":"uint64"},{"name":"deploy_list","tlb_type":"dict 64","map_to":"dict"}]}`,
		}, {
			structType: (*NFTCollectionChangeOwner)(nil),
			expected:   `{"op_name":"nft_collection_change_owner","op_code":"0x3","body":[{"name":"query_id","tlb_type":"## 64","map_to":"uint64"},{"name":"new_owner","tlb_type":"addr","map_to":"addr"}]}`,
		}, {
			structType: (*NFTCollectionChangeContent)(nil),
			expected:   `{"op_name":"nft_collection_change_content","op_code":"0x4","body":[{"name":"query_id","tlb_type":"## 64","map_to":"uint64"},{"name":"content","tlb_type":"^","map_to":"cell"}]}`,
		},
	}

	for _, test := range testCases {
		got := makeOperationDesc(t, test.structType)
		assert.Equal(t, test.expected, got)
	}
}

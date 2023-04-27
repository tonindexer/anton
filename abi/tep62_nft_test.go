package abi_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestLoadOperation_NFTCollection(t *testing.T) {
	var testCases = []*struct {
		schema   string
		boc      string
		expected string
	}{
		{
			// tx hash 3fa60549d7bdd8640f67bebb96240b3683032420691396b0dd4bff37a65c3222
			schema:   `{"op_name":"nft_collection_item_mint","op_code":"0x1","body":[{"name":"query_id","tlb_type":"## 64","map_to":"uint64"},{"name":"index","tlb_type":"## 64","map_to":"bigInt"},{"name":"ton_amount","tlb_type":".","map_to":"coins"},{"name":"content","tlb_type":"^","map_to":"struct","struct_fields":[{"name":"owner","tlb_type":"addr","map_to":"addr"},{"name":"content","tlb_type":"^","map_to":"cell"}]}]}`,
			boc:      `te6cckEBAwEAcAABMQAAAAEAAAAAAAAAAAAAAAAAAAEaQC+vCAgBAYWADp1n1HTZ6w3lG+DAt9kB/n3R2/GVH4sVu8f2JT7l9r4wAnNV2x5VpHz5NgrTTwmcnFe3yTCqMygGQXqc8N/RpSlKAgAYMjgyLzI4Mi5qc29uxWpC2Q==`,
			expected: `{"query_id":0,"index":282,"ton_amount":"50000000","content":{"owner":"EQB06z6jps9YbyjfBgW-yA_z7o7fjKj8WK3eP7Ep9y-18a0N","content":{}}}`,
		}, {
			// tx hash 0b97a525e1acbcde7832e66270675c2ce5de7ace50d517c195aa0956edc66bd7
			schema:   `{"op_name":"nft_collection_item_mint_batch","op_code":"0x2","body":[{"name":"query_id","tlb_type":"## 64","map_to":"uint64"},{"name":"deploy_list","tlb_type":"dict 64","map_to":"dict"}]}`,
			boc:      `te6cckECDQEAAWkAARkAAAACAAAe79MtoUjAAQIRnwAAAAAAAAEnAwIBCkQC+vCABwIBIAQJAQkQC+vCAgUChYAX28Uqtfy2ijm9JdYxIkuCTQuN8K678kSAG4mDqoI5prAAC0KLGhVVTkLKRqF5qaC/+TjInh0vIy7y28Gi/4qWYPYGCwAQNjA0Lmpzb24ChYAX28Uqtfy2ijm9JdYxIkuCTQuN8K678kSAG4mDqoI5prAAC0KLGhVVTkLKRqF5qaC/+TjInh0vIy7y28Gi/4qWYPYICwAQNjIzLmpzb24BCRAL68ICCgKFgBfbxSq1/LaKOb0l1jEiS4JNC43wrrvyRIAbiYOqgjmmsAALQosaFVVOQspGoXmpoL/5OMieHS8jLvLbwaL/ipZg9gwLAHOACQS5lb2Y344H77A78MhYkkcNG+wnaJs9b4Ew7o3aD70MdYDMKxOg9yg7msoAzs6+2MLq3MbQ4MLJABA2MTkuanNvbowfcO8=`,
			expected: `{"query_id":34015389000008,"deploy_list":{}}`,
		}, {
			// tx hash 88b26b816fea70e7fcfabe4c92fea7ca98ea5fa1ec5ebcf088b0c8072b0c230a
			schema:   `{"op_name":"nft_collection_change_owner","op_code":"0x3","body":[{"name":"query_id","tlb_type":"## 64","map_to":"uint64"},{"name":"new_owner","tlb_type":"addr","map_to":"addr"}]}`,
			boc:      `te6cckEBAQEAMAAAWwAAAAMAAAAAAAAAAIAAXm3EFmHhzJdBivX89fE8mLQiAsredEQ45yxDDHmrxnDHbQiG`,
			expected: `{"query_id":0,"new_owner":"EQAC824gsw8OZLoMV6_nr4nkxaEQFlbzoiHHOWIYY81eM5rQ"}`,
		}, {
			// tx hash 19a40062e31365d6ad4473aabb62562f37d04c8aa5618b7ea800885dbb5a0e70
			schema:   `{"op_name":"nft_collection_change_content","op_code":"0x4","body":[{"name":"query_id","tlb_type":"## 64","map_to":"uint64"},{"name":"content","tlb_type":"^","map_to":"cell"}]}`,
			boc:      `te6cckEBBQEAxQACGAAAAAQAAAAAAAAAAAIBAEsAAABkgAe4JggE07o9i50C5vJ2aQiIbUYwl8/YMW27aEtS1YEfMAIABAMAaGh0dHBzOi8vcy5nZXRnZW1zLmlvL25mdC9jLzYzYTgwMDVlY2MxM2M0OTE0YjMxNGIyMy8AogFodHRwczovL3MuZ2V0Z2Vtcy5pby9uZnQvYy82M2E4MDA1ZWNjMTNjNDkxNGIzMTRiMjMvZWRpdC9tZXRhLTE2NzIyODIyMDQ1ODkuanNvbml2vhQ=`,
			expected: `{"query_id":0,"content":{}}`,
		},
	}

	for _, test := range testCases {
		var d abi.OperationDesc

		err := json.Unmarshal([]byte(test.schema), &d)
		require.Nil(t, err)

		op, err := d.New()
		require.Nil(t, err)

		c, err := cell.FromBOC(mustBase64(t, test.boc))
		require.Nil(t, err)

		err = tlb.LoadFromCell(op, c.BeginParse())
		require.Nil(t, err)

		j, err := json.Marshal(op)
		require.Nil(t, err)

		assert.Equal(t, string(j), test.expected)
	}
}

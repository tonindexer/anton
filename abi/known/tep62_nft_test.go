package known_test

import (
	"encoding/base64"
	"math/big"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/stretchr/testify/assert"

	"github.com/tonindexer/anton/abi"
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
			expected:   `{"op_name":"nft_collection_item_mint","op_code":"0x1","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"index","tlb_type":"## 64","format":"bigInt"},{"name":"ton_amount","tlb_type":".","format":"coins"},{"name":"content","tlb_type":"^","format":"struct","struct_fields":[{"name":"owner","tlb_type":"addr","format":"addr"},{"name":"content","tlb_type":"^","format":"cell"}]}]}`,
		}, {
			structType: (*NFTCollectionItemMintBatch)(nil),
			expected:   `{"op_name":"nft_collection_item_mint_batch","op_code":"0x2","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"deploy_list","tlb_type":"dict 64","format":"dict"}]}`,
		}, {
			structType: (*NFTCollectionChangeOwner)(nil),
			expected:   `{"op_name":"nft_collection_change_owner","op_code":"0x3","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"new_owner","tlb_type":"addr","format":"addr"}]}`,
		}, {
			structType: (*NFTCollectionChangeContent)(nil),
			expected:   `{"op_name":"nft_collection_change_content","op_code":"0x4","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"content","tlb_type":"^","format":"cell"}]}`,
		},
	}

	for _, test := range testCases {
		d, err := abi.NewOperationDesc(test.structType)
		require.Nil(t, err)
		got, err := json.Marshal(d)
		require.Nil(t, err)
		assert.Equal(t, test.expected, string(got))
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
			schema:   `{"op_name":"nft_collection_item_mint","op_code":"0x1","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"index","tlb_type":"## 64","format":"bigInt"},{"name":"ton_amount","tlb_type":".","format":"coins"},{"name":"content","tlb_type":"^","format":"struct","struct_fields":[{"name":"owner","tlb_type":"addr","format":"addr"},{"name":"content","tlb_type":"^","format":"cell"}]}]}`,
			boc:      `te6cckEBAwEAcAABMQAAAAEAAAAAAAAAAAAAAAAAAAEaQC+vCAgBAYWADp1n1HTZ6w3lG+DAt9kB/n3R2/GVH4sVu8f2JT7l9r4wAnNV2x5VpHz5NgrTTwmcnFe3yTCqMygGQXqc8N/RpSlKAgAYMjgyLzI4Mi5qc29uxWpC2Q==`,
			expected: `{"query_id":0,"index":282,"ton_amount":"50000000","content":{"owner":"EQB06z6jps9YbyjfBgW-yA_z7o7fjKj8WK3eP7Ep9y-18a0N","content":{}}}`,
		}, {
			// tx hash 0b97a525e1acbcde7832e66270675c2ce5de7ace50d517c195aa0956edc66bd7
			schema:   `{"op_name":"nft_collection_item_mint_batch","op_code":"0x2","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"deploy_list","tlb_type":"dict 64","format":"dict"}]}`,
			boc:      `te6cckECDQEAAWkAARkAAAACAAAe79MtoUjAAQIRnwAAAAAAAAEnAwIBCkQC+vCABwIBIAQJAQkQC+vCAgUChYAX28Uqtfy2ijm9JdYxIkuCTQuN8K678kSAG4mDqoI5prAAC0KLGhVVTkLKRqF5qaC/+TjInh0vIy7y28Gi/4qWYPYGCwAQNjA0Lmpzb24ChYAX28Uqtfy2ijm9JdYxIkuCTQuN8K678kSAG4mDqoI5prAAC0KLGhVVTkLKRqF5qaC/+TjInh0vIy7y28Gi/4qWYPYICwAQNjIzLmpzb24BCRAL68ICCgKFgBfbxSq1/LaKOb0l1jEiS4JNC43wrrvyRIAbiYOqgjmmsAALQosaFVVOQspGoXmpoL/5OMieHS8jLvLbwaL/ipZg9gwLAHOACQS5lb2Y344H77A78MhYkkcNG+wnaJs9b4Ew7o3aD70MdYDMKxOg9yg7msoAzs6+2MLq3MbQ4MLJABA2MTkuanNvbowfcO8=`,
			expected: `{"query_id":34015389000008,"deploy_list":{}}`,
		}, {
			// tx hash 88b26b816fea70e7fcfabe4c92fea7ca98ea5fa1ec5ebcf088b0c8072b0c230a
			schema:   `{"op_name":"nft_collection_change_owner","op_code":"0x3","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"new_owner","tlb_type":"addr","format":"addr"}]}`,
			boc:      `te6cckEBAQEAMAAAWwAAAAMAAAAAAAAAAIAAXm3EFmHhzJdBivX89fE8mLQiAsredEQ45yxDDHmrxnDHbQiG`,
			expected: `{"query_id":0,"new_owner":"EQAC824gsw8OZLoMV6_nr4nkxaEQFlbzoiHHOWIYY81eM5rQ"}`,
		}, {
			// tx hash 19a40062e31365d6ad4473aabb62562f37d04c8aa5618b7ea800885dbb5a0e70
			schema:   `{"op_name":"nft_collection_change_content","op_code":"0x4","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"content","tlb_type":"^","format":"cell"}]}`,
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

		boc, err := base64.StdEncoding.DecodeString(test.boc)
		require.Nil(t, err)

		c, err := cell.FromBOC(boc)
		require.Nil(t, err)

		err = tlb.LoadFromCell(op, c.BeginParse())
		require.Nil(t, err)

		j, err := json.Marshal(op)
		require.Nil(t, err)

		assert.Equal(t, test.expected, string(j))
	}
}

type (
	NFTGetRoyaltyParams struct {
		_       tlb.Magic `tlb:"#693d3950"`
		QueryID uint64    `tlb:"## 64"`
	}
	NFTReportRoyaltyParams struct {
		_           tlb.Magic        `tlb:"#a8cb00ad"`
		Numerator   uint16           `tlb:"## 16"`
		Denominator uint16           `tlb:"## 16"`
		Destination *address.Address `tlb:"addr"`
	}
)

func TestNewOperationDesc_NFTRoyalty(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
		{
			structType: (*NFTGetRoyaltyParams)(nil),
			expected:   `{"op_name":"nft_get_royalty_params","op_code":"0x693d3950","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"}]}`,
		}, {
			structType: (*NFTReportRoyaltyParams)(nil),
			expected:   `{"op_name":"nft_report_royalty_params","op_code":"0xa8cb00ad","body":[{"name":"numerator","tlb_type":"## 16","format":"uint16"},{"name":"denominator","tlb_type":"## 16","format":"uint16"},{"name":"destination","tlb_type":"addr","format":"addr"}]}`,
		},
	}

	for _, test := range testCases {
		d, err := abi.NewOperationDesc(test.structType)
		require.Nil(t, err)
		got, err := json.Marshal(d)
		require.Nil(t, err)
		assert.Equal(t, test.expected, string(got))
	}
}

type (
	NFTItemTransfer struct {
		_                   tlb.Magic        `tlb:"#5fcc3d14"`
		QueryID             uint64           `tlb:"## 64"`
		NewOwner            *address.Address `tlb:"addr"`
		ResponseDestination *address.Address `tlb:"addr"`
		CustomPayload       *cell.Cell       `tlb:"maybe ^"`
		ForwardAmount       tlb.Coins        `tlb:"."`
		ForwardPayload      *cell.Cell       `tlb:"either . ^"`
	}
	NFTItemOwnershipAssigned struct {
		_              tlb.Magic        `tlb:"#05138d91"`
		QueryID        uint64           `tlb:"## 64"`
		PrevOwner      *address.Address `tlb:"addr"`
		ForwardPayload *cell.Cell       `tlb:"either . ^"`
	}
	NFTItemGetStaticData struct {
		_       tlb.Magic `tlb:"#2fcb26a2"`
		QueryID uint64    `tlb:"## 64"`
	}
	NFTItemReportStaticData struct {
		_          tlb.Magic        `tlb:"#8b771735"`
		QueryID    uint64           `tlb:"## 64"`
		Index      *big.Int         `tlb:"## 256"`
		Collection *address.Address `tlb:"addr"`
	}
	Excesses struct {
		_       tlb.Magic `tlb:"#d53276db"`
		QueryID uint64    `tlb:"## 64"`
	}
)

func TestNewOperationDesc_NFTItem(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
		{
			structType: (*NFTItemTransfer)(nil),
			expected:   `{"op_name":"nft_item_transfer","op_code":"0x5fcc3d14","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"new_owner","tlb_type":"addr","format":"addr"},{"name":"response_destination","tlb_type":"addr","format":"addr"},{"name":"custom_payload","tlb_type":"maybe ^","format":"cell"},{"name":"forward_amount","tlb_type":".","format":"coins"},{"name":"forward_payload","tlb_type":"either . ^","format":"cell"}]}`,
		}, {
			structType: (*NFTItemOwnershipAssigned)(nil),
			expected:   `{"op_name":"nft_item_ownership_assigned","op_code":"0x5138d91","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"prev_owner","tlb_type":"addr","format":"addr"},{"name":"forward_payload","tlb_type":"either . ^","format":"cell"}]}`,
		}, {
			structType: (*NFTItemGetStaticData)(nil),
			expected:   `{"op_name":"nft_item_get_static_data","op_code":"0x2fcb26a2","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"}]}`,
		}, {
			structType: (*NFTItemReportStaticData)(nil),
			expected:   `{"op_name":"nft_item_report_static_data","op_code":"0x8b771735","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"index","tlb_type":"## 256","format":"bigInt"},{"name":"collection","tlb_type":"addr","format":"addr"}]}`,
		}, {
			structType: (*Excesses)(nil),
			expected:   `{"op_name":"excesses","op_code":"0xd53276db","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"}]}`,
		},
	}

	for _, test := range testCases {
		d, err := abi.NewOperationDesc(test.structType)
		require.Nil(t, err)
		got, err := json.Marshal(d)
		require.Nil(t, err)
		assert.Equal(t, test.expected, string(got))
	}
}

type (
	NFTEditContent struct {
		_       tlb.Magic  `tlb:"#1a0b9d51"`
		QueryID uint64     `tlb:"## 64"`
		Content *cell.Cell `tlb:"^"`
	}
	NFTTransferEditorship struct {
		_                   tlb.Magic        `tlb:"#1c04412a"`
		QueryID             uint64           `tlb:"## 64"`
		NewEditor           *address.Address `tlb:"addr"`
		ResponseDestination *address.Address `tlb:"addr"`
		CustomPayload       *cell.Cell       `tlb:"maybe ^"`
		ForwardAmount       tlb.Coins        `tlb:"."`
		ForwardPayload      *cell.Cell       `tlb:"either . ^"`
	}
	NFTEditorshipAssigned struct {
		_              tlb.Magic        `tlb:"#511a4463"`
		QueryID        uint64           `tlb:"## 64"`
		PrevEditor     *address.Address `tlb:"addr"`
		ForwardPayload *cell.Cell       `tlb:"either . ^"`
	}
)

func TestNewOperationDesc_NFTEditable(t *testing.T) {
	var testCases = []*struct {
		structType any
		expected   string
	}{
		{
			structType: (*NFTEditContent)(nil),
			expected:   `{"op_name":"nft_edit_content","op_code":"0x1a0b9d51","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"content","tlb_type":"^","format":"cell"}]}`,
		}, {
			structType: (*NFTTransferEditorship)(nil),
			expected:   `{"op_name":"nft_transfer_editorship","op_code":"0x1c04412a","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"new_editor","tlb_type":"addr","format":"addr"},{"name":"response_destination","tlb_type":"addr","format":"addr"},{"name":"custom_payload","tlb_type":"maybe ^","format":"cell"},{"name":"forward_amount","tlb_type":".","format":"coins"},{"name":"forward_payload","tlb_type":"either . ^","format":"cell"}]}`,
		}, {
			structType: (*NFTEditorshipAssigned)(nil),
			expected:   `{"op_name":"nft_editorship_assigned","op_code":"0x511a4463","body":[{"name":"query_id","tlb_type":"## 64","format":"uint64"},{"name":"prev_editor","tlb_type":"addr","format":"addr"},{"name":"forward_payload","tlb_type":"either . ^","format":"cell"}]}`,
		},
	}

	for _, test := range testCases {
		d, err := abi.NewOperationDesc(test.structType)
		require.Nil(t, err)
		got, err := json.Marshal(d)
		require.Nil(t, err)
		assert.Equal(t, test.expected, string(got))
	}
}

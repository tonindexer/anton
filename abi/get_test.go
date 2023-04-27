package abi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
)

func TestHasGetMethod(t *testing.T) {
	// https://ton.cx/address/EQAiZupbLhdE7UWQgnTirCbIJRg6yxfmkvTDjxsFh33Cu5rM
	codeBOC := mustBase64(t, "te6cckECDQEAAdAAART/APSkE/S88sgLAQIBYgIDAgLOBAUACaEfn+AFAgEgBgcCASALDALXDIhxwCSXwPg0NMDAXGwkl8D4PpA+kAx+gAxcdch+gAx+gAw8AIEs44UMGwiNFIyxwXy4ZUB+kDUMBAj8APgBtMf0z+CEF/MPRRSMLqOhzIQN14yQBPgMDQ0NTWCEC/LJqISuuMCXwSED/LwgCAkAET6RDBwuvLhTYAH2UTXHBfLhkfpAIfAB+kDSADH6AIIK+vCAG6EhlFMVoKHeItcLAcMAIJIGoZE24iDC//LhkiGOPoIQBRONkchQCc8WUAvPFnEkSRRURqBwgBDIywVQB88WUAX6AhXLahLLH8s/Im6zlFjPFwGRMuIByQH7ABBHlBAqN1viCgBycIIQi3cXNQXIy/9QBM8WECSAQHCAEMjLBVAHzxZQBfoCFctqEssfyz8ibrOUWM8XAZEy4gHJAfsAAIICjjUm8AGCENUydtsQN0QAbXFwgBDIywVQB88WUAX6AhXLahLLH8s/Im6zlFjPFwGRMuIByQH7AJMwMjTiVQLwAwA7O1E0NM/+kAg10nCAJp/AfpA1DAQJBAj4DBwWW1tgAB0A8jLP1jPFgHPFszJ7VSC/dQQb")

	code, err := cell.FromBOC(codeBOC)
	assert.Nil(t, err)

	h := abi.MethodNameHash("get_nft_data")
	assert.Equal(t, int32(0x18fcf), h)
	assert.True(t, abi.HasGetMethod(code, "get_nft_data"))
}

func TestGetMethodHashes(t *testing.T) {
	// https://ton.cx/address/EQAiZupbLhdE7UWQgnTirCbIJRg6yxfmkvTDjxsFh33Cu5rM
	codeBOC := mustBase64(t, "te6cckECDQEAAdAAART/APSkE/S88sgLAQIBYgIDAgLOBAUACaEfn+AFAgEgBgcCASALDALXDIhxwCSXwPg0NMDAXGwkl8D4PpA+kAx+gAxcdch+gAx+gAw8AIEs44UMGwiNFIyxwXy4ZUB+kDUMBAj8APgBtMf0z+CEF/MPRRSMLqOhzIQN14yQBPgMDQ0NTWCEC/LJqISuuMCXwSED/LwgCAkAET6RDBwuvLhTYAH2UTXHBfLhkfpAIfAB+kDSADH6AIIK+vCAG6EhlFMVoKHeItcLAcMAIJIGoZE24iDC//LhkiGOPoIQBRONkchQCc8WUAvPFnEkSRRURqBwgBDIywVQB88WUAX6AhXLahLLH8s/Im6zlFjPFwGRMuIByQH7ABBHlBAqN1viCgBycIIQi3cXNQXIy/9QBM8WECSAQHCAEMjLBVAHzxZQBfoCFctqEssfyz8ibrOUWM8XAZEy4gHJAfsAAIICjjUm8AGCENUydtsQN0QAbXFwgBDIywVQB88WUAX6AhXLahLLH8s/Im6zlFjPFwGRMuIByQH7AJMwMjTiVQLwAwA7O1E0NM/+kAg10nCAJp/AfpA1DAQJBAj4DBwWW1tgAB0A8jLP1jPFgHPFszJ7VSC/dQQb")

	code, err := cell.FromBOC(codeBOC)
	assert.Nil(t, err)

	hashes, err := abi.GetMethodHashes(code)
	assert.Nil(t, err)
	assert.Equal(t, []int32{0x18fcf}, hashes)
}

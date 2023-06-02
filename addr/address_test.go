package addr

import (
	"database/sql/driver"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress_TypeKind(t *testing.T) {
	a, err := new(Address).FromBase64("Ef8zMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzM0vF")
	assert.Nil(t, err)
	assert.Equal(t, int8(-1), a.Workchain())

	v := reflect.ValueOf(a)
	vt := v.Type()

	assert.Equal(t, reflect.Pointer, vt.Kind())
	assert.Equal(t, reflect.Array, vt.Elem().Kind())
	assert.Equal(t, reflect.Uint8, vt.Elem().Elem().Kind())
	assert.True(t, vt.Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()))

	r, err := v.Interface().(driver.Valuer).Value()
	assert.Nil(t, err)

	rb, ok := r.([]byte)
	assert.True(t, ok)

	t.Logf("%+v\n", rb)
}

func TestAddress_FromBase64(t *testing.T) {
	var testCases = []*struct {
		b64 string
		uf  string
	}{
		{
			b64: "EQCcLpOBWyOrE_mL9C2Ss4KZVxvwSSIjE6jOba69PeFHIgt1",
			uf:  "0:9c2e93815b23ab13f98bf42d92b38299571bf049222313a8ce6daebd3de14722",
		}, {
			b64: "EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N",
			uf:  "0:0e41dc1dc3c9067ed24248580e12b3359818d83dee0304fabcf80845eafafdb2",
		}, {
			b64: "Ef8zMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzM0vF",
			uf:  "-1:3333333333333333333333333333333333333333333333333333333333333333",
		},
	}

	for _, c := range testCases {
		addr, err := new(Address).FromBase64(c.b64)
		assert.Nil(t, err)

		addrStr := addr.String()
		assert.Equal(t, c.uf, addrStr)
		assert.Equal(t, c.b64, addr.Base64())

		addrGot, err := new(Address).FromString(addrStr)
		assert.Nil(t, err)
		assert.Equal(t, c.b64, addrGot.Base64())

		addrTU, err := addrGot.ToTonutils()
		assert.Nil(t, err)
		addrFromTU := MustFromTonutils(addrTU)
		assert.Equal(t, c.b64, addrFromTU.Base64())
	}
}

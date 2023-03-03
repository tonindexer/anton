package addr

import (
	"database/sql/driver"
	"reflect"
	"testing"
)

func TestAddress_TypeKind(t *testing.T) {
	a := MustFromBase64("Ef8zMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzM0vF")
	v := reflect.ValueOf(a)
	vt := v.Type()

	if vt.Kind() != reflect.Pointer {
		t.Fatal(vt.Kind())
	}
	if vt.Elem().Kind() != reflect.Array {
		t.Fatal(vt.Elem().Kind())
	}
	if vt.Elem().Elem().Kind() != reflect.Uint8 {
		t.Fatal(vt.Elem().Elem().Kind())
	}

	if !vt.Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
		t.Fatal()
	}

	r, err := v.Interface().(driver.Valuer).Value()
	if err != nil {
		t.Fatal(err)
	}
	rb, ok := r.([]byte)
	if !ok {
		t.Fatal()
	}

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
		var addr Address

		_, err := addr.FromBase64(c.b64)
		if err != nil {
			t.Fatal(c.b64, err)
		}

		if addr.String() != c.uf {
			t.Fatal(c.uf, addr.String())
		}
		if addr.Base64() != c.b64 {
			t.Fatal(c.b64, addr.Base64())
		}
	}
}

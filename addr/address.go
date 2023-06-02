package addr

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sigurn/crc16"

	"github.com/xssnick/tonutils-go/address"
)

// Address consists of flags (1 byte), workchain_id (1 byte) and account_id (32 byte).
// https://docs.ton.org/learn/overviews/addresses#user-friendly-address-structure
type Address [33]byte

var (
	_ json.Marshaler   = (*Address)(nil)
	_ json.Unmarshaler = (*Address)(nil)

	_ sql.Scanner   = (*Address)(nil)
	_ driver.Valuer = (*Address)(nil)

	_ fmt.Stringer = (*Address)(nil)
)

func (x *Address) Workchain() int8 {
	return int8(x[0])
}

func (x *Address) ToTonutils() (*address.Address, error) {
	return address.ParseAddr(x.Base64())
}

func (x *Address) MustToTonutils() *address.Address {
	a, err := x.ToTonutils()
	if err != nil {
		panic(errors.Wrapf(err, "%s", x.String()))
	}
	return a
}

func (x *Address) FromTonutils(addr *address.Address) (*Address, error) {
	if addr.Type() == address.NoneAddress {
		return nil, nil //nolint:nilnil // no address
	}
	if addr.Type() == address.ExtAddress {
		return nil, nil //nolint:nilnil // TODO: support ext addresses
	}
	if len(addr.Data()) != 32 {
		return nil, fmt.Errorf("wrong addr data length %d", addr.Data())
	}
	x[0] = byte(addr.Workchain())
	copy(x[1:33], addr.Data())
	return x, nil
}

func MustFromTonutils(a *address.Address) *Address {
	addr, err := new(Address).FromTonutils(a)
	if err != nil {
		panic(fmt.Errorf("%s to address: %w", a.String(), err))
	}
	return addr
}

var crcTable = crc16.MakeTable(crc16.CRC16_XMODEM)

func (x *Address) Checksum() uint16 {
	var xFlags [34]byte
	xFlags[0] = 0x11
	copy(xFlags[1:34], x[:])
	return crc16.Checksum(xFlags[:], crcTable)
}

func (x *Address) String() string {
	return fmt.Sprintf("%d:%x", int8(x[0]), x[1:33])
}

func (x *Address) FromString(str string) (*Address, error) {
	split := strings.Split(str, ":")
	if len(split) != 2 {
		return nil, fmt.Errorf("wrong address format")
	}
	w, err := strconv.ParseInt(split[0], 10, 8)
	if err != nil {
		return nil, errors.Wrap(err, "parse address workchain int8")
	}
	d, err := hex.DecodeString(split[1])
	if err != nil {
		return nil, errors.Wrap(err, "parse address data hex")
	}
	return x.FromTonutils(address.NewAddress(0, byte(w), d))
}

func MustFromString(str string) *Address {
	a, err := new(Address).FromString(str)
	if err != nil {
		panic(errors.Wrapf(err, "%s", str))
	}
	return a
}

func (x *Address) Base64() string {
	var xCheck [36]byte
	xCheck[0] = 0x11
	copy(xCheck[1:34], x[:])
	binary.BigEndian.PutUint16(xCheck[34:], x.Checksum())
	return base64.RawURLEncoding.EncodeToString(xCheck[:])
}

func (x *Address) FromBase64(b64 string) (*Address, error) {
	d, err := base64.RawURLEncoding.DecodeString(b64)
	if err != nil {
		return nil, errors.Wrap(err, "decode base64")
	}
	if len(d) != 36 {
		return nil, errors.New("wrong decoded address length")
	}

	copy(x[0:33], d[1:34])

	if x.Checksum() != binary.BigEndian.Uint16(d[34:36]) {
		return nil, errors.New("wrong address checksum")
	}

	return x, nil
}

func MustFromBase64(b64 string) *Address {
	addr, err := new(Address).FromBase64(b64)
	if err != nil {
		panic(fmt.Sprintf("%s to address", addr))
	}
	return addr
}

func (x *Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Hex    string `json:"hex"`
		Base64 string `json:"base64"`
	}{
		Hex:    x.String(),
		Base64: x.Base64(),
	})
}

func (x *Address) UnmarshalJSON(raw []byte) error {
	s := strings.Replace(string(raw), "\"", "", 2)
	if _, err := x.FromBase64(s); err == nil {
		return nil
	}
	if _, err := x.FromString(s); err == nil {
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s to address", s)
}

func (x *Address) UnmarshalText(data []byte) error {
	return x.UnmarshalJSON(data)
}

func (x *Address) Value() (driver.Value, error) {
	if x == nil {
		return nil, nil
	}
	none := true
	for _, i := range x {
		if i != 0 {
			none = false
			break
		}
	}
	if none {
		return nil, nil
	}
	return x[:], nil
}

func (x *Address) Scan(value interface{}) error {
	var i sql.NullString

	if value == nil {
		return nil
	}

	if err := i.Scan(value); err != nil {
		return err
	}
	if !i.Valid {
		return fmt.Errorf("error converting type %T into address", value)
	}
	if l := len(i.String); l != 33 {
		return fmt.Errorf("wrong address length %d", l)
	}

	copy(x[0:33], i.String)
	return nil
}

func Equal(x, y *Address) bool {
	if x != nil && y != nil && bytes.Equal(x[:], y[:]) {
		return true
	}
	return false
}

package uuid

import (
	"bytes"
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type UUID string

const size = 16

var (
	Nil        UUID
	byteGroups = []int{8, 4, 4, 4, 12}
)

var uuidRegex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func FromString(str string) (UUID, error) {
	if str == "" || str == "00000000-0000-0000-0000-000000000000" {
		return Nil, nil
	}

	if !uuidRegex.MatchString(str) {
		return Nil, errors.New("invalid uuid: " + str)
	}

	return UUID(strings.ToLower(str)), nil
}

func NewV4() UUID {
	u := [size]byte{}
	if _, err := io.ReadFull(rand.Reader, u[:]); err != nil {
		panic(err)
	}

	// set version to v4
	const v4 byte = 4
	u[6] = (u[6] & 0x0f) | (v4 << 4)
	// set variant to RFC4122
	u[8] = u[8]&(0xff>>2) | (0x02 << 6)

	return UUID(string(encodeBytes(u[:])))
}

func (u UUID) String() string {
	return string(u)
}

func (u UUID) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *UUID) UnmarshalText(text []byte) error {
	uid, err := FromString(string(text))
	if err != nil {
		return err
	}

	*u = uid

	return nil
}

func (u UUID) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(u.String())), nil
}

func (u *UUID) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		return nil
	}

	str, err := strconv.Unquote(string(b))
	if err != nil {
		return errors.New("invalid json value for uuid (must be string or null): " + string(b))
	}

	uid, err := FromString(str)
	if err != nil {
		return err
	}

	*u = uid

	return nil
}

func (u *UUID) UnmarshalBinary(data []byte) error {
	return u.UnmarshalText(data)
}

func (u UUID) MarshalBinary() (data []byte, err error) {
	return u.MarshalText()
}

func (u UUID) Value() (driver.Value, error) {
	if u == Nil {
		return nil, nil
	}

	if u[8] != '-' || u[13] != '-' || u[18] != '-' || u[23] != '-' {
		return nil, fmt.Errorf("uuid: incorrect UUID format %s", u)
	}

	// the backing array for the slice
	var ba [size]byte
	src := []byte(u)
	dst := ba[:]

	for i, byteGroup := range byteGroups {
		if i > 0 {
			src = src[1:] // skip dash
		}
		_, err := hex.Decode(dst[:byteGroup/2], src[:byteGroup])
		if err != nil {
			return nil, err
		}
		src = src[byteGroup:]
		dst = dst[byteGroup/2:]
	}

	// driver.Value support only slice, not array
	return ba[:], nil
}

func (u *UUID) Scan(src interface{}) error {
	if src == nil {
		*u = Nil
		return nil
	}

	if src, ok := src.([]byte); ok && len(src) == size {
		buf := encodeBytes(src)

		var err error
		*u, err = FromString(string(buf))

		return err
	}

	return fmt.Errorf("uuid: cannot convert %T to UUID", src)
}

func encodeBytes(u []byte) []byte {
	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])

	return buf
}

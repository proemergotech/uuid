package uuid

import (
	"bytes"
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type UUID string

const size = 16

var maxTime uint64
var bigPrime *big.Int

func init() {
	bigPrime = new(big.Int)
	// 14 bytes long prime number
	_ = bigPrime.UnmarshalText([]byte("908070605040302010203040506070809"))

	maxTimeUUID, _ := FromString("ffffffff-ffff-1000-a000-000000000000")
	t, _ := maxTimeUUID.TimeUUIDToTime()
	maxTime = Timestamp(t)
}

var (
	Nil        UUID
	byteGroups = []int{8, 4, 4, 4, 12}
)

var uuidRegex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// FromString parses uuid in canonical format, eg: afe40693-8f63-4766-85f1-250a427f1db5
func FromString(str string) (UUID, error) {
	if str == "" || str == "00000000-0000-0000-0000-000000000000" {
		return Nil, nil
	}

	if !uuidRegex.MatchString(str) {
		return Nil, errors.New("invalid uuid: " + str)
	}

	return UUID(strings.ToLower(str)), nil
}

// FromHashLike parses uuid in hash format, eg: afe406938f63476685f1250a427f1db5
func FromHashLike(str string) (UUID, error) {
	if str == "" || str == "00000000000000000000000000000000" {
		return Nil, nil
	}

	if len(str) != 32 {
		return Nil, errors.New("invalid uuid: " + str)
	}

	uuid := str[0:8] + "-" + str[8:12] + "-" + str[12:16] + "-" + str[16:20] + "-" + str[20:]
	if !uuidRegex.MatchString(uuid) {
		return Nil, errors.New("invalid uuid: " + str)
	}

	return UUID(strings.ToLower(uuid)), nil
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

func NewTime(t time.Time) UUID {
	u := [size]byte{}
	if _, err := io.ReadFull(rand.Reader, u[6:]); err != nil {
		panic(err)
	}

	ms := Timestamp(t)

	if ms > maxTime {
		panic("time too big")
	}

	const v4 byte = 4

	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)
	u[6] = (u[6] & 0x0f) | (v4 << 4)
	// set variant to RFC4122
	u[8] = u[8]&(0xff>>2) | (0x02 << 6)

	return UUID(string(encodeBytes(u[:])))
}

func (u UUID) String() string {
	return string(u)
}

// TimeUUIDToTime converts UUID into UTC time.
// @warning - Handle with care.
// If you use it for single UUID you will receive random/invalid timestamp.
func (u UUID) TimeUUIDToTime() (time.Time, error) {
	tmp, err := hex.DecodeString(u.HashLike())
	if err != nil {
		return time.Time{}.UTC(), err
	}

	ms := uint64(tmp[5]) | uint64(tmp[4])<<8 |
		uint64(tmp[3])<<16 | uint64(tmp[2])<<24 |
		uint64(tmp[1])<<32 | uint64(tmp[0])<<40

	return Time(ms), nil
}

func Timestamp(t time.Time) uint64 {
	return uint64(t.Unix())*1000 +
		uint64(t.Nanosecond()/int(time.Millisecond))
}

func Time(ms uint64) time.Time {
	s := int64(ms / 1e3)
	ns := int64((ms % 1e3) * 1e6)
	return time.Unix(s, ns).UTC()
}

// Next generates a new uuid from the current one. The uuid returned is consistent,
// meaning calling Next() on a given uuid will always return the same value.
func (u UUID) Next() (UUID, error) {
	if u == Nil {
		return Nil, nil
	}

	// remove dashes
	hash := u.HashLike()

	b, err := hex.DecodeString(hash)
	if err != nil {
		return Nil, errors.New("invalid uuid: " + u.String())
	}

	// skip 6th and 8th byte as they contain version and variant bits
	newB := appendAll(b[0:6], b[7:8], b[9:16])

	// add a big prime number (actually any odd number would work)
	bInt := new(big.Int)
	bInt.SetBytes(newB)
	bInt.Add(bInt, bigPrime)

	// len(newB) will never be less than 14 because bigPrime is 14 bytes long
	newB = bInt.Bytes()
	if len(newB) > 14 {
		// cut overflow
		newB = newB[len(newB)-14:]
	}

	// add back bytes containing version and variant bits
	newB = appendAll(newB[0:6], b[6:7], newB[6:7], b[8:9], newB[7:14])

	return UUID(string(encodeBytes(newB))), nil
}

func (u UUID) XOR(v UUID) (UUID, error) {
	if u == Nil || v == Nil {
		return Nil, nil
	}

	// remove dashes
	hash1 := u.HashLike()
	hash2 := v.HashLike()

	b1, err := hex.DecodeString(hash1)
	if err != nil {
		return Nil, errors.New("invalid left side parameter: " + u.String())
	}
	b2, err := hex.DecodeString(hash2)
	if err != nil {
		return Nil, errors.New("invalid right side parameter: " + v.String())
	}

	arr := make([]byte, 16)
	for i := range b1 {
		arr[i] = b1[i] ^ b2[i]
	}

	// set version to v4
	const v4 byte = 4
	arr[6] = (arr[6] & 0x0f) | (v4 << 4)
	// set variant to RFC4122
	arr[8] = arr[8]&(0xff>>2) | (0x02 << 6)

	return UUID(string(encodeBytes(arr))), nil
}

// HashLike returns the uuid without dashes, eg: afe406938f63476685f1250a427f1db5
func (u UUID) HashLike() string {
	if u == Nil {
		return ""
	}

	return string(u[0:8] + u[9:13] + u[14:18] + u[19:23] + u[24:])
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

func appendAll(bs ...[]byte) []byte {
	l := 0
	for _, b := range bs {
		l += len(b)
	}
	res := make([]byte, 0, l)
	for _, b := range bs {
		res = append(res, b...)
	}

	return res
}

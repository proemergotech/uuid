package uuid

import (
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/ugorji/go/codec"
)

var tests = map[string]string{
	"":                                     "",
	"00000000-0000-0000-0000-000000000000": "",
	"afe40693-8f63-4766-85f1-250a427f1db5": "afe40693-8f63-4766-85f1-250a427f1db5",
	"AFE40693-8F63-4766-85F1-250a427F1DB5": "afe40693-8f63-4766-85f1-250a427f1db5",
}

var testErrors = []string{
	"asda",
	"gfe40693-8f63-4766-85f1-250a427f1db5",
	"afe406938f63476685f1250a427f1db5",
	"99999999-9999-6999-9999-250a427f1db5", // invalid version bit
	"99999999-9999-4999-1999-250a427f1db5", // invalid variant bit
}

func TestBigPrimeLength(t *testing.T) {
	l := len(bigPrime.Bytes())
	if l != 14 {
		t.Errorf("expected bigPrime to be 14 bytes long, actual length: %v", l)
	}
}

func TestFromString(t *testing.T) {
	for orig, exp := range tests {
		uid, err := FromString(orig)
		if err != nil {
			t.Fatal(err)
		}

		str := uid.String()
		if str != exp {
			t.Errorf("expected: %s, got: %s", exp, str)
		}
	}
}

func TestFromStringError(t *testing.T) {
	for _, orig := range testErrors {
		_, err := FromString(orig)
		if err == nil {
			t.Errorf("expected error, but got nothing for %v", orig)
		}
	}
}

func TestJSON(t *testing.T) {
	for orig, exp := range tests {
		// 1. marshal as string
		// 2. unmarshal into UUID
		// 3. marshal the UUID
		// 5. unmarshal into string

		origB, err := json.Marshal(orig)
		if err != nil {
			t.Fatal(err)
		}

		var uid UUID
		err = json.Unmarshal(origB, &uid)
		if err != nil {
			t.Fatal(err)
		}

		b, err := json.Marshal(uid)
		if err != nil {
			t.Fatal(err)
		}

		var str string
		err = json.Unmarshal(b, &str)
		if err != nil {
			t.Fatal(err)
		}

		if exp != str {
			t.Errorf("expected: %s, got: %s", exp, b)
		}
	}
}

func TestJSONError(t *testing.T) {
	for _, orig := range testErrors {
		origB, err := json.Marshal(orig)
		if err != nil {
			t.Fatal(err)
		}

		var uid UUID
		err = json.Unmarshal(origB, &uid)
		if err == nil {
			t.Errorf("expected error, but got nothing for %v", orig)
		}
	}
}

func TestMsgPack(t *testing.T) {
	for orig, exp := range tests {
		// 1. marshal as string
		// 2. unmarshal into UUID
		// 3. marshal the UUID
		// 5. unmarshal into string

		handle := &codec.MsgpackHandle{}

		var origB []byte
		err := codec.NewEncoderBytes(&origB, handle).Encode(orig)
		if err != nil {
			t.Fatal(err)
		}

		var uid UUID
		err = codec.NewDecoderBytes(origB, handle).Decode(&uid)
		if err != nil {
			t.Fatal(err)
		}

		var b []byte
		err = codec.NewEncoderBytes(&b, handle).Encode(uid)
		if err != nil {
			t.Fatal(err)
		}

		var str string
		err = codec.NewDecoderBytes(b, handle).Decode(&str)
		if err != nil {
			t.Fatal(err)
		}

		if exp != str {
			t.Errorf("expected: %v, got: %v", exp, str)
		}
	}
}

func TestMsgPackError(t *testing.T) {
	for _, orig := range testErrors {
		handle := &codec.MsgpackHandle{}

		var origB []byte
		err := codec.NewEncoderBytes(&origB, handle).Encode(orig)
		if err != nil {
			t.Fatal(err)
		}

		var uid UUID
		err = codec.NewDecoderBytes(origB, handle).Decode(&uid)
		if err == nil {
			t.Errorf("expected error, but got nothing for %v", orig)
		}
	}
}

func TestSql(t *testing.T) {
	for orig, exp := range tests {
		t.Run(orig, func(t *testing.T) {
			origUUID := UUID(orig)
			driverValue, err := origUUID.Value()
			if err != nil {
				t.Fatal(err)
			}

			b, ok := driverValue.([]byte)
			if !ok && orig != "" {
				t.Fatalf("value does not returned with a byte slice, returned: %T", driverValue)
			}

			var scanValue UUID

			if len(b) == 0 {
				err = scanValue.Scan(nil)
			} else {
				err = scanValue.Scan(b)
			}

			if err != nil {
				t.Fatal(err)
			}

			if scanValue.String() != exp {
				t.Fatalf("expected: %v, got: %v", exp, scanValue.String())
			}
		})
	}
}

func TestSqlError(t *testing.T) {
	for _, orig := range testErrors {
		t.Run(orig, func(t *testing.T) {
			var scanValue UUID
			err := scanValue.Scan(&scanValue)
			if err == nil {
				t.Fatalf("expected error, but got nothing for %v", orig)
			}
		})
	}
}

func TestNewV4(t *testing.T) {
	const max = 100000

	uuids := make(map[UUID]struct{}, max)
	for i := 0; i < max; i++ {
		u := NewV4()
		if _, ok := uuids[u]; ok {
			t.Errorf("NewV4 returned same uuid twice: %s", u)
		}
		uuids[u] = struct{}{}

		uid, err := uuid.FromString(u.String())
		if err != nil {
			t.Error(err)
		}

		if uuid.V4 != uid.Version() {
			t.Errorf("invalid version in generated uuid: %s, expected: %v got: %v", u.String(), uuid.V4, uid.Version())
		}

		if uuid.VariantRFC4122 != uid.Variant() {
			t.Errorf("invalid variant in generated uuid: %s, expected: %v got: %v", u.String(), uuid.VariantRFC4122, uid.Variant())
		}
	}
}

func TestNewTimeUUID(t *testing.T) {
	for _, timestamp := range []uint64{
		0,
		100,
		1569479272,
		9999999999,
		99999999999999,
		281474976710655,
	} {
		u := NewTime(Time(timestamp))

		uid, err := uuid.FromString(u.String())
		if err != nil {
			t.Error(err)
		}

		if uuid.V4 != uid.Version() {
			t.Errorf("invalid version in generated uuid: %s, expected: %v got: %v", u.String(), uuid.V4, uid.Version())
		}

		if uuid.VariantRFC4122 != uid.Variant() {
			t.Errorf("invalid variant in generated uuid: %s, expected: %v got: %v", u.String(), uuid.VariantRFC4122, uid.Variant())
		}

		revertedTime, _ := u.TimeUUIDToTime()

		if timestamp != Timestamp(revertedTime) {
			t.Errorf("want: %v, got: %v", timestamp, revertedTime)
		}
	}
}

func TestFromHashLike(t *testing.T) {
	for _, data := range []struct {
		original string
		want     UUID
	}{
		{
			original: "",
			want:     "",
		},
		{
			original: "00000000000000000000000000000000",
			want:     "",
		},
		{
			original: "afe406938f63476685f1250a427f1db5",
			want:     "afe40693-8f63-4766-85f1-250a427f1db5",
		},
		{
			original: "AFE406938F63476685F1250a427F1DB5",
			want:     "afe40693-8f63-4766-85f1-250a427f1db5",
		},
	} {
		got, err := FromHashLike(data.original)
		if err != nil {
			t.Fatal(err)
		}

		if data.want != got {
			t.Errorf("want: %s, got: %s", data.want, got)
		}
	}
}

func TestFromHashLikeError(t *testing.T) {
	for _, data := range []struct {
		name     string
		original string
	}{
		{
			name:     "too short",
			original: "asda",
		},
		{
			name:     "too long",
			original: "afe406938f63476685f1250a427f1db51",
		},
		{
			name:     "invalid character",
			original: "gfe406938f63476685f1250a427f1db5",
		},
		{
			name:     "canonical form",
			original: "afe40693-8f63-4766-85f1-250a427f1db5",
		},
		{
			name:     "invalid version bit",
			original: "99999999999969999999250a427f1db5",
		},
		{
			name:     "invalid variant bit",
			original: "99999999999949991999250a427f1db5",
		},
	} {
		t.Run(data.name, func(t *testing.T) {
			_, err := FromHashLike(data.original)
			if err == nil {
				t.Errorf("expected error, but got nothing for %v", data.original)
			}
		})
	}
}

func TestHashLike(t *testing.T) {
	for _, data := range []struct {
		original UUID
		want     string
	}{
		{
			original: "",
			want:     "",
		},
		{
			original: "afe40693-8f63-4766-85f1-250a427f1db5",
			want:     "afe406938f63476685f1250a427f1db5",
		},
	} {
		got := data.original.HashLike()

		if data.want != got {
			t.Errorf("want: %s, got: %s", data.want, got)
		}
	}
}

func TestNext(t *testing.T) {
	count := 10000
	m := make(map[UUID]struct{}, count)
	var err error

	uid := UUID("afe40693-8f63-4766-85f1-250a427f1db5")
	for i := 0; i < count; i++ {
		if _, ok := m[uid]; ok {
			t.Fatal("duplicate uuid: " + uid)
		}

		if _, err = FromString(uid.String()); err != nil {
			t.Fatal(err)
		}

		m[uid] = struct{}{}
		uid, err = uid.Next()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestNextConsistent(t *testing.T) {
	uid := UUID("afe40693-8f63-4766-85f1-250a427f1db5")
	uuidNext1, err := uid.Next()
	if err != nil {
		t.Fatal(err)
	}

	uuidNext2, err := uid.Next()
	if err != nil {
		t.Fatal(err)
	}

	if uuidNext1 != uuidNext2 {
		t.Errorf("uuid is different, %v != %v", uuidNext1, uuidNext2)
	}
}

func TestXOR(t *testing.T) {
	a := UUID("afe40693-8f63-4766-85f1-250a427f1db5")
	b := UUID("43ae2f25-802d-4aae-be57-b7acefe336ac")

	aXb, err := a.XOR(b)
	if err != nil {
		t.Fatal(err)
	}
	bXa, err := b.XOR(a)
	if err != nil {
		t.Fatal(err)
	}
	if aXb != bXa {
		t.Errorf("a xor b is different from b xor a, %v != %v", aXb, bXa)
	}

	uid, err := uuid.FromString(aXb.String())
	if err != nil {
		t.Error(err)
	}

	if uuid.V4 != uid.Version() {
		t.Errorf("invalid version in generated uuid (a xor b): %s, expected: %v got: %v", aXb.String(), uuid.V4, uid.Version())
	}

	if uuid.VariantRFC4122 != uid.Variant() {
		t.Errorf("invalid variant in generated uuid (a xor b): %s, expected: %v got: %v", aXb.String(), uuid.VariantRFC4122, uid.Variant())
	}

	aXbXa, err := aXb.XOR(a)
	if err != nil {
		t.Fatal(err)
	}
	if aXbXa != b {
		t.Errorf("(a xor b) xor a is different from b, %v != %v", aXbXa, b)
	}
	aXbXb, err := aXb.XOR(b)
	if err != nil {
		t.Fatal(err)
	}
	if aXbXb != a {
		t.Errorf("(a xor b) xor b is different from a, %v != %v", aXbXb, a)
	}
}

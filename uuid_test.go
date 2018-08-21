package uuid

import (
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/ugorji/go/codec"
)

var tests = map[string]string{
	"": "",
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

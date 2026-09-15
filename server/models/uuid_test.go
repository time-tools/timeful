package models

import (
	"encoding/json"
	"testing"
)

const testUUID = "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e"

func marshalJSONString(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %T: %v", value, err)
	}
	return string(payload)
}

func TestZeroUUIDIsTheAllZeroSentinel(t *testing.T) {
	if got := ZeroUUID().String(); got != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("zero UUID = %s", got)
	}
	var unset UUID
	if got := unset.String(); got != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("unset UUID = %s, want the zero UUID sentinel", got)
	}
	if !unset.IsZero() {
		t.Fatal("expected the unset UUID to be zero")
	}
	if !ZeroUUID().IsZero() {
		t.Fatal("expected the zero UUID sentinel to be zero")
	}
}

func TestUUIDUnmarshalJSONAcceptsCanonicalForm(t *testing.T) {
	var value UUID
	if err := json.Unmarshal([]byte(`"`+testUUID+`"`), &value); err != nil {
		t.Fatalf("unmarshal canonical UUID: %v", err)
	}
	if value != UUID(testUUID) {
		t.Fatalf("decoded UUID = %q", value)
	}

	var empty UUID
	if err := json.Unmarshal([]byte(`""`), &empty); err != nil {
		t.Fatalf("unmarshal empty UUID: %v", err)
	}
	if !empty.IsZero() {
		t.Fatalf("empty UUID did not decode as zero: %q", empty)
	}

	var null UUID
	if err := json.Unmarshal([]byte(`null`), &null); err != nil {
		t.Fatalf("unmarshal null UUID: %v", err)
	}
	if !null.IsZero() {
		t.Fatalf("null UUID did not stay zero: %q", null)
	}
}

func TestUUIDUnmarshalJSONRejectsNonCanonicalForms(t *testing.T) {
	for name, encoded := range map[string]string{
		"uppercase":     `"0198E6F0-6A3A-7C4B-9A2D-4F6A1B2C3D4E"`,
		"compact":       `"0198e6f06a3a7c4b9a2d4f6a1b2c3d4e"`,
		"braced":        `"{0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e}"`,
		"extended-json": `{"$oid":"0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e"}`,
		"legacy-24-hex": `"507f1f77bcf86cd799439011"`,
		"non-uuid":      `"not-a-uuid"`,
	} {
		t.Run(name, func(t *testing.T) {
			var value UUID
			if err := json.Unmarshal([]byte(encoded), &value); err == nil {
				t.Fatalf("expected %s to be rejected, decoded %q", name, value)
			}
		})
	}
}

func TestUUIDMapKeysUseCanonicalForm(t *testing.T) {
	encoded := marshalJSONString(t, map[UUID]string{ZeroUUID(): "zero", UUID(testUUID): "account"})
	if encoded != `{"00000000-0000-0000-0000-000000000000":"zero","`+testUUID+`":"account"}` {
		t.Fatalf("UUID map keys = %s", encoded)
	}

	var decoded map[UUID]string
	if err := json.Unmarshal([]byte(encoded), &decoded); err != nil {
		t.Fatalf("decode UUID map: %v", err)
	}
	if decoded[ZeroUUID()] != "zero" || decoded[UUID(testUUID)] != "account" {
		t.Fatalf("decoded UUID map = %#v", decoded)
	}
}

func TestParseUUIDValidatesCanonicalForm(t *testing.T) {
	if parsed, ok := ParseUUID(testUUID); !ok || parsed != UUID(testUUID) {
		t.Fatalf("parsed UUID = %q, %v", parsed, ok)
	}
	for _, invalid := range []string{"", "507f1f77bcf86cd799439011", "0198E6F0-6a3a-7c4b-9a2d-4f6a1b2c3d4e", "0198e6f06a3a7c4b9a2d4f6a1b2c3d4e", "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4g"} {
		if _, ok := ParseUUID(invalid); ok {
			t.Fatalf("ParseUUID(%q) accepted an invalid value", invalid)
		}
	}
}

func TestNewUUIDIsCanonicalAndUnique(t *testing.T) {
	first := NewUUID()
	second := NewUUID()
	if _, ok := ParseUUID(first.String()); !ok {
		t.Fatalf("generated UUID %q is not canonical", first)
	}
	if first == second {
		t.Fatalf("generated UUIDs collide: %q", first)
	}
}

package models

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"strings"
)

// Handwritten sibling beside the generated uuid.go.
//
// UUID is transpiled from uuid.gala as `opaque type UUID string`. Hash and
// Compare suppress the synthesized std-backed methods so the generated Go stays
// runtime-free; every other member stays here because it crosses the GALA
// boundary (multi-value return, fixed-size [16]byte, or a pointer receiver over
// a text unmarshaler).

const zeroUUIDValue = "00000000-0000-0000-0000-000000000000"

// Hash hashes the canonical string with the FNV-1a mixing the synthesized
// method used.
func (id UUID) Hash() uint32 {
	value := string(id)
	var h uint32 = 2166136261
	for index := 0; index < len(value); index++ {
		h = h ^ uint32(value[index])
		h = h * 16777619
	}
	return h
}

// Compare orders two identifiers lexicographically, matching the synthesized
// method.
func (id UUID) Compare(other UUID) int {
	value := string(id)
	compare := string(other)
	if value < compare {
		return -1
	}
	if value > compare {
		return 1
	}
	return 0
}

// MarshalJSON emits the canonical lowercase hyphenated string, including the
// all-zero sentinel.
func (id UUID) MarshalJSON() ([]byte, error) { return json.Marshal(id.String()) }

// MarshalText emits the canonical form so UUID keeps working as a JSON map key.
func (id UUID) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

// UnmarshalJSON accepts a canonical lowercase hyphenated UUID string, an empty
// string (the zero sentinel), or null. Uppercase, compact, braced, and extended
// JSON forms are rejected so only the canonical wire form decodes.
func (id *UUID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == "" {
		*id = zeroUUIDValue
		return nil
	}
	value, ok := ParseUUID(raw)
	if !ok {
		return errors.New("invalid canonical UUID")
	}
	*id = value
	return nil
}

// UnmarshalText accepts the canonical form when UUID is a JSON map key. It
// rejects the empty string; only UnmarshalJSON maps empty values to the zero
// sentinel.
func (id *UUID) UnmarshalText(data []byte) error {
	value, ok := ParseUUID(string(data))
	if !ok {
		return errors.New("invalid canonical UUID")
	}
	*id = value
	return nil
}

// ParseUUID validates the canonical lowercase hyphenated RFC 9562 form and
// returns it unchanged. No uppercase, compact, braced, or extended variant is
// accepted.
func ParseUUID(value string) (UUID, bool) {
	if len(value) != 36 {
		return "", false
	}
	for index, char := range value {
		switch index {
		case 8, 13, 18, 23:
			if char != '-' {
				return "", false
			}
		default:
			if !isLowerHex(char) {
				return "", false
			}
		}
	}
	return UUID(value), true
}

func isLowerHex(char rune) bool {
	return (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')
}

// NewUUID returns a fresh random version-4 UUID. Account identifiers are minted
// by the platform_identities uuidv7() default, so account creation never calls
// it; tests and other non-account identifiers use it for uniqueness.
func NewUUID() UUID {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return UUID(formatUUID(value))
}

const lowerHexDigits = "0123456789abcdef"

func formatUUID(value [16]byte) string {
	var builder strings.Builder
	builder.Grow(36)
	for index, b := range value {
		switch index {
		case 4, 6, 8, 10:
			builder.WriteByte('-')
		}
		builder.WriteByte(lowerHexDigits[b>>4])
		builder.WriteByte(lowerHexDigits[b&0x0f])
	}
	return builder.String()
}

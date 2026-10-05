package models

import (
	"encoding/json"
	"time"
)

// Handwritten sibling beside the generated datetime.go.
//
// DateTime is transpiled from datetime.gala as `opaque type DateTime int64`.
// Hash and Compare suppress the synthesized std-backed methods so the generated
// Go stays runtime-free; MarshalJSON and UnmarshalJSON stay here because each
// one crosses the GALA boundary (multi-value return, pointer receiver).

// Hash hashes the instant with the FNV-1a int64 mixing the synthesized method
// used.
func (d DateTime) Hash() uint32 {
	var h uint32 = 2166136261
	n := int64(d)
	h = h ^ uint32(n&255)
	h = h * 16777619
	h = h ^ uint32((n>>8)&255)
	h = h * 16777619
	h = h ^ uint32((n>>16)&255)
	h = h * 16777619
	h = h ^ uint32((n>>24)&255)
	h = h * 16777619
	h = h ^ uint32((n>>32)&255)
	h = h * 16777619
	h = h ^ uint32((n>>40)&255)
	h = h * 16777619
	h = h ^ uint32((n>>48)&255)
	h = h * 16777619
	h = h ^ uint32((n>>56)&255)
	h = h * 16777619
	return h
}

// Compare orders two instants, negative when d is earlier.
func (d DateTime) Compare(other DateTime) int {
	if d < other {
		return -1
	}
	if d > other {
		return 1
	}
	return 0
}

// MarshalJSON emits an RFC3339 UTC timestamp.
func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time().UTC())
}

// UnmarshalJSON accepts an RFC3339 timestamp or null, matching the driver
// behavior the wire format was built on.
func (d *DateTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var value time.Time
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*d = NewDateTimeFromTime(value)
	return nil
}

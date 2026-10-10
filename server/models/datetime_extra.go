package models

import (
	"encoding/json"
	"time"
)

// Handwritten sibling beside the generated datetime.go.
//
// DateTime is transpiled from datetime.gala as `opaque type DateTime int64`.
// Hash and Compare are synthesized by the transpiler through the adopted
// go.gala.fyi/stdlib runtime. MarshalJSON and UnmarshalJSON stay here because
// each one crosses the GALA boundary (multi-value return, pointer receiver).

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

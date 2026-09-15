package models

import (
	"encoding/json"
	"time"
)

// DateTime is a canonical millisecond-precision instant. Its JSON representation
// is an RFC3339 string for values and a millisecond integer when used as a map
// key.
type DateTime int64

// Time returns the instant.
func (d DateTime) Time() time.Time { return time.UnixMilli(int64(d)) }

// IsZero reports whether the value is the Unix epoch default.
func (d DateTime) IsZero() bool { return d == 0 }

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

// NewDateTimeFromTime converts a time into millisecond precision.
func NewDateTimeFromTime(value time.Time) DateTime {
	return DateTime(value.UnixMilli())
}

package postgres

import (
	"regexp"
	"testing"
)

// TestFormatTransferCodePreservesLeadingZeros proves a numeric code keeps its
// full six-digit width instead of being trimmed when leading digits are zero.
func TestFormatTransferCodePreservesLeadingZeros(t *testing.T) {
	for _, test := range []struct {
		value uint32
		want  string
	}{
		{value: 0, want: "000000"},
		{value: 42, want: "000042"},
		{value: 999999, want: "999999"},
	} {
		if got := formatTransferCode(test.value); got != test.want {
			t.Fatalf("formatTransferCode(%d) = %q, want %q", test.value, got, test.want)
		}
	}
}

// TestGenerateTransferCode proves every generated code is exactly six decimal
// digits and never contains letters or other Crockford alphabet characters.
func TestGenerateTransferCode(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9]{6}$`)
	for i := 0; i < 256; i++ {
		code, err := GenerateTransferCode()
		if err != nil {
			t.Fatal(err)
		}
		if !pattern.MatchString(code) {
			t.Fatalf("code %q is not six decimal digits", code)
		}
	}
}

package eventsource

import "testing"

func TestCanonical(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{name: "canonical", id: "ABCD1234", want: true},
		{name: "legacy prefixed short ID", id: "m_ABCD1234", want: false},
		{name: "legacy prefixed long ID", id: "m_64f5e4d3c2b1a09876543210", want: false},
		{name: "bare long ID", id: "64f5e4d3c2b1a09876543210", want: false},
		{name: "legacy p_ prefix", id: "p_ABCD1234", want: false},
		{name: "empty", id: "", want: false},
		{name: "seven characters", id: "ABCD123", want: false},
		{name: "nine characters", id: "ABCD12345", want: false},
		{name: "excluded I", id: "ABCI1234", want: false},
		{name: "excluded L", id: "ABCL1234", want: false},
		{name: "excluded O", id: "ABCO1234", want: false},
		{name: "excluded U", id: "ABCU1234", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Canonical(test.id); got != test.want {
				t.Fatalf("Canonical(%q) = %v, want %v", test.id, got, test.want)
			}
		})
	}
}

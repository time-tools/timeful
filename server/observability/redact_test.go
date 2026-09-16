package observability

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRedact(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{
			name:  "benign message",
			input: "failed to load account integration data",
			want:  "failed to load account integration data",
		},
		{
			name:  "error code",
			input: "event-not-found",
			want:  "event-not-found",
		},
		{
			name:  "authorization header",
			input: "Authorization: Bearer supersecrettoken",
			want:  "Authorization: [REDACTED]",
		},
		{
			name:  "basic authorization header",
			input: "authorization=Basic dXNlcjpwYXNzd29yZA==",
			want:  "authorization=[REDACTED]",
		},
		{
			name:  "standalone bearer credential",
			input: "upstream rejected Bearer abcdef123456",
			want:  "upstream rejected Bearer [REDACTED]",
		},
		{
			name:  "cookie header",
			input: "Cookie: timeful_owner_ABCD1234=anonymous-edit-token",
			want:  "Cookie: [REDACTED]",
		},
		{
			name:  "owner token field",
			input: `{"editToken":"anonymous-edit-token"}`,
			want:  `{"editToken":[REDACTED]}`,
		},
		{
			name:  "password field",
			input: "password=hunter2",
			want:  "password=[REDACTED]",
		},
		{
			name:  "python style quoted password",
			input: `password: "hunter2"`,
			want:  "password: [REDACTED]",
		},
		{
			name:  "observe query parameters",
			input: "GET /api/auth/callback?code=oauth-code&state=kept",
			want:  "GET /api/auth/callback?code=[REDACTED]&state=kept",
		},
		{
			name:  "edit token query parameter",
			input: "GET /e/ABCD1234?editToken=anonymous-edit-token",
			want:  "GET /e/ABCD1234?editToken=[REDACTED]",
		},
		{
			name:  "access token query parameter",
			input: "https://api.example.com?access_token=secret-token&limit=10",
			want:  "https://api.example.com?access_token=[REDACTED]&limit=10",
		},
		{
			name:  "url userinfo",
			input: "dial https://alice:password@example.com/token failed",
			want:  "dial https://[REDACTED]@example.com/token failed",
		},
		{
			name:  "api key header",
			input: "X-Api-Key: secret-api-key",
			want:  "X-Api-Key: [REDACTED]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Redact(test.input); got != test.want {
				t.Fatalf("Redact(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 20); got != "short" {
		t.Fatalf("truncate(short) = %q", got)
	}
	long := "failed to reach the upstream service because the connection timed out after several retries"
	got := truncate(long, 40)
	if len(got) > 43 {
		t.Fatalf("truncate() length = %d, want bounded near 40", len(got))
	}
	if got[len(got)-3:] != "..." {
		t.Fatalf("truncate() = %q, want an ellipsis suffix", got)
	}
}

func TestTruncateKeepsValidUTF8(t *testing.T) {
	splitRune := strings.Repeat("a", 39) + "é" + "tail"
	got := truncate(splitRune, 40)
	if !utf8.ValidString(got) {
		t.Fatalf("truncate() split a rune into invalid UTF-8: %q", got)
	}
	if len(got) > 43 {
		t.Fatalf("truncate() length = %d, want bounded near 40", len(got))
	}
	if got[len(got)-3:] != "..." {
		t.Fatalf("truncate() = %q, want an ellipsis suffix", got)
	}

	invalid := "caf\xc3 est"
	got = truncate(invalid, 20)
	if !utf8.ValidString(got) {
		t.Fatalf("truncate() kept invalid UTF-8: %q", got)
	}
}

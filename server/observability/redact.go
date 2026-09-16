package observability

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const redactedValue = "[REDACTED]"

var (
	// URL userinfo, for example https://user:password@host/path.
	urlUserInfoPattern = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+.\-]*://)([^/@\s]+)@`)
	// Bearer and Basic scheme credentials.
	schemeCredentialPattern = regexp.MustCompile(`(?i)\b(bearer|basic)\s+[A-Za-z0-9._~+/=\-]{6,}`)
	// Credential-bearing headers and fields, including JSON forms. Unquoted
	// values stop at a query separator so following parameters survive.
	credentialFieldPattern = regexp.MustCompile(`(?i)\b(authorization|proxy-authorization|cookie|set-cookie|x-api-key|api[-_]?key|password|passwd|secret|token|access_token|refresh_token|id_token|edit_token|editToken|owner_token)\b("?)(\s*[:=]\s*)("[^"]*"|'[^']*'|[^&\n]+)`)
	// Token, secret, credential, and OAuth-code query parameters.
	sensitiveQueryPattern = regexp.MustCompile(`(?i)([?&][^=&\s#]*(?:token|secret|password|passwd|api[-_]?key|credential|auth|code|otp)[^=&\s#]*=)[^&#\s]*`)
)

// Redact removes credentials, secrets, and token-bearing URLs and headers from
// a diagnostic string. Structured records never carry raw requests, so this is
// the defense-in-depth pass over error context that originated in a handler.
func Redact(value string) string {
	if value == "" {
		return value
	}
	redacted := urlUserInfoPattern.ReplaceAllString(value, `${1}`+redactedValue+`@`)
	redacted = schemeCredentialPattern.ReplaceAllString(redacted, `${1} `+redactedValue)
	redacted = credentialFieldPattern.ReplaceAllString(redacted, `${1}${2}${3}`+redactedValue)
	redacted = sensitiveQueryPattern.ReplaceAllString(redacted, `${1}`+redactedValue)
	return redacted
}

// truncate bounds a redacted value so a single record can never grow without
// limit from captured error context. The result is always valid UTF-8, because
// an invalid string attribute makes protobuf marshalling fail for the whole
// export batch.
func truncate(value string, limit int) string {
	value = strings.ToValidUTF8(value, "\uFFFD")
	if len(value) <= limit {
		return value
	}
	trimmed := value[:limit]
	for len(trimmed) > 0 {
		r, size := utf8.DecodeLastRuneInString(trimmed)
		if r != utf8.RuneError || size > 1 {
			break
		}
		trimmed = trimmed[:len(trimmed)-1]
	}
	if index := strings.LastIndexAny(trimmed, " \t\n"); index > limit/2 {
		trimmed = trimmed[:index]
	}
	return trimmed + "..."
}

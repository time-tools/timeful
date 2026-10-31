package routes

import (
	"strings"
	"unicode/utf8"
)

// transferActionInput carries the transfer-action request body. Its tags define
// the request wire shape, and transfers.gala constructs it through
// newTransferActionInput instead of naming the local type.
type transferActionInput struct {
	RequestID            string `json:"requestId"`
	Code                 string `json:"code"`
	ConfirmAccountSwitch bool   `json:"confirmAccountSwitch"`
}

func newTransferActionInput() transferActionInput { return transferActionInput{} }

// transferUserAgent keeps the target browser's identifying header bounded
// before it is stored; an absent or blank header stores the empty string.
// The remainder of the function is a rune slice, which has no equivalent in
// GALA, so it stays handwritten.
func transferUserAgent(raw string) string {
	const maxRunes = 512
	value := strings.TrimSpace(raw)
	if utf8.RuneCountInString(value) <= maxRunes {
		return value
	}
	return string([]rune(value)[:maxRunes])
}

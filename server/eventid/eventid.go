// Package eventid validates canonical public event identifiers. The canonical
// form is an eight-character Crockford base32 identifier.
package eventid

import "regexp"

var canonicalShortID = regexp.MustCompile(`^[0-9A-HJKMNPQRSTVWXYZ]{8}$`)

// Canonical reports whether id is an eight-character Crockford base32 event
// identifier.
func Canonical(id string) bool {
	return canonicalShortID.MatchString(id)
}

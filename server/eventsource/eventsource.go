// Package eventsource validates canonical public event identifiers. Only
// eight-character Crockford base32 identifiers are served by the canonical
// event store.
package eventsource

import "regexp"

var canonicalShortID = regexp.MustCompile(`^[0-9A-HJKMNPQRSTVWXYZ]{8}$`)

// Canonical reports whether id is an eight-character Crockford base32
// identifier of the form served by the canonical event store.
func Canonical(id string) bool {
	return canonicalShortID.MatchString(id)
}

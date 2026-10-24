package postgres

import "time"

// DailyUserLog is a PostgreSQL-owned historical daily user log. LogDate is the
// account-local month/day/year at UTC midnight, and Members holds one row per
// account that signed in that day in first-seen order.
//
// The types stay handwritten because every GALA struct declaration emits Copy,
// Equal, Unapply, and StructMeta helpers that import the GALA runtime. This is
// a mixed-package sibling; see docs/GO_INTEROP.MD Part 3 in the GALA repository.
type DailyUserLog struct {
	ID      string
	LogDate time.Time
	Members []DailyUserLogMember
}

// DailyUserLogMember is one account's membership in a daily log. The profile
// fields are rebuilt from the authoritative accounts table at read time and are
// never stored on the log.
type DailyUserLogMember struct {
	PlatformIdentityID string
	FirstName          string
	LastName           string
	Email              string
	Position           int
}

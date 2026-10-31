// Calendar integrations are authoritative. calendar.gala is the explicit
// boundary that translates between the internal API user shape (models) and the
// repository types (pgstore), so no route handler, provider, or service reaches
// the database directly.
package accounts

import (
	"timeful/server/models"
)

// CalendarIntegrations is the authoritative calendar state for one account
// expressed in the existing internal user shape, so routes, calendar providers,
// and the frontend response contract are unchanged.
//
// The type and its constructor stay handwritten together: a GALA-declared struct
// would add its synthesized Copy, Equal, and Unapply members to this exported
// type, and the GALA file constructs it through newCalendarIntegrations so the
// transpiler never has to resolve the sibling's field list.
type CalendarIntegrations struct {
	Accounts          map[string]models.CalendarAccount
	PrimaryAccountKey *string
	TokenOrigin       models.TokenOriginType
	CalendarOptions   *models.CalendarOptions
}

// newCalendarIntegrations builds the integration set from an already-sized
// account map. calendar.gala calls it instead of naming the local type.
func newCalendarIntegrations(accounts map[string]models.CalendarAccount) *CalendarIntegrations {
	return &CalendarIntegrations{Accounts: accounts}
}

// CalendarPreferences carries the writable calendar preference fields for one
// account. A nil field is written as absent.
//
// The struct stays handwritten for the same synthesized-member reason as
// CalendarIntegrations.
type CalendarPreferences struct {
	PrimaryAccountKey *string
	TokenOrigin       models.TokenOriginType
	CalendarOptions   *models.CalendarOptions
}

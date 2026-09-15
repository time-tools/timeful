package models

// Representation of an authenticated user profile.
type User struct {
	TimezoneOffset int `json:"timezoneOffset"`

	// Profile info
	Id        UUID   `json:"_id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Picture   string `json:"picture"`

	// Whether the user has set a custom name for themselves, i.e. don't change their name when they sign in
	HasCustomName *bool `json:"hasCustomName"`

	// CalendarAccounts is a mapping from {`email_CALENDARTYPE` => CalendarAccount} that contains all the
	// additional accounts the user wants to see google calendar events for
	CalendarAccounts map[string]CalendarAccount `json:"calendarAccounts"`

	// The calendarAccountKey of the account the user first signed in with
	PrimaryAccountKey *string `json:"primaryAccountKey"`

	// Google OAuth stuff
	TokenOrigin TokenOriginType `json:"-"`

	// Calendar options
	CalendarOptions *CalendarOptions `json:"calendarOptions"`

	NumEventsCreated int `json:"numEventsCreated"`
}

// Declare the possible types of TokenOrigin
type TokenOriginType string

const (
	Undefined TokenOriginType = ""
	IOS       TokenOriginType = "ios"
	ANDROID   TokenOriginType = "android"
	WEB       TokenOriginType = "web"
)

type UserStatus string

const (
	FREE UserStatus = "free"
	BUSY UserStatus = "busy"
)

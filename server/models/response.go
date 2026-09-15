package models

// A response object containing an array of times that the given user is available
type Response struct {
	// Guest information
	Name               string `json:"name"`
	Email              string `json:"email"`
	GuestId            string `json:"guestId,omitempty"`
	GuestEditToken     string `json:"-"`
	GuestEditPolicy    string `json:"guestEditPolicy,omitempty"`
	GuestOwnershipMode string `json:"guestOwnershipMode,omitempty"`

	// User information
	UserId UUID  `json:"userId"`
	User   *User `json:"user"`

	// Availability
	Availability []DateTime `json:"availability"`
	IfNeeded     []DateTime `json:"ifNeeded"`

	// Mapping from the start date of a day to the available times for that day
	ManualAvailability *map[DateTime][]DateTime `json:"manualAvailability"`

	// Calendar availability variables for Availability Groups feature
	UseCalendarAvailability *bool                `json:"useCalendarAvailability"`
	EnabledCalendars        *map[string][]string `json:"enabledCalendars"` // Maps email to an array of sub calendar ids
	CalendarOptions         *CalendarOptions     `json:"calendarOptions"`
}

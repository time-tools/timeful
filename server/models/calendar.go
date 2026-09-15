package models

// CalendarType is an enum representing the type of calendar
type CalendarType string

const (
	AppleCalendarType   CalendarType = "apple"
	GoogleCalendarType  CalendarType = "google"
	OutlookCalendarType CalendarType = "outlook"
	ICSCalendarType     CalendarType = "ics"
)

// OAuth2CalendarAuth contains necessary auth info for the user's google calendar account
type OAuth2CalendarAuth struct {
	AccessToken           string   `json:"-"`
	AccessTokenExpireDate DateTime `json:"-"`
	RefreshToken          string   `json:"-"`
	Scope                 string   `json:"-"`
}

// AppleCalendarAuth contains necessary auth info for the user's apple calendar account
type AppleCalendarAuth struct {
	Email    string `json:"-"`
	Password string `json:"-"`
}

type ICSCalendarAuth struct {
	FeedURL string `json:"-"`
	Label   string `json:"label"`
}

// CalendarAccount contains info about the user's other signed in calendar accounts
type CalendarAccount struct {
	CalendarType       CalendarType        `json:"calendarType"`
	OAuth2CalendarAuth *OAuth2CalendarAuth `json:"oAuth2CalendarAuth"`
	AppleCalendarAuth  *AppleCalendarAuth  `json:"appleCalendarAuth"`
	ICSCalendarAuth    *ICSCalendarAuth    `json:"icsCalendarAuth"`

	Email        string                  `json:"email"` // Email is required for all calendar accounts
	Picture      string                  `json:"picture"`
	Enabled      *bool                   `json:"enabled"`
	SubCalendars *map[string]SubCalendar `json:"subCalendars"`
}

// SubCalendar represents a calendar within a calendar account
type SubCalendar struct {
	Name    string `json:"name"`
	Enabled *bool  `json:"enabled"`
}

// CalendarOptions contains options for calendar autofill
type CalendarOptions struct {
	BufferTime   BufferTimeOptions   `json:"bufferTime"`
	WorkingHours WorkingHoursOptions `json:"workingHours"`
}
type BufferTimeOptions struct {
	Enabled bool `json:"enabled"`
	Time    int  `json:"time"`
}
type WorkingHoursOptions struct {
	Enabled   bool    `json:"enabled"`
	StartTime float32 `json:"startTime"`
	EndTime   float32 `json:"endTime"`
}

// Simplified representation of a Calendar event from the calendar api
type CalendarEvent struct {
	Id         string   `json:"id"`
	CalendarId string   `json:"calendarId"`
	Summary    string   `json:"summary"`
	StartDate  DateTime `json:"startDate"`
	EndDate    DateTime `json:"endDate"`

	// Whether the user is free during this event
	Free bool `json:"free"`

	// Whether the event is an all day event
	AllDay bool `json:"allDay"`
}

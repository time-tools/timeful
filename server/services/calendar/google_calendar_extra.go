package calendar

import (
	"time"

	"timeful/server/errs"
	"timeful/server/models"
)

type GoogleCalendar struct {
	models.OAuth2CalendarAuth
}

type googleCalendarListItem struct {
	Id       string `json:"id"`
	Summary  string `json:"summary"`
	Selected bool   `json:"selected"`
}

type googleCalendarListResponse struct {
	Items []googleCalendarListItem `json:"items"`
	Error *errs.GoogleAPIError     `json:"error"`
}

type googleEventAttendee struct {
	Self           bool   `json:"self"`
	ResponseStatus string `json:"responseStatus"`
}

type googleEventTime struct {
	DateTime time.Time `json:"dateTime" binding:"required"`
	Date     string    `json:"date"`
}

type googleCalendarEventItem struct {
	Id           string                `json:"id"`
	Summary      string                `json:"summary"`
	Start        googleEventTime       `json:"start"`
	End          googleEventTime       `json:"end"`
	Transparency string                `json:"transparency"`
	Attendees    []googleEventAttendee `json:"attendees"`
}

type googleCalendarEventsResponse struct {
	Items []googleCalendarEventItem `json:"items"`
	Error *errs.GoogleAPIError      `json:"error"`
}

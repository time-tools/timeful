package calendar

import (
	"timeful/server/models"
)

type OutlookCalendar struct {
	models.OAuth2CalendarAuth
}

type outlookCalendarListValue struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type outlookCalendarListResponse struct {
	Value []outlookCalendarListValue `json:"value"`
	Error map[string]any             `json:"error"`
}

type outlookEventTime struct {
	DateTime string `json:"dateTime"`
}

type outlookCalendarEventItem struct {
	Id      string           `json:"id"`
	Subject string           `json:"subject"`
	Start   outlookEventTime `json:"start"`
	End     outlookEventTime `json:"end"`
	ShowAs  string           `json:"showAs"`
}

type outlookCalendarEventsResponse struct {
	Value []outlookCalendarEventItem `json:"value"`
	Error map[string]any             `json:"error"`
}

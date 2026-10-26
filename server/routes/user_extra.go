package routes

import (
	"time"

	"timeful/server/models"
)

// The request-body structs and the shared calendar-account argument struct stay
// handwritten because their tags define the request wire shape and a GALA struct
// declaration would add exported Copy/Equal/Unapply members; user.gala
// constructs each one through its constructor below instead of naming its field
// list.

// updateNamePayload carries the profile-name update body.
type updateNamePayload struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
}

func newUpdateNamePayload() updateNamePayload { return updateNamePayload{} }

// calendarOptionsPayload carries the calendar-options update body.
type calendarOptionsPayload struct {
	BufferTime   *models.BufferTimeOptions   `json:"bufferTime"`
	WorkingHours *models.WorkingHoursOptions `json:"workingHours"`
}

func newCalendarOptionsPayload() calendarOptionsPayload {
	return calendarOptionsPayload{}
}

// setEventFolderBody carries the set-folder request body.
type setEventFolderBody struct {
	FolderId *string `json:"folderId"`
}

func newSetEventFolderBody() setEventFolderBody { return setEventFolderBody{} }

// calendarsQuery carries the calendar-list query parameters.
type calendarsQuery struct {
	TimeMin  time.Time `form:"timeMin" binding:"required"`
	TimeMax  time.Time `form:"timeMax" binding:"required"`
	Accounts string    `form:"accounts"`
}

func newCalendarsQuery() calendarsQuery { return calendarsQuery{} }

// oauthCalendarCodePayload carries the Google and Outlook authorization body.
type oauthCalendarCodePayload struct {
	Code  string `json:"code" binding:"required"`
	Scope string `json:"scope" binding:"required"`
}

func newOAuthCalendarCodePayload() oauthCalendarCodePayload {
	return oauthCalendarCodePayload{}
}

// appleCalendarCredentialsPayload carries the Apple app-password body.
type appleCalendarCredentialsPayload struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func newAppleCalendarCredentialsPayload() appleCalendarCredentialsPayload {
	return appleCalendarCredentialsPayload{}
}

// icsCalendarCredentialsPayload carries the ICS feed body.
type icsCalendarCredentialsPayload struct {
	FeedURL string `json:"feedUrl" binding:"required"`
	Label   string `json:"label" binding:"required"`
}

func newICSCalendarCredentialsPayload() icsCalendarCredentialsPayload {
	return icsCalendarCredentialsPayload{}
}

// removeCalendarAccountPayload carries the remove-account body.
type removeCalendarAccountPayload struct {
	Email        string              `json:"email" binding:"required"`
	CalendarType models.CalendarType `json:"calendarType" binding:"required"`
}

func newRemoveCalendarAccountPayload() removeCalendarAccountPayload {
	return removeCalendarAccountPayload{}
}

// toggleCalendarPayload carries the toggle-calendar body.
type toggleCalendarPayload struct {
	Email        string              `json:"email" binding:"required"`
	CalendarType models.CalendarType `json:"calendarType" binding:"required"`
	Enabled      *bool               `json:"enabled" binding:"required"`
}

func newToggleCalendarPayload() toggleCalendarPayload { return toggleCalendarPayload{} }

// toggleSubCalendarPayload carries the toggle-sub-calendar body.
type toggleSubCalendarPayload struct {
	Email         string              `json:"email" binding:"required"`
	CalendarType  models.CalendarType `json:"calendarType" binding:"required"`
	SubCalendarId string              `json:"subCalendarId" binding:"required"`
	Enabled       *bool               `json:"enabled" binding:"required"`
}

func newToggleSubCalendarPayload() toggleSubCalendarPayload {
	return toggleSubCalendarPayload{}
}

// searchContactsQuery carries the contact-search query parameters.
type searchContactsQuery struct {
	Query string `form:"query"`
}

func newSearchContactsQuery() searchContactsQuery { return searchContactsQuery{} }

// deleteUserPayload carries the account-deletion confirmation body.
type deleteUserPayload struct {
	Email string `json:"email"`
}

func newDeleteUserPayload() deleteUserPayload { return deleteUserPayload{} }

// addCalendarAccountArgs carries the per-provider inputs of addCalendarAccount.
type addCalendarAccountArgs struct {
	calendarType       models.CalendarType
	oAuth2CalendarAuth *models.OAuth2CalendarAuth
	appleCalendarAuth  *models.AppleCalendarAuth
	icsCalendarAuth    *models.ICSCalendarAuth
	email              string
	picture            string
}

func newAddCalendarAccountArgs(
	calendarType models.CalendarType,
	oAuth2CalendarAuth *models.OAuth2CalendarAuth,
	appleCalendarAuth *models.AppleCalendarAuth,
	icsCalendarAuth *models.ICSCalendarAuth,
	email string,
	picture string,
) addCalendarAccountArgs {
	return addCalendarAccountArgs{
		calendarType:       calendarType,
		oAuth2CalendarAuth: oAuth2CalendarAuth,
		appleCalendarAuth:  appleCalendarAuth,
		icsCalendarAuth:    icsCalendarAuth,
		email:              email,
		picture:            picture,
	}
}

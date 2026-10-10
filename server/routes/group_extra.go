package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"timeful/server/accounts"
	"timeful/server/errs"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
	"timeful/server/responses"
	"timeful/server/services/calendar"
	"timeful/server/utils"
)

// The members in this file stay handwritten beside the generated group.go:
// GALA has no embedded-field syntax, so eventInput's promotion is not
// expressible; GALA has no struct tags, so groupAttendeePayload's wire shape is
// not expressible; and getCalendarAvailabilities fans out over a buffered
// channel with an in-place recover, which has no GALA spelling.

// eventInput carries the legacy group attendee invite list alongside the
// embedded event payload. Attendees are persisted in their own table rather than
// on models.Event, so the outer field collects the JSON key the event type no
// longer holds while the route applies the list separately.
type eventInput struct {
	models.Event
	Attendees []string `json:"attendees"`
}

// groupAttendeePayload is the attendee wire shape the group views and
// dashboard consume. It mirrors the former models.Attendee wire shape while
// using the canonical attendee UUID as _id.
type groupAttendeePayload struct {
	ID       string `json:"_id"`
	EventID  string `json:"eventId"`
	Email    string `json:"email"`
	Declined *bool  `json:"declined,omitempty"`
}

func groupAttendeePayloads(shortID string, attendees []pgstore.Attendee) []groupAttendeePayload {
	payloads := make([]groupAttendeePayload, 0, len(attendees))
	for _, attendee := range attendees {
		payloads = append(payloads, groupAttendeePayload{
			ID:       attendee.ID,
			EventID:  shortID,
			Email:    attendee.Email,
			Declined: attendee.Declined,
		})
	}
	return payloads
}

// declineBody is the named form of the anonymous struct declineInvite decodes,
// which GALA cannot declare because the json tag defines the wire shape.
type declineBody struct {
	Declined *bool `json:"declined"`
}

// newGuestForbidden wraps the sibling guestForbidden type so the .gala file can
// return it without constructing a type declared in handwritten Go.
func newGuestForbidden(message string) error { return guestForbidden{message} }

// newGuestNameError wraps the sibling guestNameError type for the same reason.
func newGuestNameError(message string) error { return guestNameError{message} }

// getCalendarAvailabilities resolves each group response's
// account through the authoritative account, then returns the
// response's enabled calendar events keyed by the opaque response publicId so
// clients can match them to the response map. Other members' event
// names are redacted, matching legacy behavior.
// @Summary Return a map mapping user id to their calendar events that they have enabled for the given time range
// @Tags events
// @Accept json
// @Produce json
// @Param eventId path string true "Event ID"
// @Param timeMin query string true "Lower bound for event's start time to filter by"
// @Param timeMax query string true "Upper bound for event's end time to filter by"
// @Success 200 {object} map[string]map[string]calendar.CalendarEventsWithError
// @Router /events/{eventId}/calendar-availabilities [get]
func getCalendarAvailabilities(c *gin.Context) {
	query := struct {
		TimeMin time.Time `form:"timeMin" binding:"required"`
		TimeMax time.Time `form:"timeMax" binding:"required"`
	}{}
	if err := c.BindQuery(&query); err != nil {
		return
	}
	repository := defaultRepository(c)
	if repository == nil {
		return
	}
	event := loadEvent(c, repository)
	if event == nil {
		return
	}
	if event.Type != pgstore.EventTypeGroup {
		c.JSON(http.StatusBadRequest, responses.Error{Error: errs.EventNotGroup})
		return
	}
	visitor, err := resolveVisitor(c, repository, event)
	if err != nil {
		mutationError(c, err)
		return
	}
	stored, err := repository.ListResponses(c.Request.Context(), event.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.Error{Error: "failed-to-load-responses"})
		return
	}
	type calendarRequest struct {
		publicID      string
		authorized    bool
		user          *models.User
		enabledSet    models.Set[string]
		enabledSubIDs models.Set[string]
	}
	requests := make([]calendarRequest, 0, len(stored))
	for _, response := range stored {
		if response.PlatformIdentityID == nil || *response.PlatformIdentityID == "" {
			continue
		}
		var value models.Response
		if err := json.Unmarshal(response.Payload, &value); err != nil {
			continue
		}
		if !utils.Coalesce(value.UseCalendarAvailability) {
			continue
		}
		user, err := accounts.LoadSessionUserByPlatformIdentityID(c.Request.Context(), *response.PlatformIdentityID)
		if err != nil {
			continue
		}
		enabledAccounts := make([]string, 0)
		enabledSubIDs := make([]string, 0)
		for accountKey, subIDs := range utils.Coalesce(value.EnabledCalendars) {
			enabledAccounts = append(enabledAccounts, accountKey)
			enabledSubIDs = append(enabledSubIDs, subIDs...)
		}
		authorized, err := visitor.controls(c.Request.Context(), repository, response.EventVisitorIdentityID)
		if err != nil {
			mutationError(c, err)
			return
		}
		requests = append(requests, calendarRequest{
			publicID:      response.PublicID,
			authorized:    authorized,
			user:          user,
			enabledSet:    utils.ArrayToSet(enabledAccounts),
			enabledSubIDs: utils.ArrayToSet(enabledSubIDs),
		})
	}
	type calendarResult struct {
		publicID string
		events   map[string]calendar.CalendarEventsWithError
	}
	results := make(chan calendarResult, len(requests))
	for _, request := range requests {
		request := request
		go func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					results <- calendarResult{publicID: request.publicID}
				}
			}()
			events, _ := calendar.GetUsersCalendarEvents(request.user, request.enabledSet, query.TimeMin, query.TimeMax)
			results <- calendarResult{publicID: request.publicID, events: events}
		}()
	}
	authorizedByPublicID := make(map[string]bool, len(requests))
	enabledByPublicID := make(map[string]models.Set[string], len(requests))
	for _, request := range requests {
		authorizedByPublicID[request.publicID] = request.authorized
		enabledByPublicID[request.publicID] = request.enabledSubIDs
	}
	result := make(map[string][]models.CalendarEvent, len(requests))
	for i := 0; i < len(requests); i++ {
		calendarEvents := <-results
		events := make([]models.CalendarEvent, 0)
		for _, entry := range calendarEvents.events {
			events = append(events, entry.CalendarEvents...)
		}
		filtered := make([]models.CalendarEvent, 0, len(events))
		for _, event := range events {
			if _, ok := enabledByPublicID[calendarEvents.publicID][event.CalendarId]; !ok {
				continue
			}
			if !authorizedByPublicID[calendarEvents.publicID] {
				event.Summary = "BUSY"
			}
			filtered = append(filtered, event)
		}
		result[calendarEvents.publicID] = filtered
	}
	c.JSON(http.StatusOK, result)
}

package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"timeful/server/accounts"
	"timeful/server/errs"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
	"timeful/server/respondents"
	"timeful/server/responses"
	"timeful/server/services/calendar"
	"timeful/server/services/listmonk"
	"timeful/server/utils"
)

const (
	postgresGroupInviteEmailTemplate = 9
	postgresGroupUpdateEmailTemplate = 11
)

// postgresEventInput carries the legacy group attendee invite list alongside the
// embedded event payload. Attendees are persisted in their own table rather than
// on models.Event, so the outer field collects the JSON key the event type no
// longer holds while the route applies the list separately.
type postgresEventInput struct {
	models.Event
	Attendees []string `json:"attendees"`
}

// postgresGroupAttendeePayload is the attendee wire shape the group views and
// dashboard consume. It mirrors the former models.Attendee wire shape while
// using the PostgreSQL attendee UUID as _id.
type postgresGroupAttendeePayload struct {
	ID       string `json:"_id"`
	EventID  string `json:"eventId"`
	Email    string `json:"email"`
	Declined *bool  `json:"declined,omitempty"`
}

func postgresGroupAttendeePayloads(shortID string, attendees []pgstore.Attendee) []postgresGroupAttendeePayload {
	payloads := make([]postgresGroupAttendeePayload, 0, len(attendees))
	for _, attendee := range attendees {
		payloads = append(payloads, postgresGroupAttendeePayload{
			ID:       attendee.ID,
			EventID:  shortID,
			Email:    attendee.Email,
			Declined: attendee.Declined,
		})
	}
	return payloads
}

// postgresAccountEmail resolves the signed-in account's email through the
// authoritative PostgreSQL boundary, adopting a legacy session once if needed.
func postgresAccountEmail(ctx context.Context, platformIdentityID string) string {
	if platformIdentityID == "" {
		return ""
	}
	account, err := accounts.Resolve(ctx, platformIdentityID)
	if err != nil || account == nil {
		return ""
	}
	return account.Email
}

// postgresGroupViewerIsInvitee reports whether the signed-in viewer is a
// non-declined member of the group, which is required to expose respondent
// emails for matching pending attendees to respondents. The membership check is
// a repository EXISTS over event_attendees (event_id, lower(email)) rather than
// an in-memory scan of the loaded attendee list.
func postgresGroupViewerIsInvitee(ctx context.Context, repository *pgstore.Repository, eventID string, viewer *postgresVisitor) bool {
	email := postgresAccountEmail(ctx, viewer.platformIdentityID)
	if email == "" {
		return false
	}
	invitee, err := repository.HasNonDeclinedAttendeeEmail(ctx, eventID, email)
	return err == nil && invitee
}

// postgresGroupEmailVisibility keeps respondent emails visible to the owner and
// non-declined invitees so clients can match pending attendees to respondents
// when collectEmails is off, mirroring legacy group behavior. PostgreSQL does
// not persist a denormalized account snapshot, so the response email is
// promoted into the rebuilt user snapshot at read time.
func postgresGroupEmailVisibility(ctx context.Context, repository *pgstore.Repository, eventID string, value models.Event, viewer *postgresVisitor, responseMap map[string]*postgresPublicResponse) {
	if responseMap == nil {
		return
	}
	showEmails := viewer.owner && utils.Coalesce(value.CollectEmails)
	keepGroupEmails := viewer.owner || postgresGroupViewerIsInvitee(ctx, repository, eventID, viewer)
	for key, response := range responseMap {
		if response == nil {
			continue
		}
		if keepGroupEmails && response.Email != "" && response.User == nil {
			response.User = &models.User{Email: response.Email}
		}
		stripSensitiveUserFields(response.User)
		if !showEmails {
			response.Email = ""
			if response.User != nil && !keepGroupEmails {
				response.User.Email = ""
			}
		}
		responseMap[key] = response
	}
}

// postgresGroupViewerHasResponded reports whether the calling visitor owns a
// response on the group, either through a signed-in account response or the
// browser Event Visitor Identity. It backs the derived hasResponded read.
func postgresGroupViewerHasResponded(ctx context.Context, repo *pgstore.Repository, event *pgstore.Event, viewer *postgresVisitor) bool {
	if viewer == nil {
		return false
	}
	if viewer.platformIdentityID != "" {
		if _, err := repo.GetResponseByPlatformIdentityID(ctx, event.ID, viewer.platformIdentityID); err == nil {
			return true
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return false
		}
	}
	if viewer.identity == nil {
		return false
	}
	hasResponded, err := repo.EventVisitorHasResponse(ctx, event.ID, viewer.identity.ID)
	if err != nil {
		return false
	}
	return hasResponded
}

// sendPostgresGroupInviteEmails sends the existing availability-group
// invitation email to each newly added invitee.
func sendPostgresGroupInviteEmails(ownerName, groupName, groupURL string, emails []string) {
	for _, email := range emails {
		if strings.TrimSpace(email) == "" {
			continue
		}
		listmonk.SendEmailAddSubscriberIfNotExist(email, postgresGroupInviteEmailTemplate, map[string]any{
			"ownerName": ownerName,
			"groupName": groupName,
			"groupUrl":  groupURL,
		}, false)
	}
}

// sendPostgresGroupUpdateEmails sends invitation emails to added members and
// the existing group update email to kept members, matching legacy ordering.
func sendPostgresGroupUpdateEmails(ownerName, groupName, groupURL string, added, kept []string) {
	sendPostgresGroupInviteEmails(ownerName, groupName, groupURL, added)
	if len(added) == 0 {
		return
	}
	for _, email := range kept {
		if strings.TrimSpace(email) == "" {
			continue
		}
		listmonk.SendEmailAddSubscriberIfNotExist(email, postgresGroupUpdateEmailTemplate, map[string]any{
			"ownerName": ownerName,
			"groupName": groupName,
			"groupUrl":  groupURL,
			"emails":    added,
		}, false)
	}
}

func postgresGroupURL(shortID string) string {
	return fmt.Sprintf("%s/g/%s", utils.GetBaseUrl(), shortID)
}

// postgresGroupEmailPlan carries the owner-facing email work computed while the
// event row is locked so it can be sent after the transaction commits.
type postgresGroupEmailPlan struct {
	ownerName string
	groupName string
	added     []string
	kept      []string
}

// postgresApplyGroupAttendeeEdits diffs the requested attendee emails against
// the stored membership, removes departed members' account responses so the
// response count stays correct, and returns the post-commit email plan. The
// owner membership is protected from removal. The caller already holds the
// event row lock and persists the adjusted response count.
func postgresApplyGroupAttendeeEdits(ctx context.Context, tx *pgstore.Repository, event *pgstore.Event, requested []string) (postgresGroupEmailPlan, error) {
	plan := postgresGroupEmailPlan{
		ownerName: postgresGroupOwnerName(ctx, event.OwnerPlatformIdentityID),
		groupName: event.Name,
	}
	current, err := tx.ListAttendees(ctx, event.ID)
	if err != nil {
		return plan, err
	}
	currentEmails := make([]string, 0, len(current))
	byEmail := make(map[string]pgstore.Attendee, len(current))
	for _, attendee := range current {
		currentEmails = append(currentEmails, attendee.Email)
		byEmail[strings.ToLower(strings.TrimSpace(attendee.Email))] = attendee
	}
	added, removed, kept := utils.FindAddedRemovedKept(requested, currentEmails)
	removedEmails := make([]string, 0, len(removed))
	removedIdentityIDs := make([]string, 0, len(removed))
	for _, item := range removed {
		attendee := byEmail[strings.ToLower(strings.TrimSpace(item.Value))]
		if attendee.PlatformIdentityID != nil && event.OwnerPlatformIdentityID != nil && *attendee.PlatformIdentityID == *event.OwnerPlatformIdentityID {
			continue
		}
		removedEmails = append(removedEmails, attendee.Email)
		if attendee.PlatformIdentityID != nil {
			removedIdentityIDs = append(removedIdentityIDs, *attendee.PlatformIdentityID)
		}
	}
	// One statement removes every departed account response and reports the
	// deleted rows so the in-memory response count stays exact. The caller
	// persists it with the settings payload. The decrement is clamped at zero
	// because the stored count can lag the response table; a stale counter must
	// not fail the settings write on the non-negative check.
	deleted, err := tx.DeleteAccountResponses(ctx, event.ID, removedIdentityIDs)
	if err != nil {
		return plan, err
	}
	if deleted > 0 {
		event.NumResponses -= int(deleted)
		if event.NumResponses < 0 {
			event.NumResponses = 0
		}
	}
	if err := tx.RemoveAttendees(ctx, event.ID, removedEmails); err != nil {
		return plan, err
	}
	addedEmails := make([]string, 0, len(added))
	for _, item := range added {
		if strings.TrimSpace(item.Value) == "" {
			continue
		}
		addedEmails = append(addedEmails, item.Value)
		plan.added = append(plan.added, item.Value)
	}
	if err := tx.AddAttendees(ctx, event.ID, addedEmails, utils.FalsePtr()); err != nil {
		return plan, err
	}
	for _, item := range kept {
		plan.kept = append(plan.kept, item.Value)
	}
	return plan, nil
}

// postgresGroupOwnerName resolves the owner display name used in group emails.
func postgresGroupOwnerName(ctx context.Context, ownerPlatformIdentityID *string) string {
	if ownerPlatformIdentityID == nil || *ownerPlatformIdentityID == "" {
		return "Somebody"
	}
	account, err := accounts.Resolve(ctx, *ownerPlatformIdentityID)
	if err != nil || account == nil || account.FirstName == "" {
		return "Somebody"
	}
	return account.FirstName
}

// postgresDeclineInvite sets the attendee decline state for the signed-in
// member of a PostgreSQL group. An optional {"declined": false} body covers
// undecline.
// @Summary Decline the current user's invite to the event
// @Tags events
// @Accept json
// @Produce json
// @Param eventId path string true "Event ID"
// @Success 200
// @Router /events/{eventId}/decline [post]
func postgresDeclineInvite(c *gin.Context) {
	declined := true
	if body, err := io.ReadAll(c.Request.Body); err == nil && len(bytes.TrimSpace(body)) > 0 {
		var input struct {
			Declined *bool `json:"declined"`
		}
		if json.Unmarshal(body, &input) == nil && input.Declined != nil {
			declined = *input.Declined
		}
	}
	repository := postgresRepository(c)
	if repository == nil {
		return
	}
	event := postgresEvent(c, repository)
	if event == nil {
		return
	}
	if event.Type != pgstore.EventTypeGroup {
		c.JSON(http.StatusBadRequest, responses.Error{Error: errs.EventNotGroup})
		return
	}
	userInterface, _ := c.Get("authUser")
	user, _ := userInterface.(*models.User)
	if user == nil {
		c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.NotSignedIn})
		return
	}
	if _, err := repository.GetAttendeeByEmail(c.Request.Context(), event.ID, user.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, responses.Error{Error: errs.AttendeeEmailNotFound})
			return
		}
		postgresMutationError(c, err)
		return
	}
	if err := repository.SetAttendeeDeclined(c.Request.Context(), event.ID, user.Email, declined); err != nil {
		postgresMutationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// groupManualAvailability is the day-window map an availability group stores per
// response. Each key is the millisecond instant a day starts and each value
// holds the available instants inside that day.
type groupManualAvailability = map[models.DateTime][]models.DateTime

// decodeGroupManualAvailability normalizes the two JSON encodings the group
// transport can produce: the legacy millisecond-keyed map whose values are
// RFC3339 strings, and the frontend ZonedDateTime-keyed encoding whose values
// are epoch milliseconds. Omitted or null input yields a nil map so an update
// preserves the stored availability.
func decodeGroupManualAvailability(raw json.RawMessage) (groupManualAvailability, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, nil
	}
	var encoded map[string][]json.RawMessage
	if err := json.Unmarshal(trimmed, &encoded); err != nil {
		return nil, err
	}
	result := make(groupManualAvailability, len(encoded))
	for key, values := range encoded {
		day, err := parseGroupAvailabilityInstant(key)
		if err != nil {
			return nil, err
		}
		instants := make([]models.DateTime, 0, len(values))
		for _, value := range values {
			instant, ok, err := parseGroupAvailabilityValue(value)
			if err != nil {
				return nil, err
			}
			if ok {
				instants = append(instants, instant)
			}
		}
		result[day] = instants
	}
	return result, nil
}

func parseGroupAvailabilityInstant(value string) (models.DateTime, error) {
	trimmed := strings.TrimSpace(value)
	if millis, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return models.DateTime(millis), nil
	}
	instant, err := parseGroupAvailabilityTime(trimmed)
	if err != nil {
		return 0, err
	}
	return models.NewDateTimeFromTime(instant), nil
}

func parseGroupAvailabilityValue(raw json.RawMessage) (models.DateTime, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return 0, false, nil
	}
	var millis int64
	if err := json.Unmarshal(trimmed, &millis); err == nil {
		return models.DateTime(millis), true, nil
	}
	var encoded string
	if err := json.Unmarshal(trimmed, &encoded); err != nil {
		return 0, false, err
	}
	instant, err := parseGroupAvailabilityTime(encoded)
	if err != nil {
		return 0, false, err
	}
	return models.NewDateTimeFromTime(instant), true, nil
}

// parseGroupAvailabilityTime accepts RFC3339 instants and Temporal's RFC9557
// ZonedDateTime string form whose trailing [timeZone] annotation Go cannot parse
// on its own.
func parseGroupAvailabilityTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if index := strings.LastIndex(trimmed, "["); index != -1 {
		trimmed = trimmed[:index]
	}
	return time.Parse(time.RFC3339Nano, trimmed)
}

// mergeGroupManualAvailability applies the legacy day-window replacement: an
// existing day inside a payload day's [start, start+window) span is replaced by
// the payload day, and payload days that replace nothing are appended. Deletions
// during iteration match the legacy map behavior so partially overlapping days
// resolve the same way.
func mergeGroupManualAvailability(window time.Duration, existing, incoming groupManualAvailability) groupManualAvailability {
	merged := make(groupManualAvailability, len(existing)+len(incoming))
	for day, times := range existing {
		merged[day] = times
	}
	replacements := make(groupManualAvailability, len(incoming))
	for day, times := range incoming {
		replacements[day] = times
	}
	for day := range merged {
		for payloadDay, availableTimes := range replacements {
			endTime := payloadDay.Time().Add(window)
			if day.Time().Compare(payloadDay.Time()) >= 0 && day.Time().Compare(endTime) <= 0 {
				delete(merged, day)
				merged[payloadDay] = availableTimes
				delete(replacements, payloadDay)
				break
			}
		}
		if len(replacements) == 0 {
			break
		}
	}
	for day, times := range replacements {
		merged[day] = times
	}
	return merged
}

// postgresGroupDurationHours derives the availability group's legacy duration
// from the canonical slot-generation window. The group editor defines the
// duration as the wrapped local-time window (a 09:00-17:00 group is eight
// hours), and the manual availability day-window merge spans that duration
// because the transport no longer carries the legacy duration field.
func postgresGroupDurationHours(generation *models.SlotGeneration) *float32 {
	if generation == nil {
		return nil
	}
	start, err := parseLocalTime(generation.StartTimeLocal)
	if err != nil {
		return nil
	}
	end, err := parseLocalTime(generation.EndTimeLocal)
	if err != nil {
		return nil
	}
	duration := end.Sub(start)
	if duration <= 0 {
		duration += 24 * time.Hour
	}
	hours := float32(duration.Hours())
	return &hours
}

// canonicalGroupResponseName resolves the guest display name for a group
// mutation, preferring the supplied name and falling back to the stored name so
// an explicit-selection edit may omit the unchanged name.
func canonicalGroupResponseName(supplied string, value *models.Response) string {
	if name := canonicalGuestName(supplied); name != "" {
		return name
	}
	if value == nil {
		return ""
	}
	return canonicalGuestName(value.Name)
}

// postgresSetGroupDecline writes the attendee decline state for the respondent's
// account email. A respondent with no resolved account, or who is not an
// attendee, is left untouched. Responding clears the decline state and leaving
// sets it, matching legacy group behavior.
func postgresSetGroupDecline(ctx context.Context, tx *pgstore.Repository, eventID string, stored *pgstore.Response, visitor *postgresVisitor, declined bool) error {
	email := ""
	if visitor != nil && visitor.platformIdentityID != "" {
		email = postgresAccountEmail(ctx, visitor.platformIdentityID)
	}
	if email == "" && stored != nil && stored.PlatformIdentityID != nil {
		email = postgresAccountEmail(ctx, *stored.PlatformIdentityID)
	}
	if email == "" {
		return nil
	}
	if err := tx.SetAttendeeDeclined(ctx, eventID, email, declined); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	return nil
}

// postgresMutateGroupResponse applies the explicit-selection contract to an
// availability group. Account respondents persist with their account identity;
// anonymous respondents persist with the shared canonical guest name. Saving a
// response clears the respondent's attendee decline state and deleting it sets
// the decline state (leaving the group). Calendar-derived mode, selected
// calendars, copied calendar preferences, and the day-window merged manual
// availability are persisted in the response payload.
func postgresMutateGroupResponse(c *gin.Context, repository *pgstore.Repository, event *pgstore.Event, visitor *postgresVisitor, input postgresResponseInput, operation string) {
	publicID := input.ResponseID
	manualAvailability, err := decodeGroupManualAvailability(input.ManualAvailability)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.Error{Error: "invalid-manual-availability"})
		return
	}
	eventModel, err := postgresEventModel(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.Error{Error: "failed-to-serialize-event"})
		return
	}
	manualWindow := time.Duration(0)
	if eventModel.Duration != nil {
		manualWindow = time.Duration(*eventModel.Duration) * time.Hour
	}
	err = repository.WithTransaction(c.Request.Context(), func(ctx context.Context, tx *pgstore.Repository) error {
		locked, err := tx.LockEvent(ctx, event.ID)
		if err != nil {
			return err
		}
		if err := postgresWritableEvent(locked); err != nil {
			return err
		}
		stored := &pgstore.Response{EventID: event.ID, EventVisitorIdentityID: visitor.identity.ID}
		var value *models.Response
		if input.CreateResponse {
			if !visitor.authorized {
				return guestForbidden{"visitor-credential-required"}
			}
		} else {
			stored, err = tx.GetResponseByPublicID(ctx, event.ID, input.ResponseID)
			if err != nil {
				return err
			}
			authorized, err := visitor.controls(ctx, tx, stored.EventVisitorIdentityID)
			if err != nil {
				return err
			}
			if !authorized {
				return guestForbidden{"response-credential-required"}
			}
			value, _, err = postgresResponseModel(*stored)
			if err != nil {
				return err
			}
		}
		if value == nil {
			value = &models.Response{}
		}
		if operation == "delete" {
			if err := tx.DeleteResponse(ctx, stored.ID); err != nil {
				return err
			}
			if err := tx.AdjustEventResponseCount(ctx, locked.ID, -1); err != nil {
				return err
			}
			return postgresSetGroupDecline(ctx, tx, locked.ID, stored, visitor, true)
		}
		if operation == "rename" {
			validated := respondents.ValidateGuestName(input.NewName)
			if validated.Code != respondents.GuestNameValid {
				return guestNameError{guestNameValidationErrorMessage(validated.Code)}
			}
			value.Name = validated.Name
		} else {
			if input.CreateResponse {
				if visitor.platformIdentityID != "" {
					platformIdentityID := visitor.platformIdentityID
					stored.RespondentKind = pgstore.RespondentKindAccount
					stored.PlatformIdentityID = &platformIdentityID
					value.Email = postgresAccountEmail(ctx, visitor.platformIdentityID)
				} else {
					if err := applyPostgresGroupGuestName(stored, value, input.Name); err != nil {
						return err
					}
					value.Email = input.Email
				}
			} else {
				switch stored.RespondentKind {
				case pgstore.RespondentKindAccount:
					// Reattach the visitor identity only when the caller has one;
					// an anonymous credential holder editing a legacy account
					// response must not write an empty uuid.
					if (stored.PlatformIdentityID == nil || *stored.PlatformIdentityID == "") && visitor.platformIdentityID != "" {
						platformIdentityID := visitor.platformIdentityID
						stored.PlatformIdentityID = &platformIdentityID
					}
					if stored.PlatformIdentityID != nil && *stored.PlatformIdentityID != "" {
						value.Email = postgresAccountEmail(ctx, *stored.PlatformIdentityID)
					}
				default:
					if err := applyPostgresGroupGuestName(stored, value, input.Name); err != nil {
						return err
					}
					value.Email = input.Email
				}
			}
			value.Availability, value.IfNeeded = normalizeTimedResponseAvailabilitySlots(input.Availability, input.IfNeeded)
			existing := groupManualAvailability(nil)
			if value.ManualAvailability != nil {
				existing = groupManualAvailability(*value.ManualAvailability)
			}
			merged := mergeGroupManualAvailability(manualWindow, existing, manualAvailability)
			value.ManualAvailability = &merged
			value.UseCalendarAvailability = input.UseCalendarAvailability
			value.EnabledCalendars = input.EnabledCalendars
			value.CalendarOptions = input.CalendarOptions
		}
		stored.Payload, err = json.Marshal(value)
		if err != nil {
			return err
		}
		if input.CreateResponse {
			if err := tx.CreateResponse(ctx, stored); err != nil {
				return err
			}
			publicID = stored.PublicID
			if err := tx.AdjustEventResponseCount(ctx, locked.ID, 1); err != nil {
				return err
			}
		} else if err := tx.UpdateResponse(ctx, stored); err != nil {
			return err
		}
		if operation == "save" {
			return postgresSetGroupDecline(ctx, tx, locked.ID, stored, visitor, false)
		}
		return nil
	})
	if err != nil {
		postgresMutationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"responseId": publicID, "eventVisitorId": visitor.identity.PublicID})
}

func applyPostgresGroupGuestName(stored *pgstore.Response, value *models.Response, supplied string) error {
	validated := respondents.ValidateGuestName(canonicalGroupResponseName(supplied, value))
	if validated.Code != respondents.GuestNameValid {
		return guestNameError{guestNameValidationErrorMessage(validated.Code)}
	}
	stored.RespondentKind = pgstore.RespondentKindGuest
	stored.PlatformIdentityID = nil
	value.Name = validated.Name
	return nil
}

// postgresGetCalendarAvailabilities resolves each PostgreSQL group response's
// account through the authoritative PostgreSQL account, then returns the
// response's enabled calendar events keyed by the opaque response publicId so
// clients can match them to the PostgreSQL response map. Other members' event
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
func postgresGetCalendarAvailabilities(c *gin.Context) {
	query := struct {
		TimeMin time.Time `form:"timeMin" binding:"required"`
		TimeMax time.Time `form:"timeMax" binding:"required"`
	}{}
	if err := c.BindQuery(&query); err != nil {
		return
	}
	repository := postgresRepository(c)
	if repository == nil {
		return
	}
	event := postgresEvent(c, repository)
	if event == nil {
		return
	}
	if event.Type != pgstore.EventTypeGroup {
		c.JSON(http.StatusBadRequest, responses.Error{Error: errs.EventNotGroup})
		return
	}
	visitor, err := resolvePostgresVisitor(c, repository, event)
	if err != nil {
		postgresMutationError(c, err)
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
			postgresMutationError(c, err)
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

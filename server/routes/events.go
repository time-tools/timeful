/* The /events group contains all the routes to get and edit events */
package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"timeful/server/middleware"
	"timeful/server/models"
)

func rejectLegacyTimedScheduleFields(c *gin.Context) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	daysOnly := false
	if rawDaysOnly, exists := payload["daysOnly"]; exists {
		if err := json.Unmarshal(rawDaysOnly, &daysOnly); err != nil {
			return err
		}
	}
	legacyFields := []string{"duration", "times", "timeIncrement", "hasSpecificTimes", "startOnMonday"}
	if !daysOnly {
		legacyFields = append(legacyFields, "dates")
	}
	for _, field := range legacyFields {
		if _, exists := payload[field]; exists {
			return fmt.Errorf("legacy-timed-event-field:%s", field)
		}
	}
	return nil
}

func InitEvents(router *gin.RouterGroup) {
	eventRouter := router.Group("/events")

	eventRouter.POST("", createEvent)
	eventRouter.POST("/:eventId/transfers", createTransfer)
	eventRouter.POST("/:eventId/transfers/:transferId/:action", transferAction)
	eventRouter.POST("/:eventId/grant-association", grantAssociation)
	eventRouter.PUT("/:eventId", editEvent)
	eventRouter.GET("/:eventId/ids", getEventIDs)
	eventRouter.GET("/:eventId", getEvent)
	eventRouter.GET("/:eventId/responses", getResponses)
	eventRouter.POST("/:eventId/response", updateResponse)
	eventRouter.DELETE("/:eventId/response", deleteResponse)
	eventRouter.PUT("/:eventId/schedule", saveSchedule)
	eventRouter.DELETE("/:eventId/schedule", clearSchedule)
	eventRouter.POST("/:eventId/rename-user", renameUser)
	eventRouter.POST("/:eventId/decline", middleware.AuthRequired(), declineInvite)
	eventRouter.GET("/:eventId/calendar-availabilities", middleware.AuthRequired(), getCalendarAvailabilities)
	eventRouter.DELETE("/:eventId", deleteEvent)
	eventRouter.POST("/:eventId/archive", archiveEvent)
}

func normalizeTimedResponseAvailabilitySlots(
	availability []models.DateTime,
	ifNeeded []models.DateTime,
) ([]models.DateTime, []models.DateTime) {
	normalizedAvailability := make([]models.DateTime, 0, len(availability))
	availabilitySet := make(map[models.DateTime]struct{}, len(availability))
	for _, slot := range availability {
		if _, exists := availabilitySet[slot]; exists {
			continue
		}
		availabilitySet[slot] = struct{}{}
		normalizedAvailability = append(normalizedAvailability, slot)
	}

	normalizedIfNeeded := make([]models.DateTime, 0, len(ifNeeded))
	ifNeededSet := make(map[models.DateTime]struct{}, len(ifNeeded))
	for _, slot := range ifNeeded {
		if _, exists := availabilitySet[slot]; exists {
			continue
		}
		if _, exists := ifNeededSet[slot]; exists {
			continue
		}
		ifNeededSet[slot] = struct{}{}
		normalizedIfNeeded = append(normalizedIfNeeded, slot)
	}

	return normalizedAvailability, normalizedIfNeeded
}

// stripSensitiveUserFields removes fields from a User that should never be
// exposed in the event page API response (calendar accounts, etc.).
// Email is NOT stripped here as callers handle email visibility separately based
// on the collectEmails setting and owner status.
func stripSensitiveUserFields(user *models.User) {
	if user == nil {
		return
	}
	user.CalendarAccounts = nil
	user.CalendarOptions = nil
	user.PrimaryAccountKey = nil
}

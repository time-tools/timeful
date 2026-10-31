package routes

import (
	"timeful/server/models"
)

// timedEventPayloadFields is the canonical timed-event payload. The struct stays
// handwritten because its JSON tags define the stored and wire shape, which a
// GALA Codec cannot reproduce.
type timedEventPayloadFields struct {
	ActiveSlots     []models.DateTime       `json:"activeSlots"`
	EventTimezone   *string                 `json:"eventTimezone"`
	SlotGeneration  *models.SlotGeneration  `json:"slotGeneration"`
	TimedRecurrence *models.TimedRecurrence `json:"timedRecurrence"`
}

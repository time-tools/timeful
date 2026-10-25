package routes

import (
	"timeful/server/models"
)

// normalizeTimedResponseAvailabilitySlots stays handwritten because its dedup
// sets are map[models.DateTime]struct{}, which is the struct{} finding; the rest
// of this file moved to events.gala.
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

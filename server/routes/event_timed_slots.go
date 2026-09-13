package routes

import (
	"errors"
	"sort"
	"time"

	"timeful/server/models"
)

var errActiveSlotOutsideEnabled = errors.New("active-slots-must-be-enabled")
var errIncompleteTimedEventContract = errors.New("incomplete-timed-event-contract")
var errInvalidEventTimezone = errors.New("invalid-event-timezone")
var errInvalidSlotGeneration = errors.New("invalid-slot-generation")
var errInvalidSelectedDay = errors.New("invalid-selected-day")

type timedEventPayloadFields struct {
	ActiveSlots     []models.DateTime       `json:"activeSlots"`
	EventTimezone   *string                 `json:"eventTimezone"`
	SlotGeneration  *models.SlotGeneration  `json:"slotGeneration"`
	TimedRecurrence *models.TimedRecurrence `json:"timedRecurrence"`
}

func normalizeDateTimes(values []models.DateTime) []models.DateTime {
	if len(values) == 0 {
		return []models.DateTime{}
	}

	seen := make(map[int64]struct{}, len(values))
	normalized := make([]models.DateTime, 0, len(values))
	for _, value := range values {
		ms := int64(value)
		if _, exists := seen[ms]; exists {
			continue
		}
		seen[ms] = struct{}{}
		normalized = append(normalized, value)
	}

	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].Time().Before(normalized[j].Time())
	})

	return normalized
}

// normalizeTimedEventPayloadFields validates the canonical timed-event contract.
// The enabled slot domain is not persisted: it is derived as the full civil day
// of each generation day, and active slots are validated against that derived
// domain.
func normalizeTimedEventPayloadFields(
	fields timedEventPayloadFields,
) (timedEventPayloadFields, error) {
	if fields.ActiveSlots == nil ||
		fields.EventTimezone == nil || fields.SlotGeneration == nil ||
		fields.TimedRecurrence == nil {
		return timedEventPayloadFields{}, errIncompleteTimedEventContract
	}
	if *fields.EventTimezone == "" || fields.SlotGeneration.TimeIncrementMinutes <= 0 ||
		fields.SlotGeneration.StartTimeLocal == "" || fields.SlotGeneration.EndTimeLocal == "" ||
		(fields.TimedRecurrence.Kind != "specific_dates" && fields.TimedRecurrence.Kind != "weekly") ||
		fields.TimedRecurrence.StartOnMonday == nil {
		return timedEventPayloadFields{}, errIncompleteTimedEventContract
	}

	enabledSlots, err := deriveEnabledSlots(fields)
	if err != nil {
		return timedEventPayloadFields{}, err
	}
	activeSlots := normalizeDateTimes(fields.ActiveSlots)

	enabledLookup := make(map[int64]struct{}, len(enabledSlots))
	for _, slot := range enabledSlots {
		enabledLookup[int64(slot)] = struct{}{}
	}
	for _, slot := range activeSlots {
		if _, exists := enabledLookup[int64(slot)]; !exists {
			return timedEventPayloadFields{}, errActiveSlotOutsideEnabled
		}
	}

	fields.ActiveSlots = activeSlots

	return fields, nil
}

// discardActiveSlotsOutsideEnabledDomain applies the wipe rule to fields.ActiveSlots
// against the domain derived from the same fields, instead of rejecting them.
func discardActiveSlotsOutsideEnabledDomain(fields timedEventPayloadFields) ([]models.DateTime, error) {
	enabledSlots, err := deriveEnabledSlots(fields)
	if err != nil {
		return nil, err
	}
	enabled := make(map[int64]struct{}, len(enabledSlots))
	for _, slot := range enabledSlots {
		enabled[int64(slot)] = struct{}{}
	}
	kept := make([]models.DateTime, 0, len(fields.ActiveSlots))
	for _, slot := range fields.ActiveSlots {
		if _, ok := enabled[int64(slot)]; ok {
			kept = append(kept, slot)
		}
	}
	return normalizeDateTimes(kept), nil
}

// deriveEnabledSlots regenerates the enabled slot domain from picked dates
// (specific-dates) or the anchor week (weekly), the event timezone, and the
// slot-generation settings. The enabled domain is always the full civil day
// (00:00 through the next 00:00 exclusive, event timezone) of each generation
// day; the slot-generation window is validated and carried over but does not
// bound the domain. Weekly anchoring reuses the day-index conventions of the
// frontend editor (Sunday = 0, and 7 when startOnMonday is true) and anchors
// on the week of the earliest active instant.
func deriveEnabledSlots(fields timedEventPayloadFields) ([]models.DateTime, error) {
	location, err := time.LoadLocation(*fields.EventTimezone)
	if err != nil {
		return nil, errInvalidEventTimezone
	}

	if _, err := parseLocalTime(fields.SlotGeneration.StartTimeLocal); err != nil {
		return nil, errInvalidSlotGeneration
	}
	if _, err := parseLocalTime(fields.SlotGeneration.EndTimeLocal); err != nil {
		return nil, errInvalidSlotGeneration
	}
	increment := fields.SlotGeneration.TimeIncrementMinutes

	days, err := generationDays(
		fields.TimedRecurrence,
		fields.ActiveSlots,
		location,
	)
	if err != nil {
		return nil, err
	}

	var generated []models.DateTime
	for _, day := range days {
		start := time.Date(
			day.Year(), day.Month(), day.Day(),
			0, 0, 0, 0,
			location,
		)
		nextDay := day.AddDate(0, 0, 1)
		end := time.Date(
			nextDay.Year(), nextDay.Month(), nextDay.Day(),
			0, 0, 0, 0,
			location,
		)
		for current := start; current.Before(end); current = current.Add(time.Duration(increment) * time.Minute) {
			generated = append(generated, models.NewDateTimeFromTime(current.UTC()))
		}
	}

	return normalizeDateTimes(generated), nil
}

func generationDays(
	recurrence *models.TimedRecurrence,
	activeSlots []models.DateTime,
	location *time.Location,
) ([]time.Time, error) {
	switch recurrence.Kind {
	case "specific_dates":
		days := make([]time.Time, 0, len(recurrence.SelectedDays))
		for _, raw := range recurrence.SelectedDays {
			day, err := time.Parse("2006-01-02", raw)
			if err != nil {
				return nil, errInvalidSelectedDay
			}
			days = append(days, day)
		}
		return days, nil
	case "weekly":
		return weeklyAnchorDays(
			recurrence,
			weeklyAnchorInstant(activeSlots, location),
		), nil
	default:
		return nil, errIncompleteTimedEventContract
	}
}

// weeklyAnchorInstant returns the instant whose week anchors the derived
// domain: the earliest active slot, or time.Now() in the event timezone when
// there are no active slots (no anchor can be recovered in that case).
func weeklyAnchorInstant(
	activeSlots []models.DateTime,
	location *time.Location,
) time.Time {
	if len(activeSlots) == 0 {
		return time.Now().In(location)
	}

	earliest := activeSlots[0].Time()
	for _, slot := range activeSlots[1:] {
		if slot.Time().Before(earliest) {
			earliest = slot.Time()
		}
	}

	return earliest.In(location)
}

// weeklyAnchorDays ports the day-of-week math from eventEditorSchedule.ts:
// dayIndex 0 means Sunday when startOnMonday is false (and is filtered out
// when true), dayIndex 7 means Sunday when startOnMonday is true (and is
// filtered out when false). Temporal's dayOfWeek (1 = Monday ... 7 = Sunday)
// maps to Go's Weekday (0 = Sunday ... 6 = Saturday) via index % 7.
func weeklyAnchorDays(
	recurrence *models.TimedRecurrence,
	anchor time.Time,
) []time.Time {
	startOnMonday := recurrence.StartOnMonday != nil && *recurrence.StartOnMonday
	indexes := make([]int, 0, len(recurrence.SelectedDaysOfWeek))
	for _, dayIndex := range recurrence.SelectedDaysOfWeek {
		if startOnMonday && dayIndex == 0 {
			continue
		}
		if !startOnMonday && dayIndex == 7 {
			continue
		}
		indexes = append(indexes, dayIndex)
	}
	sort.Ints(indexes)

	anchorWeekday := int(anchor.Weekday())
	days := make([]time.Time, 0, len(indexes))
	for _, dayIndex := range indexes {
		targetWeekday := dayIndex % 7
		daysUntil := (targetWeekday - anchorWeekday + 7) % 7
		days = append(days, anchor.AddDate(0, 0, daysUntil))
	}

	return days
}

func parseLocalTime(value string) (time.Time, error) {
	for _, layout := range []string{"15:04:05", "15:04"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, errors.New("invalid local time")
}

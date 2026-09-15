package routes

import (
	"encoding/json"
	"testing"

	"timeful/server/models"
)

// Cross-side derivation fixtures. The full civil-day domain is asserted by
// count and bound instants, mirroring
// frontend/src/utils/timedEventSlots.test.ts so Go and Temporal derivation
// agree on the same expected instants.
func specificDatesFieldsFor(
	t *testing.T,
	timezone string,
	selectedDays []string,
	startTimeLocal string,
	endTimeLocal string,
	increment int,
) timedEventPayloadFields {
	t.Helper()
	startOnMonday := false
	return timedEventPayloadFields{
		EventTimezone: &timezone,
		SlotGeneration: &models.SlotGeneration{
			StartTimeLocal: startTimeLocal, EndTimeLocal: endTimeLocal, TimeIncrementMinutes: increment,
		},
		TimedRecurrence: &models.TimedRecurrence{
			Kind: "specific_dates", SelectedDays: selectedDays, StartOnMonday: &startOnMonday,
		},
	}
}

func weeklyFieldsFor(
	t *testing.T,
	timezone string,
	selectedDaysOfWeek []int,
	startOnMonday bool,
	startTimeLocal string,
	endTimeLocal string,
	increment int,
) timedEventPayloadFields {
	t.Helper()
	return timedEventPayloadFields{
		EventTimezone: &timezone,
		SlotGeneration: &models.SlotGeneration{
			StartTimeLocal: startTimeLocal, EndTimeLocal: endTimeLocal, TimeIncrementMinutes: increment,
		},
		TimedRecurrence: &models.TimedRecurrence{
			Kind: "weekly", SelectedDaysOfWeek: selectedDaysOfWeek, StartOnMonday: &startOnMonday,
		},
	}
}

// assertDerivedDomainBounds asserts the derived domain by count plus the
// first, a sampled middle, and the last instant, matching the frontend
// fixture assertions for the same scenario.
func assertDerivedDomainBounds(
	t *testing.T,
	derived []models.DateTime,
	expectedCount int,
	sampleIndex int,
	first string,
	sample string,
	last string,
) {
	t.Helper()
	if len(derived) != expectedCount {
		t.Fatalf("expected %d slots, got %d", expectedCount, len(derived))
	}
	if sampleIndex < 0 || sampleIndex >= len(derived) {
		t.Fatalf("invalid sample index %d for %d slots", sampleIndex, len(derived))
	}

	checks := []struct {
		index int
		raw   string
	}{
		{0, first},
		{sampleIndex, sample},
		{len(derived) - 1, last},
	}
	for _, check := range checks {
		if derived[check.index] != timedSlotDateTime(t, check.raw) {
			t.Fatalf("expected slot %d to be %v, got %v", check.index, check.raw, derived[check.index])
		}
	}
}

func assertDerivedDomainContains(t *testing.T, derived []models.DateTime, raw string) {
	t.Helper()
	target := timedSlotDateTime(t, raw)
	for _, slot := range derived {
		if slot == target {
			return
		}
	}
	t.Fatalf("expected derived domain to contain %v", raw)
}

func TestNormalizeTimedEventPayloadFieldsPreservesExplicitEmptyActiveSlots(t *testing.T) {
	fields := specificDatesFieldsFor(t, "America/New_York", []string{"2026-01-05"}, "09:00", "10:00", 15)
	fields.ActiveSlots = []models.DateTime{}
	fields, err := normalizeTimedEventPayloadFields(fields)
	if err != nil {
		t.Fatalf("normalize timed payload: %v", err)
	}

	assertDateTimesEqual(t, fields.ActiveSlots, []models.DateTime{})
}

func TestNormalizeTimedEventPayloadFieldsRejectsMissingActiveSlots(t *testing.T) {
	fields := specificDatesFieldsFor(t, "UTC", []string{"2026-01-05"}, "09:00", "10:00", 15)
	if _, err := normalizeTimedEventPayloadFields(fields); err == nil {
		t.Fatal("expected incomplete timed contract error")
	}
}

func TestDeriveEnabledSlotsSpecificDatesMatchesFrontendFixture(t *testing.T) {
	// The 09:00-10:00 window is validated but does not bound the domain: the
	// full civil day of 2026-01-05 New York spans 96 quarter-hour slots from
	// 00:00 (05:00Z) through the next 00:00 exclusive.
	fields := specificDatesFieldsFor(t, "America/New_York", []string{"2026-01-05"}, "09:00", "10:00", 15)
	derived, err := deriveEnabledSlots(fields)
	if err != nil {
		t.Fatalf("derive enabled slots: %v", err)
	}

	assertDerivedDomainBounds(t, derived, 96, 48,
		"2026-01-05T05:00:00Z",
		"2026-01-05T17:00:00Z",
		"2026-01-06T04:45:00Z",
	)
}

func TestDeriveEnabledSlotsWeeklyAnchorsOnEarliestActiveInstant(t *testing.T) {
	fields := weeklyFieldsFor(t, "America/Los_Angeles", []int{1, 3}, true, "09:00", "11:00", 30)
	fields.ActiveSlots = []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T17:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T17:30:00Z"),
		timedSlotDateTime(t, "2026-01-07T17:00:00Z"),
	}
	derived, err := deriveEnabledSlots(fields)
	if err != nil {
		t.Fatalf("derive enabled slots: %v", err)
	}

	// Earliest active is Monday 2026-01-05 09:00 LA; the anchor week is that
	// week, so the domain is the full civil days of Mon + Wed at half-hour
	// increments (48 slots each).
	assertDerivedDomainBounds(t, derived, 96, 18,
		"2026-01-05T08:00:00Z",
		"2026-01-05T17:00:00Z",
		"2026-01-08T07:30:00Z",
	)
}

func TestDeriveEnabledSlotsWeeklySaturdayAnchorUsesFollowingWeek(t *testing.T) {
	fields := weeklyFieldsFor(t, "America/Los_Angeles", []int{1, 3}, true, "09:00", "10:00", 15)
	fields.ActiveSlots = []models.DateTime{
		timedSlotDateTime(t, "2026-01-10T09:00:00Z"),
	}

	derived, err := deriveEnabledSlots(fields)
	if err != nil {
		t.Fatalf("derive enabled slots: %v", err)
	}

	// Earliest active is Saturday 2026-01-10; its week runs Mon 01-12 and Wed
	// 01-14 (daysUntil 2 and 4 from a Saturday anchor). Full civil days at
	// 15-minute increments.
	assertDerivedDomainBounds(t, derived, 192, 96,
		"2026-01-12T08:00:00Z",
		"2026-01-14T08:00:00Z",
		"2026-01-15T07:45:00Z",
	)
}

func TestDeriveEnabledSlotsStartOnMondayFiltersSundayIndexes(t *testing.T) {
	fields := weeklyFieldsFor(t, "America/Los_Angeles", []int{0, 1, 7}, true, "09:00", "10:00", 15)
	fields.ActiveSlots = []models.DateTime{
		timedSlotDateTime(t, "2026-01-11T09:00:00Z"),
	}
	derived, err := deriveEnabledSlots(fields)
	if err != nil {
		t.Fatalf("derive enabled slots: %v", err)
	}

	// Index 0 (Sunday) is filtered with startOnMonday; only Monday (1) and
	// Sunday-as-7 remain, anchored on the week of Sunday 2026-01-11.
	assertDerivedDomainBounds(t, derived, 192, 96,
		"2026-01-11T08:00:00Z",
		"2026-01-12T08:00:00Z",
		"2026-01-13T07:45:00Z",
	)
}

func TestDeriveEnabledSlotsDstSpringGapMatchesFrontendFixture(t *testing.T) {
	// America/New_York 2026-03-08 skips 02:00 -> 03:00, so the full civil day
	// has 92 quarter-hour slots instead of 96.
	fields := specificDatesFieldsFor(t, "America/New_York", []string{"2026-03-08"}, "01:30", "03:00", 15)
	derived, err := deriveEnabledSlots(fields)
	if err != nil {
		t.Fatalf("derive enabled slots: %v", err)
	}

	assertDerivedDomainBounds(t, derived, 92, 0,
		"2026-03-08T05:00:00Z",
		"2026-03-08T05:00:00Z",
		"2026-03-09T03:45:00Z",
	)
}

func TestDeriveEnabledSlotsDstFallOverlapMatchesFrontendFixture(t *testing.T) {
	// America/New_York 2026-11-01 repeats 02:00 -> 01:00, so the full civil
	// day spans 25 hours: 100 quarter-hour slots, including both fold
	// occurrences (01:30 EDT = 05:30Z and 01:30 EST = 06:30Z).
	fields := specificDatesFieldsFor(t, "America/New_York", []string{"2026-11-01"}, "01:00", "02:30", 15)
	derived, err := deriveEnabledSlots(fields)
	if err != nil {
		t.Fatalf("derive enabled slots: %v", err)
	}

	assertDerivedDomainBounds(t, derived, 100, 48,
		"2026-11-01T04:00:00Z",
		"2026-11-01T16:00:00Z",
		"2026-11-02T04:45:00Z",
	)
	assertDerivedDomainContains(t, derived, "2026-11-01T05:30:00Z")
	assertDerivedDomainContains(t, derived, "2026-11-01T06:30:00Z")
}

func TestDeriveEnabledSlotsLordHoweThirtyMinuteDtstMatchesFrontendFixture(t *testing.T) {
	// Australia/Lord_Howe DST shifts by 30 minutes: 02:00 -> 02:30 on the
	// spring-forward day (2026-10-04) and 02:30 -> 02:00 on the fall-back day
	// (2026-04-05). The full civil day starts at 00:00 and never sits inside a
	// fold, so both sides derive identically to the frontend fixtures.
	fields := specificDatesFieldsFor(t, "Australia/Lord_Howe", []string{"2026-10-04"}, "02:00", "03:30", 15)
	derived, err := deriveEnabledSlots(fields)
	if err != nil {
		t.Fatalf("derive enabled slots: %v", err)
	}
	assertDerivedDomainBounds(t, derived, 94, 0,
		"2026-10-03T13:30:00Z",
		"2026-10-03T13:30:00Z",
		"2026-10-04T12:45:00Z",
	)

	fields = specificDatesFieldsFor(t, "Australia/Lord_Howe", []string{"2026-04-05"}, "00:30", "03:00", 15)
	derived, err = deriveEnabledSlots(fields)
	if err != nil {
		t.Fatalf("derive enabled slots: %v", err)
	}
	assertDerivedDomainBounds(t, derived, 98, 0,
		"2026-04-04T13:00:00Z",
		"2026-04-04T13:00:00Z",
		"2026-04-05T13:15:00Z",
	)
}

func TestDeriveEnabledSlotsWrappedWindowDerivesFullCivilDay(t *testing.T) {
	// The wrapped 22:30-01:30 window does not bound the domain: the full
	// civil day of 2026-05-10 UTC holds 48 half-hour slots.
	fields := specificDatesFieldsFor(t, "UTC", []string{"2026-05-10"}, "22:30", "01:30", 30)
	derived, err := deriveEnabledSlots(fields)
	if err != nil {
		t.Fatalf("derive enabled slots: %v", err)
	}

	assertDerivedDomainBounds(t, derived, 48, 0,
		"2026-05-10T00:00:00Z",
		"2026-05-10T00:00:00Z",
		"2026-05-10T23:30:00Z",
	)
}

func TestDeriveEnabledSlotsDeduplicatesGeneratedSlots(t *testing.T) {
	fields := specificDatesFieldsFor(t, "UTC", []string{"2026-01-05", "2026-01-05"}, "09:00", "09:30", 15)
	derived, err := deriveEnabledSlots(fields)
	if err != nil {
		t.Fatalf("derive enabled slots: %v", err)
	}

	assertDerivedDomainBounds(t, derived, 96, 0,
		"2026-01-05T00:00:00Z",
		"2026-01-05T00:00:00Z",
		"2026-01-05T23:45:00Z",
	)
}

func TestDeriveEnabledSlotsRejectsInvalidTimezone(t *testing.T) {
	fields := specificDatesFieldsFor(t, "Not/AZone", []string{"2026-01-05"}, "09:00", "10:00", 15)
	if _, err := deriveEnabledSlots(fields); err != errInvalidEventTimezone {
		t.Fatalf("expected %v, got %v", errInvalidEventTimezone, err)
	}
}

func TestDeriveEnabledSlotsRejectsInvalidSelectedDay(t *testing.T) {
	fields := specificDatesFieldsFor(t, "UTC", []string{"not-a-day"}, "09:00", "10:00", 15)
	if _, err := deriveEnabledSlots(fields); err != errInvalidSelectedDay {
		t.Fatalf("expected %v, got %v", errInvalidSelectedDay, err)
	}
}

func TestDeriveEnabledSlotsRejectsInvalidSlotGenerationTime(t *testing.T) {
	fields := specificDatesFieldsFor(t, "UTC", []string{"2026-01-05"}, "09:00", "oops", 15)
	if _, err := deriveEnabledSlots(fields); err != errInvalidSlotGeneration {
		t.Fatalf("expected %v, got %v", errInvalidSlotGeneration, err)
	}
}

func TestNormalizeTimedEventPayloadFieldsRejectsActiveOutsideDerivedDomain(t *testing.T) {
	fields := specificDatesFieldsFor(t, "America/New_York", []string{"2026-01-05"}, "09:00", "10:00", 15)
	// 00:15 on January 6 New York time is outside the picked January 5 full
	// civil day.
	fields.ActiveSlots = []models.DateTime{
		timedSlotDateTime(t, "2026-01-06T05:15:00Z"),
	}
	if _, err := normalizeTimedEventPayloadFields(fields); err != errActiveSlotOutsideEnabled {
		t.Fatalf("expected %v, got %v", errActiveSlotOutsideEnabled, err)
	}
}

func TestNormalizeTimedEventPayloadFieldsRejectsWeeklyActiveOutsideDerivedDomain(t *testing.T) {
	fields := weeklyFieldsFor(t, "America/Los_Angeles", []int{1, 3}, true, "09:00", "11:00", 30)
	// 15:00 on Sunday January 4 Los Angeles time is inside the anchor week
	// but not a selected day (Mon/Wed).
	fields.ActiveSlots = []models.DateTime{
		timedSlotDateTime(t, "2026-01-04T23:00:00Z"),
	}
	if _, err := normalizeTimedEventPayloadFields(fields); err != errActiveSlotOutsideEnabled {
		t.Fatalf("expected %v, got %v", errActiveSlotOutsideEnabled, err)
	}
}

func TestDiscardActiveSlotsOutsideEnabledDomainDiscardsOutOfDomainActives(t *testing.T) {
	fields := specificDatesFieldsFor(t, "UTC", []string{"2026-01-04"}, "23:00", "01:00", 30)
	fields.ActiveSlots = []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T07:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T07:30:00Z"),
	}

	kept, err := discardActiveSlotsOutsideEnabledDomain(fields)
	if err != nil {
		t.Fatalf("discard active slots: %v", err)
	}

	assertDateTimesEqual(t, kept, []models.DateTime{})
}

func TestDiscardActiveSlotsOutsideEnabledDomainKeepsInDomainActives(t *testing.T) {
	fields := specificDatesFieldsFor(t, "UTC", []string{"2026-01-04"}, "23:00", "01:00", 30)
	fields.ActiveSlots = []models.DateTime{
		timedSlotDateTime(t, "2026-01-04T18:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T07:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T07:30:00Z"),
	}

	kept, err := discardActiveSlotsOutsideEnabledDomain(fields)
	if err != nil {
		t.Fatalf("discard active slots: %v", err)
	}

	assertDateTimesEqual(t, kept, []models.DateTime{
		timedSlotDateTime(t, "2026-01-04T18:00:00Z"),
	})
}

func TestDiscardActiveSlotsOutsideEnabledDomainWeeklyDropsUnselectedWeekday(t *testing.T) {
	fields := weeklyFieldsFor(t, "America/Los_Angeles", []int{1, 3}, true, "09:00", "11:00", 30)
	fields.ActiveSlots = []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T17:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T17:30:00Z"),
		timedSlotDateTime(t, "2026-01-06T17:00:00Z"),
	}

	kept, err := discardActiveSlotsOutsideEnabledDomain(fields)
	if err != nil {
		t.Fatalf("discard active slots: %v", err)
	}

	// The anchor week runs Monday 2026-01-05 and Wednesday 2026-01-07, so the
	// Tuesday instant is dropped and the Monday pair survives.
	assertDateTimesEqual(t, kept, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T17:00:00Z"),
		timedSlotDateTime(t, "2026-01-05T17:30:00Z"),
	})
}

func TestFilterResponseSlotsOutsideActiveSetFiltersOnlySlotKeys(t *testing.T) {
	payload := json.RawMessage(`{"availability":["2026-01-05T14:00:00Z","2026-01-05T14:30:00Z"],"ifNeeded":["2026-01-05T14:15:00Z","2026-01-05T14:30:00Z"],"name":"Maya"}`)
	active := map[models.DateTime]struct{}{
		timedSlotDateTime(t, "2026-01-05T14:00:00Z"): {},
		timedSlotDateTime(t, "2026-01-05T14:15:00Z"): {},
	}

	next, changed, err := filterResponseSlotsOutsideActiveSet(payload, active)
	if err != nil {
		t.Fatalf("filter response slots: %v", err)
	}
	if !changed {
		t.Fatal("expected changed payload")
	}

	var fields struct {
		Availability []models.DateTime `json:"availability"`
		IfNeeded     []models.DateTime `json:"ifNeeded"`
		Name         string            `json:"name"`
	}
	if err := json.Unmarshal(next, &fields); err != nil {
		t.Fatalf("unmarshal filtered payload: %v", err)
	}
	assertDateTimesEqual(t, fields.Availability, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T14:00:00Z"),
	})
	assertDateTimesEqual(t, fields.IfNeeded, []models.DateTime{
		timedSlotDateTime(t, "2026-01-05T14:15:00Z"),
	})
	if fields.Name != "Maya" {
		t.Fatalf("expected untouched name, got %q", fields.Name)
	}
}

func TestFilterResponseSlotsOutsideActiveSetReportsUnchanged(t *testing.T) {
	payload := json.RawMessage(`{"availability":["2026-01-05T14:00:00Z"],"name":"Maya"}`)
	active := map[models.DateTime]struct{}{
		timedSlotDateTime(t, "2026-01-05T14:00:00Z"): {},
	}

	next, changed, err := filterResponseSlotsOutsideActiveSet(payload, active)
	if err != nil {
		t.Fatalf("filter response slots: %v", err)
	}
	if changed {
		t.Fatal("expected unchanged payload")
	}
	if string(next) != string(payload) {
		t.Fatalf("expected original payload %s, got %s", payload, next)
	}
}

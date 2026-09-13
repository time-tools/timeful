package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

// newAvailabilityGroupTestRepository applies the schema migrations into a
// transaction-scoped set of temporary tables. Temp tables shadow the real
// schema so the isolated tests never mutate test-stack records. It reuses the
// accounts harness because both need the full baseline schema.
func newAvailabilityGroupTestRepository(t *testing.T) (context.Context, *Repository, pgx.Tx) {
	t.Helper()
	ctx, repo, tx := newAccountsTestRepository(t)
	return ctx, repo, tx
}

func seedAvailabilityGroupEvent(t *testing.T, ctx context.Context, tx pgx.Tx) string {
	t.Helper()
	shortID, err := GenerateShortID()
	if err != nil {
		t.Fatal(err)
	}
	var eventID string
	if err := tx.QueryRow(ctx, `INSERT INTO postgres_events (short_id, name, type)
VALUES ($1, 'Group', 'group') RETURNING id`, shortID).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	return eventID
}

func boolPointer(value bool) *bool { return &value }

// TestAvailabilityGroupSchemaConstraints proves the baseline admits the group
// kind, keeps unsupported kinds rejected, and enforces the attendee membership
// relation and uniqueness.
func TestAvailabilityGroupSchemaConstraints(t *testing.T) {
	ctx, _, tx := newAvailabilityGroupTestRepository(t)
	eventID := seedAvailabilityGroupEvent(t, ctx, tx)

	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO postgres_events (short_id, name, type) VALUES ($1, 'Bogus', 'bogus')`, signupTestShortID(t))
		return err
	})
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO event_attendees (event_id, email) VALUES ($1, '')`, eventID)
		return err
	})
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO event_attendees (event_id, email) VALUES (gen_random_uuid(), 'orphan@example.com')`)
		return err
	})
	if _, err := tx.Exec(ctx, `INSERT INTO event_attendees (event_id, email) VALUES ($1, 'dupe@example.com')`, eventID); err != nil {
		t.Fatal(err)
	}
	duplicateErr := expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO event_attendees (event_id, email) VALUES ($1, 'dupe@example.com')`, eventID)
		return err
	})
	if !IsUniqueViolation(duplicateErr) {
		t.Fatalf("duplicate membership error = %v, want a unique violation", duplicateErr)
	}
}

// TestAttendeeRepositoryMembershipLifecycle proves add, list, email lookup,
// decline/undecline, and removal are keyed by event and email and that an
// absent decline state is distinct from an explicit false.
func TestAttendeeRepositoryMembershipLifecycle(t *testing.T) {
	ctx, repo, tx := newAvailabilityGroupTestRepository(t)
	eventID := seedAvailabilityGroupEvent(t, ctx, tx)

	if err := repo.AddAttendees(ctx, eventID, []string{"invitee@example.com"}, nil); err != nil {
		t.Fatal(err)
	}
	attendee, err := repo.GetAttendeeByEmail(ctx, eventID, "invitee@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if attendee.ID == "" || attendee.Declined != nil {
		t.Fatalf("new membership is unexpected: %#v", attendee)
	}

	// Re-adding keeps the same membership and its absent decline state.
	if err := repo.AddAttendees(ctx, eventID, []string{"invitee@example.com"}, boolPointer(false)); err != nil {
		t.Fatal(err)
	}
	duplicate, err := repo.GetAttendeeByEmail(ctx, eventID, "invitee@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if duplicate.ID != attendee.ID {
		t.Fatalf("re-add created a second membership: %s vs %s", duplicate.ID, attendee.ID)
	}
	if duplicate.Declined != nil {
		t.Fatalf("re-add replaced the absent decline state: %#v", duplicate.Declined)
	}
	list, err := repo.ListAttendees(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != attendee.ID {
		t.Fatalf("membership list is unexpected: %#v", list)
	}

	if err := repo.SetAttendeeDeclined(ctx, eventID, "invitee@example.com", true); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetAttendeeByEmail(ctx, eventID, "invitee@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Declined == nil || !*stored.Declined {
		t.Fatalf("decline was not stored: %#v", stored.Declined)
	}
	if err := repo.SetAttendeeDeclined(ctx, eventID, "invitee@example.com", false); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetAttendeeByEmail(ctx, eventID, "invitee@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Declined == nil || *stored.Declined {
		t.Fatalf("undecline was not stored as explicit false: %#v", stored.Declined)
	}

	if err := repo.RemoveAttendees(ctx, eventID, []string{"invitee@example.com"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetAttendeeByEmail(ctx, eventID, "invitee@example.com"); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("removed membership lookup error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.RemoveAttendees(ctx, eventID, []string{"invitee@example.com"}); err != nil {
		t.Fatalf("repeated batch removal should be a no-op: %v", err)
	}
}

// TestAttendeeRepositoryHasNonDeclinedAttendeeEmail proves the invitee EXISTS
// matches case-insensitively only non-declined memberships of the requested
// event, mirroring the dashboard membership predicate, and reports no
// membership for an empty email.
func TestAttendeeRepositoryHasNonDeclinedAttendeeEmail(t *testing.T) {
	ctx, repo, tx := newAvailabilityGroupTestRepository(t)
	eventID := seedAvailabilityGroupEvent(t, ctx, tx)
	otherEventID := seedAvailabilityGroupEvent(t, ctx, tx)

	if err := repo.AddAttendees(ctx, eventID, []string{"invitee@example.com"}, boolPointer(false)); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddAttendees(ctx, eventID, []string{"declined@example.com"}, boolPointer(true)); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddAttendees(ctx, otherEventID, []string{"elsewhere@example.com"}, boolPointer(false)); err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		name  string
		event string
		email string
		want  bool
	}{
		{"case-insensitive match", eventID, "INVITEE@example.com", true},
		{"declined excluded", eventID, "declined@example.com", false},
		{"other event excluded", eventID, "elsewhere@example.com", false},
		{"unknown email", eventID, "stranger@example.com", false},
		{"empty email", eventID, "", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			found, err := repo.HasNonDeclinedAttendeeEmail(ctx, testCase.event, testCase.email)
			if err != nil {
				t.Fatal(err)
			}
			if found != testCase.want {
				t.Fatalf("HasNonDeclinedAttendeeEmail(%q, %q) = %v, want %v", testCase.event, testCase.email, found, testCase.want)
			}
		})
	}
}

// TestAttendeeRepositoryResolvesAccountByEmail proves email resolves to a
// PostgreSQL account case-insensitively and that an unknown email leaves the
// account relation absent.
func TestAttendeeRepositoryResolvesAccountByEmail(t *testing.T) {
	ctx, repo, tx := newAvailabilityGroupTestRepository(t)
	account, err := repo.CreateAccount(ctx, Account{Email: "member@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	eventID := seedAvailabilityGroupEvent(t, ctx, tx)

	resolvedEmail := "MEMBER@example.com"
	if err := repo.AddAttendees(ctx, eventID, []string{resolvedEmail}, nil); err != nil {
		t.Fatal(err)
	}
	resolved, err := repo.GetAttendeeByEmail(ctx, eventID, resolvedEmail)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.PlatformIdentityID == nil || *resolved.PlatformIdentityID != account.PlatformIdentityID {
		t.Fatalf("email did not resolve to the account: %#v", resolved.PlatformIdentityID)
	}

	unmatchedEmail := "stranger@example.com"
	if err := repo.AddAttendees(ctx, eventID, []string{unmatchedEmail}, nil); err != nil {
		t.Fatal(err)
	}
	unmatched, err := repo.GetAttendeeByEmail(ctx, eventID, unmatchedEmail)
	if err != nil {
		t.Fatal(err)
	}
	if unmatched.PlatformIdentityID != nil {
		t.Fatalf("unknown email resolved to an account: %#v", unmatched.PlatformIdentityID)
	}
}

// TestAddAttendeesBatchDedupesAndPreservesOrder proves the batch insert
// tolerates exact duplicates and differently-cased emails, keeps
// first-occurrence order, resolves accounts case-insensitively, and preserves
// the stored decline state and account fill-on-conflict behavior.
func TestAddAttendeesBatchDedupesAndPreservesOrder(t *testing.T) {
	ctx, repo, tx := newAvailabilityGroupTestRepository(t)
	first, err := repo.CreateAccount(ctx, Account{Email: "batch-first@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.CreateAccount(ctx, Account{Email: "batch-second@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	eventID := seedAvailabilityGroupEvent(t, ctx, tx)

	if err := repo.AddAttendees(ctx, eventID, []string{
		"batch-first@example.com",
		"batch-first@example.com",
		"batch-second@example.com",
		"BATCH-FIRST@example.com",
	}, boolPointer(false)); err != nil {
		t.Fatal(err)
	}
	list, err := repo.ListAttendees(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("batch insert stored %d memberships, want 3: %#v", len(list), list)
	}
	wantOrder := []string{"batch-first@example.com", "batch-second@example.com", "BATCH-FIRST@example.com"}
	for i, want := range wantOrder {
		if list[i].Email != want {
			t.Fatalf("membership %d email = %q, want %q", i, list[i].Email, want)
		}
	}
	if list[0].PlatformIdentityID == nil || *list[0].PlatformIdentityID != first.PlatformIdentityID {
		t.Fatalf("first membership did not resolve its account: %#v", list[0].PlatformIdentityID)
	}
	if list[1].PlatformIdentityID == nil || *list[1].PlatformIdentityID != second.PlatformIdentityID {
		t.Fatalf("second membership did not resolve its account: %#v", list[1].PlatformIdentityID)
	}
	if list[2].PlatformIdentityID == nil || *list[2].PlatformIdentityID != first.PlatformIdentityID {
		t.Fatalf("differently-cased membership did not resolve case-insensitively: %#v", list[2].PlatformIdentityID)
	}

	// Re-adding an existing email keeps its stored decline state and account.
	if err := repo.SetAttendeeDeclined(ctx, eventID, "batch-first@example.com", true); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddAttendees(ctx, eventID, []string{"batch-first@example.com"}, boolPointer(false)); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetAttendeeByEmail(ctx, eventID, "batch-first@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Declined == nil || !*stored.Declined {
		t.Fatalf("batch re-add replaced the stored decline state: %#v", stored.Declined)
	}

	// A later account creation fills the previously missing resolution.
	if err := repo.AddAttendees(ctx, eventID, []string{"batch-late@example.com"}, boolPointer(false)); err != nil {
		t.Fatal(err)
	}
	late, err := repo.CreateAccount(ctx, Account{Email: "batch-late@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.AddAttendees(ctx, eventID, []string{"batch-late@example.com"}, boolPointer(false)); err != nil {
		t.Fatal(err)
	}
	filled, err := repo.GetAttendeeByEmail(ctx, eventID, "batch-late@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if filled.PlatformIdentityID == nil || *filled.PlatformIdentityID != late.PlatformIdentityID {
		t.Fatalf("batch re-add did not fill the missing account: %#v", filled.PlatformIdentityID)
	}

	// Removing a batch leaves the untouched memberships and ignores absent ones.
	if err := repo.RemoveAttendees(ctx, eventID, []string{"batch-second@example.com", "missing@example.com"}); err != nil {
		t.Fatal(err)
	}
	remaining, err := repo.ListAttendees(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	for _, attendee := range remaining {
		if attendee.Email == "batch-second@example.com" {
			t.Fatalf("batch removal left %q: %#v", attendee.Email, remaining)
		}
	}
}

// TestDeleteAccountResponsesRemovesOnlyMatchingRows proves the set-based
// cleanup removes only account responses owned by the supplied platform
// identities and reports the deleted row count.
func TestDeleteAccountResponsesRemovesOnlyMatchingRows(t *testing.T) {
	ctx, repo, tx := newAvailabilityGroupTestRepository(t)
	eventID := seedAvailabilityGroupEvent(t, ctx, tx)
	firstVisitor, err := repo.CreateEventVisitorIdentity(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	secondVisitor, err := repo.CreateEventVisitorIdentity(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	guestVisitor, err := repo.CreateEventVisitorIdentity(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	firstAccount, err := repo.CreateAccount(ctx, Account{Email: "delete-first@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	secondAccount, err := repo.CreateAccount(ctx, Account{Email: "delete-second@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	for _, seeded := range []struct {
		visitor        *EventVisitorIdentity
		kind           string
		platformID     *string
		canonicalGuest *string
	}{
		{visitor: firstVisitor, kind: RespondentKindAccount, platformID: &firstAccount.PlatformIdentityID},
		{visitor: secondVisitor, kind: RespondentKindAccount, platformID: &secondAccount.PlatformIdentityID},
		{visitor: guestVisitor, kind: RespondentKindGuest, canonicalGuest: boolStringPointer("Guest")},
	} {
		response := &Response{
			EventID:                eventID,
			EventVisitorIdentityID: seeded.visitor.ID,
			RespondentKind:         seeded.kind,
			PlatformIdentityID:     seeded.platformID,
			CanonicalGuestName:     seeded.canonicalGuest,
			Payload:                json.RawMessage(`{"name":"Respondent"}`),
		}
		if err := repo.CreateResponse(ctx, response); err != nil {
			t.Fatal(err)
		}
	}

	deleted, err := repo.DeleteAccountResponses(ctx, eventID, []string{firstAccount.PlatformIdentityID})
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Fatalf("deleted rows = %d, want 1", deleted)
	}
	remaining, err := repo.ListResponses(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 2 {
		t.Fatalf("remaining responses = %d, want 2", len(remaining))
	}
	for _, response := range remaining {
		if response.PlatformIdentityID != nil && *response.PlatformIdentityID == firstAccount.PlatformIdentityID {
			t.Fatal("cleanup left the matching account response")
		}
	}
	deleted, err = repo.DeleteAccountResponses(ctx, eventID, []string{"086f4f9a-1b0e-4b7c-9a3e-6f4c2d1e5a03"})
	if err != nil || deleted != 0 {
		t.Fatalf("missing identity cleanup = %d, %v; want 0, nil", deleted, err)
	}
}

func boolStringPointer(value string) *string { return &value }

// TestAccountDeletionReleasesAttendeeRelations proves a deleted account's
// email-keyed memberships survive with their account relation released and
// decline state preserved, while another account's relations are untouched.
func TestAccountDeletionReleasesAttendeeRelations(t *testing.T) {
	ctx, repo, tx := newAvailabilityGroupTestRepository(t)
	deleted, err := repo.CreateAccount(ctx, Account{Email: "deleted@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := repo.CreateAccount(ctx, Account{Email: "other@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	eventID := seedAvailabilityGroupEvent(t, ctx, tx)

	if err := repo.AddAttendees(ctx, eventID, []string{"deleted@example.com"}, boolPointer(true)); err != nil {
		t.Fatal(err)
	}
	member, err := repo.GetAttendeeByEmail(ctx, eventID, "deleted@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if member.PlatformIdentityID == nil || *member.PlatformIdentityID != deleted.PlatformIdentityID {
		t.Fatalf("membership did not resolve the account: %#v", member.PlatformIdentityID)
	}
	if err := repo.AddAttendees(ctx, eventID, []string{"other@example.com"}, nil); err != nil {
		t.Fatal(err)
	}
	otherMember, err := repo.GetAttendeeByEmail(ctx, eventID, "other@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if otherMember.PlatformIdentityID == nil || *otherMember.PlatformIdentityID != other.PlatformIdentityID {
		t.Fatalf("second membership did not resolve its account: %#v", otherMember.PlatformIdentityID)
	}

	if err := repo.DeleteAccountByPlatformIdentityID(ctx, deleted.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}

	stored, err := repo.GetAttendeeByEmail(ctx, eventID, "deleted@example.com")
	if err != nil {
		t.Fatalf("membership did not survive account deletion: %v", err)
	}
	if stored.PlatformIdentityID != nil {
		t.Fatalf("account relation was not released: %#v", stored.PlatformIdentityID)
	}
	if stored.Declined == nil || !*stored.Declined {
		t.Fatalf("decline state was not preserved: %#v", stored.Declined)
	}
	untouched, err := repo.GetAttendeeByEmail(ctx, eventID, "other@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if untouched.PlatformIdentityID == nil || *untouched.PlatformIdentityID != other.PlatformIdentityID {
		t.Fatalf("another account's relation was changed: %#v", untouched.PlatformIdentityID)
	}
}

// TestGroupResponseReusesResponseStoragePayload proves a group response is
// stored in the existing postgres_event_responses table with calendar-derived
// mode, selected calendars, copied calendar preferences, and manual
// availability preserved, and that the manual availability window stays in the
// event payload.
func TestGroupResponseReusesResponseStoragePayload(t *testing.T) {
	ctx, repo, tx := newAvailabilityGroupTestRepository(t)
	event := &Event{
		Name:    "Group",
		Type:    "group",
		Payload: json.RawMessage(`{"duration":30,"manualAvailabilityWindow":{"start":"09:00","end":"17:00"}}`),
	}
	if err := repo.CreateEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	var visitorID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_visitor_identities (event_id) VALUES ($1) RETURNING id`, event.ID).Scan(&visitorID); err != nil {
		t.Fatal(err)
	}
	guestName := "Ada"
	response := &Response{
		EventID:                event.ID,
		EventVisitorIdentityID: visitorID,
		RespondentKind:         RespondentKindGuest,
		CanonicalGuestName:     &guestName,
		Payload: json.RawMessage(`{
			"useCalendarAvailability": true,
			"enabledCalendars": {"ada@example.com": ["primary", "work_google"]},
			"calendarOptions": {"weekStart": 1},
			"manualAvailability": {"1700000000000": true}
		}`),
	}
	if err := repo.CreateResponse(ctx, response); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetResponseByID(ctx, response.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"useCalendarAvailability", "enabledCalendars", "calendarOptions", "manualAvailability"} {
		if !bytes.Contains(stored.Payload, []byte(`"`+key+`"`)) {
			t.Fatalf("group response payload lost %s: %s", key, stored.Payload)
		}
	}
	var storedResponse int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM postgres_event_responses WHERE id = $1`, response.ID).Scan(&storedResponse); err != nil {
		t.Fatal(err)
	}
	if storedResponse != 1 {
		t.Fatalf("group response did not reuse the shared response storage: %d", storedResponse)
	}
	storedEvent, err := repo.GetEventByID(ctx, event.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(storedEvent.Payload, []byte(`"duration"`)) || !bytes.Contains(storedEvent.Payload, []byte(`"manualAvailabilityWindow"`)) {
		t.Fatalf("group event payload lost the manual availability window: %s", storedEvent.Payload)
	}
}

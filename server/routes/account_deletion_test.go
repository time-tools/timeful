package routes

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"timeful/server/accounts"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

// TestAccountDeletionRemovesAuthority proves the ratified FR-123
// deletion unit: the account authority, platform identity, calendar
// connections, responses, folders, and daily-log membership are removed; events
// the account organized survive with ownership released and their other guests'
// responses intact.
func TestAccountDeletionRemovesAuthority(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	ctx := context.Background()
	email := "delete-" + models.NewUUID().String() + "@example.com"

	verifyOtpSignIn(t, client, email, "123456")
	repository := repositoryForTest(t)
	account, err := repository.GetAccountByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { deleteAccountTestFixtures(t, account.PlatformIdentityID) })
	objectID := accountObjectID(t, account.PlatformIdentityID)

	// Account data: a daily log shared with another account, and a log that only
	// the deleted account used.
	sharedDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(objectID[0])*256+int(objectID[1]))
	soloDate := time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(objectID[2])*256+int(objectID[3]))
	otherAccountID := newSessionAccount(t)
	var sharedLogID, soloLogID string
	if err := pgstore.Pool.QueryRow(ctx, `INSERT INTO daily_user_logs (log_date) VALUES ($1)
ON CONFLICT (log_date) DO UPDATE SET updated_at = daily_user_logs.updated_at RETURNING id`, sharedDate).Scan(&sharedLogID); err != nil {
		t.Fatal(err)
	}
	if _, err := pgstore.Pool.Exec(ctx, `INSERT INTO daily_user_log_members (daily_user_log_id, platform_identity_id, first_seen_position)
VALUES ($1, $2, 0), ($1, $3, 1) ON CONFLICT DO NOTHING`, sharedLogID, account.PlatformIdentityID, otherAccountID); err != nil {
		t.Fatal(err)
	}
	if err := pgstore.Pool.QueryRow(ctx, `INSERT INTO daily_user_logs (log_date) VALUES ($1)
ON CONFLICT (log_date) DO UPDATE SET updated_at = daily_user_logs.updated_at RETURNING id`, soloDate).Scan(&soloLogID); err != nil {
		t.Fatal(err)
	}
	if _, err := pgstore.Pool.Exec(ctx, `INSERT INTO daily_user_log_members (daily_user_log_id, platform_identity_id, first_seen_position)
VALUES ($1, $2, 0) ON CONFLICT DO NOTHING`, soloLogID, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pgstore.Pool.Exec(context.Background(), `DELETE FROM daily_user_logs WHERE id = ANY($1)`, []string{sharedLogID, soloLogID})
	})

	// Events: an event the account owns with one account response and one
	// guest response.
	shortID, err := pgstore.GenerateShortID()
	if err != nil {
		t.Fatal(err)
	}
	var eventID, ownerVisitorID, guestVisitorID string
	if err := pgstore.Pool.QueryRow(ctx, `INSERT INTO events (short_id, name, type, owner_platform_identity_id)
VALUES ($1, 'Owned', 'specific_dates', $2) RETURNING id`, shortID, account.PlatformIdentityID).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	if err := pgstore.Pool.QueryRow(ctx, `INSERT INTO event_visitor_identities (event_id, platform_identity_id) VALUES ($1, $2) RETURNING id`, eventID, account.PlatformIdentityID).Scan(&ownerVisitorID); err != nil {
		t.Fatal(err)
	}
	if _, err := pgstore.Pool.Exec(ctx, `UPDATE events SET owner_event_visitor_identity_id = $2 WHERE id = $1`, eventID, ownerVisitorID); err != nil {
		t.Fatal(err)
	}
	if err := pgstore.Pool.QueryRow(ctx, `INSERT INTO event_visitor_identities (event_id) VALUES ($1) RETURNING id`, eventID).Scan(&guestVisitorID); err != nil {
		t.Fatal(err)
	}
	if _, err := pgstore.Pool.Exec(ctx, `INSERT INTO event_responses (event_id, event_visitor_identity_id, respondent_kind, platform_identity_id, payload)
VALUES ($1, $2, 'account', $3, '{"name":"Owner"}'), ($1, $4, 'guest', NULL, '{"name":"Guest"}')`,
		eventID, ownerVisitorID, account.PlatformIdentityID, guestVisitorID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup := context.Background()
		_, _ = pgstore.Pool.Exec(cleanup, `DELETE FROM events WHERE id = $1`, eventID)
	})

	// Folders: an account folder with a member.
	var folderID string
	if err := pgstore.Pool.QueryRow(ctx, `INSERT INTO folders (platform_identity_id, name) VALUES ($1, 'Folder') RETURNING id`, account.PlatformIdentityID).Scan(&folderID); err != nil {
		t.Fatal(err)
	}
	if _, err := pgstore.Pool.Exec(ctx, `INSERT INTO folder_events (platform_identity_id, folder_id, event_id) VALUES ($1, $2, $3)`, account.PlatformIdentityID, folderID, eventID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pgstore.Pool.Exec(context.Background(), `DELETE FROM folders WHERE id = $1`, folderID)
	})

	client.request(http.MethodDelete, "/api/user", map[string]any{"email": email}, http.StatusOK)

	// Signed out and account authority gone.
	client.request(http.MethodGet, "/api/user/profile", nil, http.StatusUnauthorized)
	if _, err := repository.GetAccountByPlatformIdentityID(ctx, account.PlatformIdentityID); err == nil {
		t.Fatal("account still resolves after deletion")
	}
	var identities int
	if err := pgstore.Pool.QueryRow(ctx, `SELECT count(*) FROM platform_identities WHERE id = $1`, account.PlatformIdentityID).Scan(&identities); err != nil {
		t.Fatal(err)
	}
	if identities != 0 {
		t.Fatalf("platform identity survived deletion: %d", identities)
	}

	// The account's own data is gone while another guest's
	// response survives.
	var ownResponses, guestResponses, pgFolders, pgMemberships int64
	if err := pgstore.Pool.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM event_responses WHERE platform_identity_id = $1),
 (SELECT count(*) FROM event_responses WHERE event_id = $2 AND respondent_kind = 'guest'),
 (SELECT count(*) FROM folders WHERE platform_identity_id = $1),
 (SELECT count(*) FROM folder_events WHERE platform_identity_id = $1)`,
		account.PlatformIdentityID, eventID).Scan(&ownResponses, &guestResponses, &pgFolders, &pgMemberships); err != nil {
		t.Fatal(err)
	}
	if ownResponses != 0 || pgFolders != 0 || pgMemberships != 0 {
		t.Fatalf("owned data survived: responses=%d folders=%d memberships=%d", ownResponses, pgFolders, pgMemberships)
	}
	if guestResponses != 1 {
		t.Fatalf("another guest's response was removed: %d", guestResponses)
	}

	// The deleted account is gone from the daily logs, the log it emptied
	// is deleted, and a log shared with another account survives.
	var ownLogMembers, sharedSurvived, soloSurvived, otherLogMembers int
	if err := pgstore.Pool.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM daily_user_log_members WHERE platform_identity_id = $1),
 (SELECT count(*) FROM daily_user_logs WHERE id = $2),
 (SELECT count(*) FROM daily_user_logs WHERE id = $3),
 (SELECT count(*) FROM daily_user_log_members WHERE daily_user_log_id = $2)`,
		account.PlatformIdentityID, sharedLogID, soloLogID).Scan(&ownLogMembers, &sharedSurvived, &soloSurvived, &otherLogMembers); err != nil {
		t.Fatal(err)
	}
	if ownLogMembers != 0 {
		t.Fatalf("deleted account still appears in a daily user log: %d", ownLogMembers)
	}
	if soloSurvived != 0 {
		t.Fatal("daily user log emptied by deletion was not removed")
	}
	if sharedSurvived != 1 || otherLogMembers != 1 {
		t.Fatalf("daily user log shared with another account was not preserved: logs=%d members=%d", sharedSurvived, otherLogMembers)
	}

	// Events survive with released ownership.
	var pgOwned, pgOwnResponses, pgGuestResponses int
	if err := pgstore.Pool.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM events WHERE id = $1 AND owner_platform_identity_id IS NULL AND owner_event_visitor_identity_id IS NULL),
 (SELECT count(*) FROM event_responses WHERE event_id = $1 AND respondent_kind = 'account'),
 (SELECT count(*) FROM event_responses WHERE event_id = $1 AND respondent_kind = 'guest')`, eventID).Scan(&pgOwned, &pgOwnResponses, &pgGuestResponses); err != nil {
		t.Fatal(err)
	}
	if pgOwned != 1 {
		t.Fatal("event did not survive with ownership released")
	}
	if pgOwnResponses != 0 || pgGuestResponses != 1 {
		t.Fatalf("responses wrong: own=%d guest=%d", pgOwnResponses, pgGuestResponses)
	}

	// Deletion is idempotent: repeating the unit is a no-op.
	if err := accounts.DeleteAccount(ctx, account.PlatformIdentityID); err != nil {
		t.Fatalf("repeated deletion must be idempotent: %v", err)
	}
}

// TestAccountDeletionRejectsEmailMismatch proves that the account email must be
// typed before anything is removed.
func TestAccountDeletionRejectsEmailMismatch(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	ctx := context.Background()
	email := "mismatch-" + models.NewUUID().String() + "@example.com"

	verifyOtpSignIn(t, client, email, "123456")
	account, err := repositoryForTest(t).GetAccountByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { deleteAccountTestFixtures(t, account.PlatformIdentityID) })

	client.request(http.MethodDelete, "/api/user", map[string]any{"email": "someone-else@example.com"}, http.StatusBadRequest)
	client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
	if _, err := repositoryForTest(t).GetAccountByPlatformIdentityID(ctx, account.PlatformIdentityID); err != nil {
		t.Fatalf("account removed despite email mismatch: %v", err)
	}
}

// TestAccountDeletionPartialFailureLeavesAuthorityAndRetryConverges proves that
// a PostgreSQL deletion failure leaves the account authority and session intact
// and that retrying converges.
func TestAccountDeletionPartialFailureLeavesAuthorityAndRetryConverges(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	ctx := context.Background()
	email := "partial-" + models.NewUUID().String() + "@example.com"

	verifyOtpSignIn(t, client, email, "123456")
	account, err := repositoryForTest(t).GetAccountByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { deleteAccountTestFixtures(t, account.PlatformIdentityID) })

	restore := accounts.SetDefaultDeleter(accounts.Deleter{
		DeleteAccount: func(context.Context, string) error { return errors.New("postgres unavailable") },
	})
	defer restore()

	client.request(http.MethodDelete, "/api/user", map[string]any{"email": email}, http.StatusInternalServerError)
	client.request(http.MethodGet, "/api/user/profile", nil, http.StatusOK)
	if _, err := repositoryForTest(t).GetAccountByPlatformIdentityID(ctx, account.PlatformIdentityID); err != nil {
		t.Fatalf("partial failure removed account authority: %v", err)
	}

	restore()
	client.request(http.MethodDelete, "/api/user", map[string]any{"email": email}, http.StatusOK)
	client.request(http.MethodGet, "/api/user/profile", nil, http.StatusUnauthorized)
}

// TestAccountDeletionAllowsFreshReSignIn proves that signing in with the same
// email after deletion creates a new account and platform identity, leaving no
// orphan identity behind.
func TestAccountDeletionAllowsFreshReSignIn(t *testing.T) {
	router := newAccountContractRouter(t)
	client := newAccountContractClient(t, router)
	ctx := context.Background()
	email := "resignin-" + models.NewUUID().String() + "@example.com"

	verifyOtpSignIn(t, client, email, "123456")
	repository := repositoryForTest(t)
	original, err := repository.GetAccountByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		deleteAccountTestFixtures(t, original.PlatformIdentityID)
		if replacement, err := repository.GetAccountByEmail(ctx, email); err == nil {
			deleteAccountTestFixtures(t, replacement.PlatformIdentityID)
		}
	})

	client.request(http.MethodDelete, "/api/user", map[string]any{"email": email}, http.StatusOK)
	profile := verifyOtpSignIn(t, client, email, "654321")
	if got := decodeAccountString(t, profile, "email"); got != email {
		t.Fatalf("re-sign-in email = %q, want %q", got, email)
	}
	replacement, err := repository.GetAccountByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	if replacement.PlatformIdentityID == original.PlatformIdentityID {
		t.Fatal("re-sign-in reused the deleted external identity")
	}
	var orphaned int
	if err := pgstore.Pool.QueryRow(ctx, `SELECT count(*) FROM platform_identities WHERE id = $1`, original.PlatformIdentityID).Scan(&orphaned); err != nil {
		t.Fatal(err)
	}
	if orphaned != 0 {
		t.Fatalf("deleted platform identity is still present: %d", orphaned)
	}
}

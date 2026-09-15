package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"timeful/server/models"
)

// newAccountsTestRepository applies the schema migrations into a
// transaction-scoped set of temporary tables. Temp tables shadow the real
// schema so the isolated test never mutates test-stack records.
func newAccountsTestRepository(t *testing.T) (context.Context, *Repository, pgx.Tx) {
	t.Helper()
	return newMigrationTestRepository(t)
}

// cleanupAccountUnitsByEmail removes the accounts a test created for one email
// together with their platform identities, so a run leaves no orphan identity
// behind. Accounts are removed first because accounts.platform_identity_id
// references platform_identities.
func cleanupAccountUnitsByEmail(t *testing.T, ctx context.Context, pool *pgxpool.Pool, email string) {
	t.Helper()
	rows, err := pool.Query(ctx, `DELETE FROM accounts WHERE lower(email) = lower($1) RETURNING platform_identity_id`, email)
	if err != nil {
		t.Errorf("delete accounts for %q: %v", email, err)
		return
	}
	identityIDs := []string{}
	for rows.Next() {
		var identityID string
		if err := rows.Scan(&identityID); err != nil {
			rows.Close()
			t.Errorf("scan deleted account identity for %q: %v", email, err)
			return
		}
		identityIDs = append(identityIDs, identityID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Errorf("iterate deleted accounts for %q: %v", email, err)
		return
	}
	if len(identityIDs) == 0 {
		return
	}
	if _, err := pool.Exec(ctx, `DELETE FROM platform_identities WHERE id::text = ANY($1)`, identityIDs); err != nil {
		t.Errorf("delete platform identities for %q: %v", email, err)
	}
}

func TestAccountRepositoryCreatesAndResolvesAccount(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	first, err := repo.CreateAccount(ctx, Account{
		Email: "Ada@example.com", FirstName: "Ada", LastName: "Lovelace",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !validPlatformIdentityID(first.PlatformIdentityID) {
		t.Fatalf("account platform identity is not a canonical UUID: %q", first.PlatformIdentityID)
	}
	if first.ID == "" {
		t.Fatalf("unexpected account %#v", first)
	}
	byPlatformIdentity, err := repo.GetAccountByPlatformIdentityID(ctx, first.PlatformIdentityID)
	if err != nil || byPlatformIdentity.ID != first.ID {
		t.Fatalf("platform identity lookup: %v %#v", err, byPlatformIdentity)
	}
	byEmail, err := repo.GetAccountByEmail(ctx, "ADA@EXAMPLE.COM")
	if err != nil || byEmail.ID != first.ID {
		t.Fatalf("case-insensitive email lookup: %v %#v", err, byEmail)
	}
	if _, err := repo.GetAccountByPlatformIdentityID(ctx, "507f1f77bcf86cd799439011"); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("legacy 24-hex account identifier resolved: %v", err)
	}
}

func TestAccountRepositoryKeepsEqualEmailAccountsDistinct(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	older, err := repo.CreateAccount(ctx, Account{Email: "same@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	newer, err := repo.CreateAccount(ctx, Account{Email: "same@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if older.ID == newer.ID {
		t.Fatal("equal-email accounts must not be merged")
	}
	resolved, err := repo.GetAccountByEmail(ctx, "same@example.com")
	if err != nil || resolved.ID != older.ID {
		t.Fatalf("ambiguous email must resolve deterministically to the oldest: %v %#v", err, resolved)
	}
}

// TestListAccountsByPlatformIdentityIDs proves the batched account read returns
// every requested live account keyed by platform identity and omits missing and
// non-canonical identifiers instead of failing the whole read.
func TestListAccountsByPlatformIdentityIDs(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	first, err := repo.CreateAccount(ctx, Account{Email: "batch-first@example.com", FirstName: "First"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.CreateAccount(ctx, Account{Email: "batch-second@example.com", FirstName: "Second"})
	if err != nil {
		t.Fatal(err)
	}

	accounts, err := repo.ListAccountsByPlatformIdentityIDs(ctx, []string{
		first.PlatformIdentityID,
		second.PlatformIdentityID,
		"507f1f77bcf86cd799439011",
		models.NewUUID().String(),
		first.PlatformIdentityID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 2 {
		t.Fatalf("batched accounts = %d, want 2", len(accounts))
	}
	if got := accounts[first.PlatformIdentityID]; got == nil || got.FirstName != "First" {
		t.Fatalf("first batched account = %#v", got)
	}
	if got := accounts[second.PlatformIdentityID]; got == nil || got.FirstName != "Second" {
		t.Fatalf("second batched account = %#v", got)
	}
	empty, err := repo.ListAccountsByPlatformIdentityIDs(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("empty batched read = %#v, want a non-nil empty map", empty)
	}
}

func TestAccountRepositoryUpdatesAndDeletesProfileAndIdentity(t *testing.T) {
	ctx, repo, tx := newAccountsTestRepository(t)
	account, err := repo.CreateAccount(ctx, Account{Email: "old@example.com", FirstName: "Old"})
	if err != nil {
		t.Fatal(err)
	}
	custom := true
	account.Email = "new@example.com"
	account.FirstName = "New"
	account.LastName = "Name"
	account.HasCustomName = &custom
	account.TimezoneOffset = -300
	if err := repo.UpdateAccountProfile(ctx, account); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetAccountByPlatformIdentityID(ctx, account.PlatformIdentityID)
	if err != nil || stored.Email != "new@example.com" || stored.FirstName != "New" || stored.HasCustomName == nil || !*stored.HasCustomName || stored.TimezoneOffset != -300 {
		t.Fatalf("profile not updated: %v %#v", err, stored)
	}
	if err := repo.DeleteAccountByPlatformIdentityID(ctx, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetAccountByPlatformIdentityID(ctx, account.PlatformIdentityID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("account still resolves after delete")
	}
	if _, err := repo.GetPlatformIdentity(ctx, account.PlatformIdentityID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("platform identity still resolves after delete: %v", err)
	}
	var identities, tombstones int
	if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM platform_identities), (SELECT count(*) FROM account_deletion_tombstones WHERE platform_identity_id = $1)`, account.PlatformIdentityID).Scan(&identities, &tombstones); err != nil {
		t.Fatal(err)
	}
	if identities != 0 {
		t.Fatalf("deleting an account must remove its platform identity, got %d", identities)
	}
	if tombstones != 1 {
		t.Fatalf("deleting an account must record a tombstone, got %d", tombstones)
	}
	// A repeated deletion is idempotent and does not disturb the tombstone.
	if err := repo.DeleteAccountByPlatformIdentityID(ctx, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM account_deletion_tombstones WHERE platform_identity_id = $1`, account.PlatformIdentityID).Scan(&tombstones); err != nil {
		t.Fatal(err)
	}
	if tombstones != 1 {
		t.Fatalf("repeated deletion changed the tombstone count: %d", tombstones)
	}
}

// TestAccountRepositoryDeletedIdentityResolvesToNoAccount proves the hard
// cutover: a deleted platform identity uuid carries a tombstone, resolves to no
// account, and a same-email sign-in mints a new identity instead of adopting
// the deleted one.
func TestAccountRepositoryDeletedIdentityResolvesToNoAccount(t *testing.T) {
	ctx, repo, tx := newAccountsTestRepository(t)
	original, _, err := repo.FindOrCreateAccountByEmail(ctx, "resurrect@example.com", Account{Email: "resurrect@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteAccountByPlatformIdentityID(ctx, original.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	var tombstones int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM account_deletion_tombstones WHERE platform_identity_id = $1`, original.PlatformIdentityID).Scan(&tombstones); err != nil {
		t.Fatal(err)
	}
	if tombstones != 1 {
		t.Fatalf("deleted identity has %d tombstones, want 1", tombstones)
	}
	if _, err := repo.GetAccountByPlatformIdentityID(ctx, original.PlatformIdentityID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("deleted identity resolved an account: %v", err)
	}
	replacement, created, err := repo.FindOrCreateAccountByEmail(ctx, "resurrect@example.com", Account{Email: "resurrect@example.com"})
	if err != nil || !created {
		t.Fatalf("re-sign-in = %v, created=%v; want a fresh account", err, created)
	}
	if replacement.PlatformIdentityID == original.PlatformIdentityID {
		t.Fatal("re-sign-in reused the deleted platform identity")
	}
}

// TestAccountRepositoryDeletionReleasesOwnershipAndRemovesOwnResponses proves
// that events the account organized survive with ownership released and their
// other guests' responses intact, while the account's own response and visitor
// identity are removed.
func TestAccountRepositoryDeletionReleasesOwnershipAndRemovesOwnResponses(t *testing.T) {
	ctx, repo, tx := newAccountsTestRepository(t)
	account, err := repo.CreateAccount(ctx, Account{Email: "owner@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	var eventID string
	if err := tx.QueryRow(ctx, `INSERT INTO events (short_id, name, type, owner_platform_identity_id)
VALUES ('AAAA0001', 'Owned', 'specific_dates', $1) RETURNING id`, account.PlatformIdentityID).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	var ownerVisitorID, guestVisitorID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_visitor_identities (event_id, platform_identity_id) VALUES ($1, $2) RETURNING id`, eventID, account.PlatformIdentityID).Scan(&ownerVisitorID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE events SET owner_event_visitor_identity_id = $2 WHERE id = $1`, eventID, ownerVisitorID); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `INSERT INTO event_visitor_identities (event_id) VALUES ($1) RETURNING id`, eventID).Scan(&guestVisitorID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO event_responses (event_id, event_visitor_identity_id, respondent_kind, platform_identity_id, payload)
VALUES ($1, $2, 'account', $3, '{"name":"Owner"}')`, eventID, ownerVisitorID, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO event_responses (event_id, event_visitor_identity_id, respondent_kind, payload)
VALUES ($1, $2, 'guest', '{"name":"Guest"}')`, eventID, guestVisitorID); err != nil {
		t.Fatal(err)
	}

	if err := repo.DeleteAccountByPlatformIdentityID(ctx, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}

	var owned, ownerVisitor, guestVisitor, ownResponses, guestResponses int
	if err := tx.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM events WHERE id = $1 AND owner_platform_identity_id IS NULL AND owner_event_visitor_identity_id IS NULL),
 (SELECT count(*) FROM event_visitor_identities WHERE id = $2),
 (SELECT count(*) FROM event_visitor_identities WHERE id = $3),
 (SELECT count(*) FROM event_responses WHERE platform_identity_id = $4),
 (SELECT count(*) FROM event_responses WHERE event_id = $1 AND respondent_kind = 'guest')`,
		eventID, ownerVisitorID, guestVisitorID, account.PlatformIdentityID).Scan(&owned, &ownerVisitor, &guestVisitor, &ownResponses, &guestResponses); err != nil {
		t.Fatal(err)
	}
	if owned != 1 {
		t.Fatal("event must survive with ownership released")
	}
	if ownerVisitor != 0 || ownResponses != 0 {
		t.Fatalf("account visitor identity and response must be removed: visitor=%d responses=%d", ownerVisitor, ownResponses)
	}
	if guestVisitor != 1 || guestResponses != 1 {
		t.Fatalf("other guests' responses must survive: visitor=%d responses=%d", guestVisitor, guestResponses)
	}
}

// TestAccountRepositoryDeletionRemovesOwnSignupResponses proves the typed
// uuid[] bind in deleteAccountAuthority removes the deleted account's signup
// responses while leaving another guest's signup response intact. The visitor
// FK cascade is dropped inside the test transaction so the
// event_visitor_identity_id = ANY($2) branch is observable on its own: the
// account's direct response covers the platform_identity_id = $1 clause, and
// the guest response on the account-owned visitor would survive without the
// ANY clause. Another visitor's response is the control.
func TestAccountRepositoryDeletionRemovesOwnSignupResponses(t *testing.T) {
	ctx, repo, tx := newAccountsTestRepository(t)
	account, err := repo.CreateAccount(ctx, Account{Email: "signup-owner@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `ALTER TABLE event_signup_responses DROP CONSTRAINT event_signup_responses_visitor_event_fk`); err != nil {
		t.Fatal(err)
	}

	var eventID, ownerVisitorID, guestVisitorID string
	if err := tx.QueryRow(ctx, `INSERT INTO events (short_id, name, type)
VALUES ($1, 'Signup', 'signup') RETURNING id`, signupTestShortID(t)).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `INSERT INTO event_visitor_identities (event_id, platform_identity_id)
VALUES ($1, $2) RETURNING id`, eventID, account.PlatformIdentityID).Scan(&ownerVisitorID); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `INSERT INTO event_visitor_identities (event_id) VALUES ($1) RETURNING id`, eventID).Scan(&guestVisitorID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO event_signup_responses
 (event_id, event_visitor_identity_id, respondent_kind, platform_identity_id, name)
VALUES ($1, $2, 'account', $3, 'Owner')`, eventID, ownerVisitorID, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO event_signup_responses
 (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name, name)
VALUES ($1, $2, 'guest', 'Owner Guest', 'Owner Guest')`, eventID, ownerVisitorID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO event_signup_responses
 (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name, name)
VALUES ($1, $2, 'guest', 'Guest', 'Guest')`, eventID, guestVisitorID); err != nil {
		t.Fatal(err)
	}

	if err := repo.DeleteAccountByPlatformIdentityID(ctx, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}

	var directResponses, visitorResponses, ownerVisitor int
	if err := tx.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM event_signup_responses WHERE event_id = $1 AND platform_identity_id = $2),
 (SELECT count(*) FROM event_signup_responses WHERE event_id = $1 AND event_visitor_identity_id = $3),
 (SELECT count(*) FROM event_visitor_identities WHERE id = $3)`,
		eventID, account.PlatformIdentityID, ownerVisitorID).Scan(&directResponses, &visitorResponses, &ownerVisitor); err != nil {
		t.Fatal(err)
	}
	if directResponses != 0 {
		t.Fatalf("account's direct signup response must be removed: %d", directResponses)
	}
	if visitorResponses != 0 {
		t.Fatalf("account visitor's signup responses must be removed through the typed array branch: %d", visitorResponses)
	}
	if ownerVisitor != 0 {
		t.Fatalf("account visitor identity must be removed: %d", ownerVisitor)
	}
	var guestStored int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM event_signup_responses WHERE event_id = $1 AND event_visitor_identity_id = $2`, eventID, guestVisitorID).Scan(&guestStored); err != nil {
		t.Fatal(err)
	}
	if guestStored != 1 {
		t.Fatalf("other guest's signup response must survive: %d", guestStored)
	}
}

func TestAccountRepositoryIncrementsUsageCounter(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	account, err := repo.CreateAccount(ctx, Account{Email: "count@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.IncrementAccountEventsCreated(ctx, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetAccountByPlatformIdentityID(ctx, account.PlatformIdentityID)
	if err != nil || stored.NumEventsCreated != 1 {
		t.Fatalf("usage counter = %v %#v", err, stored)
	}
}

// TestCreateAccountRollsBackIdentityOnFailure proves that the platform identity
// and the account are one unit: when the account insert violates a constraint,
// the identity created for the same unit is rolled back instead of leaking a
// half-applied account.
func TestCreateAccountRollsBackIdentityOnFailure(t *testing.T) {
	uri := os.Getenv("POSTGRES_APPLICATION_URI")
	if uri == "" {
		t.Skip("POSTGRES_APPLICATION_URI is required")
	}
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		t.Fatal(err)
	}
	if config.ConnConfig.Database != "timeful-test" && !strings.HasPrefix(config.ConnConfig.Database, "timeful-test-") {
		t.Fatal("requires an isolated test database")
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repo := NewRepository(pool)

	// Count orphan identities, not all identities: packages that run in parallel
	// create account-and-identity units in one transaction, so their identities
	// always have an account and only a leaked identity from this failing unit
	// changes this number.
	orphanIdentities := func() int {
		t.Helper()
		var orphans int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM platform_identities p
WHERE NOT EXISTS (SELECT 1 FROM accounts a WHERE a.platform_identity_id = p.id)`).Scan(&orphans); err != nil {
			t.Fatal(err)
		}
		return orphans
	}
	identitiesBefore := orphanIdentities()
	if _, err := repo.CreateAccount(ctx, Account{Email: "bad@example.com", NumEventsCreated: -1}); err == nil {
		t.Fatal("expected a negative usage counter to fail the account insert")
	}
	identitiesAfter := orphanIdentities()
	if identitiesAfter != identitiesBefore {
		t.Fatalf("failed unit leaked %d platform identity row(s)", identitiesAfter-identitiesBefore)
	}
}

// TestUpdateAccountProfilePreservesUsageCounter proves that a profile update
// cannot change or reset the authoritative usage counter.
func TestUpdateAccountProfilePreservesUsageCounter(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	account, err := repo.CreateAccount(ctx, Account{Email: "counter@example.com", FirstName: "Before"})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.IncrementAccountEventsCreated(ctx, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	if err := repo.IncrementAccountEventsCreated(ctx, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}
	account.FirstName = "After"
	account.NumEventsCreated = 0 // A stale merged value must not reset the stored counter.
	if err := repo.UpdateAccountProfile(ctx, account); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetAccountByPlatformIdentityID(ctx, account.PlatformIdentityID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.FirstName != "After" {
		t.Fatalf("profile update did not persist: %#v", stored)
	}
	if stored.NumEventsCreated != 2 {
		t.Fatalf("profile update changed the usage counter: got %d want 2", stored.NumEventsCreated)
	}
}

func TestFindOrCreateAccountByEmailReusesExistingAccount(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	first, created, err := repo.FindOrCreateAccountByEmail(ctx, "reuse@example.com", Account{Email: "reuse@example.com", FirstName: "First"})
	if err != nil || !created {
		t.Fatalf("first call = %v, created=%v; want a created account", err, created)
	}
	second, created, err := repo.FindOrCreateAccountByEmail(ctx, "REUSE@EXAMPLE.COM", Account{Email: "reuse@example.com", FirstName: "Second"})
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("case-insensitive repeat must reuse the existing account")
	}
	if second.ID != first.ID || second.FirstName != "First" {
		t.Fatalf("repeat changed the account: %#v vs %#v", first, second)
	}
	if second.PlatformIdentityID != first.PlatformIdentityID {
		t.Fatalf("repeat minted a second platform identity: %q vs %q", first.PlatformIdentityID, second.PlatformIdentityID)
	}
	var identities, accounts int
	if err := repo.db.QueryRow(ctx, `SELECT (SELECT count(*) FROM platform_identities), (SELECT count(*) FROM accounts)`).Scan(&identities, &accounts); err != nil {
		t.Fatal(err)
	}
	if identities != 1 || accounts != 1 {
		t.Fatalf("repeat created extra rows: identities=%d accounts=%d", identities, accounts)
	}
}

// TestFindOrCreateAccountByEmailConcurrentSignIns proves that concurrent
// first-time sign-ins for one email serialize behind the advisory lock and
// create exactly one account and platform identity, with no failed request.
func TestFindOrCreateAccountByEmailConcurrentSignIns(t *testing.T) {
	uri := os.Getenv("POSTGRES_APPLICATION_URI")
	if uri == "" {
		t.Skip("POSTGRES_APPLICATION_URI is required")
	}
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		t.Fatal(err)
	}
	if config.ConnConfig.Database != "timeful-test" && !strings.HasPrefix(config.ConnConfig.Database, "timeful-test-") {
		t.Fatal("requires an isolated test database")
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repo := NewRepository(pool)

	email := "concurrent-email-" + models.NewUUID().String() + "@example.com"
	t.Cleanup(func() {
		cleanupAccountUnitsByEmail(t, context.Background(), pool, email)
	})

	const workers = 8
	results := make([]*Account, workers)
	failures := make([]error, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i], _, failures[i] = repo.FindOrCreateAccountByEmail(ctx, email, Account{Email: email, FirstName: "Racer"})
		}(i)
	}
	close(start)
	wg.Wait()

	var winner *Account
	for i := range results {
		if failures[i] != nil {
			t.Fatalf("worker %d failed: %v", i, failures[i])
		}
		if results[i] == nil {
			t.Fatalf("worker %d returned no account", i)
		}
		if winner == nil {
			winner = results[i]
		} else if results[i].ID != winner.ID {
			t.Fatalf("concurrent sign-ins resolved different accounts: %s vs %s", winner.ID, results[i].ID)
		}
	}
	var accounts, identities int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM accounts WHERE lower(email) = lower($1)), (SELECT count(*) FROM platform_identities WHERE id = $2)`, email, winner.PlatformIdentityID).Scan(&accounts, &identities); err != nil {
		t.Fatal(err)
	}
	if accounts != 1 || identities != 1 {
		t.Fatalf("concurrent first-time sign-ins created accounts=%d identities=%d", accounts, identities)
	}
}

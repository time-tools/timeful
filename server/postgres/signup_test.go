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
)

// newSignupTestRepository applies the schema migrations into a
// transaction-scoped set of temporary tables. Temp tables shadow the real
// schema so the isolated test never mutates test-stack records.
func newSignupTestRepository(t *testing.T) (context.Context, *Repository, pgx.Tx) {
	t.Helper()
	return newMigrationTestRepository(t)
}

func signupTestShortID(t *testing.T) string {
	t.Helper()
	shortID, err := GenerateShortID()
	if err != nil {
		t.Fatal(err)
	}
	return shortID
}

// expectSavepointError runs fn and requires an error without aborting the
// enclosing test transaction, which PostgreSQL marks failed after any statement
// error.
func expectSavepointError(t *testing.T, ctx context.Context, tx pgx.Tx, fn func() error) error {
	t.Helper()
	if _, err := tx.Exec(ctx, `SAVEPOINT signup_expected_error`); err != nil {
		t.Fatal(err)
	}
	err := fn()
	if err == nil {
		t.Fatal("expected operation to fail")
	}
	if _, err := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT signup_expected_error`); err != nil {
		t.Fatal(err)
	}
	return err
}

func seedSignupEvent(t *testing.T, ctx context.Context, tx pgx.Tx, shortID string) string {
	t.Helper()
	var eventID string
	if err := tx.QueryRow(ctx, `INSERT INTO events (short_id, name, type)
VALUES ($1, 'Signup', 'signup') RETURNING id`, shortID).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	return eventID
}

func seedSignupVisitor(t *testing.T, ctx context.Context, tx pgx.Tx, eventID string) string {
	t.Helper()
	var visitorID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_visitor_identities (event_id) VALUES ($1) RETURNING id`, eventID).Scan(&visitorID); err != nil {
		t.Fatal(err)
	}
	return visitorID
}

func seedSignupBlock(t *testing.T, ctx context.Context, tx pgx.Tx, eventID, name string) string {
	t.Helper()
	var blockID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_signup_blocks (event_id, name) VALUES ($1, $2) RETURNING id::text`, eventID, name).Scan(&blockID); err != nil {
		t.Fatal(err)
	}
	return blockID
}

func assertSignupBlockIDs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("block IDs = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("block IDs = %#v, want %#v", got, want)
		}
	}
}

// TestSignupSchemaConstraints proves the baseline admits the signup kind,
// enforces the visitor/event relation, rejects negative capacity, and requires
// a canonical guest name for guest responses.
func TestSignupSchemaConstraints(t *testing.T) {
	ctx, _, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))

	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO events (short_id, name, type) VALUES ($1, 'Bogus', 'bogus')`, signupTestShortID(t))
		return err
	})
	otherEventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO event_signup_responses (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name)
VALUES ($1, $2, 'guest', 'Ada')`, otherEventID, visitorID)
		return err
	})
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO event_signup_blocks (event_id, name, capacity) VALUES ($1, 'Bad', -1)`, eventID)
		return err
	})
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO event_signup_responses (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name)
VALUES ($1, $2, 'guest', NULL)`, eventID, visitorID)
		return err
	})
	expectSavepointError(t, ctx, tx, func() error {
		_, err := tx.Exec(ctx, `INSERT INTO event_signup_responses (event_id, event_visitor_identity_id, respondent_kind, platform_identity_id, canonical_guest_name)
VALUES ($1, $2, 'account', NULL, 'Ada')`, eventID, visitorID)
		return err
	})
}

// TestSignupBlocksCreateReplaceAndListPreserveOrder covers block create, list
// ordering, and replacement, including detaching removed blocks from the
// responses that claimed them.
func TestSignupBlocksCreateReplaceAndListPreserveOrder(t *testing.T) {
	ctx, repo, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)

	capacity := 1
	first := &SignupBlock{EventID: eventID, Name: "Morning", Capacity: &capacity}
	if err := repo.CreateSignupBlock(ctx, first); err != nil {
		t.Fatal(err)
	}
	second := &SignupBlock{EventID: eventID, Name: "Afternoon"}
	if err := repo.CreateSignupBlock(ctx, second); err != nil {
		t.Fatal(err)
	}
	listed, err := repo.ListSignupBlocks(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 || listed[0].ID != first.ID || listed[1].ID != second.ID {
		t.Fatalf("blocks are not in insertion order: %#v", listed)
	}
	if listed[0].Position != 1 || listed[1].Position != 2 {
		t.Fatalf("appended positions are unexpected: %#v", listed)
	}

	response := &SignupResponse{EventID: eventID, EventVisitorIdentityID: visitorID, Name: "Ada", BlockIDs: []string{first.ID}}
	if err := repo.CreateSignupResponse(ctx, response); err != nil {
		t.Fatal(err)
	}

	replacement, err := repo.ReplaceSignupBlocks(ctx, eventID, []SignupBlock{
		{ID: second.ID, Name: "Afternoon", Position: 5},
		{Name: "Evening"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(replacement) != 2 || replacement[0].ID != second.ID || replacement[0].Name != "Afternoon" || replacement[1].Name != "Evening" {
		t.Fatalf("replace did not apply: %#v", replacement)
	}
	if replacement[0].Position != 1 || replacement[1].Position != 2 {
		t.Fatalf("replace did not renumber positions: %#v", replacement)
	}
	stored, err := repo.GetSignupResponseByPublicID(ctx, eventID, response.PublicID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.BlockIDs) != 0 {
		t.Fatalf("removed block identity stayed attached to the response: %#v", stored.BlockIDs)
	}
	if _, err := repo.ReplaceSignupBlocks(ctx, eventID, []SignupBlock{{ID: response.ID, Name: "Foreign"}}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("replacing a foreign block error = %v, want pgx.ErrNoRows", err)
	}
}

// TestSignupResponseCanonicalGuestNameAndMembership proves guest names are
// canonicalized through the shared normalizer, equivalent names collide, and
// update/delete are keyed by the opaque public ID.
func TestSignupResponseCanonicalGuestNameAndMembership(t *testing.T) {
	ctx, repo, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)
	otherVisitorID := seedSignupVisitor(t, ctx, tx, eventID)

	block := &SignupBlock{EventID: eventID, Name: "Open"}
	if err := repo.CreateSignupBlock(ctx, block); err != nil {
		t.Fatal(err)
	}
	response := &SignupResponse{
		EventID:                eventID,
		EventVisitorIdentityID: visitorID,
		Name:                   "  A\u200bda\u0301  ",
		BlockIDs:               []string{block.ID, block.ID, " "},
	}
	if err := repo.CreateSignupResponse(ctx, response); err != nil {
		t.Fatal(err)
	}
	if response.Name != "Adá" || response.CanonicalGuestName == nil || *response.CanonicalGuestName != "Adá" {
		t.Fatalf("guest name was not canonicalized: %#v", response)
	}
	if len(response.BlockIDs) != 1 || response.BlockIDs[0] != block.ID {
		t.Fatalf("block membership was not normalized: %#v", response.BlockIDs)
	}
	if response.PublicID == "" || response.ID == "" {
		t.Fatalf("response identity is missing: %#v", response)
	}

	duplicate := &SignupResponse{EventID: eventID, EventVisitorIdentityID: otherVisitorID, Name: "Adá", BlockIDs: []string{block.ID}}
	duplicateErr := expectSavepointError(t, ctx, tx, func() error {
		return repo.CreateSignupResponse(ctx, duplicate)
	})
	if !IsUniqueViolation(duplicateErr) {
		t.Fatalf("duplicate canonical guest name error = %v, want a unique violation", duplicateErr)
	}

	response.Name = "Bob"
	response.Email = "bob@example.com"
	if err := repo.UpdateSignupResponse(ctx, response); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetSignupResponseByPublicID(ctx, eventID, response.PublicID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Name != "Bob" || stored.CanonicalGuestName == nil || *stored.CanonicalGuestName != "Bob" || stored.Email != "bob@example.com" {
		t.Fatalf("update did not persist: %#v", stored)
	}
	if err := repo.DeleteSignupResponse(ctx, eventID, response.PublicID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetSignupResponseByPublicID(ctx, eventID, response.PublicID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("deleted response lookup error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.DeleteSignupResponse(ctx, eventID, response.PublicID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("repeated delete error = %v, want pgx.ErrNoRows", err)
	}
}

// TestSignupResponseAccountIdentityAndBlockValidation proves account signup
// responses resolve and deduplicate by account user ID and that a response
// cannot claim a block from another event.
func TestSignupResponseAccountIdentityAndBlockValidation(t *testing.T) {
	ctx, repo, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)
	otherVisitorID := seedSignupVisitor(t, ctx, tx, eventID)
	foreignEventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))

	var accountID string
	if err := tx.QueryRow(ctx, `INSERT INTO platform_identities DEFAULT VALUES RETURNING id`).Scan(&accountID); err != nil {
		t.Fatal(err)
	}
	response := &SignupResponse{EventID: eventID, EventVisitorIdentityID: visitorID, PlatformIdentityID: &accountID}
	if err := repo.CreateSignupResponse(ctx, response); err != nil {
		t.Fatal(err)
	}
	if response.RespondentKind != RespondentKindAccount || response.CanonicalGuestName != nil {
		t.Fatalf("account identity was not resolved: %#v", response)
	}
	duplicate := &SignupResponse{EventID: eventID, EventVisitorIdentityID: otherVisitorID, PlatformIdentityID: &accountID}
	duplicateErr := expectSavepointError(t, ctx, tx, func() error {
		return repo.CreateSignupResponse(ctx, duplicate)
	})
	if !IsUniqueViolation(duplicateErr) {
		t.Fatalf("duplicate account response error = %v, want a unique violation", duplicateErr)
	}

	foreignBlockID := ""
	if err := tx.QueryRow(ctx, `INSERT INTO event_signup_blocks (event_id, name) VALUES ($1, 'Foreign') RETURNING id::text`, foreignEventID).Scan(&foreignBlockID); err != nil {
		t.Fatal(err)
	}
	guest := &SignupResponse{EventID: eventID, EventVisitorIdentityID: otherVisitorID, Name: "Carol", BlockIDs: []string{foreignBlockID}}
	foreignErr := expectSavepointError(t, ctx, tx, func() error {
		return repo.CreateSignupResponse(ctx, guest)
	})
	if !errors.Is(foreignErr, ErrSignupBlockNotFound) {
		t.Fatalf("foreign block error = %v, want ErrSignupBlockNotFound", foreignErr)
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM event_signup_responses WHERE event_id = $1`, eventID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("a rejected signup wrote %d responses, want 1", count)
	}
}

// TestSignupCapacityReservationAtomicUnderContention proves concurrent signups
// for one limited block serialize under the event row lock and admit exactly the
// capacity count with no partial writes.
func TestSignupCapacityReservationAtomicUnderContention(t *testing.T) {
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

	shortID := signupTestShortID(t)
	var eventID, blockID string
	if err := pool.QueryRow(ctx, `INSERT INTO events (short_id, name, type)
VALUES ($1, 'Contended', 'signup') RETURNING id`, shortID).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM events WHERE id = $1`, eventID); err != nil {
			t.Errorf("delete contended event: %v", err)
		}
	})
	capacity := 2
	if err := pool.QueryRow(ctx, `INSERT INTO event_signup_blocks (event_id, name, capacity)
VALUES ($1, 'Limited', $2) RETURNING id::text`, eventID, capacity).Scan(&blockID); err != nil {
		t.Fatal(err)
	}

	const workers = 8
	visitorIDs := make([]string, workers)
	for i := range visitorIDs {
		if err := pool.QueryRow(ctx, `INSERT INTO event_visitor_identities (event_id) VALUES ($1) RETURNING id::text`, eventID).Scan(&visitorIDs[i]); err != nil {
			t.Fatal(err)
		}
	}

	failures := make([]error, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			failures[i] = repo.CreateSignupResponse(ctx, &SignupResponse{
				EventID:                eventID,
				EventVisitorIdentityID: visitorIDs[i],
				Name:                   "Racer " + string(rune('A'+i)),
				BlockIDs:               []string{blockID},
			})
		}(i)
	}
	close(start)
	wg.Wait()

	admitted, rejected := 0, 0
	for i, err := range failures {
		switch {
		case err == nil:
			admitted++
		case errors.Is(err, ErrSignupCapacityExceeded):
			rejected++
		default:
			t.Fatalf("worker %d error = %v", i, err)
		}
	}
	if admitted != capacity || rejected != workers-capacity {
		t.Fatalf("capacity contention admitted=%d rejected=%d, want %d/%d", admitted, rejected, capacity, workers-capacity)
	}
	var stored int
	if err := pool.QueryRow(ctx, `SELECT count(*)
FROM event_signup_responses response
JOIN event_signup_response_blocks membership ON membership.response_id = response.id
WHERE response.event_id = $1 AND membership.block_id::text = $2`, eventID, blockID).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != capacity {
		t.Fatalf("stored signups = %d, want %d", stored, capacity)
	}
}

// TestSignupUnlimitedBlockAdmitsAll proves a nil capacity does not reject.
func TestSignupUnlimitedBlockAdmitsAll(t *testing.T) {
	ctx, repo, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	block := &SignupBlock{EventID: eventID, Name: "Open"}
	if err := repo.CreateSignupBlock(ctx, block); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		visitorID := seedSignupVisitor(t, ctx, tx, eventID)
		response := &SignupResponse{EventID: eventID, EventVisitorIdentityID: visitorID, Name: "Guest " + string(rune('A'+i)), BlockIDs: []string{block.ID}}
		if err := repo.CreateSignupResponse(ctx, response); err != nil {
			t.Fatal(err)
		}
	}
	responses, err := repo.ListSignupResponses(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	if len(responses) != 5 {
		t.Fatalf("unlimited block admitted %d signups, want 5", len(responses))
	}
}

// TestSignupResponseBlockMembershipRoundTripOrder proves a response's claimed
// blocks round-trip through the join table in write order on create, public-ID
// read, list, and update, and that deleting the response cascades the
// memberships away.
func TestSignupResponseBlockMembershipRoundTripOrder(t *testing.T) {
	ctx, repo, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)
	firstBlockID := seedSignupBlock(t, ctx, tx, eventID, "First")
	secondBlockID := seedSignupBlock(t, ctx, tx, eventID, "Second")

	response := &SignupResponse{
		EventID:                eventID,
		EventVisitorIdentityID: visitorID,
		Name:                   "Ada",
		BlockIDs:               []string{secondBlockID, firstBlockID, secondBlockID, " "},
	}
	if err := repo.CreateSignupResponse(ctx, response); err != nil {
		t.Fatal(err)
	}
	assertSignupBlockIDs(t, response.BlockIDs, []string{secondBlockID, firstBlockID})

	stored, err := repo.GetSignupResponseByPublicID(ctx, eventID, response.PublicID)
	if err != nil {
		t.Fatal(err)
	}
	assertSignupBlockIDs(t, stored.BlockIDs, []string{secondBlockID, firstBlockID})

	listed, err := repo.ListSignupResponses(ctx, eventID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("listed %d responses, want 1", len(listed))
	}
	assertSignupBlockIDs(t, listed[0].BlockIDs, []string{secondBlockID, firstBlockID})

	if err := repo.UpdateSignupResponse(ctx, &SignupResponse{
		ID:       response.ID,
		EventID:  eventID,
		Name:     "Ada",
		BlockIDs: []string{firstBlockID, secondBlockID},
	}); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetSignupResponseByPublicID(ctx, eventID, response.PublicID)
	if err != nil {
		t.Fatal(err)
	}
	assertSignupBlockIDs(t, stored.BlockIDs, []string{firstBlockID, secondBlockID})

	if err := repo.DeleteSignupResponse(ctx, eventID, response.PublicID); err != nil {
		t.Fatal(err)
	}
	var memberships int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM event_signup_response_blocks WHERE response_id = $1`, response.ID).Scan(&memberships); err != nil {
		t.Fatal(err)
	}
	if memberships != 0 {
		t.Fatalf("deleting the response left %d memberships", memberships)
	}
}

// TestSignupCapacityClaimsAndMissingBlocks proves the grouped capacity count
// reads the join table, excludes the response being updated, and maps foreign,
// nonexistent, and non-canonical block identifiers to ErrSignupBlockNotFound
// instead of a cast error.
func TestSignupCapacityClaimsAndMissingBlocks(t *testing.T) {
	ctx, repo, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	otherEventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	foreignBlockID := seedSignupBlock(t, ctx, tx, otherEventID, "Foreign")
	capacity := 1
	block := &SignupBlock{EventID: eventID, Name: "One seat", Capacity: &capacity}
	if err := repo.CreateSignupBlock(ctx, block); err != nil {
		t.Fatal(err)
	}

	first := &SignupResponse{EventID: eventID, EventVisitorIdentityID: seedSignupVisitor(t, ctx, tx, eventID), Name: "Ada", BlockIDs: []string{block.ID}}
	if err := repo.CreateSignupResponse(ctx, first); err != nil {
		t.Fatal(err)
	}
	// Re-saving the claiming response excludes its own membership from the count.
	first.Email = "ada@example.com"
	if err := repo.UpdateSignupResponse(ctx, first); err != nil {
		t.Fatal(err)
	}

	full := &SignupResponse{EventID: eventID, EventVisitorIdentityID: seedSignupVisitor(t, ctx, tx, eventID), Name: "Grace", BlockIDs: []string{block.ID}}
	if err := repo.CreateSignupResponse(ctx, full); !errors.Is(err, ErrSignupCapacityExceeded) {
		t.Fatalf("second claim error = %v, want ErrSignupCapacityExceeded", err)
	}

	missing := []string{foreignBlockID, "00000000-0000-0000-0000-000000000000", "not-a-uuid"}
	for i, blockID := range missing {
		guest := &SignupResponse{
			EventID:                eventID,
			EventVisitorIdentityID: seedSignupVisitor(t, ctx, tx, eventID),
			Name:                   "Missing " + string(rune('A'+i)),
			BlockIDs:               []string{blockID},
		}
		if err := repo.CreateSignupResponse(ctx, guest); !errors.Is(err, ErrSignupBlockNotFound) {
			t.Fatalf("block %q error = %v, want ErrSignupBlockNotFound", blockID, err)
		}
	}
}

// TestSignupCapacityIgnoresOtherEventMemberships proves the grouped capacity
// count only credits claims made by responses of the same event: a cross-event
// membership inserted by direct SQL must not consume the block's seats.
func TestSignupCapacityIgnoresOtherEventMemberships(t *testing.T) {
	ctx, repo, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	otherEventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	capacity := 2
	block := &SignupBlock{EventID: eventID, Name: "Two seats", Capacity: &capacity}
	if err := repo.CreateSignupBlock(ctx, block); err != nil {
		t.Fatal(err)
	}
	claimant := &SignupResponse{EventID: eventID, EventVisitorIdentityID: seedSignupVisitor(t, ctx, tx, eventID), Name: "Ada", BlockIDs: []string{block.ID}}
	if err := repo.CreateSignupResponse(ctx, claimant); err != nil {
		t.Fatal(err)
	}

	// The repository filters memberships by event, so the cross-event row only
	// exists through direct SQL.
	var foreignResponseID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_signup_responses (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name)
VALUES ($1, $2, 'guest', 'Grace') RETURNING id`, otherEventID, seedSignupVisitor(t, ctx, tx, otherEventID)).Scan(&foreignResponseID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO event_signup_response_blocks (response_id, block_id, position)
VALUES ($1, $2, 1)`, foreignResponseID, block.ID); err != nil {
		t.Fatal(err)
	}

	second := &SignupResponse{EventID: eventID, EventVisitorIdentityID: seedSignupVisitor(t, ctx, tx, eventID), Name: "Carol", BlockIDs: []string{block.ID}}
	if err := repo.CreateSignupResponse(ctx, second); err != nil {
		t.Fatalf("second same-event claim error = %v, want admission", err)
	}
	assertSignupBlockIDs(t, second.BlockIDs, []string{block.ID})
}

// TestWriteSignupResponseMembershipsRejectsUnwrittenIDs proves the set-based
// membership write reports ErrSignupBlockNotFound instead of silently dropping
// an identity that does not name a same-event block, even without the capacity
// pre-check.
func TestWriteSignupResponseMembershipsRejectsUnwrittenIDs(t *testing.T) {
	ctx, repo, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	otherEventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	blockID := seedSignupBlock(t, ctx, tx, eventID, "Open")
	foreignBlockID := seedSignupBlock(t, ctx, tx, otherEventID, "Foreign")

	var responseID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_signup_responses (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name)
VALUES ($1, $2, 'guest', 'Ada') RETURNING id`, eventID, seedSignupVisitor(t, ctx, tx, eventID)).Scan(&responseID); err != nil {
		t.Fatal(err)
	}
	if err := repo.writeSignupResponseMemberships(ctx, responseID, eventID, []string{blockID, foreignBlockID}); !errors.Is(err, ErrSignupBlockNotFound) {
		t.Fatalf("foreign membership error = %v, want ErrSignupBlockNotFound", err)
	}
}

// TestSignupResponseBlockJoinTableSchema proves the migration created the join
// table with cascading foreign keys on both sides, a unique (response_id,
// block_id) constraint, a NOT NULL position, and the (block_id, response_id)
// and (response_id, position, block_id) indexes that serve capacity counts and
// ordered membership reads.
func TestSignupResponseBlockJoinTableSchema(t *testing.T) {
	ctx, _, tx := newSignupTestRepository(t)
	var positionNullable string
	if err := tx.QueryRow(ctx, `SELECT is_nullable FROM information_schema.columns
WHERE table_name = 'event_signup_response_blocks' AND column_name = 'position'`).Scan(&positionNullable); err != nil {
		t.Fatal(err)
	}
	if positionNullable != "NO" {
		t.Fatalf("position is_nullable = %q, want NO", positionNullable)
	}
	var cascadeCount int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM pg_constraint
WHERE conrelid = 'event_signup_response_blocks'::regclass
  AND contype = 'f'
  AND confdeltype = 'c'`).Scan(&cascadeCount); err != nil {
		t.Fatal(err)
	}
	if cascadeCount != 2 {
		t.Fatalf("cascading foreign keys = %d, want 2", cascadeCount)
	}
	var uniqueExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conrelid = 'event_signup_response_blocks'::regclass
      AND contype = 'u'
      AND conkey = ARRAY[
          (SELECT attnum FROM pg_attribute WHERE attrelid = 'event_signup_response_blocks'::regclass AND attname = 'response_id'),
          (SELECT attnum FROM pg_attribute WHERE attrelid = 'event_signup_response_blocks'::regclass AND attname = 'block_id')
      ]::smallint[]
)`).Scan(&uniqueExists); err != nil {
		t.Fatal(err)
	}
	if !uniqueExists {
		t.Fatal("unique (response_id, block_id) constraint is missing")
	}
	for _, indexName := range []string{
		"event_signup_response_blocks_block_id_idx",
		"event_signup_response_blocks_response_position_idx",
	} {
		var resolved string
		if err := tx.QueryRow(ctx, `SELECT COALESCE(to_regclass($1)::text, '')`, indexName).Scan(&resolved); err != nil {
			t.Fatal(err)
		}
		if resolved == "" {
			t.Fatalf("%s is missing", indexName)
		}
	}
}

// TestReplaceSignupBlocksDuplicatesCollapseToLastOccurrence proves a repeated
// supplied block identity is applied once with its last occurrence, so the
// set-based upsert cannot raise PostgreSQL 21000, and that mixed new and kept
// blocks still renumber in supplied order.
func TestReplaceSignupBlocksDuplicatesCollapseToLastOccurrence(t *testing.T) {
	ctx, repo, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	first := &SignupBlock{EventID: eventID, Name: "First"}
	if err := repo.CreateSignupBlock(ctx, first); err != nil {
		t.Fatal(err)
	}
	second := &SignupBlock{EventID: eventID, Name: "Second"}
	if err := repo.CreateSignupBlock(ctx, second); err != nil {
		t.Fatal(err)
	}
	replacement, err := repo.ReplaceSignupBlocks(ctx, eventID, []SignupBlock{
		{ID: second.ID, Name: "Second renamed first"},
		{ID: first.ID, Name: "First renamed"},
		{Name: "Third"},
		{ID: second.ID, Name: "Second renamed last"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(replacement) != 3 || replacement[0].ID != first.ID || replacement[0].Name != "First renamed" ||
		replacement[1].Name != "Third" || replacement[2].ID != second.ID || replacement[2].Name != "Second renamed last" {
		t.Fatalf("duplicate replacement = %#v", replacement)
	}
	if replacement[0].Position != 1 || replacement[1].Position != 2 || replacement[2].Position != 3 {
		t.Fatalf("duplicate replacement positions = %#v", replacement)
	}
}

// TestSignupResponseBlockMembershipBackfill proves the join-table migration
// copies legacy block_ids in stored array order, collapses a repeated identity
// to its first occurrence, skips values that do not name a same-event block, and
// drops the retired column.
func TestSignupResponseBlockMembershipBackfill(t *testing.T) {
	ctx, tx := newMigrationTestTransaction(t)
	applyMigration(t, ctx, tx, "20260912000000_baseline_schema.sql")

	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	otherEventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)
	firstBlockID := seedSignupBlock(t, ctx, tx, eventID, "First")
	secondBlockID := seedSignupBlock(t, ctx, tx, eventID, "Second")
	foreignBlockID := seedSignupBlock(t, ctx, tx, otherEventID, "Foreign")

	var responseID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_signup_responses (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name, block_ids)
VALUES ($1, $2, 'guest', 'Ada', $3) RETURNING id`,
		eventID, visitorID, []string{secondBlockID, firstBlockID, secondBlockID, foreignBlockID, "not-a-uuid"}).Scan(&responseID); err != nil {
		t.Fatal(err)
	}
	var emptyResponseID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_signup_responses (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name)
VALUES ($1, $2, 'guest', 'Grace') RETURNING id`, eventID, seedSignupVisitor(t, ctx, tx, eventID)).Scan(&emptyResponseID); err != nil {
		t.Fatal(err)
	}

	applyMigration(t, ctx, tx, "20260913000000_signup_response_blocks.sql")

	if hasColumn(t, ctx, tx, "event_signup_responses", "block_ids") {
		t.Fatal("block_ids survived the join-table migration")
	}
	rows, err := tx.Query(ctx, `SELECT block_id::text, position
FROM event_signup_response_blocks WHERE response_id = $1 ORDER BY position`, responseID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	blockIDs := []string{}
	positions := []int{}
	for rows.Next() {
		var blockID string
		var position int
		if err := rows.Scan(&blockID, &position); err != nil {
			t.Fatal(err)
		}
		blockIDs = append(blockIDs, blockID)
		positions = append(positions, position)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	assertSignupBlockIDs(t, blockIDs, []string{secondBlockID, firstBlockID})
	if len(positions) != 2 || positions[0] != 1 || positions[1] != 2 {
		t.Fatalf("backfilled positions = %#v, want [1 2]", positions)
	}
	var emptyCount int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM event_signup_response_blocks WHERE response_id = $1`, emptyResponseID).Scan(&emptyCount); err != nil {
		t.Fatal(err)
	}
	if emptyCount != 0 {
		t.Fatalf("empty block_ids backfilled %d memberships", emptyCount)
	}
}

// TestSignupResponseBlockMembershipDownMigration proves the down migration
// restores block_ids in stored position order, reinstates the empty-array
// default, and drops the join table.
func TestSignupResponseBlockMembershipDownMigration(t *testing.T) {
	ctx, tx := newMigrationTestTransaction(t)
	applyMigration(t, ctx, tx, "20260912000000_baseline_schema.sql")

	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	firstBlockID := seedSignupBlock(t, ctx, tx, eventID, "First")
	secondBlockID := seedSignupBlock(t, ctx, tx, eventID, "Second")

	var responseID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_signup_responses (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name, block_ids)
VALUES ($1, $2, 'guest', 'Ada', $3) RETURNING id`,
		eventID, seedSignupVisitor(t, ctx, tx, eventID), []string{secondBlockID, firstBlockID}).Scan(&responseID); err != nil {
		t.Fatal(err)
	}

	applyMigration(t, ctx, tx, "20260913000000_signup_response_blocks.sql")
	applyMigrationDown(t, ctx, tx, "20260913000000_signup_response_blocks.sql")

	if !hasColumn(t, ctx, tx, "event_signup_responses", "block_ids") {
		t.Fatal("block_ids was not restored by the down migration")
	}
	var restored []string
	if err := tx.QueryRow(ctx, `SELECT block_ids FROM event_signup_responses WHERE id = $1`, responseID).Scan(&restored); err != nil {
		t.Fatal(err)
	}
	assertSignupBlockIDs(t, restored, []string{secondBlockID, firstBlockID})

	var tempJoinTableExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (
    SELECT 1 FROM pg_class
    WHERE relname = 'event_signup_response_blocks'
      AND relnamespace = pg_my_temp_schema()
)`).Scan(&tempJoinTableExists); err != nil {
		t.Fatal(err)
	}
	if tempJoinTableExists {
		t.Fatal("the down migration left the temporary join table in place")
	}

	var defaultResponseID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_signup_responses (event_id, event_visitor_identity_id, respondent_kind, canonical_guest_name)
VALUES ($1, $2, 'guest', 'Grace') RETURNING id`, eventID, seedSignupVisitor(t, ctx, tx, eventID)).Scan(&defaultResponseID); err != nil {
		t.Fatal(err)
	}
	var defaulted []string
	if err := tx.QueryRow(ctx, `SELECT block_ids FROM event_signup_responses WHERE id = $1`, defaultResponseID).Scan(&defaulted); err != nil {
		t.Fatal(err)
	}
	if len(defaulted) != 0 {
		t.Fatalf("restored block_ids default = %#v, want empty", defaulted)
	}
}

// TestSignupPublicIDLookupsRejectNonCanonicalIdentifiers proves the signup
// public-ID lookups validate the opaque identifier before binding the uuid
// column, so a non-canonical value keeps reporting pgx.ErrNoRows for reads,
// deletes, and updates instead of a PostgreSQL 22P02 cast error.
func TestSignupPublicIDLookupsRejectNonCanonicalIdentifiers(t *testing.T) {
	ctx, repo, tx := newSignupTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)
	response := &SignupResponse{EventID: eventID, EventVisitorIdentityID: visitorID, Name: "Ada"}
	if err := repo.CreateSignupResponse(ctx, response); err != nil {
		t.Fatal(err)
	}

	for _, input := range []string{
		"not-a-uuid",
		"507f1f77bcf86cd799439011",
		"0198E6F0-6A3A-7C4B-9A2D-4F6A1B2C3D4E",
		"0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e-",
	} {
		if _, err := repo.GetSignupResponseByPublicID(ctx, eventID, input); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("signup public ID %q read error = %v, want pgx.ErrNoRows", input, err)
		}
		if err := repo.DeleteSignupResponse(ctx, eventID, input); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("signup public ID %q delete error = %v, want pgx.ErrNoRows", input, err)
		}
		update := &SignupResponse{PublicID: input, EventID: eventID, Name: "Ada"}
		if err := repo.UpdateSignupResponse(ctx, update); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("signup public ID %q update error = %v, want pgx.ErrNoRows", input, err)
		}
	}

	if _, err := repo.GetSignupResponseByPublicID(ctx, eventID, response.PublicID); err != nil {
		t.Fatalf("canonical lookup after rejected identifiers: %v", err)
	}
}

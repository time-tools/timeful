package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestVisitorIdentityRepository(t *testing.T) {
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
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	repo := &Repository{db: tx}
	first := &Event{Name: "Identity repository", Type: EventTypeSpecificDates}
	second := &Event{Name: "Other event", Type: EventTypeSpecificDates}
	for _, event := range []*Event{first, second} {
		if err := repo.CreateEvent(ctx, event); err != nil {
			t.Fatal(err)
		}
	}
	visitor, err := repo.CreateEventVisitorIdentity(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if visitor.ID == visitor.PublicID || visitor.PublicID == "" {
		t.Fatal("public identity must be separate")
	}
	platform, err := repo.CreatePlatformIdentity(ctx)
	if err != nil {
		t.Fatal(err)
	}
	same, err := repo.GetPlatformIdentity(ctx, platform.ID)
	if err != nil || same.ID != platform.ID {
		t.Fatalf("unstable platform identity: %v", err)
	}
	if _, err := repo.GetPlatformIdentity(ctx, "507f1f77bcf86cd799439011"); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("legacy 24-hex identifier resolved a platform identity: %v", err)
	}
	if err := repo.AssociateEventVisitorIdentity(ctx, visitor.ID, platform.ID); err != nil {
		t.Fatal(err)
	}
	if ok, err := repo.VisitorBelongsToAccount(ctx, visitor.ID, platform.ID); err != nil || !ok {
		t.Fatalf("association: %v", err)
	}
	if ok, err := repo.VisitorBelongsToAccount(ctx, visitor.ID, "0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e"); err != nil || ok {
		t.Fatalf("unrelated account authorized: %v", err)
	}
	if ok, err := repo.VisitorBelongsToAccount(ctx, visitor.ID, "unrelated"); err != nil || ok {
		t.Fatalf("non-canonical account authorized: %v", err)
	}
	otherVisitor, err := repo.CreateEventVisitorIdentity(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	owned, err := repo.EventVisitorIdentitiesBelongingToAccount(ctx, platform.ID, []string{visitor.ID, otherVisitor.ID, visitor.ID})
	if err != nil || !owned[visitor.ID] || owned[otherVisitor.ID] {
		t.Fatalf("batched association = %#v, %v", owned, err)
	}
	if owned, err := repo.EventVisitorIdentitiesBelongingToAccount(ctx, "unrelated", []string{visitor.ID}); err != nil || len(owned) != 0 {
		t.Fatalf("non-canonical batched association = %#v, %v", owned, err)
	}
	if hasResponse, err := repo.EventVisitorHasResponse(ctx, first.ID, visitor.ID); err != nil || hasResponse {
		t.Fatalf("unexpected pre-existing response: %v, %v", hasResponse, err)
	}
	locked, err := repo.LockEvent(ctx, first.ID)
	if err != nil || locked.ID != first.ID || locked.ShortID != first.ShortID {
		t.Fatalf("lock event = %#v, %v", locked, err)
	}
	hash := sha256.Sum256([]byte("test credential"))
	credential := &EventVisitorCredential{EventVisitorIdentityID: visitor.ID, CredentialHash: hash[:]}
	if err := repo.CreateEventVisitorCredential(ctx, credential); err != nil {
		t.Fatal(err)
	}
	if err := repo.RevokeEventVisitorCredentials(ctx, visitor.ID); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetEventVisitorCredential(ctx, visitor.ID, credential.ID)
	if err != nil || stored.RevokedAt == nil {
		t.Fatalf("revocation: %v", err)
	}
	for i := 0; i < 2; i++ {
		response := &Response{EventID: first.ID, EventVisitorIdentityID: visitor.ID, RespondentKind: RespondentKindGuest}
		if err := repo.CreateResponse(ctx, response); err != nil {
			t.Fatal(err)
		}
		fetched, err := repo.GetResponseByPublicID(ctx, first.ID, response.PublicID)
		if err != nil || fetched.ID != response.ID {
			t.Fatalf("opaque response lookup: %v", err)
		}
	}
	rows, err := repo.ListResponses(ctx, first.ID)
	if err != nil || len(rows) != 2 {
		t.Fatalf("multiple responses: %d %v", len(rows), err)
	}
	if hasResponse, err := repo.EventVisitorHasResponse(ctx, first.ID, visitor.ID); err != nil || !hasResponse {
		t.Fatalf("owned response not found: %v, %v", hasResponse, err)
	}
	if hasResponse, err := repo.EventVisitorHasResponse(ctx, second.ID, visitor.ID); err != nil || hasResponse {
		t.Fatalf("cross-event response found: %v, %v", hasResponse, err)
	}
	mismatch := &Response{EventID: second.ID, EventVisitorIdentityID: visitor.ID, RespondentKind: RespondentKindGuest}
	if err := repo.CreateResponse(ctx, mismatch); err == nil {
		t.Fatal("accepted visitor from another event")
	}
}

// TestOpaqueUUIDLookupsRejectNonCanonicalIdentifiers proves the repository
// validates opaque client identifiers before binding a uuid column, so a
// non-canonical event-visitor public ID, credential ID, or response public ID
// keeps reporting pgx.ErrNoRows instead of a PostgreSQL 22P02 cast error.
func TestOpaqueUUIDLookupsRejectNonCanonicalIdentifiers(t *testing.T) {
	ctx, repo, tx := newMigrationTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)
	var visitorPublicID string
	if err := tx.QueryRow(ctx, `SELECT public_id::text FROM event_visitor_identities WHERE id = $1`, visitorID).Scan(&visitorPublicID); err != nil {
		t.Fatal(err)
	}

	hash := sha256.Sum256([]byte("opaque lookup"))
	credential := &EventVisitorCredential{EventVisitorIdentityID: visitorID, CredentialHash: hash[:]}
	if err := repo.CreateEventVisitorCredential(ctx, credential); err != nil {
		t.Fatal(err)
	}
	response := &Response{EventID: eventID, EventVisitorIdentityID: visitorID, RespondentKind: RespondentKindGuest}
	if err := repo.CreateResponse(ctx, response); err != nil {
		t.Fatal(err)
	}

	for _, input := range []string{
		"",
		"not-a-uuid",
		"507f1f77bcf86cd799439011",
		"0198E6F0-6A3A-7C4B-9A2D-4F6A1B2C3D4E",
		"0198e6f0-6a3a-7c4b-9a2d-4f6a1b2c3d4e-",
	} {
		if _, err := repo.GetEventVisitorIdentity(ctx, eventID, input); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("event visitor public ID %q error = %v, want pgx.ErrNoRows", input, err)
		}
		if _, err := repo.GetEventVisitorCredential(ctx, visitorID, input); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("credential ID %q error = %v, want pgx.ErrNoRows", input, err)
		}
		if _, err := repo.GetResponseByPublicID(ctx, eventID, input); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("response public ID %q error = %v, want pgx.ErrNoRows", input, err)
		}
	}

	if _, err := repo.GetEventVisitorIdentity(ctx, eventID, visitorPublicID); err != nil {
		t.Fatalf("canonical visitor lookup after rejected identifiers: %v", err)
	}
	if _, err := repo.GetEventVisitorCredential(ctx, visitorID, credential.ID); err != nil {
		t.Fatalf("canonical credential lookup after rejected identifiers: %v", err)
	}
	if _, err := repo.GetResponseByPublicID(ctx, eventID, response.PublicID); err != nil {
		t.Fatalf("canonical response lookup after rejected identifiers: %v", err)
	}
}

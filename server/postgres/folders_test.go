package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// TestUpdateFolderStaticNoOpPreservesUpdatedAt proves the static update writes
// only supplied fields and that supplying neither is an existence check that
// does not bump updated_at.
func TestUpdateFolderStaticNoOpPreservesUpdatedAt(t *testing.T) {
	ctx, repo, _ := newMigrationTestRepository(t)
	identity, err := repo.CreatePlatformIdentity(ctx)
	if err != nil {
		t.Fatal(err)
	}
	folder := &Folder{PlatformIdentityID: identity.ID, Name: "Original"}
	if err := repo.CreateFolder(ctx, folder); err != nil {
		t.Fatal(err)
	}

	// Neither field supplied is a no-op that still resolves the folder.
	before := folder.UpdatedAt
	time.Sleep(2 * time.Millisecond)
	if err := repo.UpdateFolder(ctx, folder.ID, identity.ID, nil, nil); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetFolderByID(ctx, folder.ID, identity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Name != "Original" || stored.UpdatedAt.After(before) {
		t.Fatalf("no-op update changed the folder: %#v", stored)
	}

	// A supplied name updates only the name and advances updated_at.
	name := "Renamed"
	if err := repo.UpdateFolder(ctx, folder.ID, identity.ID, &name, nil); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetFolderByID(ctx, folder.ID, identity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Name != "Renamed" || !stored.UpdatedAt.After(before) {
		t.Fatalf("name update did not persist or bump updated_at: %#v", stored)
	}

	// A supplied color leaves the stored name untouched.
	color := "#123456"
	if err := repo.UpdateFolder(ctx, folder.ID, identity.ID, nil, &color); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.GetFolderByID(ctx, folder.ID, identity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Color == nil || *stored.Color != color || stored.Name != "Renamed" {
		t.Fatalf("color update clobbered fields: %#v", stored)
	}

	// A missing or foreign folder stays pgx.ErrNoRows for both branches.
	if err := repo.UpdateFolder(ctx, "086f4f9a-1b0e-4b7c-9a3e-6f4c2d1e5a01", identity.ID, nil, nil); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing folder no-op error = %v, want pgx.ErrNoRows", err)
	}
	if err := repo.UpdateFolder(ctx, "086f4f9a-1b0e-4b7c-9a3e-6f4c2d1e5a01", identity.ID, &name, nil); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing folder update error = %v, want pgx.ErrNoRows", err)
	}
}

package postgres

import (
	"context"
	"time"
)

// The former folders.go's folderSelect query, scanFolders row scanner, and the
// method bodies live in folders.gala. Folder, FolderMember, and the
// *Repository wrappers stay here because a method declared on Repository in a
// .gala file makes the GALA-declared set Repository's complete method set,
// which hides the eighteen methods repository_methods.go declares, and because
// a GALA struct declaration emits Copy, Equal, Unapply, and StructMeta helpers
// that import the GALA runtime.

// Folder is an account-owned container for member events. ID is a hidden
// UUIDv7; PlatformIdentityID is the owning account's platform identity uuid.
//
// The type stays handwritten because every GALA struct declaration emits Copy,
// Equal, Unapply, and StructMeta helpers that import the GALA runtime, which
// would change the Go-facing surface of this exported type.
type Folder struct {
	ID                 string
	PlatformIdentityID string
	Name               string
	Color              *string
	IsDeleted          *bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Members            []FolderMember
}

// FolderMember is one event reference stored by folder_events. EventID is the
// member's events.id. EventShortID is the member event's
// public short identifier and is empty when the referenced event row is
// missing.
//
// The type stays handwritten for the same GALA-surface reason as Folder.
type FolderMember struct {
	EventID      *string
	EventShortID *string
}

// CreateFolder inserts an account-scoped folder and returns its hidden identity
// and timestamps.
func (r *Repository) CreateFolder(ctx context.Context, folder *Folder) error {
	return createFolder(r, ctx, folder)
}

// ListFolders returns the account's non-deleted folders with their members. The
// account scope is enforced in the query, so another account's folders are
// never visible.
func (r *Repository) ListFolders(ctx context.Context, platformIdentityID string) ([]Folder, error) {
	return listFolders(r, ctx, platformIdentityID)
}

// GetFolderByID returns one account-scoped folder and its members. A missing or
// foreign folder is reported as pgx.ErrNoRows so the caller cannot distinguish
// non-existence from another account's folder.
func (r *Repository) GetFolderByID(ctx context.Context, folderID, platformIdentityID string) (*Folder, error) {
	return getFolderByID(r, ctx, folderID, platformIdentityID)
}

// UpdateFolder writes only the supplied fields for one account-scoped folder. A
// missing or foreign folder is reported as pgx.ErrNoRows. Supplying neither
// field is an existence check that leaves updated_at untouched.
func (r *Repository) UpdateFolder(ctx context.Context, folderID, platformIdentityID string, name, color *string) error {
	return updateFolder(r, ctx, folderID, platformIdentityID, name, color)
}

// DeleteFolder soft-deletes one account-scoped folder, soft-deletes the
// account's own PostgreSQL member events, and removes the memberships.
func (r *Repository) DeleteFolder(ctx context.Context, folderID, platformIdentityID string) error {
	return deleteFolder(r, ctx, folderID, platformIdentityID)
}

// AssignEventToFolder moves one event reference into an account-scoped folder,
// or removes it when folderID is nil. A given account holds at most one
// membership per event because the existing membership is deleted before the
// new one is inserted in the same transaction.
func (r *Repository) AssignEventToFolder(ctx context.Context, platformIdentityID string, folderID *string, member FolderMember) error {
	return assignEventToFolder(r, ctx, platformIdentityID, folderID, member)
}

package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Folder is an account-owned container for member events. ID is a hidden
// UUIDv7; PlatformIdentityID is the owning account's platform identity uuid.
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
type FolderMember struct {
	EventID      *string
	EventShortID *string
}

const folderSelect = `SELECT f.id, f.platform_identity_id, f.name, f.color, f.is_deleted, f.created_at, f.updated_at,
       fe.event_id, e.short_id
FROM folders f
LEFT JOIN folder_events fe ON fe.folder_id = f.id
LEFT JOIN events e ON e.id = fe.event_id`

// CreateFolder inserts an account-scoped folder and returns its hidden identity
// and timestamps.
func (r *Repository) CreateFolder(ctx context.Context, folder *Folder) error {
	if folder == nil || folder.PlatformIdentityID == "" {
		return errors.New("folder platform identity ID is required")
	}
	return r.db.QueryRow(ctx, `INSERT INTO folders (platform_identity_id, name, color)
VALUES ($1, $2, $3)
RETURNING id, is_deleted, created_at, updated_at`, folder.PlatformIdentityID, folder.Name, folder.Color).Scan(&folder.ID, &folder.IsDeleted, &folder.CreatedAt, &folder.UpdatedAt)
}

// ListFolders returns the account's non-deleted folders with their members. The
// account scope is enforced in the query, so another account's folders are
// never visible.
func (r *Repository) ListFolders(ctx context.Context, platformIdentityID string) ([]Folder, error) {
	if platformIdentityID == "" {
		return nil, errors.New("folder platform identity ID is required")
	}
	rows, err := r.db.Query(ctx, folderSelect+`
WHERE f.platform_identity_id = $1 AND f.is_deleted IS DISTINCT FROM TRUE
ORDER BY f.created_at, f.id, fe.created_at, fe.id`, platformIdentityID)
	if err != nil {
		return nil, err
	}
	return scanFolders(rows)
}

// GetFolderByID returns one account-scoped folder and its members. A missing or
// foreign folder is reported as pgx.ErrNoRows so the caller cannot distinguish
// non-existence from another account's folder.
func (r *Repository) GetFolderByID(ctx context.Context, folderID, platformIdentityID string) (*Folder, error) {
	if folderID == "" || platformIdentityID == "" {
		return nil, errors.New("folder ID and platform identity ID are required")
	}
	rows, err := r.db.Query(ctx, folderSelect+`
WHERE f.id = $1 AND f.platform_identity_id = $2 AND f.is_deleted IS DISTINCT FROM TRUE
ORDER BY fe.created_at, fe.id`, folderID, platformIdentityID)
	if err != nil {
		return nil, err
	}
	folders, err := scanFolders(rows)
	if err != nil {
		return nil, err
	}
	if len(folders) == 0 {
		return nil, pgx.ErrNoRows
	}
	return &folders[0], nil
}

// UpdateFolder writes only the supplied fields for one account-scoped folder. A
// missing or foreign folder is reported as pgx.ErrNoRows. Supplying neither
// field is an existence check that leaves updated_at untouched.
func (r *Repository) UpdateFolder(ctx context.Context, folderID, platformIdentityID string, name, color *string) error {
	if folderID == "" || platformIdentityID == "" {
		return errors.New("folder ID and platform identity ID are required")
	}
	var id string
	return r.db.QueryRow(ctx, `UPDATE folders
SET name = COALESCE($3, name),
    color = COALESCE($4, color),
    updated_at = CASE WHEN $3 IS NULL AND $4 IS NULL THEN updated_at ELSE clock_timestamp() END
WHERE id = $1 AND platform_identity_id = $2 AND is_deleted IS DISTINCT FROM TRUE
RETURNING id`, folderID, platformIdentityID, name, color).Scan(&id)
}

// DeleteFolder soft-deletes one account-scoped folder, soft-deletes the
// account's own PostgreSQL member events, and removes the memberships.
func (r *Repository) DeleteFolder(ctx context.Context, folderID, platformIdentityID string) error {
	if folderID == "" || platformIdentityID == "" {
		return errors.New("folder ID and platform identity ID are required")
	}
	return r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		var id string
		if err := tx.db.QueryRow(ctx, `UPDATE folders SET is_deleted = TRUE, updated_at = clock_timestamp()
WHERE id = $1 AND platform_identity_id = $2 AND is_deleted IS DISTINCT FROM TRUE
RETURNING id`, folderID, platformIdentityID).Scan(&id); err != nil {
			return err
		}
		if _, err := tx.db.Exec(ctx, `UPDATE events SET is_deleted = TRUE, updated_at = clock_timestamp()
WHERE id IN (
    SELECT fe.event_id FROM folder_events fe
    WHERE fe.folder_id = $1 AND fe.platform_identity_id = $2
)
AND owner_platform_identity_id = $2`, folderID, platformIdentityID); err != nil {
			return err
		}
		_, err := tx.db.Exec(ctx, `DELETE FROM folder_events WHERE folder_id = $1 AND platform_identity_id = $2`, folderID, platformIdentityID)
		return err
	})
}

// AssignEventToFolder moves one event reference into an account-scoped folder,
// or removes it when folderID is nil. A given account holds at most one
// membership per event because the existing membership is deleted before the
// new one is inserted in the same transaction.
func (r *Repository) AssignEventToFolder(ctx context.Context, platformIdentityID string, folderID *string, member FolderMember) error {
	if platformIdentityID == "" {
		return errors.New("folder platform identity ID is required")
	}
	if member.EventID == nil {
		return errors.New("folder member event reference is required")
	}
	return r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		if _, err := tx.db.Exec(ctx, `DELETE FROM folder_events WHERE platform_identity_id = $1 AND event_id = $2`, platformIdentityID, *member.EventID); err != nil {
			return err
		}
		if folderID == nil {
			return nil
		}
		var exists bool
		if err := tx.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM folders
WHERE id = $1 AND platform_identity_id = $2 AND is_deleted IS DISTINCT FROM TRUE)`, *folderID, platformIdentityID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return pgx.ErrNoRows
		}
		_, err := tx.db.Exec(ctx, `INSERT INTO folder_events (platform_identity_id, folder_id, event_id)
VALUES ($1, $2, $3)`, platformIdentityID, *folderID, *member.EventID)
		return err
	})
}

func scanFolders(rows pgx.Rows) ([]Folder, error) {
	defer rows.Close()
	folders := []Folder{}
	indexByID := map[string]int{}
	for rows.Next() {
		var (
			folder           Folder
			eventID, shortID *string
		)
		if err := rows.Scan(&folder.ID, &folder.PlatformIdentityID, &folder.Name, &folder.Color, &folder.IsDeleted, &folder.CreatedAt, &folder.UpdatedAt, &eventID, &shortID); err != nil {
			return nil, err
		}
		index, ok := indexByID[folder.ID]
		if !ok {
			folder.Members = []FolderMember{}
			folders = append(folders, folder)
			index = len(folders) - 1
			indexByID[folder.ID] = index
		}
		if eventID != nil {
			folders[index].Members = append(folders[index].Members, FolderMember{EventID: eventID, EventShortID: shortID})
		}
	}
	return folders, rows.Err()
}

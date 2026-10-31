package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handwritten sibling beside the generated repository.go.
//
// The method bodies live in repository.gala as free functions that take the
// repository first; the *Repository methods here are one-line wrappers, because
// a method declared on Repository in a .gala file makes the GALA-declared set
// Repository's complete method set, which hides the methods the other
// handwritten files of this package declare. WithTransaction and withTransaction
// keep their bodies here: their defer tx.Rollback has no spelling that also lets
// a callback panic propagate as a panic rather than as a returned error.

func (r *Repository) WithTransaction(ctx context.Context, fn func(context.Context, *Repository) error) error {
	pool, ok := r.db.(*pgxpool.Pool)
	if !ok {
		return errors.New("repository is already transaction-scoped")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // A successful commit makes this a no-op.
	if err := fn(ctx, &Repository{db: tx}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// withTransaction runs fn in a single transaction. A repository that is already
// transaction-scoped runs fn in its existing transaction so a lock taken by fn
// spans the work instead of being rejected or silently split across statements.
func (r *Repository) withTransaction(ctx context.Context, fn func(context.Context, *Repository) error) error {
	if _, ok := r.db.(*pgxpool.Pool); !ok {
		return fn(ctx, r)
	}
	return r.WithTransaction(ctx, fn)
}

func (r *Repository) CreateEvent(ctx context.Context, event *Event) error {
	return createEvent(r, ctx, event)
}

func (r *Repository) GetEventByShortID(ctx context.Context, shortID string) (*Event, error) {
	return getEventByShortID(r, ctx, shortID)
}

func (r *Repository) GetEventByID(ctx context.Context, id string) (*Event, error) {
	return getEventByID(r, ctx, id)
}

func (r *Repository) getEvent(ctx context.Context, column, value string) (*Event, error) {
	return getEvent(r, ctx, column, value)
}

func (r *Repository) UpdateEvent(ctx context.Context, event *Event) error {
	return updateEvent(r, ctx, event)
}

// AdjustEventResponseCount applies a relative change to an event's response
// count without rewriting the payload. The result is clamped at zero so a stale
// count cannot violate the non-negative check when a response row is removed.
// Callers hold the event row lock so the adjustment cannot race another
// response mutation.
func (r *Repository) AdjustEventResponseCount(ctx context.Context, eventID string, delta int) error {
	return adjustEventResponseCount(r, ctx, eventID, delta)
}

// SetEventArchived writes only the archive flag, leaving the event payload
// untouched. A missing event is reported as pgx.ErrNoRows.
func (r *Repository) SetEventArchived(ctx context.Context, eventID string, archived bool) error {
	return setEventArchived(r, ctx, eventID, archived)
}

// SetEventDeleted writes only the soft-delete flag, leaving the event payload
// untouched. A missing event is reported as pgx.ErrNoRows.
func (r *Repository) SetEventDeleted(ctx context.Context, eventID string, deleted bool) error {
	return setEventDeleted(r, ctx, eventID, deleted)
}

func (r *Repository) CreateResponse(ctx context.Context, response *Response) error {
	return createResponse(r, ctx, response)
}

func (r *Repository) GetResponseByID(ctx context.Context, id string) (*Response, error) {
	return getResponseByID(r, ctx, id)
}

func (r *Repository) GetResponseByPlatformIdentityID(ctx context.Context, eventID, platformIdentityID string) (*Response, error) {
	return getResponseByPlatformIdentityID(r, ctx, eventID, platformIdentityID)
}

func (r *Repository) getResponse(ctx context.Context, predicate string, values ...any) (*Response, error) {
	return getResponse(r, ctx, predicate, values...)
}

func (r *Repository) ListResponses(ctx context.Context, eventID string) ([]Response, error) {
	return listResponses(r, ctx, eventID)
}

func (r *Repository) UpdateResponse(ctx context.Context, response *Response) error {
	return updateResponse(r, ctx, response)
}

// DeleteResponse removes one response by its hidden primary key.
func (r *Repository) DeleteResponse(ctx context.Context, id string) error {
	return deleteResponse(r, ctx, id)
}

// DeleteAccountResponses removes every account response owned by the given
// platform identities on one event in one statement and reports the deleted row
// count so the caller can adjust the event response count exactly. Unlike the
// former per-identity lookup and delete, this removes every matching response,
// including more than one response for the same platform identity.
func (r *Repository) DeleteAccountResponses(ctx context.Context, eventID string, platformIdentityIDs []string) (int64, error) {
	return deleteAccountResponses(r, ctx, eventID, platformIdentityIDs)
}

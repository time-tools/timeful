package postgres

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	crockfordBase32 = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
)

var ErrPoolUninitialized = errors.New("postgresql pool is not initialized")

type dbtx interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// Repository persists PostgreSQL-owned compatibility events and responses.
// A transaction callback receives a repository bound to the same transaction.
type Repository struct {
	db dbtx
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{db: pool}
}

// NewRepositoryFromTx binds a repository to an existing transaction so a caller
// that owns the transaction can compose repository operations with its own SQL
// in one atomic unit of work.
func NewRepositoryFromTx(tx pgx.Tx) *Repository {
	return &Repository{db: tx}
}

// DefaultRepository uses the package-global pool initialized by Init.
func DefaultRepository() (*Repository, error) {
	if Pool == nil {
		return nil, ErrPoolUninitialized
	}
	return NewRepository(Pool), nil
}

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

// WithTransaction runs fn against the package-global pool in one transaction.
func WithTransaction(ctx context.Context, fn func(context.Context, *Repository) error) error {
	repository, err := DefaultRepository()
	if err != nil {
		return err
	}
	return repository.WithTransaction(ctx, fn)
}

func GenerateShortID() (string, error) {
	var value [5]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate short event ID: %w", err)
	}
	return encodeCrockford(value[:], 8), nil
}

func GenerateEventShortID() (string, error) { return GenerateShortID() }

// GenerateTransferCode returns a six-digit decimal matching code. Rejection
// sampling keeps the value uniform without modulo bias, and formatting keeps
// leading zeros so every code is exactly six digits.
func GenerateTransferCode() (string, error) {
	var raw [4]byte
	for {
		if _, err := rand.Read(raw[:]); err != nil {
			return "", fmt.Errorf("generate transfer code: %w", err)
		}
		value := uint32(raw[0])<<24 | uint32(raw[1])<<16 | uint32(raw[2])<<8 | uint32(raw[3])
		if value >= 4_000_000_000 {
			continue
		}
		return formatTransferCode(value % 1_000_000), nil
	}
}

func formatTransferCode(value uint32) string {
	return fmt.Sprintf("%06d", value)
}

func encodeCrockford(value []byte, length int) string {
	result := make([]byte, length)
	var buffer uint64
	bits := uint(0)
	index := 0
	for _, b := range value {
		buffer = buffer<<8 | uint64(b)
		bits += 8
		for bits >= 5 && index < length {
			bits -= 5
			result[index] = crockfordBase32[(buffer>>bits)&31]
			index++
		}
	}
	if index < length && bits > 0 {
		result[index] = crockfordBase32[(buffer<<(5-bits))&31]
	}
	return string(result)
}

func (r *Repository) CreateEvent(ctx context.Context, event *Event) error {
	if event == nil {
		return errors.New("event is nil")
	}
	payload, err := encodePayload(event.Payload)
	if err != nil {
		return err
	}
	if event.ScheduleVersion == 0 {
		event.ScheduleVersion = 1
	}
	for attempt := 0; attempt < 5; attempt++ {
		generatedShortID := event.ShortID == ""
		if event.ShortID == "" {
			event.ShortID, err = GenerateShortID()
			if err != nil {
				return err
			}
		}
		err = r.db.QueryRow(ctx, `INSERT INTO events (short_id, name, type, is_archived, is_deleted, num_responses, schedule_version, creator_posthog_id, payload)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, created_at, updated_at`, event.ShortID, event.Name, event.Type, event.IsArchived, event.IsDeleted, event.NumResponses, event.ScheduleVersion, event.CreatorPosthogID, payload).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
		if err == nil || !isUniqueViolation(err) || !generatedShortID {
			event.Payload = decodePayload(payload)
			return err
		}
		event.ShortID = ""
	}
	return errors.New("generate unique event identifiers")
}

func (r *Repository) GetEventByShortID(ctx context.Context, shortID string) (*Event, error) {
	return r.getEvent(ctx, "short_id", shortID)
}

func (r *Repository) GetEventByID(ctx context.Context, id string) (*Event, error) {
	return r.getEvent(ctx, "id", id)
}

const eventColumns = `id, short_id, owner_edit_token_hash, owner_platform_identity_id, owner_event_visitor_identity_id, name, type, is_archived, is_deleted, num_responses, schedule_version, creator_posthog_id, created_at, updated_at, payload`

func scanEvent(row interface{ Scan(...any) error }) (*Event, error) {
	event := &Event{}
	err := row.Scan(&event.ID, &event.ShortID, &event.OwnerEditTokenHash, &event.OwnerPlatformIdentityID, &event.OwnerEventVisitorIdentityID, &event.Name, &event.Type, &event.IsArchived, &event.IsDeleted, &event.NumResponses, &event.ScheduleVersion, &event.CreatorPosthogID, &event.CreatedAt, &event.UpdatedAt, &event.Payload)
	if err != nil {
		return nil, err
	}
	event.Payload = decodePayload(event.Payload)
	return event, nil
}

func (r *Repository) getEvent(ctx context.Context, column, value string) (*Event, error) {
	return scanEvent(r.db.QueryRow(ctx, `SELECT `+eventColumns+` FROM events WHERE `+column+` = $1`, value))
}

func (r *Repository) UpdateEvent(ctx context.Context, event *Event) error {
	if event == nil || event.ID == "" {
		return errors.New("event ID is required")
	}
	payload, err := encodePayload(event.Payload)
	if err != nil {
		return err
	}
	err = r.db.QueryRow(ctx, `UPDATE events SET name = $2, type = $3, is_archived = $4, is_deleted = $5, num_responses = $6, schedule_version = $7, creator_posthog_id = $8, payload = $9, owner_event_visitor_identity_id = $10, updated_at = clock_timestamp() WHERE id = $1 RETURNING updated_at`, event.ID, event.Name, event.Type, event.IsArchived, event.IsDeleted, event.NumResponses, event.ScheduleVersion, event.CreatorPosthogID, payload, event.OwnerEventVisitorIdentityID).Scan(&event.UpdatedAt)
	if err != nil {
		return err
	}
	event.Payload = decodePayload(payload)
	return nil
}

// AdjustEventResponseCount applies a relative change to an event's response
// count without rewriting the payload. The result is clamped at zero so a stale
// count cannot violate the non-negative check when a response row is removed.
// Callers hold the event row lock so the adjustment cannot race another
// response mutation.
func (r *Repository) AdjustEventResponseCount(ctx context.Context, eventID string, delta int) error {
	if eventID == "" {
		return errors.New("event ID is required")
	}
	tag, err := r.db.Exec(ctx, `UPDATE events
SET num_responses = GREATEST(num_responses + $2, 0), updated_at = clock_timestamp()
WHERE id = $1`, eventID, delta)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// SetEventArchived writes only the archive flag, leaving the event payload
// untouched. A missing event is reported as pgx.ErrNoRows.
func (r *Repository) SetEventArchived(ctx context.Context, eventID string, archived bool) error {
	if eventID == "" {
		return errors.New("event ID is required")
	}
	tag, err := r.db.Exec(ctx, `UPDATE events
SET is_archived = $2, updated_at = clock_timestamp()
WHERE id = $1`, eventID, archived)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// SetEventDeleted writes only the soft-delete flag, leaving the event payload
// untouched. A missing event is reported as pgx.ErrNoRows.
func (r *Repository) SetEventDeleted(ctx context.Context, eventID string, deleted bool) error {
	if eventID == "" {
		return errors.New("event ID is required")
	}
	tag, err := r.db.Exec(ctx, `UPDATE events
SET is_deleted = $2, updated_at = clock_timestamp()
WHERE id = $1`, eventID, deleted)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) CreateResponse(ctx context.Context, response *Response) error {
	if response == nil || response.EventID == "" {
		return errors.New("response event ID is required")
	}
	payload, err := encodePayload(response.Payload)
	if err != nil {
		return err
	}
	err = r.db.QueryRow(ctx, `INSERT INTO event_responses (event_id, event_visitor_identity_id, respondent_kind, platform_identity_id, payload)
VALUES ($1, $2, $3, $4, $5) RETURNING id, public_id, created_at, updated_at`, response.EventID, response.EventVisitorIdentityID, response.RespondentKind, response.PlatformIdentityID, payload).Scan(&response.ID, &response.PublicID, &response.CreatedAt, &response.UpdatedAt)
	if err == nil {
		response.Payload = decodePayload(payload)
	}
	return err
}

func (r *Repository) GetResponseByID(ctx context.Context, id string) (*Response, error) {
	return r.getResponse(ctx, `id = $1`, id)
}

func (r *Repository) GetResponseByPlatformIdentityID(ctx context.Context, eventID, platformIdentityID string) (*Response, error) {
	return r.getResponse(ctx, `event_id = $1 AND respondent_kind = 'account' AND platform_identity_id = $2`, eventID, platformIdentityID)
}

func (r *Repository) getResponse(ctx context.Context, predicate string, values ...any) (*Response, error) {
	response := &Response{}
	err := r.db.QueryRow(ctx, `SELECT id, public_id, event_visitor_identity_id, event_id, respondent_kind, platform_identity_id, payload, created_at, updated_at FROM event_responses WHERE `+predicate, values...).Scan(&response.ID, &response.PublicID, &response.EventVisitorIdentityID, &response.EventID, &response.RespondentKind, &response.PlatformIdentityID, &response.Payload, &response.CreatedAt, &response.UpdatedAt)
	if err != nil {
		return nil, err
	}
	response.Payload = decodePayload(response.Payload)
	return response, nil
}

// listResponsesQuery lists one event's responses in write order. The
// supporting-index forced-plan test runs this statement directly.
const listResponsesQuery = `SELECT id, public_id, event_visitor_identity_id, event_id, respondent_kind, platform_identity_id, payload, created_at, updated_at FROM event_responses WHERE event_id = $1 ORDER BY created_at, id`

func (r *Repository) ListResponses(ctx context.Context, eventID string) ([]Response, error) {
	rows, err := r.db.Query(ctx, listResponsesQuery, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	responses := []Response{}
	for rows.Next() {
		var response Response
		if err := rows.Scan(&response.ID, &response.PublicID, &response.EventVisitorIdentityID, &response.EventID, &response.RespondentKind, &response.PlatformIdentityID, &response.Payload, &response.CreatedAt, &response.UpdatedAt); err != nil {
			return nil, err
		}
		response.Payload = decodePayload(response.Payload)
		responses = append(responses, response)
	}
	return responses, rows.Err()
}

func (r *Repository) UpdateResponse(ctx context.Context, response *Response) error {
	if response == nil || response.ID == "" {
		return errors.New("response ID is required")
	}
	payload, err := encodePayload(response.Payload)
	if err != nil {
		return err
	}
	err = r.db.QueryRow(ctx, `UPDATE event_responses SET respondent_kind = $2, platform_identity_id = $3, payload = $4, updated_at = clock_timestamp() WHERE id = $1 RETURNING updated_at`, response.ID, response.RespondentKind, response.PlatformIdentityID, payload).Scan(&response.UpdatedAt)
	if err != nil {
		return err
	}
	response.Payload = decodePayload(payload)
	return nil
}

// DeleteResponse removes one response by its hidden primary key.
func (r *Repository) DeleteResponse(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM event_responses WHERE id = $1`, id)
	return err
}

// DeleteAccountResponses removes every account response owned by the given
// platform identities on one event in one statement and reports the deleted row
// count so the caller can adjust the event response count exactly. Unlike the
// former per-identity lookup and delete, this removes every matching response,
// including more than one response for the same platform identity.
func (r *Repository) DeleteAccountResponses(ctx context.Context, eventID string, platformIdentityIDs []string) (int64, error) {
	if eventID == "" {
		return 0, errors.New("response event ID is required")
	}
	if len(platformIdentityIDs) == 0 {
		return 0, nil
	}
	tag, err := r.db.Exec(ctx, `DELETE FROM event_responses
WHERE event_id = $1 AND respondent_kind = 'account' AND platform_identity_id = ANY($2::uuid[])`, eventID, platformIdentityIDs)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}

// IsUniqueViolation reports PostgreSQL unique-index conflicts to route adapters.
func IsUniqueViolation(err error) bool { return isUniqueViolation(err) }

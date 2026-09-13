package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"timeful/server/respondents"
)

// ErrSignupCapacityExceeded reports that at least one requested block is already
// at capacity. The mutation is rejected atomically, so no partial response or
// block membership is written.
var ErrSignupCapacityExceeded = errors.New("signup block capacity exceeded")

// ErrSignupBlockNotFound reports that a requested block does not belong to the
// signup form event.
var ErrSignupBlockNotFound = errors.New("signup block does not belong to the event")

// SignupBlock is one ordered, capacity-limited slot on a signup form event. ID
// is a fresh PostgreSQL UUIDv7 block identifier. Capacity is nil when the block
// is unlimited. Position preserves the block array order.
type SignupBlock struct {
	ID        string
	EventID   string
	Name      string
	Capacity  *int
	StartDate *time.Time
	EndDate   *time.Time
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

const signupBlockColumns = `id, event_id, name, capacity, start_date, end_date, position, created_at, updated_at`

func scanSignupBlock(row interface{ Scan(...any) error }) (*SignupBlock, error) {
	block := &SignupBlock{}
	err := row.Scan(&block.ID, &block.EventID, &block.Name, &block.Capacity, &block.StartDate, &block.EndDate, &block.Position, &block.CreatedAt, &block.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// SignupResponse is one respondent's signup on a signup form event. PublicID is
// the opaque identifier exposed to clients; ID and the owning Event Visitor
// Identity stay internal. PlatformIdentityID or CanonicalGuestName identifies
// the respondent, matching the account-uuid-or-canonical-guest-name key.
// BlockIDs holds the claimed event_signup_blocks identities in their stored
// claim order.
type SignupResponse struct {
	ID                     string
	PublicID               string
	EventID                string
	EventVisitorIdentityID string
	RespondentKind         string
	PlatformIdentityID     *string
	CanonicalGuestName     *string
	Name                   string
	Email                  string
	BlockIDs               []string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// signupResponseColumns lists the response columns without membership. Full
// response reads append signupResponseBlockIDsColumn and end with BlockIDs.
const signupResponseColumns = `id, public_id, event_id, event_visitor_identity_id, respondent_kind, platform_identity_id, canonical_guest_name, name, email, created_at, updated_at`

// signupResponseBlockIDsColumn aggregates a response's claimed block identities
// in their stored position order so reads reproduce the write order. Position
// alone is not unique, so the block identity is the deterministic tie-break.
const signupResponseBlockIDsColumn = `COALESCE(ARRAY(
    SELECT membership.block_id::text
    FROM event_signup_response_blocks membership
    WHERE membership.response_id = event_signup_responses.id
    ORDER BY membership.position, membership.block_id
), '{}'::text[])`

const signupResponseSelectColumns = signupResponseColumns + `, ` + signupResponseBlockIDsColumn

func scanSignupResponse(row interface{ Scan(...any) error }) (*SignupResponse, error) {
	response := &SignupResponse{}
	err := row.Scan(&response.ID, &response.PublicID, &response.EventID, &response.EventVisitorIdentityID, &response.RespondentKind, &response.PlatformIdentityID, &response.CanonicalGuestName, &response.Name, &response.Email, &response.CreatedAt, &response.UpdatedAt, &response.BlockIDs)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// readSignupResponseByID reads one response with its claimed blocks aggregated
// in position order.
func (r *Repository) readSignupResponseByID(ctx context.Context, responseID string) (*SignupResponse, error) {
	return scanSignupResponse(r.db.QueryRow(ctx, `SELECT `+signupResponseSelectColumns+`
FROM event_signup_responses WHERE id = $1`, responseID))
}

// writeSignupResponseMemberships replaces a response's claimed blocks with the
// supplied normalized identities in their given order. The delete and the
// set-based insert both target the join table; the capacity check in the same
// transaction has already proven every identity names a block of the event, and
// the unique constraint backstops a repeated claim. The affected-row check keeps
// a caller that skips the capacity check from silently dropping a claim.
func (r *Repository) writeSignupResponseMemberships(ctx context.Context, responseID, eventID string, blockIDs []string) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM event_signup_response_blocks WHERE response_id = $1`, responseID); err != nil {
		return err
	}
	if len(blockIDs) == 0 {
		return nil
	}
	tag, err := r.db.Exec(ctx, `INSERT INTO event_signup_response_blocks (response_id, block_id, position)
SELECT $1, block.id, claimed.ordinality
FROM unnest($2::text[]) WITH ORDINALITY AS claimed(block_text, ordinality)
JOIN event_signup_blocks block
    ON block.id::text = claimed.block_text AND block.event_id = $3`, responseID, blockIDs, eventID)
	if err != nil {
		return err
	}
	if int(tag.RowsAffected()) != len(blockIDs) {
		return ErrSignupBlockNotFound
	}
	return nil
}

// CreateSignupBlock appends one block to a signup form event. A non-positive
// position appends after the existing blocks. The event row is locked so a
// concurrent replace or signup cannot interleave block ordering.
func (r *Repository) CreateSignupBlock(ctx context.Context, block *SignupBlock) error {
	if block == nil || block.EventID == "" {
		return errors.New("signup block event ID is required")
	}
	if block.Name == "" {
		return errors.New("signup block name is required")
	}
	if block.Capacity != nil && *block.Capacity < 0 {
		return errors.New("signup block capacity must not be negative")
	}
	return r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		if _, err := tx.LockEvent(ctx, block.EventID); err != nil {
			return err
		}
		return tx.db.QueryRow(ctx, `INSERT INTO event_signup_blocks (event_id, name, capacity, start_date, end_date, position)
VALUES ($1, $2, $3, $4, $5, CASE WHEN $6 > 0 THEN $6 ELSE (SELECT COALESCE(MAX(position), 0) + 1 FROM event_signup_blocks WHERE event_id = $1) END)
RETURNING `+signupBlockColumns, block.EventID, block.Name, block.Capacity, block.StartDate, block.EndDate, block.Position).
			Scan(&block.ID, &block.EventID, &block.Name, &block.Capacity, &block.StartDate, &block.EndDate, &block.Position, &block.CreatedAt, &block.UpdatedAt)
	})
}

// ReplaceSignupBlocks writes the supplied blocks as the complete ordered block
// set for an event. A block with an ID must belong to the event and keeps its
// identity and response memberships; a block without an ID is inserted.
// Removed blocks are deleted and their memberships cascade away in the same
// transaction, so the block/response relation never dangles. A removal still
// bumps the affected responses' updated_at, which the retired detach path did
// and a cascade alone does not. The event row is locked for the duration.
// Repeated non-empty identities collapse to their last occurrence because a
// set-based ON CONFLICT DO UPDATE cannot touch one conflict row twice; that
// matches the old per-block loop, which applied the updates in order.
func (r *Repository) ReplaceSignupBlocks(ctx context.Context, eventID string, blocks []SignupBlock) ([]SignupBlock, error) {
	if eventID == "" {
		return nil, errors.New("signup block event ID is required")
	}
	lastIndex := make(map[string]int, len(blocks))
	for i := range blocks {
		if blocks[i].Capacity != nil && *blocks[i].Capacity < 0 {
			return nil, errors.New("signup block capacity must not be negative")
		}
		if blocks[i].ID != "" {
			lastIndex[blocks[i].ID] = i
		}
	}
	positioned := make([]SignupBlock, 0, len(blocks))
	for i := range blocks {
		if blocks[i].ID != "" && lastIndex[blocks[i].ID] != i {
			continue
		}
		positioned = append(positioned, blocks[i])
	}
	ids := make([]string, 0, len(positioned))
	names := make([]string, 0, len(positioned))
	capacities := make([]*int, 0, len(positioned))
	startDates := make([]*time.Time, 0, len(positioned))
	endDates := make([]*time.Time, 0, len(positioned))
	positions := make([]int, 0, len(positioned))
	keepIDs := make([]string, 0, len(positioned))
	for i := range positioned {
		block := &positioned[i]
		ids = append(ids, block.ID)
		names = append(names, block.Name)
		capacities = append(capacities, block.Capacity)
		startDates = append(startDates, block.StartDate)
		endDates = append(endDates, block.EndDate)
		positions = append(positions, i+1)
		if block.ID != "" {
			keepIDs = append(keepIDs, block.ID)
		}
	}
	stored := []SignupBlock{}
	err := r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		if _, err := tx.LockEvent(ctx, eventID); err != nil {
			return err
		}
		if _, err := tx.db.Exec(ctx, `UPDATE event_signup_responses response
SET updated_at = clock_timestamp()
WHERE response.event_id = $1
  AND EXISTS (
      SELECT 1
      FROM event_signup_response_blocks membership
      JOIN event_signup_blocks block ON block.id = membership.block_id
      WHERE membership.response_id = response.id
        AND block.event_id = $1
        AND block.id::text <> ALL($2::text[])
  )`, eventID, keepIDs); err != nil {
			return err
		}
		if _, err := tx.db.Exec(ctx, `DELETE FROM event_signup_blocks
WHERE event_id = $1 AND id::text <> ALL($2::text[])`, eventID, keepIDs); err != nil {
			return err
		}
		tag, err := tx.db.Exec(ctx, `INSERT INTO event_signup_blocks (id, event_id, name, capacity, start_date, end_date, position)
SELECT COALESCE(existing.id, uuidv7()), $1, incoming.name, incoming.capacity, incoming.start_date, incoming.end_date, incoming.position
FROM unnest($2::text[], $3::text[], $4::int[], $5::timestamptz[], $6::timestamptz[], $7::int[]) AS incoming(id_text, name, capacity, start_date, end_date, position)
LEFT JOIN event_signup_blocks existing
    ON existing.id::text = incoming.id_text AND existing.event_id = $1
WHERE incoming.id_text = '' OR existing.id IS NOT NULL
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name, capacity = EXCLUDED.capacity, start_date = EXCLUDED.start_date, end_date = EXCLUDED.end_date, position = EXCLUDED.position, updated_at = clock_timestamp()`,
			eventID, ids, names, capacities, startDates, endDates, positions)
		if err != nil {
			return err
		}
		if int(tag.RowsAffected()) != len(positioned) {
			return pgx.ErrNoRows
		}
		list, err := tx.ListSignupBlocks(ctx, eventID)
		if err != nil {
			return err
		}
		stored = list
		return nil
	})
	if err != nil {
		return nil, err
	}
	return stored, nil
}

// ListSignupBlocks returns the event's blocks in their stored order.
func (r *Repository) ListSignupBlocks(ctx context.Context, eventID string) ([]SignupBlock, error) {
	if eventID == "" {
		return nil, errors.New("signup block event ID is required")
	}
	rows, err := r.db.Query(ctx, `SELECT `+signupBlockColumns+`
FROM event_signup_blocks WHERE event_id = $1 ORDER BY position, created_at, id`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks := []SignupBlock{}
	for rows.Next() {
		block, err := scanSignupBlock(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, *block)
	}
	return blocks, rows.Err()
}

// CreateSignupResponse inserts one response for an Event Visitor Identity. The
// respondent identity is inferred when RespondentKind is empty. Guest names are
// canonicalized through the shared respondents normalizer; a duplicate canonical
// guest name or account response surfaces as a unique violation. Capacity is
// reserved atomically under the event row lock, so an over-capacity signup is
// rejected before any row is written.
func (r *Repository) CreateSignupResponse(ctx context.Context, response *SignupResponse) error {
	if response == nil || response.EventID == "" {
		return errors.New("signup response event ID is required")
	}
	if response.EventVisitorIdentityID == "" {
		return errors.New("signup response event visitor identity is required")
	}
	if err := normalizeSignupResponseIdentity(response); err != nil {
		return err
	}
	blockIDs := normalizeSignupBlockIDs(response.BlockIDs)
	return r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		if _, err := tx.LockEvent(ctx, response.EventID); err != nil {
			return err
		}
		if err := tx.reserveSignupCapacity(ctx, response.EventID, blockIDs, ""); err != nil {
			return err
		}
		if err := tx.db.QueryRow(ctx, `INSERT INTO event_signup_responses
 (event_id, event_visitor_identity_id, respondent_kind, platform_identity_id, canonical_guest_name, name, email)
 VALUES ($1, $2, $3, $4, $5, $6, $7)
 RETURNING `+signupResponseColumns,
			response.EventID, response.EventVisitorIdentityID, response.RespondentKind, response.PlatformIdentityID, response.CanonicalGuestName, response.Name, response.Email).
			Scan(&response.ID, &response.PublicID, &response.EventID, &response.EventVisitorIdentityID, &response.RespondentKind, &response.PlatformIdentityID, &response.CanonicalGuestName, &response.Name, &response.Email, &response.CreatedAt, &response.UpdatedAt); err != nil {
			return err
		}
		if err := tx.writeSignupResponseMemberships(ctx, response.ID, response.EventID, blockIDs); err != nil {
			return err
		}
		stored, err := tx.readSignupResponseByID(ctx, response.ID)
		if err != nil {
			return err
		}
		*response = *stored
		return nil
	})
}

// GetSignupResponseByPublicID resolves one response by its opaque public
// identifier within an event. A public ID that is not a canonical UUID resolves
// to no response instead of being forwarded to the uuid column.
func (r *Repository) GetSignupResponseByPublicID(ctx context.Context, eventID, publicID string) (*SignupResponse, error) {
	if eventID == "" || publicID == "" {
		return nil, errors.New("signup response event ID and public ID are required")
	}
	return scanSignupResponse(r.db.QueryRow(ctx, `SELECT `+signupResponseSelectColumns+`
FROM event_signup_responses WHERE event_id = $1 AND public_id::text = $2`, eventID, publicID))
}

// ListSignupResponses returns every signup response for an event in write order.
func (r *Repository) ListSignupResponses(ctx context.Context, eventID string) ([]SignupResponse, error) {
	if eventID == "" {
		return nil, errors.New("signup response event ID is required")
	}
	rows, err := r.db.Query(ctx, `SELECT `+signupResponseSelectColumns+`
FROM event_signup_responses WHERE event_id = $1 ORDER BY created_at, id`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	responses := []SignupResponse{}
	for rows.Next() {
		response, err := scanSignupResponse(rows)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *response)
	}
	return responses, rows.Err()
}

// UpdateSignupResponse writes the respondent identity, name, email, and block
// membership for one response selected by internal ID or opaque public ID. The
// existing row is excluded from the capacity count, so re-saving a response
// never consumes an extra seat. The event row is locked, keeping capacity
// reservation atomic.
func (r *Repository) UpdateSignupResponse(ctx context.Context, response *SignupResponse) error {
	if response == nil || response.EventID == "" {
		return errors.New("signup response event ID is required")
	}
	if response.ID == "" && response.PublicID == "" {
		return errors.New("signup response ID is required")
	}
	if err := normalizeSignupResponseIdentity(response); err != nil {
		return err
	}
	blockIDs := normalizeSignupBlockIDs(response.BlockIDs)
	return r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		if _, err := tx.LockEvent(ctx, response.EventID); err != nil {
			return err
		}
		var responseID string
		var err error
		if response.ID != "" {
			err = tx.db.QueryRow(ctx, `SELECT id FROM event_signup_responses
WHERE id = $1 AND event_id = $2`, response.ID, response.EventID).Scan(&responseID)
		} else {
			err = tx.db.QueryRow(ctx, `SELECT id FROM event_signup_responses
WHERE public_id = $1 AND event_id = $2`, response.PublicID, response.EventID).Scan(&responseID)
		}
		if err != nil {
			return err
		}
		if err := tx.reserveSignupCapacity(ctx, response.EventID, blockIDs, responseID); err != nil {
			return err
		}
		if err := tx.db.QueryRow(ctx, `UPDATE event_signup_responses
SET respondent_kind = $2, platform_identity_id = $3, canonical_guest_name = $4, name = $5, email = $6, updated_at = clock_timestamp()
WHERE id = $1
RETURNING `+signupResponseColumns,
			responseID, response.RespondentKind, response.PlatformIdentityID, response.CanonicalGuestName, response.Name, response.Email).
			Scan(&response.ID, &response.PublicID, &response.EventID, &response.EventVisitorIdentityID, &response.RespondentKind, &response.PlatformIdentityID, &response.CanonicalGuestName, &response.Name, &response.Email, &response.CreatedAt, &response.UpdatedAt); err != nil {
			return err
		}
		if err := tx.writeSignupResponseMemberships(ctx, response.ID, response.EventID, blockIDs); err != nil {
			return err
		}
		stored, err := tx.readSignupResponseByID(ctx, response.ID)
		if err != nil {
			return err
		}
		*response = *stored
		return nil
	})
}

// DeleteSignupResponse removes one response selected by opaque public ID. The
// event row is locked so a deletion cannot interleave with a capacity
// reservation. A missing or non-canonical response is reported as pgx.ErrNoRows.
func (r *Repository) DeleteSignupResponse(ctx context.Context, eventID, publicID string) error {
	if eventID == "" || publicID == "" {
		return errors.New("signup response event ID and public ID are required")
	}
	return r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		if _, err := tx.LockEvent(ctx, eventID); err != nil {
			return err
		}
		tag, err := tx.db.Exec(ctx, `DELETE FROM event_signup_responses
WHERE event_id = $1 AND public_id::text = $2`, eventID, publicID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return nil
	})
}

// reserveSignupCapacity verifies every requested block belongs to the event and
// that no limited block is already at capacity, excluding the response being
// rewritten. It must run inside the transaction that holds the event row lock,
// which serializes concurrent reservations for the event. The grouped join
// counts claims on the join table, but only memberships of responses in the same
// event; comparing block identities as text keeps a non-canonical identifier on
// the ErrSignupBlockNotFound path instead of a cast error.
func (r *Repository) reserveSignupCapacity(ctx context.Context, eventID string, blockIDs []string, excludeResponseID string) error {
	if len(blockIDs) == 0 {
		return nil
	}
	rows, err := r.db.Query(ctx, `SELECT block.id::text, block.capacity, count(response.id)
FROM event_signup_blocks block
LEFT JOIN event_signup_response_blocks membership
    ON membership.block_id = block.id
   AND membership.response_id::text <> $3
LEFT JOIN event_signup_responses response
    ON response.id = membership.response_id
   AND response.event_id = $1
WHERE block.event_id = $1 AND block.id::text = ANY($2)
GROUP BY block.id, block.capacity`, eventID, blockIDs, excludeResponseID)
	if err != nil {
		return err
	}
	type blockClaim struct {
		capacity *int
		claimed  int
	}
	claims := make(map[string]blockClaim, len(blockIDs))
	for rows.Next() {
		var id string
		var capacity *int
		var claimed int
		if err := rows.Scan(&id, &capacity, &claimed); err != nil {
			rows.Close()
			return err
		}
		claims[id] = blockClaim{capacity: capacity, claimed: claimed}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(claims) != len(blockIDs) {
		return ErrSignupBlockNotFound
	}
	for _, blockID := range blockIDs {
		claim := claims[blockID]
		if claim.capacity == nil {
			continue
		}
		if claim.claimed >= *claim.capacity {
			return ErrSignupCapacityExceeded
		}
	}
	return nil
}

// normalizeSignupResponseIdentity resolves the respondent kind and enforces the
// identity rules: account responses carry a platform identity uuid, and guest
// responses carry a canonical guest name produced by the shared normalizer.
func normalizeSignupResponseIdentity(response *SignupResponse) error {
	switch response.RespondentKind {
	case RespondentKindAccount, RespondentKindGuest:
	case "":
		if response.PlatformIdentityID != nil && *response.PlatformIdentityID != "" {
			response.RespondentKind = RespondentKindAccount
		} else {
			response.RespondentKind = RespondentKindGuest
		}
	default:
		return fmt.Errorf("unsupported signup respondent kind %q", response.RespondentKind)
	}
	if response.RespondentKind == RespondentKindAccount {
		if response.PlatformIdentityID == nil || *response.PlatformIdentityID == "" {
			return errors.New("account signup response requires a platform identity ID")
		}
		response.CanonicalGuestName = nil
		return nil
	}
	response.PlatformIdentityID = nil
	guestName := response.Name
	if guestName == "" && response.CanonicalGuestName != nil {
		guestName = *response.CanonicalGuestName
	}
	canonical := respondents.NormalizeGuestName(guestName)
	if canonical == "" {
		return errors.New("guest signup response requires a name")
	}
	response.Name = canonical
	response.CanonicalGuestName = &canonical
	return nil
}

// normalizeSignupBlockIDs trims, drops empty entries, and deduplicates block
// identities so a response cannot claim the same block twice.
func normalizeSignupBlockIDs(blockIDs []string) []string {
	seen := make(map[string]struct{}, len(blockIDs))
	result := make([]string, 0, len(blockIDs))
	for _, blockID := range blockIDs {
		blockID = strings.TrimSpace(blockID)
		if blockID == "" {
			continue
		}
		if _, ok := seen[blockID]; ok {
			continue
		}
		seen[blockID] = struct{}{}
		result = append(result, blockID)
	}
	return result
}

package postgres

import (
	"context"
	"time"
)

// AccessTransfer is one access transfer row. ID, ExpiresAt, and State are
// minted or defaulted by PostgreSQL; the transfer's source credential and the
// target platform identity are nullable because a transfer can be opened
// before either side is known.
//
// The struct stays handwritten because a GALA struct declaration also emits
// Copy, Equal, Unapply, a NewAccessTransfer constructor, and
// StructMeta_AccessTransfer on this exported type. transfers.gala carries the
// free functions that read and write it.
type AccessTransfer struct {
	ID                 string
	EventID            string
	SourceHash         []byte
	SourceCredentialID *string
	PlatformIdentityID *string
	GrantsOwner        bool
	ExpiresAt          time.Time
	State              string
	ApprovedRequestID  *string
	GrantID            *string
}

// TransferRequest is one target request against a transfer. TargetHash is the
// target credential's digest and never leaves the server.
//
// The struct stays handwritten for the same exported-struct reason as
// AccessTransfer, and its json tags define the wire shape the transfer status
// endpoint returns.
type TransferRequest struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	UserAgent  string `json:"userAgent"`
	TargetHash []byte `json:"-"`
}

// CreateAccessTransfer inserts one transfer and fills its generated identity,
// deadline, and initial state.
func (r *Repository) CreateAccessTransfer(ctx context.Context, v *AccessTransfer) error {
	return createAccessTransfer(r, ctx, v)
}

// LockAccessTransfer locks one transfer row for a lifecycle transition. All
// lifecycle transitions use this row lock, including opening target requests.
// A non-canonical transfer identifier resolves to no transfer instead of
// reaching the uuid column as an invalid literal.
func (r *Repository) LockAccessTransfer(ctx context.Context, eventID, id string) (*AccessTransfer, error) {
	return lockAccessTransfer(r, ctx, eventID, id)
}

// SaveAccessTransfer writes the transfer's lifecycle state and its approved
// request and grant links.
func (r *Repository) SaveAccessTransfer(ctx context.Context, v *AccessTransfer) error {
	return saveAccessTransfer(r, ctx, v)
}

// CreateTransferRequest records one target request against a transfer.
func (r *Repository) CreateTransferRequest(ctx context.Context, transferID string, v *TransferRequest) error {
	return createTransferRequest(r, ctx, transferID, v)
}

// PruneExpiredAccessTransfers deletes past-deadline transfers that can no
// longer progress, together with their cascade-deleted requests. Redeemed
// transfers are retained so the source can revoke issued grants.
func (r *Repository) PruneExpiredAccessTransfers(ctx context.Context) error {
	return pruneExpiredAccessTransfers(r, ctx)
}

// ListTransferRequests returns the target requests opened against a transfer in
// creation order.
//
// The method stays handwritten rather than delegating to transfers.gala because
// its final return is `result, rows.Err()`: an iteration error still carries the
// requests already scanned, and the transfer-status caller publishes that list
// without checking the error, so the GALA failure path's nil slice would be an
// observable change. Every other member of this file delegates.
func (r *Repository) ListTransferRequests(ctx context.Context, transferID string) ([]TransferRequest, error) {
	rows, err := r.db.Query(ctx, `SELECT id,code,target_hash,user_agent FROM access_transfer_requests WHERE transfer_id=$1 ORDER BY id`, transferID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []TransferRequest{}
	for rows.Next() {
		var v TransferRequest
		if err := rows.Scan(&v.ID, &v.Code, &v.TargetHash, &v.UserAgent); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

// RevokeGrantedCredential revokes one granted credential. The kind filter keeps
// a base credential from being revoked through the transfer path.
func (r *Repository) RevokeGrantedCredential(ctx context.Context, id string) error {
	return revokeGrantedCredential(r, ctx, id)
}

// GetTransferSourceCredential reads the transfer's source credential by its
// primary key in one query. The visitor association needs no re-filter, so the
// credential lookup and visitor lookup share a single row read.
func (r *Repository) GetTransferSourceCredential(ctx context.Context, id string) (*EventVisitorCredential, error) {
	return getTransferSourceCredential(r, ctx, id)
}

// GetTransferSourceVisitor reads the visitor identity a transfer's source
// credential belongs to.
func (r *Repository) GetTransferSourceVisitor(ctx context.Context, id string) (*EventVisitorIdentity, error) {
	return getTransferSourceVisitor(r, ctx, id)
}

package accounts

import (
	"context"

	pgstore "timeful/server/postgres"
)

// Deleter removes the authoritative account. The deletion transaction removes
// the account profile, platform identity, calendar connections, and daily-log
// membership, and records the deletion tombstone. It is atomic and idempotent
// by construction: a failure leaves the account authority intact and a retry
// converges. The deletion step is injectable so a test can force a failure.
//
// The type, its constructor, and its method stay handwritten together: delete.go
// is generated from delete.gala and only carries the package-level entry points,
// so a GALA-declared Deleter would also expose its synthesized Copy, Equal, and
// Unapply members on this exported type.
type Deleter struct {
	DeleteAccount func(ctx context.Context, platformIdentityID string) error
}

// DefaultDeleter returns the production deletion step.
func DefaultDeleter() Deleter {
	return Deleter{
		DeleteAccount: func(ctx context.Context, platformIdentityID string) error {
			repository, err := pgstore.DefaultRepository()
			if err != nil {
				return err
			}
			return repository.DeleteAccountByPlatformIdentityID(ctx, platformIdentityID)
		},
	}
}

// Delete applies the deletion. A nil step is skipped so a test can isolate the
// deletion boundary.
func (d Deleter) Delete(ctx context.Context, platformIdentityID string) error {
	if d.DeleteAccount == nil {
		return nil
	}
	return d.DeleteAccount(ctx, platformIdentityID)
}

var defaultDeleter = DefaultDeleter()

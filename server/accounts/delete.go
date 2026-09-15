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

// DeleteAccount permanently deletes the account and all data it owns.
func DeleteAccount(ctx context.Context, platformIdentityID string) error {
	return defaultDeleter.Delete(ctx, platformIdentityID)
}

// SetDefaultDeleter replaces the package deletion step and returns a function
// that restores the previous one. It exists so tests can force a deletion
// failure; it is not safe for concurrent use.
func SetDefaultDeleter(deleter Deleter) func() {
	previous := defaultDeleter
	defaultDeleter = deleter
	return func() { defaultDeleter = previous }
}

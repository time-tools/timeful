// Package accounts is the explicit boundary for the authoritative account.
// Profile and identity resolve through this package, and calendar connections,
// provider tokens, and calendar preferences resolve through its calendar
// boundary.
package accounts

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	pgstore "timeful/server/postgres"
	"timeful/server/utils"
)

// ErrNotFound reports that no account exists for an identifier.
var ErrNotFound = errors.New("account not found")

// Profile carries the profile fields a sign-in provider or OTP flow supplies.
type Profile struct {
	Email          string
	FirstName      string
	LastName       string
	Picture        string
	TimezoneOffset int
}

// Lookup returns the authoritative account.
func Lookup(ctx context.Context, platformIdentityID string) (*pgstore.Account, error) {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return nil, err
	}
	account, err := repository.GetAccountByPlatformIdentityID(ctx, platformIdentityID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return account, err
}

// Resolve returns the account for an existing sign-in session. A session whose
// account no longer exists reports ErrNotFound.
func Resolve(ctx context.Context, platformIdentityID string) (*pgstore.Account, error) {
	if platformIdentityID == "" {
		return nil, ErrNotFound
	}
	return Lookup(ctx, platformIdentityID)
}

// ResolveForSignIn returns the account for an OAuth or OTP sign-in. It prefers
// the authoritative account by email and creates a fresh account otherwise. The
// boolean reports whether the account was created during this call.
func ResolveForSignIn(ctx context.Context, profile Profile) (*pgstore.Account, bool, error) {
	email := utils.NormalizeEmail(profile.Email)
	if email == "" {
		return nil, false, errors.New("account email is required")
	}
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return nil, false, err
	}
	account, err := repository.GetAccountByEmail(ctx, email)
	if err == nil {
		return account, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	initial := pgstore.Account{
		Email:          email,
		FirstName:      strings.TrimSpace(profile.FirstName),
		LastName:       strings.TrimSpace(profile.LastName),
		Picture:        profile.Picture,
		TimezoneOffset: profile.TimezoneOffset,
	}
	return repository.FindOrCreateAccountByEmail(ctx, email, initial)
}

// IsNewUser reports whether no account exists for the email. A lookup failure
// is returned as an error rather than reported as account-exists or
// account-missing, so callers never treat a transient database failure as an
// existence result.
func IsNewUser(email string) (bool, error) {
	email = utils.NormalizeEmail(email)
	if email == "" {
		return true, nil
	}
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return false, err
	}
	if _, err := repository.GetAccountByEmail(context.Background(), email); err == nil {
		return false, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}
	return true, nil
}

// UpdateProfile writes authoritative profile fields.
func UpdateProfile(ctx context.Context, account *pgstore.Account) error {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return err
	}
	return repository.UpdateAccountProfile(ctx, account)
}

// IncrementEventsCreated advances the usage counter on the account.
func IncrementEventsCreated(ctx context.Context, platformIdentityID string) error {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return err
	}
	return repository.IncrementAccountEventsCreated(ctx, platformIdentityID)
}

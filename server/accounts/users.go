package accounts

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"timeful/server/logger"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

// UserFromAccount builds the internal user shape from the authoritative
// account. It performs no fallback lookup.
func UserFromAccount(account *pgstore.Account) *models.User {
	if account == nil {
		return nil
	}
	user := &models.User{Id: models.UUID(account.PlatformIdentityID)}
	user.Email = account.Email
	user.FirstName = account.FirstName
	user.LastName = account.LastName
	user.Picture = account.Picture
	user.HasCustomName = account.HasCustomName
	user.TimezoneOffset = account.TimezoneOffset
	user.NumEventsCreated = account.NumEventsCreated
	return user
}

// UserByPlatformIdentityID returns the authoritative account profile. A lookup
// that fails for any reason other than a genuine no-row result yields nil
// instead of an inferred account, so a transient database failure can never be
// mistaken for account authority.
func UserByPlatformIdentityID(platformIdentityID string) *models.User {
	account, err := accountByPlatformIdentityID(platformIdentityID)
	if err != nil {
		logger.StdErr.Printf("account lookup failed for %s: %v", platformIdentityID, err)
		return nil
	}
	return UserFromAccount(account)
}

// UserByEmail resolves the authoritative account by case-insensitive email and
// returns its profile. A lookup failure other than a genuine no-row result
// yields nil instead of an inferred account.
func UserByEmail(email string) *models.User {
	emailQuery := strings.TrimSpace(email)
	if emailQuery == "" {
		return nil
	}
	account, err := accountByEmail(emailQuery)
	if err != nil {
		logger.StdErr.Printf("account lookup failed for %s: %v", emailQuery, err)
		return nil
	}
	return UserFromAccount(account)
}

// accountByPlatformIdentityID resolves the authoritative account. A
// deliberately uninitialized pool and a missing account row both return no
// account; every other failure is returned so callers never mistake the error
// for absence.
func accountByPlatformIdentityID(platformIdentityID string) (*pgstore.Account, error) {
	if platformIdentityID == "" {
		return nil, nil
	}
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		if errors.Is(err, pgstore.ErrPoolUninitialized) {
			return nil, nil
		}
		return nil, err
	}
	account, err := repository.GetAccountByPlatformIdentityID(context.Background(), platformIdentityID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

// accountByEmail applies the same outcome classification as
// accountByPlatformIdentityID for a case-insensitive email lookup.
func accountByEmail(email string) (*pgstore.Account, error) {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		if errors.Is(err, pgstore.ErrPoolUninitialized) {
			return nil, nil
		}
		return nil, err
	}
	account, err := repository.GetAccountByEmail(context.Background(), email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

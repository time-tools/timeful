// Calendar integrations are authoritative. This file is the explicit boundary
// that translates between the internal API user shape (models) and the
// repository types (pgstore), so no route handler, provider, or service reaches
// the database directly.
package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

// CalendarIntegrations is the authoritative calendar state for one account
// expressed in the existing internal user shape, so routes, calendar providers,
// and the frontend response contract are unchanged.
type CalendarIntegrations struct {
	Accounts          map[string]models.CalendarAccount
	PrimaryAccountKey *string
	TokenOrigin       models.TokenOriginType
	CalendarOptions   *models.CalendarOptions
}

// CalendarPreferences carries the writable calendar preference fields for one
// account. A nil field is written as absent.
type CalendarPreferences struct {
	PrimaryAccountKey *string
	TokenOrigin       models.TokenOriginType
	CalendarOptions   *models.CalendarOptions
}

// LoadCalendarIntegrations reads every calendar connection, sub-calendar, and
// preference for an account. A missing preference row is an empty preference,
// not an error, and any other failure is returned.
func LoadCalendarIntegrations(ctx context.Context, platformIdentityID string) (*CalendarIntegrations, error) {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return nil, err
	}
	stored, err := repository.ListCalendarAccountsForUser(ctx, platformIdentityID)
	if err != nil {
		return nil, err
	}
	integrations := &CalendarIntegrations{Accounts: make(map[string]models.CalendarAccount, len(stored))}
	for _, account := range stored {
		integrations.Accounts[account.CalendarKey] = calendarAccountFromStorage(account)
	}
	preferences, err := repository.GetCalendarPreferences(ctx, platformIdentityID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return integrations, nil
	}
	integrations.PrimaryAccountKey = preferences.PrimaryAccountKey
	if preferences.TokenOrigin != nil {
		integrations.TokenOrigin = models.TokenOriginType(*preferences.TokenOrigin)
	}
	if len(preferences.CalendarOptions) > 0 {
		options, err := calendarOptionsFromJSON(preferences.CalendarOptions)
		if err != nil {
			return nil, err
		}
		integrations.CalendarOptions = options
	}
	return integrations, nil
}

// LoadSessionUser builds the authenticated user from the authoritative account
// profile and the calendar state.
func LoadSessionUser(ctx context.Context, account *pgstore.Account) (*models.User, error) {
	integrations, err := LoadCalendarIntegrations(ctx, account.PlatformIdentityID)
	if err != nil {
		return nil, err
	}
	return CalendarUser(account, integrations), nil
}

// LoadSessionUserByPlatformIdentityID resolves the authoritative account for a
// platform identity UUID and loads its calendar state.
func LoadSessionUserByPlatformIdentityID(ctx context.Context, platformIdentityID string) (*models.User, error) {
	account, err := Lookup(ctx, platformIdentityID)
	if err != nil {
		return nil, err
	}
	return LoadSessionUser(ctx, account)
}

// CalendarUser overlays the calendar state onto the authoritative account
// profile. An empty integration set leaves the calendar fields absent so the
// serialized response shape is unchanged.
func CalendarUser(account *pgstore.Account, integrations *CalendarIntegrations) *models.User {
	user := UserFromAccount(account)
	if integrations == nil {
		return user
	}
	if len(integrations.Accounts) > 0 {
		user.CalendarAccounts = integrations.Accounts
	}
	user.PrimaryAccountKey = integrations.PrimaryAccountKey
	user.TokenOrigin = integrations.TokenOrigin
	user.CalendarOptions = integrations.CalendarOptions
	return user
}

// SaveCalendarAccount creates or replaces one calendar connection together with
// its credentials and sub-calendars in a single transaction. The caller supplies
// the runtime calendar key so the email_CALENDARTYPE key semantics are
// preserved.
func SaveCalendarAccount(ctx context.Context, platformIdentityID, calendarKey string, account models.CalendarAccount) error {
	if calendarKey == "" {
		return errors.New("calendar account key is required")
	}
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return err
	}
	converted := calendarAccountToStorage(calendarKey, account)
	return repository.WithTransaction(ctx, func(ctx context.Context, tx *pgstore.Repository) error {
		if err := tx.UpsertCalendarAccount(ctx, platformIdentityID, converted); err != nil {
			return err
		}
		if account.SubCalendars == nil {
			return nil
		}
		return syncCalendarSubCalendars(ctx, tx, platformIdentityID, calendarKey, *account.SubCalendars)
	})
}

// SyncCalendarSubCalendars reconciles the stored sub-calendar set with a
// provider refresh: present sub-calendars are upserted and stored sub-calendars
// the provider no longer reports are removed. An empty set removes every stored
// sub-calendar.
func SyncCalendarSubCalendars(ctx context.Context, platformIdentityID, calendarKey string, subCalendars map[string]models.SubCalendar) error {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return err
	}
	return syncCalendarSubCalendars(ctx, repository, platformIdentityID, calendarKey, subCalendars)
}

// syncCalendarSubCalendars converts the internal sub-calendar map into the
// set-based sync. The batched repository path updates only the sub-calendar
// table and deliberately does not re-read the connection's encrypted
// credentials, so a provider refresh can no longer fail on credential
// decryption it does not need.
func syncCalendarSubCalendars(ctx context.Context, repository *pgstore.Repository, platformIdentityID, calendarKey string, subCalendars map[string]models.SubCalendar) error {
	converted := make([]pgstore.CalendarSubCalendar, 0, len(subCalendars))
	for id, sub := range subCalendars {
		if id == "" {
			return errors.New("sub-calendar ID is required")
		}
		converted = append(converted, pgstore.CalendarSubCalendar{
			SubCalendarID: id,
			Name:          sub.Name,
			Enabled:       sub.Enabled,
		})
	}
	return repository.SyncCalendarSubCalendars(ctx, platformIdentityID, calendarKey, converted)
}

// DeleteCalendarAccount removes one connection and its credentials and
// sub-calendars.
func DeleteCalendarAccount(ctx context.Context, platformIdentityID, calendarKey string) error {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return err
	}
	return repository.DeleteCalendarAccount(ctx, platformIdentityID, calendarKey)
}

// SetCalendarAccountEnabled writes the explicit connection enabled state.
func SetCalendarAccountEnabled(ctx context.Context, platformIdentityID, calendarKey string, enabled bool) error {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return err
	}
	return repository.SetCalendarAccountEnabled(ctx, platformIdentityID, calendarKey, enabled)
}

// SetSubCalendarEnabled writes the explicit sub-calendar enabled state.
func SetSubCalendarEnabled(ctx context.Context, platformIdentityID, calendarKey, subCalendarID string, enabled bool) error {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return err
	}
	return repository.SetCalendarSubCalendarEnabled(ctx, platformIdentityID, calendarKey, subCalendarID, enabled)
}

// SaveCalendarPreferences replaces the account's calendar preferences. The
// caller supplies the complete preference state, so a nil field clears it.
func SaveCalendarPreferences(ctx context.Context, platformIdentityID string, preferences CalendarPreferences) error {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return err
	}
	var options json.RawMessage
	if preferences.CalendarOptions != nil {
		encoded, err := json.Marshal(preferences.CalendarOptions)
		if err != nil {
			return err
		}
		options = encoded
	}
	var tokenOrigin *string
	if preferences.TokenOrigin != "" {
		value := string(preferences.TokenOrigin)
		tokenOrigin = &value
	}
	return repository.UpsertCalendarPreferences(ctx, platformIdentityID, &pgstore.CalendarPreferences{
		PrimaryAccountKey: preferences.PrimaryAccountKey,
		TokenOrigin:       tokenOrigin,
		CalendarOptions:   options,
	})
}

// UpdateCalendarAccessToken persists a refreshed OAuth2 access token and expiry
// for one connection.
func UpdateCalendarAccessToken(ctx context.Context, platformIdentityID, calendarKey, accessToken string, expiresAt time.Time) error {
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		return err
	}
	return repository.UpdateCalendarOAuthAccessToken(ctx, platformIdentityID, calendarKey, accessToken, expiresAt)
}

// calendarAccountToStorage translates the internal user-shaped connection into
// its storage form. Provider secrets stay plaintext here; the repository
// encrypts them at rest.
func calendarAccountToStorage(calendarKey string, account models.CalendarAccount) *pgstore.CalendarAccount {
	converted := &pgstore.CalendarAccount{
		CalendarKey:  calendarKey,
		CalendarType: string(account.CalendarType),
		Email:        account.Email,
		Picture:      account.Picture,
		Enabled:      account.Enabled,
	}
	if account.OAuth2CalendarAuth != nil {
		credentials := &pgstore.CalendarOAuth2Credentials{
			AccessToken:  account.OAuth2CalendarAuth.AccessToken,
			RefreshToken: account.OAuth2CalendarAuth.RefreshToken,
			Scope:        account.OAuth2CalendarAuth.Scope,
		}
		if account.OAuth2CalendarAuth.AccessTokenExpireDate > 0 {
			expiresAt := account.OAuth2CalendarAuth.AccessTokenExpireDate.Time()
			credentials.AccessTokenExpiresAt = &expiresAt
		}
		converted.OAuth2 = credentials
	}
	if account.AppleCalendarAuth != nil {
		converted.Apple = &pgstore.CalendarAppleCredentials{Password: account.AppleCalendarAuth.Password}
	}
	if account.ICSCalendarAuth != nil {
		converted.ICS = &pgstore.CalendarICSCredentials{FeedURL: account.ICSCalendarAuth.FeedURL}
	}
	if account.SubCalendars != nil {
		converted.SubCalendars = make(map[string]pgstore.CalendarSubCalendar, len(*account.SubCalendars))
		for id, sub := range *account.SubCalendars {
			converted.SubCalendars[id] = pgstore.CalendarSubCalendar{SubCalendarID: id, Name: sub.Name, Enabled: sub.Enabled}
		}
	}
	return converted
}

// calendarAccountFromStorage translates a stored connection into the internal
// user shape.
func calendarAccountFromStorage(account pgstore.CalendarAccount) models.CalendarAccount {
	converted := models.CalendarAccount{
		CalendarType: models.CalendarType(account.CalendarType),
		Email:        account.Email,
		Picture:      account.Picture,
		Enabled:      account.Enabled,
	}
	if account.OAuth2 != nil {
		credentials := &models.OAuth2CalendarAuth{
			AccessToken:  account.OAuth2.AccessToken,
			RefreshToken: account.OAuth2.RefreshToken,
			Scope:        account.OAuth2.Scope,
		}
		if account.OAuth2.AccessTokenExpiresAt != nil {
			credentials.AccessTokenExpireDate = models.NewDateTimeFromTime(*account.OAuth2.AccessTokenExpiresAt)
		}
		converted.OAuth2CalendarAuth = credentials
	}
	if account.Apple != nil {
		converted.AppleCalendarAuth = &models.AppleCalendarAuth{Email: account.Email, Password: account.Apple.Password}
	}
	if account.ICS != nil {
		converted.ICSCalendarAuth = &models.ICSCalendarAuth{FeedURL: account.ICS.FeedURL, Label: account.Email}
	}
	if len(account.SubCalendars) > 0 {
		subCalendars := make(map[string]models.SubCalendar, len(account.SubCalendars))
		for id, sub := range account.SubCalendars {
			subCalendars[id] = models.SubCalendar{Name: sub.Name, Enabled: sub.Enabled}
		}
		converted.SubCalendars = &subCalendars
	}
	return converted
}

func calendarOptionsFromJSON(raw json.RawMessage) (*models.CalendarOptions, error) {
	var options models.CalendarOptions
	if err := json.Unmarshal(raw, &options); err != nil {
		return nil, err
	}
	return &options, nil
}

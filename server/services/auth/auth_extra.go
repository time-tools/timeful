package auth

import (
	"context"
	"time"

	"timeful/server/accounts"
	"timeful/server/logger"
	"timeful/server/models"
	"timeful/server/utils"
)

// GoogleIdTokenInfo holds the claims returned by Google's tokeninfo endpoint
// after it has validated an ID token's signature and expiry. Google returns
// the numeric claims (aud, exp, ...) as strings.
type GoogleIdTokenInfo struct {
	// Aud is the OAuth client ID the token was issued for.
	Aud string `json:"aud"`
	// Iss is the token issuer (accounts.google.com).
	Iss string `json:"iss"`
	// Email is the verified email address on the account.
	Email string `json:"email"`
	// EmailVerified is "true" when Google has verified the email.
	EmailVerified string `json:"email_verified"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	// Error is set by tokeninfo when the token is invalid.
	Error string `json:"error"`
}

type RefreshAccessTokenData struct {
	TokenResponse AccessTokenResponse
	Email         string
	CalendarType  models.CalendarType
	Error         *interface{}
}

func RefreshAccessTokenAsync(email string, accountAuth *models.OAuth2CalendarAuth, calendarType models.CalendarType, c chan RefreshAccessTokenData) {
	// Recover from panics
	defer func() {
		if err := recover(); err != nil {
			c <- RefreshAccessTokenData{Error: &err}
		}
	}()

	tokenResponse := RefreshAccessToken(accountAuth, calendarType)

	c <- RefreshAccessTokenData{tokenResponse, email, calendarType, nil}
}

// If access token has expired, get a new token for the primary account as well as all other calendar accounts, update the user object, and save it to the database
// `calendarAccounts` specifies for which accounts to refresh access tokens. If `calendarAccounts` is nil or empty, then update tokens for all accounts
func RefreshUserTokenIfNecessary(u *models.User, calendarAccounts models.Set[string]) {
	refreshTokenChan := make(chan RefreshAccessTokenData)
	numAccountsToUpdate := 0

	// If `calendarAccounts` is nil, then update tokens for all accounts
	updateAllAccounts := len(calendarAccounts) == 0

	// Refresh calendar account access tokens if necessary
	for accountKey, account := range u.CalendarAccounts {
		if account.OAuth2CalendarAuth != nil { // Only refresh access tokens for OAuth2 calendar accounts
			accountAuth := account.OAuth2CalendarAuth

			if _, ok := calendarAccounts[accountKey]; ok || updateAllAccounts {
				if time.Now().After(accountAuth.AccessTokenExpireDate.Time()) && len(accountAuth.RefreshToken) > 0 {
					go RefreshAccessTokenAsync(account.Email, accountAuth, account.CalendarType, refreshTokenChan)
					numAccountsToUpdate++
				}
			}
		}
	}

	// Update access tokens as responses are received
	type refreshedAccessToken struct {
		calendarKey string
		accessToken string
		expiresAt   time.Time
	}
	refreshed := make([]refreshedAccessToken, 0, numAccountsToUpdate)
	for i := 0; i < numAccountsToUpdate; i++ {
		res := <-refreshTokenChan

		if res.Error != nil {
			continue
		}

		accessTokenExpireDate := utils.GetAccessTokenExpireDate(res.TokenResponse.ExpiresIn)

		calendarAccountKey := utils.GetCalendarAccountKey(res.Email, res.CalendarType)
		if calendarAccount, ok := u.CalendarAccounts[calendarAccountKey]; ok {
			calendarAccount.OAuth2CalendarAuth.AccessToken = res.TokenResponse.AccessToken
			calendarAccount.OAuth2CalendarAuth.AccessTokenExpireDate = models.NewDateTimeFromTime(accessTokenExpireDate)
			u.CalendarAccounts[calendarAccountKey] = calendarAccount
			refreshed = append(refreshed, refreshedAccessToken{
				calendarKey: calendarAccountKey,
				accessToken: res.TokenResponse.AccessToken,
				expiresAt:   accessTokenExpireDate,
			})
		}
	}

	// Persist refreshed access tokens to PostgreSQL; the refresh path does not
	// write the account profile.
	if len(refreshed) > 0 {
		platformIdentityID := u.Id.String()
		for _, token := range refreshed {
			if err := accounts.UpdateCalendarAccessToken(context.Background(), platformIdentityID, token.calendarKey, token.accessToken, token.expiresAt); err != nil {
				logger.StdErr.Println(err)
			}
		}
	}
}

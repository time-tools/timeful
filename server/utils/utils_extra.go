package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
)

// Returns the currently signed in user
//
// The member stays handwritten: a single-value type assertion that must panic on
// a missing value has no GALA spelling that preserves that panic.
func GetAuthUser(c *gin.Context) *models.User {
	userInterface, _ := c.Get("authUser")
	user := userInterface.(*models.User)
	return user
}

// Returns the authoritative PostgreSQL account for the current session.
//
// The comma-ok assertion stays handwritten for the same reason as GetAuthUser.
func GetAuthAccount(c *gin.Context) *pgstore.Account {
	accountInterface, ok := c.Get("authAccount")
	if !ok {
		return nil
	}
	account, _ := accountInterface.(*pgstore.Account)
	return account
}

// CORSOrigins returns the canonical application origin and any configured additional origins.
//
// The dedup set stays handwritten because map[string]struct{} is the struct{}
// finding; the origin normalization it calls lives in utils.gala.
func CORSOrigins(additionalOrigins string) ([]string, error) {
	baseUrl, err := normalizedBaseUrl(os.Getenv("APP_BASE_URL"))
	if err != nil {
		return nil, err
	}

	origins := []string{baseUrl}
	seen := map[string]struct{}{baseUrl: {}}
	for _, origin := range strings.Split(additionalOrigins, ",") {
		if strings.TrimSpace(origin) == "" {
			continue
		}

		normalizedOrigin, err := normalizedBaseUrl(origin)
		if err != nil {
			return nil, fmt.Errorf("CORS_ORIGINS contains an invalid origin: %w", err)
		}
		if _, exists := seen[normalizedOrigin]; exists {
			continue
		}

		seen[normalizedOrigin] = struct{}{}
		origins = append(origins, normalizedOrigin)
	}

	return origins, nil
}

package utils

import (
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

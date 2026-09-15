package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"timeful/server/accounts"
	"timeful/server/errs"
	"timeful/server/responses"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// The session carries the account's platform identity UUID.
		session := sessions.Default(c)
		platformIdentityID, ok := session.Get("userId").(string)
		if !ok || platformIdentityID == "" {
			c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.NotSignedIn})
			c.Abort()
			return
		}

		// Resolve the authoritative PostgreSQL account for the session.
		account, err := accounts.Resolve(c.Request.Context(), platformIdentityID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.UserDoesNotExist})
			c.Abort()
			return
		}

		// Calendar connections, provider tokens, sub-calendars, and preferences
		// are PostgreSQL-authoritative and loaded through the accounts boundary.
		user, err := accounts.LoadSessionUser(c.Request.Context(), account)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.Error{Error: "failed to load account integration data"})
			c.Abort()
			return
		}

		c.Set("authUser", user)
		c.Set("authAccount", account)

		c.Next()
	}
}

package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"timeful/server/accounts"
	"timeful/server/errs"
	"timeful/server/models"
	"timeful/server/responses"
)

func InitUsers(router *gin.RouterGroup) {
	usersRouter := router.Group("/users")

	// Public profile for invite screens, ads, etc. (no auth). Must be registered
	// after more specific /:userId/... routes.
	usersRouter.GET("/:userId", getPublicUserProfile)
}

// @Summary Returns a minimal public user profile (safe for unauthenticated clients)
// @Tags users
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} models.User
// @Router /users/{userId} [get]
func getPublicUserProfile(c *gin.Context) {
	platformIdentityID := c.Param("userId")
	user := accounts.UserByPlatformIdentityID(platformIdentityID)
	if user == nil {
		c.JSON(http.StatusNotFound, responses.Error{Error: errs.UserDoesNotExist})
		return
	}

	public := models.User{
		Id:        user.Id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Picture:   user.Picture,
	}
	c.JSON(http.StatusOK, public)
}

package routes

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// grantAssociationInput carries the grant-association request body. It stays
// handwritten because its tag defines the request wire shape, and grants.gala
// constructs it through newGrantAssociationInput instead of naming the local
// type.
type grantAssociationInput struct {
	Confirm bool `json:"confirm"`
}

func newGrantAssociationInput() *grantAssociationInput {
	return &grantAssociationInput{}
}

// grantPlatformIdentityID reads the session's userId as a string. The comma-ok
// assertion stays handwritten; grants.gala reads the result as a plain value.
func grantPlatformIdentityID(c *gin.Context) string {
	platformIdentityID, _ := sessions.Default(c).Get("userId").(string)
	return platformIdentityID
}

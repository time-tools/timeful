package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// sessionIdentityID returns the platform identity UUID stored on the session.
// The empty string means the session value is absent or is not a string, the
// same two cases AuthRequired rejects.
//
// The type assertion stays in Go because a typed pattern match in GALA lowers
// to std.As and would tie the generated twin to the GALA runtime. This is a
// mixed-package sibling; see docs/GO_INTEROP.MD Part 3 in the GALA repository.
func sessionIdentityID(c *gin.Context) string {
	platformIdentityID, ok := sessions.Default(c).Get("userId").(string)
	if !ok {
		return ""
	}
	return platformIdentityID
}

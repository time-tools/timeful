package routes

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var routesReadFiltersTestDBOnce sync.Once

// initRoutesReadFiltersTestDB prepares the shared route-test environment. It
// does not initialize any database: route-test callers initialize the pool
// through routeTestDBOnce.
func initRoutesReadFiltersTestDB(t *testing.T) {
	t.Helper()

	routesReadFiltersTestDBOnce.Do(func() {
		gin.SetMode(gin.TestMode)
		if os.Getenv("SESSION_SECRET") == "" {
			_ = os.Setenv("SESSION_SECRET", "01234567890123456789012345678901")
		}
		// Calendar credentials are encrypted at rest with a 32-byte
		// ENCRYPTION_KEY; the isolated route tests supply one.
		if os.Getenv("ENCRYPTION_KEY") == "" {
			_ = os.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
		}
	})
}

func newEventsReadFiltersTestRouter() *gin.Engine {
	router := gin.New()
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	router.Use(sessions.Sessions("session", store))

	apiRouter := router.Group("/api")
	InitEvents(apiRouter)

	return router
}

func decodeJSONBody[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()

	var payload T
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	return payload
}

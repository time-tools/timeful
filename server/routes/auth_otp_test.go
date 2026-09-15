package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"timeful/server/logger"
)

func initAuthOtpTestEnv(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	if os.Getenv("SESSION_SECRET") == "" {
		_ = os.Setenv("SESSION_SECRET", "01234567890123456789012345678901")
	}
	logger.Init(io.Discard)
}

func newAuthOtpTestRouter() *gin.Engine {
	router := gin.New()
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	router.Use(gin.Recovery())
	router.Use(sessions.Sessions("session", store))
	apiRouter := router.Group("/api")
	InitAuth(apiRouter)
	return router
}

func sendOtpRequest(t *testing.T, router *gin.Engine) *httptest.ResponseRecorder {
	t.Helper()

	response := httptest.NewRecorder()
	router.ServeHTTP(
		response,
		httptest.NewRequest(
			http.MethodPost,
			"/api/auth/otp/send",
			strings.NewReader(`{"email":"jks@d.com"}`),
		),
	)
	return response
}

func TestSendOtpReturnsJSONErrorWhenOtpTemplateIdMissing(t *testing.T) {
	initAuthOtpTestEnv(t)

	t.Setenv("LISTMONK_OTP_EMAIL_TEMPLATE_ID", "")
	t.Setenv("LISTMONK_OTP_FROM_ADDRESS", "")

	response := sendOtpRequest(t, newAuthOtpTestRouter())

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON error body, got %q: %v", response.Body.String(), err)
	}
	if body["error"] == "" {
		t.Fatalf("expected error message in body, got %q", response.Body.String())
	}
}

func TestSendOtpReturnsJSONErrorWhenOtpFromAddressMissing(t *testing.T) {
	initAuthOtpTestEnv(t)

	t.Setenv("LISTMONK_OTP_EMAIL_TEMPLATE_ID", "1")
	t.Setenv("LISTMONK_OTP_FROM_ADDRESS", "")

	response := sendOtpRequest(t, newAuthOtpTestRouter())

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON error body, got %q: %v", response.Body.String(), err)
	}
	if body["error"] == "" {
		t.Fatalf("expected error message in body, got %q", response.Body.String())
	}
}

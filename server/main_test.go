package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestHealthRouteReportsPostgresAvailability proves that /api/health depends on
// PostgreSQL alone and succeeds with no retained-store dependency.
func TestHealthRouteReportsPostgresAvailability(t *testing.T) {
	previousPostgresPing := postgresPing
	t.Cleanup(func() {
		postgresPing = previousPostgresPing
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	initHealthRoute(router.Group("/api"))

	t.Run("available", func(t *testing.T) {
		postgresPing = func(context.Context) error {
			return nil
		}

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/health", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
	})

	t.Run("postgres unavailable", func(t *testing.T) {
		postgresPing = func(context.Context) error {
			return errors.New("postgres unavailable")
		}

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/health", nil))

		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
		}
	})

	t.Run("live", func(t *testing.T) {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/health/live", nil))

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
	})
}

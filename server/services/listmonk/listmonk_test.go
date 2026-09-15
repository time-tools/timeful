//go:build integration

package listmonk

import (
	"log"
	"os"
	"testing"

	"timeful/server/logger"
)

func TestSendEmail(t *testing.T) {
	// Init logfile
	logFile, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Init logger
	logger.Init(logFile)

	SendEmail("timeful.team@example.com", 8, map[string]any{
		"eventName": "casablanca",
		"eventUrl":  "http://localhost:8080/e/65e636bb760d3ea2e113e161",
	})
}

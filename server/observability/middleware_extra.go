package observability

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
)

const maxCapturedErrorBytes = 1024

// errorCapture stays handwritten: it embeds a net/http response writer, and the
// body field, the bounded buffer and the tagged JSON probe in errorContext are
// members the .gala source cannot declare. The middleware in middleware.gala
// constructs it through newErrorCapture.
type errorCapture struct {
	gin.ResponseWriter
	body  bytes.Buffer
	limit int
}

// newRequestCompletion builds the canonical finished-request value for
// middleware.gala, whose source never names this package's sibling types.
func newRequestCompletion(requestID string, method string, route string, status int, latency time.Duration, startTime time.Time, readiness ReadinessState) RequestCompletion {
	return RequestCompletion{
		RequestID: requestID,
		Method:    method,
		Route:     route,
		Status:    status,
		Latency:   latency,
		StartTime: startTime,
		Readiness: readiness,
	}
}

func newErrorCapture(writer gin.ResponseWriter) *errorCapture {
	return &errorCapture{ResponseWriter: writer, limit: maxCapturedErrorBytes}
}

func (w *errorCapture) Write(data []byte) (int, error) {
	w.capture(data)
	return w.ResponseWriter.Write(data)
}

func (w *errorCapture) WriteString(data string) (int, error) {
	w.capture([]byte(data))
	return w.ResponseWriter.WriteString(data)
}

func (w *errorCapture) capture(data []byte) {
	if w.Status() < 400 || w.body.Len() >= w.limit {
		return
	}
	remaining := w.limit - w.body.Len()
	if len(data) > remaining {
		data = data[:remaining]
	}
	w.body.Write(data)
}

// errorContext extracts the JSON error field from a failed response. A short
// code-like value becomes error.type; free-form text becomes a redacted
// error.message. Anything unrecognized falls back to the status class.
func (w *errorCapture) errorContext(status int) (string, string) {
	if status < 400 || w.body.Len() == 0 {
		return statusErrorType(status), ""
	}
	var payload struct {
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(w.body.Bytes(), &payload); err != nil || len(payload.Error) == 0 {
		return statusErrorType(status), ""
	}
	var message string
	if err := json.Unmarshal(payload.Error, &message); err != nil {
		return statusErrorType(status), ""
	}
	return classifyError(status, message)
}

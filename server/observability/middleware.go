package observability

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	requestIDContextKey = "observability.requestID"
	// RequestIDHeader lets an operator correlate a failed response with its
	// shipped record (QR-010).
	RequestIDHeader       = "X-Request-ID"
	maxCapturedErrorBytes = 1024
	unmatchedRoute        = "unmatched"
)

var errorCodePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// RequestMiddleware records one structured diagnostic record and one server
// span per finished request, and records the request duration metric. It
// generates the request correlation identifier, starts the trace span that the
// log record inherits, captures only the bounded JSON error field of failed
// responses, and hands the completion to the recorder, which enqueues it
// without waiting on OpenObserve.
func RequestMiddleware(recorder *Recorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		requestID := newRequestID()
		c.Set(requestIDContextKey, requestID)
		c.Header(RequestIDHeader, requestID)

		spanContext, span := recorder.StartRequestSpan(c.Request.Context(), c.Request.Method)
		// Handlers and the repository layer read the request context, so the
		// span travels with it and database work joins the request trace.
		c.Request = c.Request.WithContext(spanContext)

		capture := newErrorCapture(c.Writer)
		c.Writer = capture

		c.Next()

		completion := RequestCompletion{
			RequestID: requestID,
			Method:    c.Request.Method,
			Route:     requestRoute(c),
			Status:    capture.Status(),
			Latency:   time.Since(started),
			StartTime: started,
			Readiness: recorder.readinessState(),
		}
		completion.ErrorType, completion.ErrorMessage = capture.errorContext(completion.Status)
		if completion.ErrorMessage == "" && len(c.Errors) > 0 {
			completion.ErrorMessage = c.Errors.String()
		}
		completion.ApplyToSpan(span)
		recorder.RecordRequestMetrics(spanContext, completion)
		recorder.EmitRequest(spanContext, completion)
		span.End()
	}
}

func requestRoute(c *gin.Context) string {
	if route := c.FullPath(); route != "" {
		return route
	}
	return unmatchedRoute
}

func newRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("request-%d", time.Now().UnixNano())
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16])
}

// errorCapture mirrors writes so failed responses can contribute their error
// field to the record without buffering successful response payloads.
type errorCapture struct {
	gin.ResponseWriter
	body  bytes.Buffer
	limit int
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

func classifyError(status int, message string) (string, string) {
	message = strings.TrimSpace(Redact(message))
	if message == "" {
		return statusErrorType(status), ""
	}
	if errorCodePattern.MatchString(message) {
		return message, ""
	}
	return statusErrorType(status), message
}

func statusErrorType(status int) string {
	if status >= 500 {
		return "http_5xx"
	}
	if status >= 400 {
		return "http_4xx"
	}
	return ""
}

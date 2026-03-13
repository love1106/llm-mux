package executor

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nghyane/llm-mux/internal/provider"
	log "github.com/nghyane/llm-mux/internal/logging"
)

type HTTPErrorResult struct {
	Error      error
	StatusCode int
	Body       []byte
}

func HandleHTTPError(resp *http.Response, executorName string) HTTPErrorResult {
	body, readErr := io.ReadAll(resp.Body)

	if readErr != nil {
		return HTTPErrorResult{
			Error:      fmt.Errorf("%s: failed to read error response body: %w", executorName, readErr),
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}

	LogUpstreamError(executorName, resp.StatusCode, summarizeErrorBody(resp.Header.Get("Content-Type"), body))

	return HTTPErrorResult{
		Error:      NewStatusError(resp.StatusCode, string(body), nil),
		StatusCode: resp.StatusCode,
		Body:       body,
	}
}

type StatusError struct {
	code       int
	msg        string
	retryAfter *time.Duration
	category   provider.ErrorCategory
}

func (e StatusError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return fmt.Sprintf("status %d", e.code)
}

func (e StatusError) StatusCode() int { return e.code }

func (e StatusError) RetryAfter() *time.Duration { return e.retryAfter }

func (e StatusError) Category() provider.ErrorCategory { return e.category }

func (e StatusError) Unwrap() error { return nil }

func NewStatusError(code int, msg string, retryAfter *time.Duration) StatusError {
	return StatusError{
		code:       code,
		msg:        msg,
		retryAfter: retryAfter,
		category:   provider.CategorizeError(code, msg),
	}
}

func NewAuthError(msg string) StatusError {
	return NewStatusError(http.StatusUnauthorized, msg, nil)
}

func NewInternalError(msg string) StatusError {
	return NewStatusError(http.StatusInternalServerError, msg, nil)
}

func NewNotImplementedError(msg string) StatusError {
	return NewStatusError(http.StatusNotImplemented, msg, nil)
}

func NewTimeoutError(msg string) StatusError {
	return NewStatusError(http.StatusRequestTimeout, msg, nil)
}

// LogUpstreamError logs upstream provider errors at the appropriate severity.
// Server errors (5xx), auth errors (401/403), and quota errors (429) are logged
// at Warn level since they indicate issues needing operator attention.
// User errors (4xx) are logged at Debug level since they are caused by client requests.
func LogUpstreamError(executorName string, statusCode int, body string) {
	switch {
	case statusCode >= 500, statusCode == 401, statusCode == 403, statusCode == 429:
		log.Warnf("%s: upstream error: HTTP %d: %s", executorName, statusCode, body)
	default:
		log.Debugf("%s: upstream error: HTTP %d: %s", executorName, statusCode, body)
	}
}

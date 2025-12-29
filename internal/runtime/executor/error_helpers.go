package executor

import (
	"fmt"
	"io"
	"net/http"
	"time"

	cliproxyauth "github.com/nghyane/llm-mux/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
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

	log.Debugf("%s: error status: %d, body: %s", executorName, resp.StatusCode,
		summarizeErrorBody(resp.Header.Get("Content-Type"), body))

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
	category   cliproxyauth.ErrorCategory
}

func (e StatusError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return fmt.Sprintf("status %d", e.code)
}

func (e StatusError) StatusCode() int { return e.code }

func (e StatusError) RetryAfter() *time.Duration { return e.retryAfter }

func (e StatusError) Category() cliproxyauth.ErrorCategory { return e.category }

func (e StatusError) Unwrap() error { return nil }

func NewStatusError(code int, msg string, retryAfter *time.Duration) StatusError {
	return StatusError{
		code:       code,
		msg:        msg,
		retryAfter: retryAfter,
		category:   cliproxyauth.CategorizeError(code, msg),
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

package management

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nghyane/llm-mux/internal/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewHandler(t *testing.T) {
	cfg := &config.Config{}
	h := NewHandler(cfg, "/tmp/config.yaml", nil)

	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
	if h.cfg != cfg {
		t.Error("Handler config not set correctly")
	}
	if h.configFilePath != "/tmp/config.yaml" {
		t.Error("Handler configFilePath not set correctly")
	}
	if h.failedAttempts == nil {
		t.Error("Handler failedAttempts map not initialized")
	}
}

func TestHandler_SetConfig(t *testing.T) {
	h := NewHandler(&config.Config{}, "", nil)
	newCfg := &config.Config{Port: 9999}

	h.SetConfig(newCfg)

	if h.getConfig().Port != 9999 {
		t.Error("SetConfig did not update config")
	}
}

func TestHandler_SetLocalPassword(t *testing.T) {
	h := NewHandler(&config.Config{}, "", nil)
	h.SetLocalPassword("secret123")

	if h.localPassword != "secret123" {
		t.Error("SetLocalPassword did not set password")
	}
}

func TestHandler_SetLogDirectory(t *testing.T) {
	h := NewHandler(&config.Config{}, "", nil)

	h.SetLogDirectory("")
	if h.logDir != "" {
		t.Error("SetLogDirectory should ignore empty string")
	}

	h.SetLogDirectory("/var/log/test")
	if h.logDir != "/var/log/test" {
		t.Errorf("Expected /var/log/test, got %s", h.logDir)
	}
}

func TestRespondOK(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondOK(c, gin.H{"status": "ok"})

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"status":"ok"`) {
		t.Errorf("Response body missing status: %s", body)
	}
	if !strings.Contains(body, `"data"`) {
		t.Errorf("Response body missing data envelope: %s", body)
	}
	if !strings.Contains(body, `"meta"`) {
		t.Errorf("Response body missing meta envelope: %s", body)
	}
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"code":"INVALID_REQUEST"`) {
		t.Errorf("Response body missing error code: %s", body)
	}
	if !strings.Contains(body, `"message":"test error"`) {
		t.Errorf("Response body missing error message: %s", body)
	}
}

func TestRespondBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondBadRequest(c, "bad request message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRespondNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondNotFound(c, "not found message")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestRespondInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondInternalError(c, "internal error message")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRespondUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondUnauthorized(c, "unauthorized message")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestBindBoolValue_Valid(t *testing.T) {
	h := NewHandler(&config.Config{}, "", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"value": true}`))
	c.Request.Header.Set("Content-Type", "application/json")

	val, ok := h.bindBoolValue(c)
	if !ok {
		t.Error("bindBoolValue should return ok=true for valid input")
	}
	if !val {
		t.Error("bindBoolValue should return true")
	}
}

func TestBindBoolValue_Invalid(t *testing.T) {
	h := NewHandler(&config.Config{}, "", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"value": "not a bool"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	_, ok := h.bindBoolValue(c)
	if ok {
		t.Error("bindBoolValue should return ok=false for invalid input")
	}
}

func TestBindIntValue_Valid(t *testing.T) {
	h := NewHandler(&config.Config{}, "", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"value": 42}`))
	c.Request.Header.Set("Content-Type", "application/json")

	val, ok := h.bindIntValue(c)
	if !ok {
		t.Error("bindIntValue should return ok=true for valid input")
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestBindIntValue_Invalid(t *testing.T) {
	h := NewHandler(&config.Config{}, "", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"value": "not an int"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	_, ok := h.bindIntValue(c)
	if ok {
		t.Error("bindIntValue should return ok=false for invalid input")
	}
}

func TestBindStringValue_Valid(t *testing.T) {
	h := NewHandler(&config.Config{}, "", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"value": "hello"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	val, ok := h.bindStringValue(c)
	if !ok {
		t.Error("bindStringValue should return ok=true for valid input")
	}
	if val != "hello" {
		t.Errorf("Expected 'hello', got '%s'", val)
	}
}

func TestBindStringValue_Invalid(t *testing.T) {
	h := NewHandler(&config.Config{}, "", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"value": 123}`))
	c.Request.Header.Set("Content-Type", "application/json")

	_, ok := h.bindStringValue(c)
	if ok {
		t.Error("bindStringValue should return ok=false for invalid input")
	}
}

func TestErrorCodes(t *testing.T) {
	codes := []string{
		ErrCodeInvalidRequest,
		ErrCodeInvalidConfig,
		ErrCodeNotFound,
		ErrCodeUnauthorized,
		ErrCodeForbidden,
		ErrCodeInternalError,
		ErrCodeWriteFailed,
		ErrCodeReloadFailed,
		ErrCodeValidation,
	}

	for _, code := range codes {
		if code == "" {
			t.Error("Error code should not be empty")
		}
	}
}

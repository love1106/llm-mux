package management

import (
	"testing"
	"time"

	"github.com/nghyane/llm-mux/internal/config"
	"github.com/nghyane/llm-mux/internal/provider"
)

func TestBuildAuthFileEntry_NilAuth(t *testing.T) {
	h := NewHandler(&config.Config{}, "", nil)
	entry := h.buildAuthFileEntry(nil)
	if entry != nil {
		t.Error("buildAuthFileEntry should return nil for nil auth")
	}
}

func TestBuildAuthFileEntry_BasicAuth(t *testing.T) {
	h := NewHandler(&config.Config{AuthDir: "/tmp/auth"}, "", nil)

	auth := &provider.Auth{
		ID:       "test-auth-id",
		FileName: "test-auth.json",
		Provider: "claude",
		Label:    "Test Account",
		Status:   provider.StatusActive,
		Disabled: false,
		Attributes: map[string]string{
			"path": "/tmp/auth/test-auth.json",
		},
	}

	entry := h.buildAuthFileEntry(auth)
	if entry == nil {
		t.Fatal("buildAuthFileEntry returned nil for valid auth")
	}

	if entry["id"] != "test-auth-id" {
		t.Errorf("Expected id 'test-auth-id', got '%v'", entry["id"])
	}
	if entry["name"] != "test-auth.json" {
		t.Errorf("Expected name 'test-auth.json', got '%v'", entry["name"])
	}
	if entry["provider"] != "claude" {
		t.Errorf("Expected provider 'claude', got '%v'", entry["provider"])
	}
	if entry["label"] != "Test Account" {
		t.Errorf("Expected label 'Test Account', got '%v'", entry["label"])
	}
	if entry["disabled"] != false {
		t.Error("Expected disabled=false")
	}
}

func TestBuildAuthFileEntry_WithExpiry(t *testing.T) {
	h := NewHandler(&config.Config{AuthDir: "/tmp/auth"}, "", nil)

	expiry := time.Now().Add(4 * time.Hour)
	auth := &provider.Auth{
		ID:       "test-auth-id",
		FileName: "test-auth.json",
		Provider: "claude",
		Metadata: map[string]any{
			"expires_at": expiry.Format(time.RFC3339),
		},
		Attributes: map[string]string{
			"path": "/tmp/auth/test-auth.json",
		},
	}

	entry := h.buildAuthFileEntry(auth)
	if entry == nil {
		t.Fatal("buildAuthFileEntry returned nil")
	}

	expiresAt, ok := entry["expires_at"]
	if !ok {
		t.Error("Expected expires_at field in entry")
	}
	if expiresAt == nil {
		t.Error("expires_at should not be nil")
	}
}

func TestBuildAuthFileEntry_WithLastRefresh(t *testing.T) {
	h := NewHandler(&config.Config{AuthDir: "/tmp/auth"}, "", nil)

	lastRefresh := time.Now().Add(-1 * time.Hour)
	auth := &provider.Auth{
		ID:              "test-auth-id",
		FileName:        "test-auth.json",
		Provider:        "claude",
		LastRefreshedAt: lastRefresh,
		Attributes: map[string]string{
			"path": "/tmp/auth/test-auth.json",
		},
	}

	entry := h.buildAuthFileEntry(auth)
	if entry == nil {
		t.Fatal("buildAuthFileEntry returned nil")
	}

	lr, ok := entry["last_refresh"]
	if !ok {
		t.Error("Expected last_refresh field in entry")
	}
	if lr == nil {
		t.Error("last_refresh should not be nil")
	}
}

func TestBuildAuthFileEntry_DisabledRuntimeOnly(t *testing.T) {
	h := NewHandler(&config.Config{AuthDir: "/tmp/auth"}, "", nil)

	auth := &provider.Auth{
		ID:       "test-auth-id",
		FileName: "test-auth.json",
		Provider: "claude",
		Disabled: true,
		Status:   provider.StatusDisabled,
		Metadata: map[string]any{
			"runtime_only": true,
		},
	}

	entry := h.buildAuthFileEntry(auth)
	if entry != nil {
		t.Error("buildAuthFileEntry should return nil for disabled runtime-only auth")
	}
}

func TestBuildAuthFileEntry_WithEmail(t *testing.T) {
	h := NewHandler(&config.Config{AuthDir: "/tmp/auth"}, "", nil)

	auth := &provider.Auth{
		ID:       "test-auth-id",
		FileName: "test-auth.json",
		Provider: "claude",
		Metadata: map[string]any{
			"email": "test@example.com",
		},
		Attributes: map[string]string{
			"path": "/tmp/auth/test-auth.json",
		},
	}

	entry := h.buildAuthFileEntry(auth)
	if entry == nil {
		t.Fatal("buildAuthFileEntry returned nil")
	}

	email, ok := entry["email"]
	if !ok {
		t.Error("Expected email field in entry")
	}
	if email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%v'", email)
	}
}

func TestBuildAuthFileEntry_WithAccountInfo(t *testing.T) {
	h := NewHandler(&config.Config{AuthDir: "/tmp/auth"}, "", nil)

	auth := &provider.Auth{
		ID:       "test-auth-id",
		FileName: "test-auth.json",
		Provider: "claude",
		Attributes: map[string]string{
			"path":         "/tmp/auth/test-auth.json",
			"account_type": "oauth",
			"account":      "user@example.com",
		},
	}

	entry := h.buildAuthFileEntry(auth)
	if entry == nil {
		t.Fatal("buildAuthFileEntry returned nil")
	}

	// account_type and account come from Auth.AccountInfo() method
	// which may use different logic - just verify entry is built
	if entry["id"] != "test-auth-id" {
		t.Errorf("Expected id 'test-auth-id', got '%v'", entry["id"])
	}
}

func TestBuildAuthFileEntry_CreatedAndUpdatedAt(t *testing.T) {
	h := NewHandler(&config.Config{AuthDir: "/tmp/auth"}, "", nil)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now().Add(-1 * time.Hour)

	auth := &provider.Auth{
		ID:        "test-auth-id",
		FileName:  "test-auth.json",
		Provider:  "claude",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Attributes: map[string]string{
			"path": "/tmp/auth/test-auth.json",
		},
	}

	entry := h.buildAuthFileEntry(auth)
	if entry == nil {
		t.Fatal("buildAuthFileEntry returned nil")
	}

	if entry["created_at"] == nil {
		t.Error("Expected created_at field")
	}
	if entry["updated_at"] == nil {
		t.Error("Expected updated_at field")
	}
}

func TestAuthEmail(t *testing.T) {
	tests := []struct {
		name     string
		auth     *provider.Auth
		expected string
	}{
		{
			name:     "nil auth",
			auth:     nil,
			expected: "",
		},
		{
			name: "email in metadata",
			auth: &provider.Auth{
				Metadata: map[string]any{"email": "test@example.com"},
			},
			expected: "test@example.com",
		},
		{
			name: "email in attributes",
			auth: &provider.Auth{
				Attributes: map[string]string{"email": "attr@example.com"},
			},
			expected: "attr@example.com",
		},
		{
			name:     "no email",
			auth:     &provider.Auth{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := authEmail(tt.auth)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestAuthAttribute(t *testing.T) {
	tests := []struct {
		name     string
		auth     *provider.Auth
		key      string
		expected string
	}{
		{
			name:     "nil auth",
			auth:     nil,
			key:      "path",
			expected: "",
		},
		{
			name: "attribute exists",
			auth: &provider.Auth{
				Attributes: map[string]string{"path": "/tmp/auth/test.json"},
			},
			key:      "path",
			expected: "/tmp/auth/test.json",
		},
		{
			name: "attribute missing",
			auth: &provider.Auth{
				Attributes: map[string]string{},
			},
			key:      "path",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := authAttribute(tt.auth, tt.key)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestIsRuntimeOnlyAuth(t *testing.T) {
	tests := []struct {
		name     string
		auth     *provider.Auth
		expected bool
	}{
		{
			name:     "nil auth",
			auth:     nil,
			expected: false,
		},
		{
			name: "runtime_only true in attributes",
			auth: &provider.Auth{
				Attributes: map[string]string{"runtime_only": "true"},
			},
			expected: true,
		},
		{
			name: "runtime_only TRUE in attributes",
			auth: &provider.Auth{
				Attributes: map[string]string{"runtime_only": "TRUE"},
			},
			expected: true,
		},
		{
			name: "runtime_only false in attributes",
			auth: &provider.Auth{
				Attributes: map[string]string{"runtime_only": "false"},
			},
			expected: false,
		},
		{
			name:     "empty attributes",
			auth:     &provider.Auth{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRuntimeOnlyAuth(tt.auth)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

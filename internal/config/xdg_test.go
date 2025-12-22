package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCredentialsDir_XDGConfigHome(t *testing.T) {
	// Save original env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home dir: %v", err)
	}

	tests := []struct {
		name       string
		xdgEnv     string
		setXDG     bool
		wantPrefix string
	}{
		{
			name:       "XDG set - uses XDG path",
			xdgEnv:     "/custom/config",
			setXDG:     true,
			wantPrefix: "/custom/config/llm-mux",
		},
		{
			name:       "XDG not set - falls back to ~/.config",
			xdgEnv:     "",
			setXDG:     false,
			wantPrefix: filepath.Join(home, ".config", "llm-mux"),
		},
		{
			name:       "XDG empty - falls back to ~/.config",
			xdgEnv:     "",
			setXDG:     true,
			wantPrefix: filepath.Join(home, ".config", "llm-mux"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setXDG {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgEnv)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			got := CredentialsDir()

			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("CredentialsDir() = %q, want prefix %q", got, tt.wantPrefix)
			}
		})
	}
}

func TestCredentialsFilePath_XDGConfigHome(t *testing.T) {
	// Save original env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	tests := []struct {
		name        string
		xdgEnv      string
		setXDG      bool
		wantContain string
	}{
		{
			name:        "XDG set - path contains XDG dir",
			xdgEnv:      "/custom/config",
			setXDG:      true,
			wantContain: "/custom/config/llm-mux/credentials.json",
		},
		{
			name:        "XDG not set - path contains .config",
			xdgEnv:      "",
			setXDG:      false,
			wantContain: ".config/llm-mux/credentials.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setXDG {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgEnv)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			got := CredentialsFilePath()

			// Normalize for comparison
			normalizedGot := filepath.ToSlash(got)
			if !strings.Contains(normalizedGot, tt.wantContain) {
				t.Errorf("CredentialsFilePath() = %q, want to contain %q", got, tt.wantContain)
			}

			// Should end with credentials.json
			if !strings.HasSuffix(got, CredentialsFileName) {
				t.Errorf("CredentialsFilePath() = %q, should end with %q", got, CredentialsFileName)
			}
		})
	}
}

func TestCredentialsDir_PathSeparators(t *testing.T) {
	// Save original env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)
	os.Unsetenv("XDG_CONFIG_HOME")

	got := CredentialsDir()

	// Check for correct OS separators
	if runtime.GOOS == "windows" {
		// On Windows, should use backslashes (but filepath.Join handles this)
		// Just verify it doesn't have Unix-style absolute path
		if strings.HasPrefix(got, "/") && !strings.HasPrefix(got, "//") {
			t.Errorf("CredentialsDir() = %q, looks like Unix path on Windows", got)
		}
	} else {
		// On Unix, should not have backslashes
		if strings.Contains(got, "\\") {
			t.Errorf("CredentialsDir() = %q, contains backslashes on Unix", got)
		}
	}
}

func TestCredentialsDir_WindowsStyleXDG(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Test with Windows-style path
	os.Setenv("XDG_CONFIG_HOME", "C:\\Users\\TestUser\\.config")

	got := CredentialsDir()

	if !strings.HasPrefix(got, "C:\\Users\\TestUser\\.config\\llm-mux") {
		t.Errorf("CredentialsDir() = %q, want Windows-style path", got)
	}
}

func TestCredentialsDir_PathWithSpaces(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Test with spaces in path
	os.Setenv("XDG_CONFIG_HOME", "/path with spaces/config")

	got := CredentialsDir()

	if !strings.Contains(got, "path with spaces") {
		t.Errorf("CredentialsDir() = %q, should preserve spaces", got)
	}
}

func TestNewDefaultConfig_AuthDir(t *testing.T) {
	cfg := NewDefaultConfig()

	// Should use XDG variable
	if !strings.HasPrefix(cfg.AuthDir, "$XDG_CONFIG_HOME") {
		t.Errorf("NewDefaultConfig().AuthDir = %q, want $XDG_CONFIG_HOME prefix", cfg.AuthDir)
	}

	// Should contain llm-mux/auth
	if !strings.Contains(cfg.AuthDir, "llm-mux") || !strings.Contains(cfg.AuthDir, "auth") {
		t.Errorf("NewDefaultConfig().AuthDir = %q, should contain llm-mux/auth", cfg.AuthDir)
	}
}

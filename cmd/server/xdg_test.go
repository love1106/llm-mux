package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExpandPath_XDGConfigHome(t *testing.T) {
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
		input      string
		wantPrefix string
	}{
		{
			name:       "XDG set - uses XDG path",
			xdgEnv:     "/custom/config",
			setXDG:     true,
			input:      "$XDG_CONFIG_HOME/llm-mux/config.yaml",
			wantPrefix: "/custom/config",
		},
		{
			name:       "XDG not set - falls back to ~/.config",
			xdgEnv:     "",
			setXDG:     false,
			input:      "$XDG_CONFIG_HOME/llm-mux/config.yaml",
			wantPrefix: filepath.Join(home, ".config"),
		},
		{
			name:       "XDG empty - falls back to ~/.config",
			xdgEnv:     "",
			setXDG:     true,
			input:      "$XDG_CONFIG_HOME/llm-mux/config.yaml",
			wantPrefix: filepath.Join(home, ".config"),
		},
		{
			name:       "Legacy tilde path still works",
			xdgEnv:     "/custom/config",
			setXDG:     true,
			input:      "~/.config/llm-mux/config.yaml",
			wantPrefix: home,
		},
		{
			name:       "Absolute path unchanged",
			xdgEnv:     "/custom/config",
			setXDG:     true,
			input:      "/absolute/path/config.yaml",
			wantPrefix: "/absolute",
		},
		{
			name:       "Relative path unchanged",
			xdgEnv:     "/custom/config",
			setXDG:     true,
			input:      "relative/path/config.yaml",
			wantPrefix: "relative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setXDG {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgEnv)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			got := expandPath(tt.input)

			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("expandPath(%q) = %q, want prefix %q", tt.input, got, tt.wantPrefix)
			}
		})
	}
}

func TestExpandPath_PathSeparators(t *testing.T) {
	// Save original env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)
	os.Unsetenv("XDG_CONFIG_HOME")

	tests := []struct {
		name  string
		input string
	}{
		{"forward slashes", "$XDG_CONFIG_HOME/llm-mux/config.yaml"},
		{"backslashes", "$XDG_CONFIG_HOME\\llm-mux\\config.yaml"},
		{"mixed slashes", "$XDG_CONFIG_HOME/llm-mux\\config.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.input)

			// Result should use OS-native separators
			if runtime.GOOS == "windows" {
				if strings.Contains(got, "/") {
					t.Errorf("expandPath(%q) = %q, contains forward slashes on Windows", tt.input, got)
				}
			} else {
				if strings.Contains(got, "\\") {
					t.Errorf("expandPath(%q) = %q, contains backslashes on Unix", tt.input, got)
				}
			}
		})
	}
}

func TestExpandPath_XDGWithTrailingSlash(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", "/custom/config/")

	got := expandPath("$XDG_CONFIG_HOME/llm-mux/config.yaml")

	// Should not have double slashes
	if strings.Contains(got, "//") {
		t.Errorf("expandPath result %q contains double slashes", got)
	}
}

func TestExpandPath_XDGWithSpaces(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", "/path with spaces/config")

	got := expandPath("$XDG_CONFIG_HOME/llm-mux/config.yaml")

	if !strings.Contains(got, "path with spaces") {
		t.Errorf("expandPath = %q, should preserve spaces in path", got)
	}
}

func TestExpandPath_XDGOnly(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", "/custom/config")

	got := expandPath("$XDG_CONFIG_HOME")

	expected := filepath.Clean("/custom/config")
	if got != expected {
		t.Errorf("expandPath(\"$XDG_CONFIG_HOME\") = %q, want %q", got, expected)
	}
}

func TestExpandPath_TildeOnly(t *testing.T) {
	// Note: expandPath only handles "~/" not "~" alone
	// This tests the current behavior
	got := expandPath("~")
	if got != "~" {
		t.Errorf("expandPath(\"~\") = %q, want \"~\" (unchanged)", got)
	}
}

func TestExpandPath_TildeSlash(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home dir: %v", err)
	}

	got := expandPath("~/test")

	expected := filepath.Join(home, "test")
	if got != expected {
		t.Errorf("expandPath(\"~/test\") = %q, want %q", got, expected)
	}
}

func TestDefaultConfigPath_UsesXDG(t *testing.T) {
	if !strings.HasPrefix(DefaultConfigPath, "$XDG_CONFIG_HOME") {
		t.Errorf("DefaultConfigPath = %q, should start with $XDG_CONFIG_HOME", DefaultConfigPath)
	}
}

func TestExpandPath_DefaultConfigPath(t *testing.T) {
	// Save original env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home dir: %v", err)
	}

	// Test with XDG not set
	os.Unsetenv("XDG_CONFIG_HOME")
	got := expandPath(DefaultConfigPath)
	expected := filepath.Join(home, ".config", "llm-mux", "config.yaml")
	if got != expected {
		t.Errorf("expandPath(DefaultConfigPath) without XDG = %q, want %q", got, expected)
	}

	// Test with XDG set
	os.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	got = expandPath(DefaultConfigPath)
	expected = filepath.Clean("/custom/xdg/llm-mux/config.yaml")
	if got != expected {
		t.Errorf("expandPath(DefaultConfigPath) with XDG = %q, want %q", got, expected)
	}
}

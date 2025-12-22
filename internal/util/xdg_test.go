package util

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveAuthDir_XDGConfigHome(t *testing.T) {
	// Save original env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home dir: %v", err)
	}

	tests := []struct {
		name        string
		xdgEnv      string // empty string means unset
		setXDG      bool   // whether to set the env var
		input       string
		wantContain string // substring that should be in result
		wantPrefix  string // prefix that result should start with
	}{
		{
			name:        "XDG set - uses XDG path",
			xdgEnv:      "/custom/config",
			setXDG:      true,
			input:       "$XDG_CONFIG_HOME/llm-mux/auth",
			wantContain: "custom",
			wantPrefix:  "/custom/config",
		},
		{
			name:        "XDG not set - falls back to ~/.config",
			xdgEnv:      "",
			setXDG:      false,
			input:       "$XDG_CONFIG_HOME/llm-mux/auth",
			wantContain: ".config",
			wantPrefix:  filepath.Join(home, ".config"),
		},
		{
			name:        "XDG empty string - falls back to ~/.config",
			xdgEnv:      "",
			setXDG:      true,
			input:       "$XDG_CONFIG_HOME/llm-mux/auth",
			wantContain: ".config",
			wantPrefix:  filepath.Join(home, ".config"),
		},
		{
			name:        "Legacy tilde path still works",
			xdgEnv:      "/custom/config",
			setXDG:      true,
			input:       "~/.config/llm-mux/auth",
			wantContain: ".config",
			wantPrefix:  home,
		},
		{
			name:        "Absolute path unchanged",
			xdgEnv:      "/custom/config",
			setXDG:      true,
			input:       "/absolute/path/to/auth",
			wantContain: "absolute",
			wantPrefix:  "/absolute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setXDG {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgEnv)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			got, err := ResolveAuthDir(tt.input)
			if err != nil {
				t.Fatalf("ResolveAuthDir(%q) error = %v", tt.input, err)
			}

			if !strings.Contains(got, tt.wantContain) {
				t.Errorf("ResolveAuthDir(%q) = %q, want to contain %q", tt.input, got, tt.wantContain)
			}

			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("ResolveAuthDir(%q) = %q, want prefix %q", tt.input, got, tt.wantPrefix)
			}
		})
	}
}

func TestResolveAuthDir_PathSeparators(t *testing.T) {
	// Save original env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)
	os.Unsetenv("XDG_CONFIG_HOME")

	tests := []struct {
		name  string
		input string
	}{
		{"forward slashes", "$XDG_CONFIG_HOME/llm-mux/auth"},
		{"backslashes", "$XDG_CONFIG_HOME\\llm-mux\\auth"},
		{"mixed slashes", "$XDG_CONFIG_HOME/llm-mux\\auth"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveAuthDir(tt.input)
			if err != nil {
				t.Fatalf("ResolveAuthDir(%q) error = %v", tt.input, err)
			}

			// Result should use OS-native separators
			if runtime.GOOS == "windows" {
				if strings.Contains(got, "/") {
					t.Errorf("ResolveAuthDir(%q) = %q, contains forward slashes on Windows", tt.input, got)
				}
			} else {
				if strings.Contains(got, "\\") {
					t.Errorf("ResolveAuthDir(%q) = %q, contains backslashes on Unix", tt.input, got)
				}
			}

			// Should contain llm-mux and auth
			if !strings.Contains(got, "llm-mux") {
				t.Errorf("ResolveAuthDir(%q) = %q, missing 'llm-mux'", tt.input, got)
			}
			if !strings.Contains(got, "auth") {
				t.Errorf("ResolveAuthDir(%q) = %q, missing 'auth'", tt.input, got)
			}
		})
	}
}

func TestResolveAuthDir_EmptyInput(t *testing.T) {
	got, err := ResolveAuthDir("")
	if err != nil {
		t.Fatalf("ResolveAuthDir(\"\") error = %v", err)
	}
	if got != "" {
		t.Errorf("ResolveAuthDir(\"\") = %q, want empty string", got)
	}
}

func TestResolveAuthDir_XDGWithTrailingSlash(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Test with trailing slash in XDG_CONFIG_HOME
	os.Setenv("XDG_CONFIG_HOME", "/custom/config/")

	got, err := ResolveAuthDir("$XDG_CONFIG_HOME/llm-mux/auth")
	if err != nil {
		t.Fatalf("ResolveAuthDir error = %v", err)
	}

	// Should not have double slashes
	if strings.Contains(got, "//") {
		t.Errorf("ResolveAuthDir result %q contains double slashes", got)
	}

	// Should be properly cleaned
	expected := filepath.Clean("/custom/config/llm-mux/auth")
	if got != expected {
		t.Errorf("ResolveAuthDir = %q, want %q", got, expected)
	}
}

func TestResolveAuthDir_XDGWithSpaces(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Test with spaces in path (common on Windows)
	os.Setenv("XDG_CONFIG_HOME", "/path with spaces/config")

	got, err := ResolveAuthDir("$XDG_CONFIG_HOME/llm-mux/auth")
	if err != nil {
		t.Fatalf("ResolveAuthDir error = %v", err)
	}

	if !strings.Contains(got, "path with spaces") {
		t.Errorf("ResolveAuthDir = %q, should preserve spaces in path", got)
	}
}

func TestResolveAuthDir_TildeOnly(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home dir: %v", err)
	}

	got, err := ResolveAuthDir("~")
	if err != nil {
		t.Fatalf("ResolveAuthDir(\"~\") error = %v", err)
	}

	if got != filepath.Clean(home) {
		t.Errorf("ResolveAuthDir(\"~\") = %q, want %q", got, filepath.Clean(home))
	}
}

func TestResolveAuthDir_XDGOnly(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", "/custom/config")

	got, err := ResolveAuthDir("$XDG_CONFIG_HOME")
	if err != nil {
		t.Fatalf("ResolveAuthDir(\"$XDG_CONFIG_HOME\") error = %v", err)
	}

	expected := filepath.Clean("/custom/config")
	if got != expected {
		t.Errorf("ResolveAuthDir(\"$XDG_CONFIG_HOME\") = %q, want %q", got, expected)
	}
}

func TestResolveAuthDir_RelativePath(t *testing.T) {
	got, err := ResolveAuthDir("relative/path/auth")
	if err != nil {
		t.Fatalf("ResolveAuthDir error = %v", err)
	}

	expected := filepath.Clean("relative/path/auth")
	if got != expected {
		t.Errorf("ResolveAuthDir = %q, want %q", got, expected)
	}
}

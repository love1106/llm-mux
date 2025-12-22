// Package util provides utility functions for the CLI Proxy API server.
// It includes helper functions for logging configuration, file system operations,
// and other common utilities used throughout the application.
package util

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/nghyane/llm-mux/internal/config"
	log "github.com/sirupsen/logrus"
)

// SetLogLevel configures the logrus log level based on the configuration.
// It sets the log level to DebugLevel if debug mode is enabled, otherwise to InfoLevel.
func SetLogLevel(cfg *config.Config) {
	currentLevel := log.GetLevel()
	var newLevel log.Level
	if cfg.Debug {
		newLevel = log.DebugLevel
	} else {
		newLevel = log.InfoLevel
	}

	if currentLevel != newLevel {
		log.SetLevel(newLevel)
		log.Infof("log level changed from %s to %s (debug=%t)", currentLevel, newLevel, cfg.Debug)
	}
}

// ResolveAuthDir normalizes the auth directory path for consistent reuse throughout the app.
// It handles:
//   - "$XDG_CONFIG_HOME/..." -> expands XDG_CONFIG_HOME env var
//   - "~..." -> expands to user's home directory
//   - Returns a cleaned absolute path
func ResolveAuthDir(authDir string) (string, error) {
	if authDir == "" {
		return "", nil
	}

	// Handle $XDG_CONFIG_HOME prefix
	if strings.HasPrefix(authDir, "$XDG_CONFIG_HOME") {
		xdg := os.Getenv("XDG_CONFIG_HOME")
		if xdg == "" {
			// Fallback to ~/.config if XDG_CONFIG_HOME not set
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolve auth dir: %w", err)
			}
			xdg = filepath.Join(home, ".config")
		}
		remainder := strings.TrimPrefix(authDir, "$XDG_CONFIG_HOME")
		remainder = strings.TrimLeft(remainder, "/\\")
		if remainder == "" {
			return filepath.Clean(xdg), nil
		}
		normalized := strings.ReplaceAll(remainder, "\\", "/")
		return filepath.Clean(filepath.Join(xdg, filepath.FromSlash(normalized))), nil
	}

	// Handle ~ prefix (legacy support)
	if strings.HasPrefix(authDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve auth dir: %w", err)
		}
		remainder := strings.TrimPrefix(authDir, "~")
		remainder = strings.TrimLeft(remainder, "/\\")
		if remainder == "" {
			return filepath.Clean(home), nil
		}
		normalized := strings.ReplaceAll(remainder, "\\", "/")
		return filepath.Clean(filepath.Join(home, filepath.FromSlash(normalized))), nil
	}
	return filepath.Clean(authDir), nil
}

// CountAuthFiles returns the number of JSON auth files located under the provided directory.
// The function resolves leading tildes to the user's home directory and performs a case-insensitive
// match on the ".json" suffix so that files saved with uppercase extensions are also counted.
func CountAuthFiles(authDir string) int {
	dir, err := ResolveAuthDir(authDir)
	if err != nil {
		log.Debugf("countAuthFiles: failed to resolve auth directory: %v", err)
		return 0
	}
	if dir == "" {
		return 0
	}
	count := 0
	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Debugf("countAuthFiles: error accessing %s: %v", path, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			count++
		}
		return nil
	})
	if walkErr != nil {
		log.Debugf("countAuthFiles: walk error: %v", walkErr)
	}
	return count
}

// WritablePath returns the cleaned WRITABLE_PATH environment variable when it is set.
// It accepts both uppercase and lowercase variants for compatibility with existing conventions.
func WritablePath() string {
	for _, key := range []string{"WRITABLE_PATH", "writable_path"} {
		if value, ok := os.LookupEnv(key); ok {
			trimmed := strings.TrimSpace(value)
			if trimmed != "" {
				return filepath.Clean(trimmed)
			}
		}
	}
	return ""
}

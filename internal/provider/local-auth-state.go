// Package provider implements local auth state persistence for file-based deployments.
// When no external store (postgres/object/git) is configured, runtime state like
// disabled/enabled is lost on restart. This file persists that state to a local JSON file.
package provider

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/nghyane/llm-mux/internal/config"
	"github.com/nghyane/llm-mux/internal/json"
	log "github.com/nghyane/llm-mux/internal/logging"
)

const authStateFileName = "auth-state.json"

// localAuthState holds persisted auth runtime state.
type localAuthState struct {
	Disabled []string `json:"disabled"`
}

var localStateMu sync.Mutex

// localAuthStatePath returns the path to the local auth state file.
func localAuthStatePath() string {
	dir := config.CredentialsDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, authStateFileName)
}

// loadLocalAuthState reads the disabled auth IDs from the local state file.
func loadLocalAuthState() *localAuthState {
	path := localAuthStatePath()
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var state localAuthState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Warnf("local-auth-state: failed to parse %s: %v", path, err)
		return nil
	}
	return &state
}

// saveLocalAuthState writes the disabled auth IDs to the local state file.
func saveLocalAuthState(state *localAuthState) error {
	path := localAuthStatePath()
	if path == "" {
		return nil
	}
	localStateMu.Lock()
	defer localStateMu.Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// persistDisabledState saves the current set of disabled auth IDs from the manager.
// Always persists to local file as a fallback, since the primary store may skip
// saving for auths with nil Metadata (e.g., file-based token entries).
func (m *Manager) persistDisabledState() {
	m.mu.RLock()
	var disabled []string
	for id, auth := range m.auths {
		if auth != nil && auth.Disabled {
			disabled = append(disabled, id)
		}
	}
	m.mu.RUnlock()

	newState := &localAuthState{Disabled: disabled}
	if err := saveLocalAuthState(newState); err != nil {
		log.Warnf("local-auth-state: failed to save: %v", err)
		return
	}
	// Update cache so subsequent Register calls see fresh state.
	m.localStateMu.Lock()
	m.localStateCache = newState
	m.localStateMu.Unlock()
}

// applyLocalAuthState applies persisted disabled state to a single auth.
// Uses a cached snapshot to avoid re-reading the file on every Register call.
func (m *Manager) applyLocalAuthState(auth *Auth) {
	if auth == nil {
		return
	}
	m.localStateMu.RLock()
	state := m.localStateCache
	m.localStateMu.RUnlock()
	if state == nil {
		// First call: load from disk and cache.
		state = loadLocalAuthState()
		m.localStateMu.Lock()
		m.localStateCache = state
		m.localStateMu.Unlock()
	}
	if state == nil {
		return
	}
	for _, id := range state.Disabled {
		if id == auth.ID {
			auth.Disabled = true
			auth.Status = StatusDisabled
			auth.StatusMessage = "disabled via management API"
			return
		}
	}
}

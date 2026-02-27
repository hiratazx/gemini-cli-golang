// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings holds user preferences persisted to disk.
type Settings struct {
	// DefaultModel is the model used when no --model flag is specified.
	DefaultModel string `json:"default_model,omitempty"`
	// Theme can be "dark" or "light".
	Theme string `json:"theme,omitempty"`
	// AutoApproveTools auto-approves tool executions.
	AutoApproveTools bool `json:"auto_approve_tools,omitempty"`
	// SaveHistory enables conversation history saving.
	SaveHistory bool `json:"save_history,omitempty"`
	// MCPServers configures MCP server connections.
	MCPServers map[string]MCPServerConfig `json:"mcp_servers,omitempty"`
}

// DefaultSettings returns settings with sensible defaults.
func DefaultSettings() Settings {
	return Settings{
		DefaultModel: DefaultGeminiModel,
		Theme:        "dark",
		SaveHistory:  true,
	}
}

// settingsDir returns the path to the settings directory.
func settingsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".gemini")
}

// settingsFile returns the path to the settings file.
func settingsFile() string {
	dir := settingsDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "settings.json")
}

// LoadSettings loads settings from ~/.gemini/settings.json.
// Returns defaults if file doesn't exist.
func LoadSettings() Settings {
	settings := DefaultSettings()

	path := settingsFile()
	if path == "" {
		return settings
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return settings
	}

	// Unmarshal into defaults (preserves defaults for missing keys)
	json.Unmarshal(data, &settings)
	return settings
}

// SaveSettings saves settings to ~/.gemini/settings.json.
func SaveSettings(settings Settings) error {
	dir := settingsDir()
	if dir == "" {
		return nil // silently skip if no home dir
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsFile(), data, 0o644)
}

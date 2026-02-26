// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package history provides conversation history persistence.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Message represents a single message in a conversation.
type Message struct {
	Role      string    `json:"role"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// Session represents a single conversation session.
type Session struct {
	ID        string    `json:"id"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Messages  []Message `json:"messages"`
}

// historyDir returns the path to the history directory.
func historyDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".gemini", "history")
}

// NewSession creates a new conversation session.
func NewSession(model string) *Session {
	now := time.Now()
	return &Session{
		ID:        fmt.Sprintf("%d", now.UnixNano()),
		Model:     model,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddMessage adds a message to the session.
func (s *Session) AddMessage(role, text string) {
	s.Messages = append(s.Messages, Message{
		Role:      role,
		Text:      text,
		Timestamp: time.Now(),
	})
	s.UpdatedAt = time.Now()
}

// Save persists the session to disk.
func (s *Session) Save() error {
	dir := historyDir()
	if dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(dir, s.ID+".json")
	return os.WriteFile(filename, data, 0o644)
}

// SessionSummary is a lightweight view of a session for listing.
type SessionSummary struct {
	ID        string    `json:"id"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Preview   string    `json:"preview"`
	MsgCount  int       `json:"msg_count"`
}

// ListSessions returns recent session summaries (newest first).
func ListSessions(limit int) ([]SessionSummary, error) {
	dir := historyDir()
	if dir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var summaries []SessionSummary
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		preview := ""
		if len(session.Messages) > 0 {
			preview = session.Messages[0].Text
			if len(preview) > 60 {
				preview = preview[:57] + "..."
			}
		}

		summaries = append(summaries, SessionSummary{
			ID:        session.ID,
			Model:     session.Model,
			CreatedAt: session.CreatedAt,
			Preview:   preview,
			MsgCount:  len(session.Messages),
		})
	}

	// Sort newest first
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].CreatedAt.After(summaries[j].CreatedAt)
	})

	if limit > 0 && len(summaries) > limit {
		summaries = summaries[:limit]
	}

	return summaries, nil
}

// LoadSession loads a session by ID.
func LoadSession(id string) (*Session, error) {
	dir := historyDir()
	if dir == "" {
		return nil, fmt.Errorf("no history directory")
	}

	data, err := os.ReadFile(filepath.Join(dir, id+".json"))
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

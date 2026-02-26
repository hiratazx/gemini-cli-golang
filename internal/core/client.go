// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/google-gemini/gemini-cli/internal/config"
)

// MaxTurns is the maximum number of turns in a single conversation sequence.
const MaxTurns = 100

// GeminiClient is the main orchestrator for Gemini API interactions.
// It manages the chat session, tool execution, and conversation flow.
type GeminiClient struct {
	mu sync.Mutex

	config           *config.Config
	chat             *GeminiChat
	contentGenerator ContentGenerator
	sessionTurnCount int

	// currentSequenceModel tracks the model used in the current multi-turn sequence.
	currentSequenceModel string
}

// NewGeminiClient creates a new GeminiClient.
func NewGeminiClient(cfg *config.Config) *GeminiClient {
	return &GeminiClient{
		config: cfg,
	}
}

// Initialize sets up the chat session with the content generator.
func (c *GeminiClient) Initialize(ctx context.Context, generator ContentGenerator) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.contentGenerator = generator

	chat, err := c.startChat(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize chat: %w", err)
	}
	c.chat = chat
	return nil
}

// IsInitialized returns whether the client has been initialized.
func (c *GeminiClient) IsInitialized() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.chat != nil
}

// GetChat returns the current chat session.
func (c *GeminiClient) GetChat() *GeminiChat {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.chat
}

// GetHistory returns the conversation history.
func (c *GeminiClient) GetHistory() []Content {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.chat == nil {
		return nil
	}
	return c.chat.GetHistory()
}

// SetHistory replaces the conversation history.
func (c *GeminiClient) SetHistory(history []Content) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.chat != nil {
		c.chat.SetHistory(history)
	}
}

// AddHistory appends content to the conversation history.
func (c *GeminiClient) AddHistory(content Content) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.chat != nil {
		c.chat.AddHistory(content)
	}
}

// ResetChat creates a new chat session, discarding the current history.
func (c *GeminiClient) ResetChat(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	chat, err := c.startChat(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to reset chat: %w", err)
	}
	c.chat = chat
	c.currentSequenceModel = ""
	return nil
}

// ResumeChat resumes a chat with existing history.
func (c *GeminiClient) ResumeChat(ctx context.Context, history []Content) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	chat, err := c.startChat(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to resume chat: %w", err)
	}
	c.chat = chat
	return nil
}

// UpdateSystemInstruction refreshes the system instruction based on current config.
func (c *GeminiClient) UpdateSystemInstruction() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.chat == nil {
		return
	}

	userMemory := c.config.GetUserMemory()
	systemInstruction := GetCoreSystemPrompt(c.config, userMemory)
	c.chat.SetSystemInstruction(systemInstruction)
}

// GetCurrentSequenceModel returns the model being used in the current sequence.
func (c *GeminiClient) GetCurrentSequenceModel() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentSequenceModel
}

// SendMessage sends a user message and returns a channel of streaming events.
func (c *GeminiClient) SendMessage(ctx context.Context, message string, promptID string) <-chan ServerGeminiStreamEvent {
	return c.SendMessageParts(ctx, []Part{{Text: message}}, promptID, MaxTurns)
}

// SendMessageParts sends user message parts and returns a channel of streaming events.
func (c *GeminiClient) SendMessageParts(ctx context.Context, parts []Part, promptID string, maxTurns int) <-chan ServerGeminiStreamEvent {
	events := make(chan ServerGeminiStreamEvent, 100)

	go func() {
		defer close(events)

		c.mu.Lock()
		if c.chat == nil {
			c.mu.Unlock()
			events <- ServerGeminiStreamEvent{
				Type:  GeminiEventError,
				Value: "chat not initialized",
			}
			return
		}

		c.config.ResetTurn()
		c.sessionTurnCount++

		if c.config.GetMaxSessionTurns() > 0 && c.sessionTurnCount > c.config.GetMaxSessionTurns() {
			c.mu.Unlock()
			events <- ServerGeminiStreamEvent{Type: GeminiEventMaxSessionTurns}
			return
		}

		// Determine model
		modelToUse := c.currentSequenceModel
		if modelToUse == "" {
			modelToUse = config.ResolveModel(c.config.GetActiveModel(), false, false)
		}

		if c.currentSequenceModel == "" {
			events <- ServerGeminiStreamEvent{
				Type:  GeminiEventModelInfo,
				Value: modelToUse,
			}
		}
		c.currentSequenceModel = modelToUse
		chat := c.chat
		c.mu.Unlock()

		// Create and run the turn
		turn := NewTurn(chat, promptID)
		systemInstruction := chat.GetSystemInstruction()

		turnEvents := turn.Run(modelToUse, parts, systemInstruction)
		for event := range turnEvents {
			events <- event
		}
	}()

	return events
}

// Dispose cleans up resources.
func (c *GeminiClient) Dispose() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.chat = nil
	c.contentGenerator = nil
}

// startChat creates a new GeminiChat with the given history.
func (c *GeminiClient) startChat(ctx context.Context, extraHistory []Content) (*GeminiChat, error) {
	if c.contentGenerator == nil {
		return nil, fmt.Errorf("content generator not initialized")
	}

	userMemory := c.config.GetUserMemory()
	systemInstruction := GetCoreSystemPrompt(c.config, userMemory)

	history := make([]Content, 0)
	if extraHistory != nil {
		history = append(history, extraHistory...)
	}

	return NewGeminiChat(
		ctx,
		c.contentGenerator,
		systemInstruction,
		nil, // tools will be set separately
		history,
	), nil
}

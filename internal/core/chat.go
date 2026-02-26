// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"sync"
)

// GeminiChat manages a conversation session with the Gemini API.
type GeminiChat struct {
	mu sync.Mutex

	ctx              context.Context
	contentGenerator ContentGenerator
	systemInstruction string
	tools            []Tool
	history          []Content
	promptTokenCount int
}

// NewGeminiChat creates a new chat session.
func NewGeminiChat(
	ctx context.Context,
	contentGenerator ContentGenerator,
	systemInstruction string,
	tools []Tool,
	history []Content,
) *GeminiChat {
	if history == nil {
		history = []Content{}
	}
	return &GeminiChat{
		ctx:              ctx,
		contentGenerator: contentGenerator,
		systemInstruction: systemInstruction,
		tools:            tools,
		history:          history,
	}
}

// GetHistory returns the current conversation history.
func (c *GeminiChat) GetHistory() []Content {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]Content, len(c.history))
	copy(result, c.history)
	return result
}

// SetHistory replaces the conversation history.
func (c *GeminiChat) SetHistory(history []Content) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.history = make([]Content, len(history))
	copy(c.history, history)
}

// AddHistory appends content to the conversation history.
func (c *GeminiChat) AddHistory(content Content) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.history = append(c.history, content)
}

// SetSystemInstruction updates the system instruction.
func (c *GeminiChat) SetSystemInstruction(instruction string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.systemInstruction = instruction
}

// GetSystemInstruction returns the current system instruction.
func (c *GeminiChat) GetSystemInstruction() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.systemInstruction
}

// SetTools updates the tool declarations.
func (c *GeminiChat) SetTools(tools []Tool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tools = tools
}

// GetTools returns the current tool declarations.
func (c *GeminiChat) GetTools() []Tool {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]Tool, len(c.tools))
	copy(result, c.tools)
	return result
}

// GetLastPromptTokenCount returns the token count of the last prompt.
func (c *GeminiChat) GetLastPromptTokenCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.promptTokenCount
}

// SetLastPromptTokenCount updates the token count of the last prompt.
func (c *GeminiChat) SetLastPromptTokenCount(count int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.promptTokenCount = count
}

// StripThoughtsFromHistory removes thought parts from the conversation history.
func (c *GeminiChat) StripThoughtsFromHistory() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.history {
		var filteredParts []Part
		for _, p := range c.history[i].Parts {
			if p.Thought == "" {
				filteredParts = append(filteredParts, p)
			}
		}
		c.history[i].Parts = filteredParts
	}
}

// HistoryLength returns the number of messages in the conversation history.
func (c *GeminiChat) HistoryLength() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.history)
}

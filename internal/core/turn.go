// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"sync"
)

// GeminiEventType represents the types of events during a conversation turn.
type GeminiEventType int

const (
	// GeminiEventText is emitted when text content is received.
	GeminiEventText GeminiEventType = iota
	// GeminiEventThought is emitted when thinking/reasoning content is received.
	GeminiEventThought
	// GeminiEventFunctionCall is emitted when the model requests a tool call.
	GeminiEventFunctionCall
	// GeminiEventFunctionResponse is emitted when a tool result is provided.
	GeminiEventFunctionResponse
	// GeminiEventError is emitted when an error occurs.
	GeminiEventError
	// GeminiEventModelInfo is emitted with the model being used.
	GeminiEventModelInfo
	// GeminiEventLoopDetected is emitted when a response loop is detected.
	GeminiEventLoopDetected
	// GeminiEventMaxSessionTurns is emitted when max session turns is reached.
	GeminiEventMaxSessionTurns
	// GeminiEventChatCompressed is emitted when chat history is compressed.
	GeminiEventChatCompressed
	// GeminiEventContextWindowWillOverflow is emitted when context window is about to overflow.
	GeminiEventContextWindowWillOverflow
	// GeminiEventInvalidStream is emitted when the response stream is invalid.
	GeminiEventInvalidStream
	// GeminiEventAgentExecutionStopped is emitted when agent execution is stopped by a hook.
	GeminiEventAgentExecutionStopped
	// GeminiEventAgentExecutionBlocked is emitted when agent execution is blocked by a hook.
	GeminiEventAgentExecutionBlocked
	// GeminiEventDone is emitted when the turn is complete.
	GeminiEventDone
)

// ServerGeminiStreamEvent is an event emitted during conversation processing.
type ServerGeminiStreamEvent struct {
	Type  GeminiEventType
	Value interface{}
}

// CompressionStatus indicates the result of a compression attempt.
type CompressionStatus int

const (
	// CompressionStatusNone means no compression was attempted.
	CompressionStatusNone CompressionStatus = iota
	// CompressionStatusCompressed means compression was successful.
	CompressionStatusCompressed
	// CompressionStatusFailed means compression failed.
	CompressionStatusFailed
)

// ChatCompressionInfo holds information about a compression event.
type ChatCompressionInfo struct {
	CompressionStatus   CompressionStatus
	OriginalTokenCount  int
	CompressedTokenCount int
}

// PendingToolCall represents a tool call waiting to be executed.
type PendingToolCall struct {
	FunctionCall *FunctionCall
	ID           string
}

// Turn represents a single conversation turn, managing the streaming interaction.
type Turn struct {
	mu sync.Mutex

	chat     *GeminiChat
	promptID string

	// PendingToolCalls holds any tool calls that need to be executed.
	PendingToolCalls []PendingToolCall

	// responseText accumulates the text response.
	responseText string

	// finishReason from the API.
	FinishReason string
}

// NewTurn creates a new Turn instance.
func NewTurn(chat *GeminiChat, promptID string) *Turn {
	return &Turn{
		chat:     chat,
		promptID: promptID,
	}
}

// GetResponseText returns the accumulated response text.
func (t *Turn) GetResponseText() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.responseText
}

// appendText adds text to the response accumulator.
func (t *Turn) appendText(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.responseText += text
}

// Run executes the turn by sending the request through the content generator and
// processing the streaming response.
func (t *Turn) Run(model string, request []Part, systemInstruction string) <-chan ServerGeminiStreamEvent {
	events := make(chan ServerGeminiStreamEvent, 100)

	go func() {
		defer close(events)

		// Build the content from request parts
		userContent := Content{
			Role:  "user",
			Parts: request,
		}

		// Add user message to history
		t.chat.AddHistory(userContent)

		// Get history for the API call
		history := t.chat.GetHistory()

		// Stream the response
		streamCh, err := t.chat.contentGenerator.GenerateContentStream(
			t.chat.ctx,
			model,
			history,
			nil, // config
			t.chat.tools,
			systemInstruction,
		)

		if err != nil {
			events <- ServerGeminiStreamEvent{
				Type:  GeminiEventError,
				Value: err.Error(),
			}
			return
		}

		for chunk := range streamCh {
			if chunk.Error != nil {
				events <- ServerGeminiStreamEvent{
					Type:  GeminiEventError,
					Value: chunk.Error.Error(),
				}
				continue
			}

			if chunk.Response != nil {
				t.processResponse(chunk.Response, events)
			}

			if chunk.Done {
				// Add model response to history
				if t.responseText != "" {
					modelContent := CreateModelContent(t.responseText)
					t.chat.AddHistory(modelContent)
				}
				events <- ServerGeminiStreamEvent{Type: GeminiEventDone}
			}
		}
	}()

	return events
}

// processResponse handles a response chunk from the API.
func (t *Turn) processResponse(resp *GenerateContentResponse, events chan<- ServerGeminiStreamEvent) {
	if len(resp.Candidates) == 0 {
		return
	}

	candidate := resp.Candidates[0]
	if candidate.FinishReason != "" {
		t.FinishReason = candidate.FinishReason
	}

	if candidate.Content == nil {
		return
	}

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			t.appendText(part.Text)
			events <- ServerGeminiStreamEvent{
				Type:  GeminiEventText,
				Value: part.Text,
			}
		}

		if part.Thought != "" {
			events <- ServerGeminiStreamEvent{
				Type:  GeminiEventThought,
				Value: part.Thought,
			}
		}

		if part.FunctionCall != nil {
			t.mu.Lock()
			t.PendingToolCalls = append(t.PendingToolCalls, PendingToolCall{
				FunctionCall: part.FunctionCall,
				ID:           part.FunctionCall.ID,
			})
			t.mu.Unlock()

			events <- ServerGeminiStreamEvent{
				Type:  GeminiEventFunctionCall,
				Value: part.FunctionCall,
			}
		}
	}
}

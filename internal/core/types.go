// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package core provides the Gemini API client, chat, and turn orchestration.
package core

// Content represents a message in the conversation history.
type Content struct {
	// Role is "user" or "model".
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// Part represents a piece of content within a message.
type Part struct {
	// Text content (mutually exclusive with other fields).
	Text string `json:"text,omitempty"`

	// FunctionCall from the model.
	FunctionCall *FunctionCall `json:"functionCall,omitempty"`

	// FunctionResponse from tool execution.
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`

	// InlineData for binary content (e.g., images).
	InlineData *InlineData `json:"inlineData,omitempty"`

	// Thought represents model's internal reasoning.
	Thought string `json:"thought,omitempty"`
}

// FunctionCall represents a tool call from the model.
type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
	ID   string                 `json:"id,omitempty"`
}

// FunctionResponse represents the result of a tool execution.
type FunctionResponse struct {
	Name     string      `json:"name"`
	Response interface{} `json:"response"`
	ID       string      `json:"id,omitempty"`
}

// InlineData represents binary content embedded in a message.
type InlineData struct {
	MIMEType string `json:"mimeType"`
	Data     string `json:"data"` // base64 encoded
}

// FunctionDeclaration describes a tool that the model can call.
type FunctionDeclaration struct {
	Name                 string                 `json:"name"`
	Description          string                 `json:"description"`
	ParametersJSONSchema map[string]interface{} `json:"parameters,omitempty"`
}

// Tool wraps a set of function declarations for the API.
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations,omitempty"`
}

// GenerateContentConfig holds configuration for content generation.
type GenerateContentConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
	ThinkingConfig  *ThinkingConfig `json:"thinkingConfig,omitempty"`
}

// ThinkingConfig configures model thinking behavior.
type ThinkingConfig struct {
	ThinkingBudget int `json:"thinkingBudget,omitempty"`
}

// GenerateContentResponse represents a response from the Gemini API.
type GenerateContentResponse struct {
	Candidates    []Candidate    `json:"candidates,omitempty"`
	UsageMetadata *UsageMetadata `json:"usageMetadata,omitempty"`
}

// Candidate represents a single candidate response.
type Candidate struct {
	Content      *Content `json:"content,omitempty"`
	FinishReason string   `json:"finishReason,omitempty"`
}

// UsageMetadata contains token usage information.
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount,omitempty"`
	CandidatesTokenCount int `json:"candidatesTokenCount,omitempty"`
	TotalTokenCount      int `json:"totalTokenCount,omitempty"`
	ThoughtsTokenCount   int `json:"thoughtsTokenCount,omitempty"`
}

// FinishReason constants.
const (
	FinishReasonStop         = "STOP"
	FinishReasonMaxTokens    = "MAX_TOKENS"
	FinishReasonSafety       = "SAFETY"
	FinishReasonRecitation   = "RECITATION"
	FinishReasonOther        = "OTHER"
	FinishReasonMalformed    = "MALFORMED_FUNCTION_CALL"
)

// CreateUserContent creates a Content with role "user" from text.
func CreateUserContent(text string) Content {
	return Content{
		Role:  "user",
		Parts: []Part{{Text: text}},
	}
}

// CreateModelContent creates a Content with role "model" from text.
func CreateModelContent(text string) Content {
	return Content{
		Role:  "model",
		Parts: []Part{{Text: text}},
	}
}

// HasFunctionCalls checks if any part has a function call.
func (c *Content) HasFunctionCalls() bool {
	for _, p := range c.Parts {
		if p.FunctionCall != nil {
			return true
		}
	}
	return false
}

// GetText extracts the text content from the parts.
func (c *Content) GetText() string {
	var text string
	for _, p := range c.Parts {
		if p.Text != "" {
			text += p.Text
		}
	}
	return text
}

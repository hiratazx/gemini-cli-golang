// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package tools provides the tool system for Gemini CLI, including
// tool interfaces, built-in tool implementations, and the tool registry.
package tools

import (
	"context"

	"github.com/google-gemini/gemini-cli/internal/core"
)

// ToolResult represents the outcome of a tool execution.
type ToolResult struct {
	// LLMContent is the content to send back to the model.
	LLMContent string `json:"llmContent"`
	// ReturnDisplay is the content to display to the user.
	ReturnDisplay string `json:"returnDisplay,omitempty"`
	// Error contains error information if the tool failed.
	Error *ToolError `json:"error,omitempty"`
}

// ToolLocation represents a file/line location that a tool affects.
type ToolLocation struct {
	Path string `json:"path"`
	Line int    `json:"line,omitempty"`
}

// Kind represents the category of a tool operation.
type Kind int

const (
	// KindRead indicates a read-only operation.
	KindRead Kind = iota
	// KindWrite indicates a write/modify operation.
	KindWrite
	// KindExecute indicates a command execution operation.
	KindExecute
)

// ToolInvocation represents a validated and ready-to-execute tool call.
type ToolInvocation interface {
	// GetDescription returns a pre-execution description of the tool operation.
	GetDescription() string
	// ToolLocations returns the file paths the tool will affect.
	ToolLocations() []ToolLocation
	// Execute runs the tool with validated parameters.
	Execute(ctx context.Context) (*ToolResult, error)
}

// DeclarativeTool declares a tool's schema and creates invocations.
type DeclarativeTool interface {
	// Name returns the tool's identifier.
	Name() string
	// DisplayName returns the human-readable tool name.
	DisplayName() string
	// Description returns the tool's description.
	Description() string
	// Kind returns the operation category.
	Kind() Kind
	// GetSchema returns the function declaration for the API.
	GetSchema(modelID string) core.FunctionDeclaration
	// CreateInvocation creates a validated tool invocation from parameters.
	CreateInvocation(params map[string]interface{}) (ToolInvocation, error)
}

// BaseToolInvocation provides common functionality for tool invocations.
type BaseToolInvocation struct {
	Params   map[string]interface{}
	ToolName string
}

// NewBaseToolInvocation creates a new base invocation.
func NewBaseToolInvocation(params map[string]interface{}, toolName string) BaseToolInvocation {
	return BaseToolInvocation{
		Params:   params,
		ToolName: toolName,
	}
}

// GetStringParam extracts a string parameter with a default value.
func (b *BaseToolInvocation) GetStringParam(key string, defaultVal string) string {
	if v, ok := b.Params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

// GetIntParam extracts an integer parameter with a default value.
func (b *BaseToolInvocation) GetIntParam(key string, defaultVal int) int {
	if v, ok := b.Params[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case int64:
			return int(n)
		}
	}
	return defaultVal
}

// GetBoolParam extracts a boolean parameter with a default value.
func (b *BaseToolInvocation) GetBoolParam(key string, defaultVal bool) bool {
	if v, ok := b.Params[key]; ok {
		if bv, ok := v.(bool); ok {
			return bv
		}
	}
	return defaultVal
}

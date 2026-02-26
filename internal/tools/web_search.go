// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/google-gemini/gemini-cli/internal/core"
)

// WebSearchTool implements the google_web_search tool.
// This is a placeholder that leverages the Gemini API's built-in grounding.
type WebSearchTool struct{}

// NewWebSearchTool creates a new WebSearchTool.
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{}
}

func (t *WebSearchTool) Name() string        { return WebSearchToolName }
func (t *WebSearchTool) DisplayName() string  { return "WebSearch" }
func (t *WebSearchTool) Description() string  { return "Search the web using Google Search" }
func (t *WebSearchTool) Kind() Kind           { return KindRead }

func (t *WebSearchTool) GetSchema(modelID string) core.FunctionDeclaration {
	return core.FunctionDeclaration{
		Name:        WebSearchToolName,
		Description: "Search the web using Google Search. Returns relevant search results and snippets.",
		ParametersJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query.",
				},
			},
			"required": []string{"query"},
		},
	}
}

func (t *WebSearchTool) CreateInvocation(params map[string]interface{}) (ToolInvocation, error) {
	base := NewBaseToolInvocation(params, WebSearchToolName)
	query := base.GetStringParam("query", "")
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("the 'query' parameter must be non-empty")
	}

	return &webSearchInvocation{
		BaseToolInvocation: base,
		query:              query,
	}, nil
}

type webSearchInvocation struct {
	BaseToolInvocation
	query string
}

func (inv *webSearchInvocation) GetDescription() string {
	return fmt.Sprintf("Search for '%s'", inv.query)
}

func (inv *webSearchInvocation) ToolLocations() []ToolLocation {
	return nil
}

func (inv *webSearchInvocation) Execute(ctx context.Context) (*ToolResult, error) {
	// Web search is typically handled by the Gemini API's built-in grounding.
	// This tool serves as a declaration to enable the grounding capability.
	return &ToolResult{
		LLMContent:    fmt.Sprintf("Web search for '%s' is handled via Gemini API grounding.", inv.query),
		ReturnDisplay: fmt.Sprintf("Searching: %s", inv.query),
	}, nil
}

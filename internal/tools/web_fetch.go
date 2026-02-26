// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google-gemini/gemini-cli/internal/config"
	"github.com/google-gemini/gemini-cli/internal/core"
)

// WebFetchTool implements the web_fetch tool.
type WebFetchTool struct {
	config *config.Config
}

// NewWebFetchTool creates a new WebFetchTool.
func NewWebFetchTool(cfg *config.Config) *WebFetchTool {
	return &WebFetchTool{config: cfg}
}

func (t *WebFetchTool) Name() string        { return WebFetchToolName }
func (t *WebFetchTool) DisplayName() string  { return "WebFetch" }
func (t *WebFetchTool) Description() string  { return "Fetch content from a URL" }
func (t *WebFetchTool) Kind() Kind           { return KindRead }

func (t *WebFetchTool) GetSchema(modelID string) core.FunctionDeclaration {
	return core.FunctionDeclaration{
		Name:        WebFetchToolName,
		Description: "Fetch the content of a web page or API endpoint. Returns the text content of the response.",
		ParametersJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "The URL to fetch.",
				},
			},
			"required": []string{"url"},
		},
	}
}

func (t *WebFetchTool) CreateInvocation(params map[string]interface{}) (ToolInvocation, error) {
	base := NewBaseToolInvocation(params, WebFetchToolName)
	url := base.GetStringParam("url", "")
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("the 'url' parameter must be non-empty")
	}

	return &webFetchInvocation{
		BaseToolInvocation: base,
		url:                url,
	}, nil
}

type webFetchInvocation struct {
	BaseToolInvocation
	url string
}

func (inv *webFetchInvocation) GetDescription() string {
	return fmt.Sprintf("Fetch %s", inv.url)
}

func (inv *webFetchInvocation) ToolLocations() []ToolLocation {
	return nil
}

func (inv *webFetchInvocation) Execute(ctx context.Context) (*ToolResult, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", inv.url, nil)
	if err != nil {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("Invalid URL: %s", err)), nil
	}

	req.Header.Set("User-Agent", "Gemini-CLI/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/json,text/plain,*/*")

	resp, err := client.Do(req)
	if err != nil {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("Failed to fetch URL: %s", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("HTTP error %d: %s", resp.StatusCode, resp.Status)), nil
	}

	// Limit response size
	const maxBodySize = 500_000
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("Failed to read response: %s", err)), nil
	}

	content := string(body)
	contentType := resp.Header.Get("Content-Type")

	return &ToolResult{
		LLMContent:    fmt.Sprintf("URL: %s\nContent-Type: %s\nStatus: %d\n\n%s", inv.url, contentType, resp.StatusCode, content),
		ReturnDisplay: fmt.Sprintf("Fetched %s (%d bytes)", inv.url, len(content)),
	}, nil
}

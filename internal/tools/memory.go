// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google-gemini/gemini-cli/internal/config"
	"github.com/google-gemini/gemini-cli/internal/core"
)

// MemoryTool implements the save_memory tool.
type MemoryTool struct {
	config *config.Config
}

// NewMemoryTool creates a new MemoryTool.
func NewMemoryTool(cfg *config.Config) *MemoryTool {
	return &MemoryTool{config: cfg}
}

func (t *MemoryTool) Name() string        { return MemoryToolName }
func (t *MemoryTool) DisplayName() string  { return "SaveMemory" }
func (t *MemoryTool) Description() string  { return "Save important context to GEMINI.md memory file" }
func (t *MemoryTool) Kind() Kind           { return KindWrite }

func (t *MemoryTool) GetSchema(modelID string) core.FunctionDeclaration {
	return core.FunctionDeclaration{
		Name:        MemoryToolName,
		Description: "Save important context, instructions, or project conventions to the GEMINI.md memory file. This information persists across sessions.",
		ParametersJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The content to save to the memory file.",
				},
			},
			"required": []string{"content"},
		},
	}
}

func (t *MemoryTool) CreateInvocation(params map[string]interface{}) (ToolInvocation, error) {
	base := NewBaseToolInvocation(params, MemoryToolName)
	content := base.GetStringParam("content", "")
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("the 'content' parameter must be non-empty")
	}

	return &memoryInvocation{
		BaseToolInvocation: base,
		config:             t.config,
		content:            content,
	}, nil
}

type memoryInvocation struct {
	BaseToolInvocation
	config  *config.Config
	content string
}

func (inv *memoryInvocation) GetDescription() string {
	return "Save to GEMINI.md"
}

func (inv *memoryInvocation) ToolLocations() []ToolLocation {
	memoryPath := filepath.Join(inv.config.GetTargetDir(), "GEMINI.md")
	return []ToolLocation{{Path: memoryPath}}
}

func (inv *memoryInvocation) Execute(ctx context.Context) (*ToolResult, error) {
	memoryPath := filepath.Join(inv.config.GetTargetDir(), "GEMINI.md")

	// Read existing content
	existing := ""
	if data, err := os.ReadFile(memoryPath); err == nil {
		existing = string(data)
	}

	// Append new content
	var newContent string
	if existing != "" {
		newContent = existing + "\n\n" + inv.content
	} else {
		newContent = inv.content
	}

	// Write back
	if err := os.WriteFile(memoryPath, []byte(newContent), 0o644); err != nil {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("Failed to write memory file: %s", err)), nil
	}

	return &ToolResult{
		LLMContent:    fmt.Sprintf("Successfully saved content to GEMINI.md"),
		ReturnDisplay: "Saved to GEMINI.md",
	}, nil
}

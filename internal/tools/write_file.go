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

// WriteFileTool implements the write_file tool.
type WriteFileTool struct {
	config *config.Config
}

// NewWriteFileTool creates a new WriteFileTool.
func NewWriteFileTool(cfg *config.Config) *WriteFileTool {
	return &WriteFileTool{config: cfg}
}

func (t *WriteFileTool) Name() string        { return WriteFileToolName }
func (t *WriteFileTool) DisplayName() string  { return WriteFileDisplayName }
func (t *WriteFileTool) Description() string  { return "Create or overwrite a file with the given content" }
func (t *WriteFileTool) Kind() Kind           { return KindWrite }

func (t *WriteFileTool) GetSchema(modelID string) core.FunctionDeclaration {
	return core.FunctionDeclaration{
		Name:        WriteFileToolName,
		Description: "Create a new file or overwrite an existing file with the provided content. Use this for new files; for modifying existing files, prefer the replace tool.",
		ParametersJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "The path where the file should be written, relative to the working directory.",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The full content to write to the file.",
				},
			},
			"required": []string{"file_path", "content"},
		},
	}
}

func (t *WriteFileTool) CreateInvocation(params map[string]interface{}) (ToolInvocation, error) {
	base := NewBaseToolInvocation(params, WriteFileToolName)
	filePath := base.GetStringParam("file_path", "")
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("the 'file_path' parameter must be non-empty")
	}

	resolvedPath := filepath.Join(t.config.GetTargetDir(), filePath)
	if filepath.IsAbs(filePath) {
		resolvedPath = filePath
	}

	if err := t.config.ValidatePathAccess(resolvedPath, "write"); err != "" {
		return nil, fmt.Errorf("%s", err)
	}

	return &writeFileInvocation{
		BaseToolInvocation: base,
		config:             t.config,
		resolvedPath:       resolvedPath,
		content:            base.GetStringParam("content", ""),
	}, nil
}

type writeFileInvocation struct {
	BaseToolInvocation
	config       *config.Config
	resolvedPath string
	content      string
}

func (inv *writeFileInvocation) GetDescription() string {
	rel, _ := filepath.Rel(inv.config.GetTargetDir(), inv.resolvedPath)
	if rel == "" {
		rel = inv.resolvedPath
	}
	return fmt.Sprintf("Write to %s", rel)
}

func (inv *writeFileInvocation) ToolLocations() []ToolLocation {
	return []ToolLocation{{Path: inv.resolvedPath}}
}

func (inv *writeFileInvocation) Execute(ctx context.Context) (*ToolResult, error) {
	// Create parent directories
	dir := filepath.Dir(inv.resolvedPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("Failed to create directories: %s", err)), nil
	}

	// Check if file exists for reporting
	_, existErr := os.Stat(inv.resolvedPath)
	isNew := os.IsNotExist(existErr)

	// Write the file
	if err := os.WriteFile(inv.resolvedPath, []byte(inv.content), 0o644); err != nil {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("Failed to write file: %s", err)), nil
	}

	lines := strings.Count(inv.content, "\n") + 1
	action := "Updated"
	if isNew {
		action = "Created"
	}

	rel, _ := filepath.Rel(inv.config.GetTargetDir(), inv.resolvedPath)
	return &ToolResult{
		LLMContent:    fmt.Sprintf("%s file %s (%d lines)", action, rel, lines),
		ReturnDisplay: fmt.Sprintf("%s %s", action, rel),
	}, nil
}

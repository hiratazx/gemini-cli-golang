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

// EditTool implements the replace/edit tool for targeted content replacement.
type EditTool struct {
	config *config.Config
}

// NewEditTool creates a new EditTool.
func NewEditTool(cfg *config.Config) *EditTool {
	return &EditTool{config: cfg}
}

func (t *EditTool) Name() string        { return EditToolName }
func (t *EditTool) DisplayName() string  { return EditDisplayName }
func (t *EditTool) Description() string  { return "Replace specific content in a file" }
func (t *EditTool) Kind() Kind           { return KindWrite }

func (t *EditTool) GetSchema(modelID string) core.FunctionDeclaration {
	return core.FunctionDeclaration{
		Name:        EditToolName,
		Description: "Replace exact string occurrences in a file. Provide the old_string to find and new_string to replace it with. The old_string must match exactly, including whitespace and indentation.",
		ParametersJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the file to edit, relative to the working directory.",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "The exact string to find in the file. Must match exactly including whitespace.",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "The replacement string.",
				},
			},
			"required": []string{"file_path", "old_string", "new_string"},
		},
	}
}

func (t *EditTool) CreateInvocation(params map[string]interface{}) (ToolInvocation, error) {
	base := NewBaseToolInvocation(params, EditToolName)
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

	return &editInvocation{
		BaseToolInvocation: base,
		config:             t.config,
		resolvedPath:       resolvedPath,
		oldString:          base.GetStringParam("old_string", ""),
		newString:          base.GetStringParam("new_string", ""),
	}, nil
}

type editInvocation struct {
	BaseToolInvocation
	config       *config.Config
	resolvedPath string
	oldString    string
	newString    string
}

func (inv *editInvocation) GetDescription() string {
	rel, _ := filepath.Rel(inv.config.GetTargetDir(), inv.resolvedPath)
	if rel == "" {
		rel = inv.resolvedPath
	}
	return fmt.Sprintf("Edit %s", rel)
}

func (inv *editInvocation) ToolLocations() []ToolLocation {
	return []ToolLocation{{Path: inv.resolvedPath}}
}

func (inv *editInvocation) Execute(ctx context.Context) (*ToolResult, error) {
	// Read current content
	data, err := os.ReadFile(inv.resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrorResult(ToolErrorTypeFileNotFound,
				fmt.Sprintf("File not found: %s", inv.resolvedPath)), nil
		}
		return ErrorResult(ToolErrorTypeGeneral, err.Error()), nil
	}

	content := string(data)

	// Verify old_string exists
	count := strings.Count(content, inv.oldString)
	if count == 0 {
		return ErrorResult(ToolErrorTypeTargetNotFound,
			fmt.Sprintf("The old_string was not found in the file. Make sure it matches exactly, including whitespace and indentation.")), nil
	}

	if count > 1 {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("The old_string was found %d times in the file. It must be unique. Provide more context to make it unique.", count)), nil
	}

	// Perform replacement
	newContent := strings.Replace(content, inv.oldString, inv.newString, 1)

	// Write back
	if err := os.WriteFile(inv.resolvedPath, []byte(newContent), 0o644); err != nil {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("Failed to write file: %s", err)), nil
	}

	rel, _ := filepath.Rel(inv.config.GetTargetDir(), inv.resolvedPath)
	return &ToolResult{
		LLMContent:    fmt.Sprintf("Successfully edited %s", rel),
		ReturnDisplay: fmt.Sprintf("Edited %s", rel),
	}, nil
}

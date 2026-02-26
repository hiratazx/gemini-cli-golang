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

// LSTool implements the list_directory tool.
type LSTool struct {
	config *config.Config
}

// NewLSTool creates a new LSTool.
func NewLSTool(cfg *config.Config) *LSTool {
	return &LSTool{config: cfg}
}

func (t *LSTool) Name() string        { return LSToolName }
func (t *LSTool) DisplayName() string  { return "ListDirectory" }
func (t *LSTool) Description() string  { return "List the contents of a directory" }
func (t *LSTool) Kind() Kind           { return KindRead }

func (t *LSTool) GetSchema(modelID string) core.FunctionDeclaration {
	return core.FunctionDeclaration{
		Name:        LSToolName,
		Description: "List the contents of a directory, showing files and subdirectories with sizes.",
		ParametersJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The directory path to list (relative to working directory). Defaults to '.'.",
				},
			},
		},
	}
}

func (t *LSTool) CreateInvocation(params map[string]interface{}) (ToolInvocation, error) {
	base := NewBaseToolInvocation(params, LSToolName)
	dirPath := base.GetStringParam("path", ".")

	resolvedPath := filepath.Join(t.config.GetTargetDir(), dirPath)
	if filepath.IsAbs(dirPath) {
		resolvedPath = dirPath
	}

	return &lsInvocation{
		BaseToolInvocation: base,
		config:             t.config,
		resolvedPath:       resolvedPath,
	}, nil
}

type lsInvocation struct {
	BaseToolInvocation
	config       *config.Config
	resolvedPath string
}

func (inv *lsInvocation) GetDescription() string {
	rel, _ := filepath.Rel(inv.config.GetTargetDir(), inv.resolvedPath)
	if rel == "" || rel == "." {
		return "List ."
	}
	return fmt.Sprintf("List %s", rel)
}

func (inv *lsInvocation) ToolLocations() []ToolLocation {
	return []ToolLocation{{Path: inv.resolvedPath}}
}

func (inv *lsInvocation) Execute(ctx context.Context) (*ToolResult, error) {
	info, err := os.Stat(inv.resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrorResult(ToolErrorTypeFileNotFound,
				fmt.Sprintf("Directory not found: %s", inv.resolvedPath)), nil
		}
		return ErrorResult(ToolErrorTypeGeneral, err.Error()), nil
	}

	if !info.IsDir() {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("Path is a file, not a directory: %s", inv.resolvedPath)), nil
	}

	entries, err := os.ReadDir(inv.resolvedPath)
	if err != nil {
		return ErrorResult(ToolErrorTypePermissionDenied, err.Error()), nil
	}

	var result strings.Builder
	dirs := 0
	files := 0

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			result.WriteString(fmt.Sprintf("  %s/\n", entry.Name()))
			dirs++
		} else {
			result.WriteString(fmt.Sprintf("  %s (%s)\n", entry.Name(), formatSize(info.Size())))
			files++
		}
	}

	summary := fmt.Sprintf("%d files, %d directories", files, dirs)
	result.WriteString(fmt.Sprintf("\n%s\n", summary))

	return &ToolResult{
		LLMContent:    result.String(),
		ReturnDisplay: summary,
	}, nil
}

// formatSize returns a human-readable file size.
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

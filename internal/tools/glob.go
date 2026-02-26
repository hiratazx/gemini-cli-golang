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

// GlobTool implements the glob/find files tool.
type GlobTool struct {
	config *config.Config
}

// NewGlobTool creates a new GlobTool.
func NewGlobTool(cfg *config.Config) *GlobTool {
	return &GlobTool{config: cfg}
}

func (t *GlobTool) Name() string        { return GlobToolName }
func (t *GlobTool) DisplayName() string  { return GlobDisplayName }
func (t *GlobTool) Description() string  { return "Find files matching a glob pattern" }
func (t *GlobTool) Kind() Kind           { return KindRead }

func (t *GlobTool) GetSchema(modelID string) core.FunctionDeclaration {
	return core.FunctionDeclaration{
		Name:        GlobToolName,
		Description: "Find files matching a glob pattern in the working directory. Returns file paths relative to the working directory.",
		ParametersJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "The glob pattern to match files (e.g., '**/*.go', 'src/**/*.ts').",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Base directory to search in (relative to working directory). Defaults to '.'.",
				},
			},
			"required": []string{"pattern"},
		},
	}
}

func (t *GlobTool) CreateInvocation(params map[string]interface{}) (ToolInvocation, error) {
	base := NewBaseToolInvocation(params, GlobToolName)
	pattern := base.GetStringParam("pattern", "")
	if strings.TrimSpace(pattern) == "" {
		return nil, fmt.Errorf("the 'pattern' parameter must be non-empty")
	}

	return &globInvocation{
		BaseToolInvocation: base,
		config:             t.config,
		pattern:            pattern,
		basePath:           base.GetStringParam("path", "."),
	}, nil
}

type globInvocation struct {
	BaseToolInvocation
	config   *config.Config
	pattern  string
	basePath string
}

func (inv *globInvocation) GetDescription() string {
	return fmt.Sprintf("Find files matching '%s'", inv.pattern)
}

func (inv *globInvocation) ToolLocations() []ToolLocation {
	return nil
}

func (inv *globInvocation) Execute(ctx context.Context) (*ToolResult, error) {
	searchDir := filepath.Join(inv.config.GetTargetDir(), inv.basePath)
	if filepath.IsAbs(inv.basePath) {
		searchDir = inv.basePath
	}

	fullPattern := filepath.Join(searchDir, inv.pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return ErrorResult(ToolErrorTypeValidation,
			fmt.Sprintf("Invalid glob pattern: %s", err)), nil
	}

	// Also try recursive matching if pattern contains **
	if strings.Contains(inv.pattern, "**") {
		matches = recursiveGlob(searchDir, inv.pattern)
	}

	if len(matches) == 0 {
		return &ToolResult{
			LLMContent:    "No files found matching the pattern.",
			ReturnDisplay: "No matches",
		}, nil
	}

	// Limit results
	const maxResults = 100
	truncated := len(matches) > maxResults
	if truncated {
		matches = matches[:maxResults]
	}

	var result strings.Builder
	for _, m := range matches {
		rel, _ := filepath.Rel(inv.config.GetTargetDir(), m)
		info, err := os.Stat(m)
		if err == nil {
			if info.IsDir() {
				result.WriteString(fmt.Sprintf("%s/\n", rel))
			} else {
				result.WriteString(fmt.Sprintf("%s (%d bytes)\n", rel, info.Size()))
			}
		} else {
			result.WriteString(rel + "\n")
		}
	}

	if truncated {
		result.WriteString(fmt.Sprintf("\n... Results capped at %d files.", maxResults))
	}

	return &ToolResult{
		LLMContent:    result.String(),
		ReturnDisplay: fmt.Sprintf("%d file(s) found", len(matches)),
	}, nil
}

// recursiveGlob performs a recursive glob search supporting ** patterns.
func recursiveGlob(root string, pattern string) []string {
	var matches []string

	// Split pattern at ** boundaries
	parts := strings.Split(pattern, "**")
	if len(parts) < 2 {
		m, _ := filepath.Glob(filepath.Join(root, pattern))
		return m
	}

	suffix := parts[len(parts)-1]
	if strings.HasPrefix(suffix, "/") || strings.HasPrefix(suffix, string(filepath.Separator)) {
		suffix = suffix[1:]
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if suffix == "" {
			matches = append(matches, path)
			return nil
		}

		matched, _ := filepath.Match(suffix, filepath.Base(path))
		if matched {
			matches = append(matches, path)
		}
		return nil
	})

	return matches
}

// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google-gemini/gemini-cli/internal/config"
	"github.com/google-gemini/gemini-cli/internal/core"
)

// GrepTool implements the grep_search tool.
type GrepTool struct {
	config *config.Config
}

// NewGrepTool creates a new GrepTool.
func NewGrepTool(cfg *config.Config) *GrepTool {
	return &GrepTool{config: cfg}
}

func (t *GrepTool) Name() string        { return GrepToolName }
func (t *GrepTool) DisplayName() string  { return "Grep" }
func (t *GrepTool) Description() string  { return "Search for a pattern in files" }
func (t *GrepTool) Kind() Kind           { return KindRead }

func (t *GrepTool) GetSchema(modelID string) core.FunctionDeclaration {
	return core.FunctionDeclaration{
		Name:        GrepToolName,
		Description: "Search for a text pattern in files within the working directory. Returns matching lines with file paths and line numbers.",
		ParametersJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "The search pattern (regular expression).",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory or file path to search in (relative to working directory). Defaults to '.'.",
				},
				"include": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern to filter files (e.g., '*.go', '*.ts').",
				},
				"case_insensitive": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to perform case-insensitive search.",
				},
			},
			"required": []string{"pattern"},
		},
	}
}

func (t *GrepTool) CreateInvocation(params map[string]interface{}) (ToolInvocation, error) {
	base := NewBaseToolInvocation(params, GrepToolName)
	pattern := base.GetStringParam("pattern", "")
	if strings.TrimSpace(pattern) == "" {
		return nil, fmt.Errorf("the 'pattern' parameter must be non-empty")
	}

	return &grepInvocation{
		BaseToolInvocation: base,
		config:             t.config,
		pattern:            pattern,
		searchPath:         base.GetStringParam("path", "."),
		include:            base.GetStringParam("include", ""),
		caseInsensitive:    base.GetBoolParam("case_insensitive", false),
	}, nil
}

type grepInvocation struct {
	BaseToolInvocation
	config          *config.Config
	pattern         string
	searchPath      string
	include         string
	caseInsensitive bool
}

func (inv *grepInvocation) GetDescription() string {
	return fmt.Sprintf("Search for '%s'", inv.pattern)
}

func (inv *grepInvocation) ToolLocations() []ToolLocation {
	return nil
}

func (inv *grepInvocation) Execute(ctx context.Context) (*ToolResult, error) {
	searchDir := filepath.Join(inv.config.GetTargetDir(), inv.searchPath)
	if filepath.IsAbs(inv.searchPath) {
		searchDir = inv.searchPath
	}

	flags := ""
	if inv.caseInsensitive {
		flags = "(?i)"
	}

	re, err := regexp.Compile(flags + inv.pattern)
	if err != nil {
		return ErrorResult(ToolErrorTypeValidation,
			fmt.Sprintf("Invalid regex pattern: %s", err)), nil
	}

	type match struct {
		file    string
		lineNum int
		line    string
	}

	var matches []match
	const maxMatches = 50

	err = filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if len(matches) >= maxMatches {
			return filepath.SkipAll
		}

		// Apply include filter
		if inv.include != "" {
			matched, _ := filepath.Match(inv.include, filepath.Base(path))
			if !matched {
				return nil
			}
		}

		// Skip binary files (simple heuristic)
		if isBinaryFile(path) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if re.MatchString(line) {
				rel, _ := filepath.Rel(inv.config.GetTargetDir(), path)
				matches = append(matches, match{
					file:    rel,
					lineNum: lineNum,
					line:    line,
				})
				if len(matches) >= maxMatches {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return ErrorResult(ToolErrorTypeGeneral, err.Error()), nil
	}

	if len(matches) == 0 {
		return &ToolResult{
			LLMContent:    "No matches found.",
			ReturnDisplay: "No matches",
		}, nil
	}

	var result strings.Builder
	for _, m := range matches {
		result.WriteString(fmt.Sprintf("%s:%d: %s\n", m.file, m.lineNum, m.line))
	}

	if len(matches) >= maxMatches {
		result.WriteString(fmt.Sprintf("\n... Results capped at %d matches.", maxMatches))
	}

	return &ToolResult{
		LLMContent:    result.String(),
		ReturnDisplay: fmt.Sprintf("%d match(es)", len(matches)),
	}, nil
}

// isBinaryFile performs a simple binary file detection.
func isBinaryFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil || n == 0 {
		return false
	}

	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}

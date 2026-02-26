// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google-gemini/gemini-cli/internal/config"
	"github.com/google-gemini/gemini-cli/internal/core"
)

// ReadFileTool implements the read_file tool.
type ReadFileTool struct {
	config *config.Config
}

// NewReadFileTool creates a new ReadFileTool.
func NewReadFileTool(cfg *config.Config) *ReadFileTool {
	return &ReadFileTool{config: cfg}
}

func (t *ReadFileTool) Name() string        { return ReadFileToolName }
func (t *ReadFileTool) DisplayName() string  { return ReadFileDisplayName }
func (t *ReadFileTool) Description() string  { return "Read the contents of a file" }
func (t *ReadFileTool) Kind() Kind           { return KindRead }

func (t *ReadFileTool) GetSchema(modelID string) core.FunctionDeclaration {
	return core.FunctionDeclaration{
		Name:        ReadFileToolName,
		Description: "Read the contents of a file. Use start_line and end_line to read specific line ranges for large files.",
		ParametersJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the file to read, relative to the working directory.",
				},
				"start_line": map[string]interface{}{
					"type":        "integer",
					"description": "The 1-based line number to start reading from (inclusive).",
				},
				"end_line": map[string]interface{}{
					"type":        "integer",
					"description": "The 1-based line number to stop reading at (inclusive).",
				},
			},
			"required": []string{"file_path"},
		},
	}
}

func (t *ReadFileTool) CreateInvocation(params map[string]interface{}) (ToolInvocation, error) {
	base := NewBaseToolInvocation(params, ReadFileToolName)
	filePath := base.GetStringParam("file_path", "")
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("the 'file_path' parameter must be non-empty")
	}

	resolvedPath := filepath.Join(t.config.GetTargetDir(), filePath)
	if filepath.IsAbs(filePath) {
		resolvedPath = filePath
	}

	if err := t.config.ValidatePathAccess(resolvedPath, "read"); err != "" {
		return nil, fmt.Errorf("%s", err)
	}

	return &readFileInvocation{
		BaseToolInvocation: base,
		config:             t.config,
		resolvedPath:       resolvedPath,
		startLine:          base.GetIntParam("start_line", 0),
		endLine:            base.GetIntParam("end_line", 0),
	}, nil
}

type readFileInvocation struct {
	BaseToolInvocation
	config       *config.Config
	resolvedPath string
	startLine    int
	endLine      int
}

func (inv *readFileInvocation) GetDescription() string {
	rel, _ := filepath.Rel(inv.config.GetTargetDir(), inv.resolvedPath)
	if rel == "" {
		rel = inv.resolvedPath
	}
	return rel
}

func (inv *readFileInvocation) ToolLocations() []ToolLocation {
	return []ToolLocation{{Path: inv.resolvedPath, Line: inv.startLine}}
}

func (inv *readFileInvocation) Execute(ctx context.Context) (*ToolResult, error) {
	info, err := os.Stat(inv.resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrorResult(ToolErrorTypeFileNotFound,
				fmt.Sprintf("File not found: %s", inv.resolvedPath)), nil
		}
		return ErrorResult(ToolErrorTypeGeneral, err.Error()), nil
	}

	if info.IsDir() {
		return ErrorResult(ToolErrorTypeGeneral,
			fmt.Sprintf("Path is a directory, not a file: %s", inv.resolvedPath)), nil
	}

	file, err := os.Open(inv.resolvedPath)
	if err != nil {
		return ErrorResult(ToolErrorTypePermissionDenied, err.Error()), nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if inv.startLine > 0 && lineNum < inv.startLine {
			continue
		}
		if inv.endLine > 0 && lineNum > inv.endLine {
			break
		}
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return ErrorResult(ToolErrorTypeGeneral, err.Error()), nil
	}

	content := strings.Join(lines, "\n")

	// Truncate if too large
	const maxContentLen = 100_000
	isTruncated := len(content) > maxContentLen
	if isTruncated {
		content = content[:maxContentLen]
	}

	var llmContent string
	if isTruncated {
		llmContent = fmt.Sprintf(`IMPORTANT: The file content has been truncated.
Status: Showing partial content of %s.
Action: Use 'start_line' and 'end_line' parameters to read specific sections.

--- FILE CONTENT (truncated) ---
%s`, inv.resolvedPath, content)
	} else {
		llmContent = content
	}

	rel, _ := filepath.Rel(inv.config.GetTargetDir(), inv.resolvedPath)
	return &ToolResult{
		LLMContent:    llmContent,
		ReturnDisplay: rel,
	}, nil
}

// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google-gemini/gemini-cli/internal/config"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	tmpDir := t.TempDir()
	return config.NewConfig(config.ConfigParameters{
		TargetDir: tmpDir,
		Cwd:       tmpDir,
	})
}

func TestReadFileTool_BasicRead(t *testing.T) {
	cfg := testConfig(t)
	tool := NewReadFileTool(cfg)

	// Create test file
	filePath := filepath.Join(cfg.GetTargetDir(), "test.txt")
	os.WriteFile(filePath, []byte("line 1\nline 2\nline 3\n"), 0o644)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"file_path": "test.txt",
	})
	if err != nil {
		t.Fatalf("CreateInvocation failed: %v", err)
	}

	result, err := inv.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.LLMContent == "" {
		t.Error("expected non-empty content")
	}
}

func TestReadFileTool_LineRange(t *testing.T) {
	cfg := testConfig(t)
	tool := NewReadFileTool(cfg)

	// Create test file
	filePath := filepath.Join(cfg.GetTargetDir(), "lines.txt")
	os.WriteFile(filePath, []byte("line 1\nline 2\nline 3\nline 4\nline 5\n"), 0o644)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"file_path":  "lines.txt",
		"start_line": 2,
		"end_line":   4,
	})
	if err != nil {
		t.Fatalf("CreateInvocation failed: %v", err)
	}

	result, err := inv.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.LLMContent != "line 2\nline 3\nline 4" {
		t.Errorf("unexpected content: %q", result.LLMContent)
	}
}

func TestReadFileTool_FileNotFound(t *testing.T) {
	cfg := testConfig(t)
	tool := NewReadFileTool(cfg)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"file_path": "nonexistent.txt",
	})
	if err != nil {
		t.Fatalf("CreateInvocation failed: %v", err)
	}

	result, err := inv.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Error == nil {
		t.Error("expected error for missing file")
	}
	if result.Error.Type != ToolErrorTypeFileNotFound {
		t.Errorf("expected FileNotFound error, got %s", result.Error.Type)
	}
}

func TestReadFileTool_Directory(t *testing.T) {
	cfg := testConfig(t)
	tool := NewReadFileTool(cfg)

	dirPath := filepath.Join(cfg.GetTargetDir(), "subdir")
	os.MkdirAll(dirPath, 0o755)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"file_path": "subdir",
	})
	if err != nil {
		t.Fatalf("CreateInvocation failed: %v", err)
	}

	result, err := inv.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Error == nil {
		t.Error("expected error for directory")
	}
}

func TestReadFileTool_EmptyPath(t *testing.T) {
	cfg := testConfig(t)
	tool := NewReadFileTool(cfg)

	_, err := tool.CreateInvocation(map[string]interface{}{
		"file_path": "",
	})
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestReadFileTool_Metadata(t *testing.T) {
	cfg := testConfig(t)
	tool := NewReadFileTool(cfg)

	if tool.Name() != ReadFileToolName {
		t.Errorf("unexpected name: %s", tool.Name())
	}
	if tool.Kind() != KindRead {
		t.Errorf("unexpected kind: %v", tool.Kind())
	}
}

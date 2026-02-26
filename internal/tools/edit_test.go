// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestEditTool_SuccessfulEdit(t *testing.T) {
	cfg := testConfig(t)
	tool := NewEditTool(cfg)

	filePath := filepath.Join(cfg.GetTargetDir(), "edit_me.txt")
	os.WriteFile(filePath, []byte("hello world"), 0o644)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"file_path":  "edit_me.txt",
		"old_string": "hello",
		"new_string": "goodbye",
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

	// Verify file was modified
	data, _ := os.ReadFile(filePath)
	if string(data) != "goodbye world" {
		t.Errorf("file content = %q, want %q", string(data), "goodbye world")
	}
}

func TestEditTool_OldStringNotFound(t *testing.T) {
	cfg := testConfig(t)
	tool := NewEditTool(cfg)

	filePath := filepath.Join(cfg.GetTargetDir(), "edit_me.txt")
	os.WriteFile(filePath, []byte("hello world"), 0o644)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"file_path":  "edit_me.txt",
		"old_string": "nonexistent",
		"new_string": "replacement",
	})
	if err != nil {
		t.Fatalf("CreateInvocation failed: %v", err)
	}

	result, err := inv.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Error == nil {
		t.Error("expected error for old_string not found")
	}
	if result.Error.Type != ToolErrorTypeTargetNotFound {
		t.Errorf("expected TargetNotFound error, got %s", result.Error.Type)
	}
}

func TestEditTool_MultiplOccurrences(t *testing.T) {
	cfg := testConfig(t)
	tool := NewEditTool(cfg)

	filePath := filepath.Join(cfg.GetTargetDir(), "edit_me.txt")
	os.WriteFile(filePath, []byte("hello hello hello"), 0o644)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"file_path":  "edit_me.txt",
		"old_string": "hello",
		"new_string": "bye",
	})
	if err != nil {
		t.Fatalf("CreateInvocation failed: %v", err)
	}

	result, err := inv.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Error == nil {
		t.Error("expected error for multiple occurrences")
	}
}

func TestEditTool_FileNotFound(t *testing.T) {
	cfg := testConfig(t)
	tool := NewEditTool(cfg)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"file_path":  "nonexistent.txt",
		"old_string": "hello",
		"new_string": "bye",
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
}

func TestEditTool_Metadata(t *testing.T) {
	cfg := testConfig(t)
	tool := NewEditTool(cfg)

	if tool.Name() != EditToolName {
		t.Errorf("unexpected name: %s", tool.Name())
	}
	if tool.Kind() != KindWrite {
		t.Errorf("unexpected kind: %v", tool.Kind())
	}
}

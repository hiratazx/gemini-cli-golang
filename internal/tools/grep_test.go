// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGrepTool_BasicSearch(t *testing.T) {
	cfg := testConfig(t)
	tool := NewGrepTool(cfg)

	// Create test files
	dir := cfg.GetTargetDir()
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("hello world\nfoo bar\nhello again"), 0o644)
	os.WriteFile(filepath.Join(dir, "file2.go"), []byte("package main\nfunc hello() {}"), 0o644)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"pattern": "hello",
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
	if !strings.Contains(result.LLMContent, "hello") {
		t.Error("expected results to contain 'hello'")
	}
}

func TestGrepTool_WithInclude(t *testing.T) {
	cfg := testConfig(t)
	tool := NewGrepTool(cfg)

	dir := cfg.GetTargetDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello txt"), 0o644)
	os.WriteFile(filepath.Join(dir, "file.go"), []byte("hello go"), 0o644)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"pattern": "hello",
		"include": "*.go",
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
	if !strings.Contains(result.LLMContent, "file.go") {
		t.Error("expected results to include .go file")
	}
}

func TestGrepTool_NoResults(t *testing.T) {
	cfg := testConfig(t)
	tool := NewGrepTool(cfg)

	dir := cfg.GetTargetDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello world"), 0o644)

	inv, err := tool.CreateInvocation(map[string]interface{}{
		"pattern": "nonexistent_string_xyz",
	})
	if err != nil {
		t.Fatalf("CreateInvocation failed: %v", err)
	}

	result, err := inv.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if strings.Contains(result.LLMContent, "nonexistent_string_xyz") {
		t.Error("should not find nonexistent string")
	}
}

func TestGrepTool_EmptyPattern(t *testing.T) {
	cfg := testConfig(t)
	tool := NewGrepTool(cfg)

	_, err := tool.CreateInvocation(map[string]interface{}{
		"pattern": "",
	})
	if err == nil {
		t.Error("expected error for empty pattern")
	}
}

func TestGrepTool_Metadata(t *testing.T) {
	cfg := testConfig(t)
	tool := NewGrepTool(cfg)

	if tool.Name() != GrepToolName {
		t.Errorf("unexpected name: %s", tool.Name())
	}
	if tool.Kind() != KindRead {
		t.Errorf("unexpected kind: %v", tool.Kind())
	}
}

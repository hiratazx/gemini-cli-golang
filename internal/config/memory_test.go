// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"
	"testing"
)

func TestFlattenMemory_Nil(t *testing.T) {
	result := FlattenMemory(nil)
	if result != "" {
		t.Errorf("FlattenMemory(nil) = %q, want empty", result)
	}
}

func TestFlattenMemory_String(t *testing.T) {
	result := FlattenMemory("hello world")
	if result != "hello world" {
		t.Errorf("FlattenMemory(string) = %q, want %q", result, "hello world")
	}
}

func TestFlattenMemory_HierarchicalPointer(t *testing.T) {
	mem := &HierarchicalMemory{
		Global:  "global context",
		Project: "project context",
	}
	result := FlattenMemory(mem)

	if !strings.Contains(result, "Global") {
		t.Error("expected result to contain 'Global'")
	}
	if !strings.Contains(result, "global context") {
		t.Error("expected result to contain global context")
	}
	if !strings.Contains(result, "Project") {
		t.Error("expected result to contain 'Project'")
	}
	if !strings.Contains(result, "project context") {
		t.Error("expected result to contain project context")
	}
}

func TestFlattenMemory_NilPointer(t *testing.T) {
	var mem *HierarchicalMemory
	result := FlattenMemory(mem)
	if result != "" {
		t.Errorf("FlattenMemory(nil pointer) = %q, want empty", result)
	}
}

func TestHierarchicalMemory_Flatten_AllSections(t *testing.T) {
	mem := HierarchicalMemory{
		Global:    "global",
		Extension: "extension",
		Project:   "project",
	}
	result := mem.Flatten()

	if !strings.Contains(result, "--- Global ---") {
		t.Error("expected Global section header")
	}
	if !strings.Contains(result, "--- Extension ---") {
		t.Error("expected Extension section header")
	}
	if !strings.Contains(result, "--- Project ---") {
		t.Error("expected Project section header")
	}
}

func TestHierarchicalMemory_Flatten_EmptySections(t *testing.T) {
	mem := HierarchicalMemory{
		Global: "only global",
	}
	result := mem.Flatten()

	if !strings.Contains(result, "Global") {
		t.Error("expected Global section")
	}
	if strings.Contains(result, "Extension") {
		t.Error("did not expect Extension section")
	}
	if strings.Contains(result, "Project") {
		t.Error("did not expect Project section")
	}
}

func TestHierarchicalMemory_Flatten_AllEmpty(t *testing.T) {
	mem := HierarchicalMemory{}
	result := mem.Flatten()
	if result != "" {
		t.Errorf("expected empty string for empty memory, got %q", result)
	}
}

func TestHierarchicalMemory_Flatten_WhitespaceOnly(t *testing.T) {
	mem := HierarchicalMemory{
		Global:  "  \n  ",
		Project: "  ",
	}
	result := mem.Flatten()
	if result != "" {
		t.Errorf("expected empty string for whitespace-only memory, got %q", result)
	}
}

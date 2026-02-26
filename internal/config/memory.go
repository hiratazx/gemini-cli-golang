// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"strings"
)

// HierarchicalMemory represents memory organized by scope level.
type HierarchicalMemory struct {
	// Global memory shared across all projects.
	Global string
	// Extension memory from installed extensions.
	Extension string
	// Project memory specific to the current project.
	Project string
}

// FlattenMemory converts hierarchical memory (or a plain string) into a single string.
func FlattenMemory(memory interface{}) string {
	switch m := memory.(type) {
	case nil:
		return ""
	case string:
		return m
	case *HierarchicalMemory:
		if m == nil {
			return ""
		}
		return m.Flatten()
	case HierarchicalMemory:
		return m.Flatten()
	default:
		return fmt.Sprintf("%v", m)
	}
}

// Flatten converts the hierarchical memory into a single string with section headers.
func (h HierarchicalMemory) Flatten() string {
	type section struct {
		name    string
		content string
	}

	var sections []section

	if trimmed := strings.TrimSpace(h.Global); trimmed != "" {
		sections = append(sections, section{name: "Global", content: trimmed})
	}
	if trimmed := strings.TrimSpace(h.Extension); trimmed != "" {
		sections = append(sections, section{name: "Extension", content: trimmed})
	}
	if trimmed := strings.TrimSpace(h.Project); trimmed != "" {
		sections = append(sections, section{name: "Project", content: trimmed})
	}

	if len(sections) == 0 {
		return ""
	}

	parts := make([]string, len(sections))
	for i, s := range sections {
		parts[i] = fmt.Sprintf("--- %s ---\n%s", s.name, s.content)
	}

	return strings.Join(parts, "\n\n")
}

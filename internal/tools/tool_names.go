// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"regexp"
	"strings"
)

// Tool name constants.
const (
	ReadFileToolName      = "read_file"
	WriteFileToolName     = "write_file"
	EditToolName          = "replace"
	ShellToolName         = "run_shell_command"
	GrepToolName          = "grep_search"
	GlobToolName          = "glob"
	LSToolName            = "list_directory"
	WebFetchToolName      = "web_fetch"
	WebSearchToolName     = "google_web_search"
	MemoryToolName        = "save_memory"
	ReadManyFilesToolName = "read_many_files"
	WriteTodosToolName    = "write_todos"
	ActivateSkillToolName = "activate_skill"
	AskUserToolName       = "ask_user"
	EnterPlanModeToolName = "enter_plan_mode"
	ExitPlanModeToolName  = "exit_plan_mode"
	GetInternalDocsToolName = "get_internal_docs"
)

// Display names for tools.
const (
	ReadFileDisplayName  = "ReadFile"
	WriteFileDisplayName = "WriteFile"
	EditDisplayName      = "Edit"
	AskUserDisplayName   = "Ask User"
	GlobDisplayName      = "FindFiles"
)

// DiscoveredToolPrefix is the prefix for dynamically discovered tools.
const DiscoveredToolPrefix = "discovered_tool_"

// EditToolNames contains all tool names that modify file content.
var EditToolNames = map[string]bool{
	EditToolName:      true,
	WriteFileToolName: true,
}

// ToolLegacyAliases maps old tool names to current names for backward compatibility.
var ToolLegacyAliases = map[string]string{
	"search_file_content": GrepToolName,
}

// AllBuiltinToolNames is the list of all built-in tool names.
var AllBuiltinToolNames = []string{
	GlobToolName,
	WriteTodosToolName,
	WriteFileToolName,
	WebSearchToolName,
	WebFetchToolName,
	EditToolName,
	ShellToolName,
	GrepToolName,
	ReadManyFilesToolName,
	ReadFileToolName,
	LSToolName,
	MemoryToolName,
	ActivateSkillToolName,
	AskUserToolName,
	GetInternalDocsToolName,
	EnterPlanModeToolName,
	ExitPlanModeToolName,
}

// GetToolAliases returns all associated names for a tool (including legacy aliases).
func GetToolAliases(name string) []string {
	aliases := map[string]bool{name: true}

	// Determine canonical name
	canonicalName := name
	if current, ok := ToolLegacyAliases[name]; ok {
		canonicalName = current
	}
	aliases[canonicalName] = true

	// Find all legacy aliases pointing to the same canonical name
	for legacyName, currentName := range ToolLegacyAliases {
		if currentName == canonicalName {
			aliases[legacyName] = true
		}
	}

	result := make([]string, 0, len(aliases))
	for alias := range aliases {
		result = append(result, alias)
	}
	return result
}

var slugRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

// IsValidToolName validates if a tool name is syntactically valid.
func IsValidToolName(name string, allowWildcards bool) bool {
	// Built-in tools
	for _, builtIn := range AllBuiltinToolNames {
		if name == builtIn {
			return true
		}
	}

	// Legacy aliases
	if _, ok := ToolLegacyAliases[name]; ok {
		return true
	}

	// Discovered tools
	if strings.HasPrefix(name, DiscoveredToolPrefix) {
		return true
	}

	// Wildcards
	if allowWildcards && name == "*" {
		return true
	}

	// MCP tools (format: server__tool)
	if strings.Contains(name, "__") {
		parts := strings.SplitN(name, "__", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return false
		}

		if parts[1] == "*" {
			return allowWildcards
		}

		return slugRegex.MatchString(parts[0]) && slugRegex.MatchString(parts[1])
	}

	return false
}

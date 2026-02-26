// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/google-gemini/gemini-cli/internal/config"
)

// GetCoreSystemPrompt generates the system prompt for Gemini.
func GetCoreSystemPrompt(cfg *config.Config, userMemory string) string {
	var builder strings.Builder

	builder.WriteString("You are an interactive CLI assistant powered by Gemini.\n")
	builder.WriteString("You help users with coding tasks, file operations, and general questions.\n\n")

	// Environment context
	builder.WriteString("## Environment\n")
	builder.WriteString(fmt.Sprintf("- Working directory: %s\n", cfg.GetTargetDir()))
	builder.WriteString(fmt.Sprintf("- Operating system: %s/%s\n", runtime.GOOS, runtime.GOARCH))
	builder.WriteString(fmt.Sprintf("- Current time: %s\n", time.Now().Format(time.RFC3339)))
	builder.WriteString("\n")

	// Tool usage instructions
	builder.WriteString("## Tool Usage Guidelines\n")
	builder.WriteString("- Use the available tools to help users accomplish their tasks.\n")
	builder.WriteString("- Always use absolute paths when working with files.\n")
	builder.WriteString("- Before modifying files, read them first to understand the current content.\n")
	builder.WriteString("- When running shell commands, prefer non-interactive commands.\n")
	builder.WriteString("- Provide clear explanations of what you're doing and why.\n")
	builder.WriteString("\n")

	// User memory / context
	if userMemory != "" {
		builder.WriteString("## User Context (GEMINI.md)\n")
		builder.WriteString("The following is context provided by the user. Follow these instructions:\n\n")
		builder.WriteString(userMemory)
		builder.WriteString("\n\n")
	}

	// Approval mode context
	switch cfg.GetApprovalMode() {
	case config.ApprovalModeYolo:
		builder.WriteString("## Approval Mode\n")
		builder.WriteString("Auto-approve mode is enabled. All tool calls will be executed without confirmation.\n\n")
	case config.ApprovalModePlan:
		builder.WriteString("## Approval Mode\n")
		builder.WriteString("Plan mode is enabled. Only read-only operations are allowed. ")
		builder.WriteString("Do not attempt to write or modify files.\n\n")
	}

	return builder.String()
}

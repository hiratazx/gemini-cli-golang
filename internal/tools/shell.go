// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google-gemini/gemini-cli/internal/config"
	"github.com/google-gemini/gemini-cli/internal/core"
)

// ShellTool implements the run_shell_command tool.
type ShellTool struct {
	config *config.Config
}

// NewShellTool creates a new ShellTool.
func NewShellTool(cfg *config.Config) *ShellTool {
	return &ShellTool{config: cfg}
}

func (t *ShellTool) Name() string        { return ShellToolName }
func (t *ShellTool) DisplayName() string  { return "Shell" }
func (t *ShellTool) Description() string  { return "Execute a shell command" }
func (t *ShellTool) Kind() Kind           { return KindExecute }

func (t *ShellTool) GetSchema(modelID string) core.FunctionDeclaration {
	return core.FunctionDeclaration{
		Name:        ShellToolName,
		Description: "Execute a shell command in the working directory. Use for running builds, tests, git operations, installing packages, or any other system tasks.",
		ParametersJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The shell command to execute.",
				},
				"timeout": map[string]interface{}{
					"type":        "integer",
					"description": "Timeout in seconds (default: 120).",
				},
			},
			"required": []string{"command"},
		},
	}
}

func (t *ShellTool) CreateInvocation(params map[string]interface{}) (ToolInvocation, error) {
	base := NewBaseToolInvocation(params, ShellToolName)
	command := base.GetStringParam("command", "")
	if strings.TrimSpace(command) == "" {
		return nil, fmt.Errorf("the 'command' parameter must be non-empty")
	}

	timeout := base.GetIntParam("timeout", 120)

	return &shellInvocation{
		BaseToolInvocation: base,
		config:             t.config,
		command:            command,
		timeout:            time.Duration(timeout) * time.Second,
	}, nil
}

type shellInvocation struct {
	BaseToolInvocation
	config  *config.Config
	command string
	timeout time.Duration
}

func (inv *shellInvocation) GetDescription() string {
	cmd := inv.command
	if len(cmd) > 80 {
		cmd = cmd[:80] + "..."
	}
	return fmt.Sprintf("$ %s", cmd)
}

func (inv *shellInvocation) ToolLocations() []ToolLocation {
	return nil
}

func (inv *shellInvocation) Execute(ctx context.Context) (*ToolResult, error) {
	ctx, cancel := context.WithTimeout(ctx, inv.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", inv.command)
	cmd.Dir = inv.config.GetTargetDir()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var result strings.Builder
	if stdout.Len() > 0 {
		result.WriteString("STDOUT:\n")
		result.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString("STDERR:\n")
		result.WriteString(stderr.String())
	}

	output := result.String()

	// Truncate if too large
	const maxOutputLen = 50_000
	if len(output) > maxOutputLen {
		output = output[:maxOutputLen] + "\n... [output truncated]"
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ToolResult{
				LLMContent:    fmt.Sprintf("Command timed out after %s.\n%s", inv.timeout, output),
				ReturnDisplay: "Command timed out",
				Error: &ToolError{
					Type:    ToolErrorTypeTimeout,
					Message: "Command execution timed out",
				},
			}, nil
		}

		exitCode := -1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}

		return &ToolResult{
			LLMContent:    fmt.Sprintf("Command exited with code %d.\n%s", exitCode, output),
			ReturnDisplay: fmt.Sprintf("Exit code: %d", exitCode),
		}, nil
	}

	if output == "" {
		output = "(no output)"
	}

	return &ToolResult{
		LLMContent:    output,
		ReturnDisplay: fmt.Sprintf("$ %s", inv.command),
	}, nil
}

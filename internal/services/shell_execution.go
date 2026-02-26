// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// ShellExecutor provides shell command execution capabilities.
type ShellExecutor struct {
	workDir        string
	defaultTimeout time.Duration
	shell          string
}

// NewShellExecutor creates a new ShellExecutor.
func NewShellExecutor(workDir string) *ShellExecutor {
	return &ShellExecutor{
		workDir:        workDir,
		defaultTimeout: 120 * time.Second,
		shell:          "bash",
	}
}

// ShellResult represents the result of a shell command execution.
type ShellResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
}

// Execute runs a shell command and returns the result.
func (s *ShellExecutor) Execute(ctx context.Context, command string) (*ShellResult, error) {
	return s.ExecuteWithTimeout(ctx, command, s.defaultTimeout)
}

// ExecuteWithTimeout runs a shell command with a specific timeout.
func (s *ShellExecutor) ExecuteWithTimeout(ctx context.Context, command string, timeout time.Duration) (*ShellResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.shell, "-c", command)
	cmd.Dir = s.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &ShellResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.TimedOut = true
			return result, nil
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return result, nil
}

// SetShell sets the shell to use for execution (default: bash).
func (s *ShellExecutor) SetShell(shell string) {
	s.shell = shell
}

// SetDefaultTimeout sets the default timeout for command execution.
func (s *ShellExecutor) SetDefaultTimeout(timeout time.Duration) {
	s.defaultTimeout = timeout
}

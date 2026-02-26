// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GitService provides git operations via command execution.
type GitService struct {
	workDir string
}

// NewGitService creates a new GitService.
func NewGitService(workDir string) *GitService {
	return &GitService{workDir: workDir}
}

// IsGitRepo checks if the working directory is inside a git repository.
func (g *GitService) IsGitRepo() bool {
	_, err := g.run("rev-parse", "--is-inside-work-tree")
	return err == nil
}

// GetRepoRoot returns the root directory of the git repository.
func (g *GitService) GetRepoRoot() (string, error) {
	out, err := g.run("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetCurrentBranch returns the current branch name.
func (g *GitService) GetCurrentBranch() (string, error) {
	out, err := g.run("branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetStatus returns the git status output.
func (g *GitService) GetStatus() (string, error) {
	return g.run("status", "--porcelain")
}

// GetDiff returns the diff of uncommitted changes.
func (g *GitService) GetDiff(staged bool) (string, error) {
	if staged {
		return g.run("diff", "--cached")
	}
	return g.run("diff")
}

// GetLog returns recent commit logs.
func (g *GitService) GetLog(count int) (string, error) {
	return g.run("log", fmt.Sprintf("-n%d", count), "--oneline")
}

// run executes a git command and returns the output.
func (g *GitService) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", strings.Join(args, " "), stderr.String(), err)
	}

	return stdout.String(), nil
}

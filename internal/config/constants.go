// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration types and constants for Gemini CLI.
package config

// FileFilteringOptions controls how files are discovered and filtered.
type FileFilteringOptions struct {
	// RespectGitIgnore indicates whether to honor .gitignore patterns.
	RespectGitIgnore bool
	// RespectGeminiIgnore indicates whether to honor .geminiignore patterns.
	RespectGeminiIgnore bool
	// MaxFileCount is the maximum number of files to discover.
	MaxFileCount int
	// SearchTimeout is the timeout in milliseconds for file search operations.
	SearchTimeout int
	// CustomIgnoreFilePaths is a list of additional ignore file paths.
	CustomIgnoreFilePaths []string
}

// DefaultMemoryFileFilteringOptions provides defaults for memory file filtering.
var DefaultMemoryFileFilteringOptions = FileFilteringOptions{
	RespectGitIgnore:      false,
	RespectGeminiIgnore:   true,
	MaxFileCount:          20000,
	SearchTimeout:         5000,
	CustomIgnoreFilePaths: []string{},
}

// DefaultFileFilteringOptions provides defaults for general file filtering.
var DefaultFileFilteringOptions = FileFilteringOptions{
	RespectGitIgnore:      true,
	RespectGeminiIgnore:   true,
	MaxFileCount:          20000,
	SearchTimeout:         5000,
	CustomIgnoreFilePaths: []string{},
}

// GeminiIgnoreFileName is the name of the Gemini-specific ignore file.
const GeminiIgnoreFileName = ".geminiignore"

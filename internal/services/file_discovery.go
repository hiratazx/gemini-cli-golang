// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package services provides service implementations for file discovery,
// git operations, and shell execution.
package services

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google-gemini/gemini-cli/internal/config"
)

// FileDiscoveryService handles file discovery and ignore pattern matching.
type FileDiscoveryService struct {
	rootDir  string
	options  config.FileFilteringOptions
	patterns []string
}

// NewFileDiscoveryService creates a new FileDiscoveryService.
func NewFileDiscoveryService(rootDir string, options config.FileFilteringOptions) *FileDiscoveryService {
	svc := &FileDiscoveryService{
		rootDir: rootDir,
		options: options,
	}
	svc.loadIgnorePatterns()
	return svc
}

// loadIgnorePatterns loads patterns from .gitignore and .geminiignore.
func (s *FileDiscoveryService) loadIgnorePatterns() {
	if s.options.RespectGitIgnore {
		s.loadPatternsFromFile(filepath.Join(s.rootDir, ".gitignore"))
	}
	if s.options.RespectGeminiIgnore {
		s.loadPatternsFromFile(filepath.Join(s.rootDir, config.GeminiIgnoreFileName))
	}
	for _, p := range s.options.CustomIgnoreFilePaths {
		s.loadPatternsFromFile(p)
	}
}

// loadPatternsFromFile reads patterns from an ignore file.
func (s *FileDiscoveryService) loadPatternsFromFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		s.patterns = append(s.patterns, line)
	}
}

// ShouldIgnoreFile checks if a file path should be ignored based on patterns.
func (s *FileDiscoveryService) ShouldIgnoreFile(path string, opts config.FileFilteringOptions) bool {
	rel, err := filepath.Rel(s.rootDir, path)
	if err != nil {
		return false
	}

	for _, pattern := range s.patterns {
		negation := false
		p := pattern
		if strings.HasPrefix(p, "!") {
			negation = true
			p = p[1:]
		}

		matched, _ := filepath.Match(p, filepath.Base(rel))
		if !matched {
			matched, _ = filepath.Match(p, rel)
		}

		if matched {
			if negation {
				return false
			}
			return true
		}
	}

	return false
}

// DiscoverFiles returns a list of files in the root directory respecting ignore patterns.
func (s *FileDiscoveryService) DiscoverFiles() ([]string, error) {
	var files []string
	count := 0

	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories
		base := filepath.Base(path)
		if info.IsDir() && strings.HasPrefix(base, ".") && base != "." {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		if s.ShouldIgnoreFile(path, s.options) {
			return nil
		}

		count++
		if s.options.MaxFileCount > 0 && count > s.options.MaxFileCount {
			return filepath.SkipAll
		}

		rel, _ := filepath.Rel(s.rootDir, path)
		files = append(files, rel)
		return nil
	})

	return files, err
}

// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// MakeRelative returns path relative to base, or the original path on error.
func MakeRelative(path, base string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return rel
}

// ShortenPath truncates a path to a reasonable display length.
func ShortenPath(path string) string {
	const maxLen = 60
	if len(path) <= maxLen {
		return path
	}

	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) <= 2 {
		return path
	}

	// Show first and last parts with ... in between
	shortened := parts[0] + "/.../" + parts[len(parts)-1]
	if len(shortened) > maxLen {
		return ".../" + parts[len(parts)-1]
	}
	return shortened
}

// IsSubpath checks if child is a subpath of parent.
func IsSubpath(parent, child string) bool {
	absParent, err := filepath.Abs(parent)
	if err != nil {
		return false
	}
	absChild, err := filepath.Abs(child)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(absParent, absChild)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

// Homedir returns the user's home directory.
func Homedir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// ExpandPath expands ~ to the home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home := Homedir()
		if home != "" {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

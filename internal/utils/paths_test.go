// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMakeRelative(t *testing.T) {
	tests := []struct {
		path string
		base string
		want string
	}{
		{"/home/user/project/file.go", "/home/user/project", "file.go"},
		{"/home/user/project/sub/file.go", "/home/user/project", filepath.Join("sub", "file.go")},
		{"/home/user/other/file.go", "/home/user/project", filepath.Join("..", "other", "file.go")},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := MakeRelative(tt.path, tt.base)
			if got != tt.want {
				t.Errorf("MakeRelative(%q, %q) = %q, want %q", tt.path, tt.base, got, tt.want)
			}
		})
	}
}

func TestShortenPath(t *testing.T) {
	// Short path stays the same
	short := "/home/user/file.go"
	if got := ShortenPath(short); got != short {
		t.Errorf("ShortenPath(%q) = %q, want unchanged", short, got)
	}

	// Long path gets shortened
	long := "/very/long/path/with/many/segments/that/exceeds/the/maximum/display/length/file.go"
	got := ShortenPath(long)
	if len(got) > 60 {
		t.Errorf("ShortenPath result too long: %d chars: %q", len(got), got)
	}
	if got == long {
		t.Error("expected path to be shortened")
	}
}

func TestIsSubpath(t *testing.T) {
	// Create temp dirs for absolute path testing
	tmpDir := t.TempDir()
	childDir := filepath.Join(tmpDir, "child")
	os.MkdirAll(childDir, 0o755)

	if !IsSubpath(tmpDir, childDir) {
		t.Error("expected child to be subpath of parent")
	}
	if !IsSubpath(tmpDir, filepath.Join(tmpDir, "file.go")) {
		t.Error("expected file to be subpath of parent")
	}
	if IsSubpath(childDir, tmpDir) {
		t.Error("parent should not be subpath of child")
	}
}

func TestExpandPath(t *testing.T) {
	home := Homedir()
	if home == "" {
		t.Skip("no home directory")
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/file.go", filepath.Join(home, "file.go")},
		{"~/sub/dir/file.go", filepath.Join(home, "sub", "dir", "file.go")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExpandPath(tt.input)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

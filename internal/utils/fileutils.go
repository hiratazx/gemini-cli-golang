// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bufio"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ProcessFileContent reads a file and returns its content as a string.
// If start/end lines are specified, only that range is returned.
func ProcessFileContent(path string, startLine, endLine int) (string, int, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, false, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	lineNum := 0
	totalLines := 0
	for scanner.Scan() {
		totalLines++
		lineNum++
		if startLine > 0 && lineNum < startLine {
			continue
		}
		if endLine > 0 && lineNum > endLine {
			continue
		}
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", 0, false, err
	}

	content := strings.Join(lines, "\n")

	// Check if content was truncated
	const maxContentSize = 100_000
	truncated := len(content) > maxContentSize
	if truncated {
		content = content[:maxContentSize]
	}

	return content, totalLines, truncated, nil
}

// DetectMIMEType detects the MIME type of a file based on extension and content.
func DetectMIMEType(path string) string {
	// Try by extension first
	ext := filepath.Ext(path)
	if ext != "" {
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			return mimeType
		}
	}

	// Try by content
	file, err := os.Open(path)
	if err != nil {
		return "application/octet-stream"
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil || n == 0 {
		return "application/octet-stream"
	}

	return http.DetectContentType(buf[:n])
}

// IsBinaryFile checks if a file appears to be binary.
func IsBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil || n == 0 {
		return false
	}

	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}

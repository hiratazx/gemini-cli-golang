// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package utils provides common utility functions for the Gemini CLI.
package utils

import "fmt"

// GetErrorMessage extracts a message from an error, returning a default if nil.
func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// WrapError wraps an error with additional context.
func WrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

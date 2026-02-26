// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

// ToolErrorType identifies the category of tool error.
type ToolErrorType string

const (
	// ToolErrorTypeGeneral is a generic tool error.
	ToolErrorTypeGeneral ToolErrorType = "GENERAL"
	// ToolErrorTypeValidation is a parameter validation error.
	ToolErrorTypeValidation ToolErrorType = "VALIDATION"
	// ToolErrorTypePathNotInWorkspace is when a path is outside the workspace.
	ToolErrorTypePathNotInWorkspace ToolErrorType = "PATH_NOT_IN_WORKSPACE"
	// ToolErrorTypeFileNotFound is when a file doesn't exist.
	ToolErrorTypeFileNotFound ToolErrorType = "FILE_NOT_FOUND"
	// ToolErrorTypePermissionDenied is when permission is denied.
	ToolErrorTypePermissionDenied ToolErrorType = "PERMISSION_DENIED"
	// ToolErrorTypeTimeout is when a tool execution times out.
	ToolErrorTypeTimeout ToolErrorType = "TIMEOUT"
	// ToolErrorTypeAborted is when execution is aborted.
	ToolErrorTypeAborted ToolErrorType = "ABORTED"
	// ToolErrorTypeBinary is when a binary file is encountered.
	ToolErrorTypeBinary ToolErrorType = "BINARY_FILE"
	// ToolErrorTypeTargetNotFound is when the edit target content is not found.
	ToolErrorTypeTargetNotFound ToolErrorType = "TARGET_NOT_FOUND"
)

// ToolError represents an error from a tool execution.
type ToolError struct {
	Message string        `json:"message"`
	Type    ToolErrorType `json:"type"`
}

// Error implements the error interface.
func (e *ToolError) Error() string {
	return e.Message
}

// NewToolError creates a new ToolError.
func NewToolError(errType ToolErrorType, message string) *ToolError {
	return &ToolError{
		Type:    errType,
		Message: message,
	}
}

// ErrorResult creates a ToolResult representing an error.
func ErrorResult(errType ToolErrorType, message string) *ToolResult {
	return &ToolResult{
		LLMContent:    message,
		ReturnDisplay: message,
		Error: &ToolError{
			Type:    errType,
			Message: message,
		},
	}
}

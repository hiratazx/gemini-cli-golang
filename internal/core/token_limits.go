// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package core

// Model token limit mappings.
// These represent the maximum input context window size for each model.
var modelTokenLimits = map[string]int{
	"gemini-2.5-pro":                        1048576,
	"gemini-2.5-flash":                      1048576,
	"gemini-2.5-flash-lite":                 1048576,
	"gemini-3-pro-preview":                  1048576,
	"gemini-3.1-pro-preview":                1048576,
	"gemini-3.1-pro-preview-customtools":    1048576,
	"gemini-3-flash-preview":                1048576,
}

// defaultTokenLimit is used for unknown models.
const defaultTokenLimit = 1048576

// TokenLimit returns the input token limit for the given model.
func TokenLimit(model string) int {
	if limit, ok := modelTokenLimits[model]; ok {
		return limit
	}
	return defaultTokenLimit
}

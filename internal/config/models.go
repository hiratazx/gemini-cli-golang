// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"regexp"
	"strings"
)

// Model name constants.
const (
	PreviewGeminiModel              = "gemini-3-pro-preview"
	PreviewGemini31Model            = "gemini-3.1-pro-preview"
	PreviewGemini31CustomToolsModel = "gemini-3.1-pro-preview-customtools"
	PreviewGeminiFlashModel         = "gemini-3-flash-preview"
	DefaultGeminiModel              = "gemini-2.5-pro"
	DefaultGeminiFlashModel         = "gemini-2.5-flash"
	DefaultGeminiFlashLiteModel     = "gemini-2.5-flash-lite"
)

// Auto model names.
const (
	PreviewGeminiModelAuto = "auto-gemini-3"
	DefaultGeminiModelAuto = "auto-gemini-2.5"
)

// Model aliases for user convenience.
const (
	GeminiModelAliasAuto      = "auto"
	GeminiModelAliasPro       = "pro"
	GeminiModelAliasFlash     = "flash"
	GeminiModelAliasFlashLite = "flash-lite"
)

// DefaultGeminiEmbeddingModel is the default embedding model.
const DefaultGeminiEmbeddingModel = "gemini-embedding-001"

// DefaultThinkingMode caps thinking at 8192 to prevent run-away thinking loops.
const DefaultThinkingMode = 8192

// ValidGeminiModels is the set of valid Gemini model names.
var ValidGeminiModels = map[string]bool{
	PreviewGeminiModel:              true,
	PreviewGemini31Model:            true,
	PreviewGemini31CustomToolsModel: true,
	PreviewGeminiFlashModel:         true,
	DefaultGeminiModel:              true,
	DefaultGeminiFlashModel:         true,
	DefaultGeminiFlashLiteModel:     true,
}

// ResolveModel resolves a model alias to a concrete model name.
func ResolveModel(requestedModel string, useGemini31 bool, useCustomToolModel bool) string {
	switch requestedModel {
	case PreviewGeminiModel, PreviewGeminiModelAuto, GeminiModelAliasAuto, GeminiModelAliasPro:
		if useGemini31 {
			if useCustomToolModel {
				return PreviewGemini31CustomToolsModel
			}
			return PreviewGemini31Model
		}
		return PreviewGeminiModel
	case DefaultGeminiModelAuto:
		return DefaultGeminiModel
	case GeminiModelAliasFlash:
		return PreviewGeminiFlashModel
	case GeminiModelAliasFlashLite:
		return DefaultGeminiFlashLiteModel
	default:
		return requestedModel
	}
}

// ResolveClassifierModel resolves the appropriate model based on the classifier's decision.
func ResolveClassifierModel(requestedModel, modelAlias string, useGemini31 bool, useCustomToolModel bool) string {
	if modelAlias == GeminiModelAliasFlash {
		switch requestedModel {
		case DefaultGeminiModelAuto, DefaultGeminiModel:
			return DefaultGeminiFlashModel
		case PreviewGeminiModelAuto, PreviewGeminiModel:
			return PreviewGeminiFlashModel
		default:
			return ResolveModel(GeminiModelAliasFlash, false, false)
		}
	}
	return ResolveModel(requestedModel, useGemini31, useCustomToolModel)
}

// GetDisplayString returns a user-friendly display string for a model.
func GetDisplayString(model string) string {
	switch model {
	case PreviewGeminiModelAuto:
		return "Auto (Gemini 3)"
	case DefaultGeminiModelAuto:
		return "Auto (Gemini 2.5)"
	case GeminiModelAliasPro:
		return PreviewGeminiModel
	case GeminiModelAliasFlash:
		return PreviewGeminiFlashModel
	case PreviewGemini31CustomToolsModel:
		return PreviewGemini31Model
	default:
		return model
	}
}

// IsPreviewModel checks if the model is a preview model.
func IsPreviewModel(model string) bool {
	return model == PreviewGeminiModel ||
		model == PreviewGemini31Model ||
		model == PreviewGemini31CustomToolsModel ||
		model == PreviewGeminiFlashModel ||
		model == PreviewGeminiModelAuto
}

// IsProModel checks if the model is a Pro model.
func IsProModel(model string) bool {
	return strings.Contains(strings.ToLower(model), "pro")
}

var gemini3Regex = regexp.MustCompile(`^gemini-3(\.|-)`)

// IsGemini3Model checks if the model is a Gemini 3 model.
func IsGemini3Model(model string) bool {
	resolved := ResolveModel(model, false, false)
	return gemini3Regex.MatchString(resolved) || resolved == "gemini-3"
}

var gemini2Regex = regexp.MustCompile(`^gemini-2(\.|$)`)

// IsGemini2Model checks if the model is a Gemini 2.x model.
func IsGemini2Model(model string) bool {
	return gemini2Regex.MatchString(model)
}

// IsCustomModel checks if the model is not a Gemini branded model.
func IsCustomModel(model string) bool {
	resolved := ResolveModel(model, false, false)
	return !strings.HasPrefix(resolved, "gemini-")
}

// SupportsModernFeatures checks if the model supports modern features like thoughts.
func SupportsModernFeatures(model string) bool {
	if IsGemini3Model(model) {
		return true
	}
	return IsCustomModel(model)
}

// IsAutoModel checks if the model is an auto model.
func IsAutoModel(model string) bool {
	return model == GeminiModelAliasAuto ||
		model == PreviewGeminiModelAuto ||
		model == DefaultGeminiModelAuto
}

// SupportsMultimodalFunctionResponse checks if the model supports multimodal function responses.
func SupportsMultimodalFunctionResponse(model string) bool {
	return strings.HasPrefix(model, "gemini-3-")
}

// IsActiveModel checks if the given model is considered active based on configuration.
func IsActiveModel(model string, useGemini31 bool, useCustomToolModel bool) bool {
	if !ValidGeminiModels[model] {
		return false
	}
	if useGemini31 {
		if model == PreviewGeminiModel {
			return false
		}
		if useCustomToolModel {
			return model != PreviewGemini31Model
		}
		return model != PreviewGemini31CustomToolsModel
	}
	return model != PreviewGemini31Model && model != PreviewGemini31CustomToolsModel
}

// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"
)

func TestResolveModel(t *testing.T) {
	tests := []struct {
		name               string
		requested          string
		useGemini31        bool
		useCustomToolModel bool
		want               string
	}{
		{"auto alias resolves to preview pro", GeminiModelAliasAuto, false, false, PreviewGeminiModel},
		{"pro alias resolves to preview pro", GeminiModelAliasPro, false, false, PreviewGeminiModel},
		{"flash alias resolves to preview flash", GeminiModelAliasFlash, false, false, PreviewGeminiFlashModel},
		{"flash-lite alias resolves", GeminiModelAliasFlashLite, false, false, DefaultGeminiFlashLiteModel},
		{"auto-gemini-2.5 resolves to default", DefaultGeminiModelAuto, false, false, DefaultGeminiModel},
		{"custom model passes through", "custom-model-v1", false, false, "custom-model-v1"},
		{"gemini-2.5-pro passes through", DefaultGeminiModel, false, false, DefaultGeminiModel},
		{"gemini-2.5-flash passes through", DefaultGeminiFlashModel, false, false, DefaultGeminiFlashModel},
		{"auto with gemini31 → 3.1 model", GeminiModelAliasAuto, true, false, PreviewGemini31Model},
		{"auto with gemini31 + custom tools", GeminiModelAliasAuto, true, true, PreviewGemini31CustomToolsModel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveModel(tt.requested, tt.useGemini31, tt.useCustomToolModel)
			if got != tt.want {
				t.Errorf("ResolveModel(%q, %v, %v) = %q, want %q",
					tt.requested, tt.useGemini31, tt.useCustomToolModel, got, tt.want)
			}
		})
	}
}

func TestIsGemini3Model(t *testing.T) {
	trueModels := []string{"gemini-3-pro-preview", "gemini-3-flash-preview", "gemini-3.1-pro-preview"}
	falseModels := []string{"gemini-2.5-pro", "gemini-2.5-flash", "gpt-4", ""}

	for _, m := range trueModels {
		if !IsGemini3Model(m) {
			t.Errorf("IsGemini3Model(%q) = false, want true", m)
		}
	}
	for _, m := range falseModels {
		if IsGemini3Model(m) {
			t.Errorf("IsGemini3Model(%q) = true, want false", m)
		}
	}
}

func TestIsGemini2Model(t *testing.T) {
	trueModels := []string{"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.5-flash-lite"}
	falseModels := []string{"gemini-3-pro-preview", "gpt-4", "custom-model"}

	for _, m := range trueModels {
		if !IsGemini2Model(m) {
			t.Errorf("IsGemini2Model(%q) = false, want true", m)
		}
	}
	for _, m := range falseModels {
		if IsGemini2Model(m) {
			t.Errorf("IsGemini2Model(%q) = true, want false", m)
		}
	}
}

func TestIsCustomModel(t *testing.T) {
	custom := []string{"gpt-4", "claude-3", "custom-model"}
	notCustom := []string{"gemini-2.5-pro", "auto", "pro", "flash"}

	for _, m := range custom {
		if !IsCustomModel(m) {
			t.Errorf("IsCustomModel(%q) = false, want true", m)
		}
	}
	for _, m := range notCustom {
		if IsCustomModel(m) {
			t.Errorf("IsCustomModel(%q) = true, want false", m)
		}
	}
}

func TestIsProModel(t *testing.T) {
	if !IsProModel("gemini-3-pro-preview") {
		t.Error("expected true for gemini-3-pro-preview")
	}
	if !IsProModel("gemini-2.5-pro") {
		t.Error("expected true for gemini-2.5-pro")
	}
	if IsProModel("gemini-2.5-flash") {
		t.Error("expected false for gemini-2.5-flash")
	}
}

func TestIsPreviewModel(t *testing.T) {
	if !IsPreviewModel(PreviewGeminiModel) {
		t.Error("expected true for preview model")
	}
	if IsPreviewModel(DefaultGeminiModel) {
		t.Error("expected false for default model")
	}
}

func TestGetDisplayString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{PreviewGeminiModelAuto, "Auto (Gemini 3)"},
		{DefaultGeminiModelAuto, "Auto (Gemini 2.5)"},
		{GeminiModelAliasPro, PreviewGeminiModel},
		{GeminiModelAliasFlash, PreviewGeminiFlashModel},
		{DefaultGeminiModel, DefaultGeminiModel},
		{"custom-model", "custom-model"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := GetDisplayString(tt.input)
			if got != tt.want {
				t.Errorf("GetDisplayString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsAutoModel(t *testing.T) {
	if !IsAutoModel(GeminiModelAliasAuto) {
		t.Error("expected true for auto alias")
	}
	if !IsAutoModel(PreviewGeminiModelAuto) {
		t.Error("expected true for preview auto")
	}
	if IsAutoModel(DefaultGeminiModel) {
		t.Error("expected false for concrete model")
	}
}

func TestSupportsMultimodalFunctionResponse(t *testing.T) {
	if !SupportsMultimodalFunctionResponse("gemini-3-pro") {
		t.Error("expected true for gemini-3-pro")
	}
	if SupportsMultimodalFunctionResponse("gemini-2.5-pro") {
		t.Error("expected false for gemini-2.5-pro")
	}
}

func TestResolveClassifierModel(t *testing.T) {
	// Flash alias with default auto → default flash
	got := ResolveClassifierModel(DefaultGeminiModelAuto, GeminiModelAliasFlash, false, false)
	if got != DefaultGeminiFlashModel {
		t.Errorf("got %q, want %q", got, DefaultGeminiFlashModel)
	}

	// Flash alias with preview auto → preview flash
	got = ResolveClassifierModel(PreviewGeminiModelAuto, GeminiModelAliasFlash, false, false)
	if got != PreviewGeminiFlashModel {
		t.Errorf("got %q, want %q", got, PreviewGeminiFlashModel)
	}

	// Pro alias → delegates to ResolveModel
	got = ResolveClassifierModel(PreviewGeminiModelAuto, GeminiModelAliasPro, false, false)
	if got != PreviewGeminiModel {
		t.Errorf("got %q, want %q", got, PreviewGeminiModel)
	}
}

func TestIsActiveModel(t *testing.T) {
	// Default config: no gemini31
	if !IsActiveModel(DefaultGeminiModel, false, false) {
		t.Error("expected default model to be active")
	}
	if !IsActiveModel(PreviewGeminiModel, false, false) {
		t.Error("expected preview model to be active without gemini31")
	}
	if IsActiveModel(PreviewGemini31Model, false, false) {
		t.Error("expected 3.1 model to be inactive without gemini31")
	}
	if IsActiveModel("random-model", false, false) {
		t.Error("expected unknown model to be inactive")
	}
}

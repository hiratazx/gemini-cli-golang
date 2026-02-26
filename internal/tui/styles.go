// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package tui provides the interactive terminal user interface for Gemini CLI.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	colorPrimary    = lipgloss.Color("#4285F4") // Google Blue
	colorSecondary  = lipgloss.Color("#34A853") // Google Green
	colorAccent     = lipgloss.Color("#FBBC05") // Google Yellow
	colorDanger     = lipgloss.Color("#EA4335") // Google Red
	colorText       = lipgloss.Color("#E8EAED")
	colorSubtle     = lipgloss.Color("#9AA0A6")
	colorBackground = lipgloss.Color("#202124")
	colorSurface    = lipgloss.Color("#303134")
	colorBorder     = lipgloss.Color("#5F6368")
)

// Logo styles
var (
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	logoGemini = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Render("✦ Gemini CLI")

	versionStyle = lipgloss.NewStyle().
			Foreground(colorSubtle).
			PaddingLeft(1)
)

// Title styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Italic(true)
)

// Input styles
var (
	promptStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(colorText)
)

// Message styles
var (
	userMessageStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Bold(true).
				PaddingLeft(0)

	modelMessageStyle = lipgloss.NewStyle().
				Foreground(colorText).
				PaddingLeft(0)

	thinkingStyle = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)
)

// Selection styles
var (
	selectedStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	unselectedStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(colorSubtle).
				PaddingLeft(2)

	cursorStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)
)

// Status bar styles
var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorSubtle).
			PaddingLeft(1).
			PaddingRight(1)

	statusModelStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	statusInfoStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	statusEmailStyle = lipgloss.NewStyle().
				Foreground(colorSecondary)
)

// Box / panel styles
var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	highlightBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(1, 2)
)

// Spinner styles
var (
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)
)

// Help styles
var (
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)
)

// Divider
func divider(width int) string {
	line := ""
	for i := 0; i < width; i++ {
		line += "─"
	}
	return lipgloss.NewStyle().Foreground(colorBorder).Render(line)
}

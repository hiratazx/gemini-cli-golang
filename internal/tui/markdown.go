// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Markdown rendering patterns
var (
	headerRegex    = regexp.MustCompile(`(?m)^(#{1,3})\s+(.+)$`)
	boldRegex      = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRegex    = regexp.MustCompile(`\*(.+?)\*`)
	inlineCode     = regexp.MustCompile("`([^`]+)`")
	codeBlockRegex = regexp.MustCompile("(?s)```(\\w*)\\n(.*?)```")
	bulletRegex    = regexp.MustCompile(`(?m)^(\s*)[-*]\s+(.+)$`)
	numberedRegex  = regexp.MustCompile(`(?m)^(\s*)\d+\.\s+(.+)$`)
)

// Markdown styles
var (
	mdHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4285F4"))

	mdH1Style = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4285F4")).
			MarginBottom(1)

	mdH2Style = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#34A853"))

	mdBoldStyle = lipgloss.NewStyle().
			Bold(true)

	mdItalicStyle = lipgloss.NewStyle().
			Italic(true)

	mdCodeInline = lipgloss.NewStyle().
			Background(lipgloss.Color("#2d2d2d")).
			Foreground(lipgloss.Color("#e6db74")).
			Padding(0, 1)

	mdCodeBlock = lipgloss.NewStyle().
			Background(lipgloss.Color("#1e1e1e")).
			Foreground(lipgloss.Color("#d4d4d4")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	mdCodeLang = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	mdBulletStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#34A853"))
)

// RenderMarkdown converts markdown to styled terminal text.
func RenderMarkdown(text string) string {
	if text == "" {
		return ""
	}

	// Handle code blocks first (before other processing)
	text = codeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := codeBlockRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		lang := parts[1]
		code := strings.TrimRight(parts[2], "\n")

		var header string
		if lang != "" {
			header = mdCodeLang.Render("  "+lang) + "\n"
		}
		return header + mdCodeBlock.Render(code)
	})

	// Process line by line for headers and lists
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		// Headers
		if headerRegex.MatchString(line) {
			parts := headerRegex.FindStringSubmatch(line)
			level := len(parts[1])
			content := parts[2]
			switch level {
			case 1:
				line = mdH1Style.Render("█ " + content)
			case 2:
				line = mdH2Style.Render("▌ " + content)
			default:
				line = mdHeaderStyle.Render("▎ " + content)
			}
		}

		// Bullet lists
		if bulletRegex.MatchString(line) {
			parts := bulletRegex.FindStringSubmatch(line)
			indent := parts[1]
			content := parts[2]
			line = indent + mdBulletStyle.Render("●") + " " + content
		}

		// Numbered lists
		if numberedRegex.MatchString(line) {
			parts := numberedRegex.FindStringSubmatch(line)
			indent := parts[1]
			content := parts[2]
			line = indent + mdBulletStyle.Render("▸") + " " + content
		}

		// Inline code
		line = inlineCode.ReplaceAllStringFunc(line, func(match string) string {
			code := inlineCode.FindStringSubmatch(match)[1]
			return mdCodeInline.Render(code)
		})

		// Bold
		line = boldRegex.ReplaceAllStringFunc(line, func(match string) string {
			text := boldRegex.FindStringSubmatch(match)[1]
			return mdBoldStyle.Render(text)
		})

		// Italic (only non-bold asterisks)
		line = italicRegex.ReplaceAllStringFunc(line, func(match string) string {
			text := italicRegex.FindStringSubmatch(match)[1]
			return mdItalicStyle.Render(text)
		})

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

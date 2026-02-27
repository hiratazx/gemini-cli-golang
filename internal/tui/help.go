// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpContent returns the help overlay text.
func HelpContent() string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4285F4"))

	cmdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E8EAED")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9AA0A6"))

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#34A853")).
		Bold(true)

	sb.WriteString(titleStyle.Render("✦ Gemini CLI Help") + "\n\n")

	sb.WriteString(sectionStyle.Render("Slash Commands") + "\n")
	commands := []struct {
		cmd  string
		desc string
	}{
		{"/help", "Show this help message"},
		{"/model", "Switch to a different model"},
		{"/clear", "Clear conversation history"},
		{"/cd <path>", "Change working directory"},
		{"/history", "Show conversation history"},
		{"/quit", "Exit the application"},
	}
	for _, c := range commands {
		sb.WriteString("  " + cmdStyle.Render(c.cmd))
		padding := strings.Repeat(" ", 14-len(c.cmd))
		sb.WriteString(padding + descStyle.Render(c.desc) + "\n")
	}

	sb.WriteString("\n" + sectionStyle.Render("Keyboard Shortcuts") + "\n")
	shortcuts := []struct {
		key  string
		desc string
	}{
		{"Enter", "Send message"},
		{"PgUp/PgDn", "Scroll chat history"},
		{"Ctrl+C", "Quit application"},
		{"↑/↓", "Navigate menus"},
	}
	for _, s := range shortcuts {
		sb.WriteString("  " + cmdStyle.Render(s.key))
		padding := strings.Repeat(" ", 14-len(s.key))
		sb.WriteString(padding + descStyle.Render(s.desc) + "\n")
	}

	sb.WriteString("\n" + sectionStyle.Render("Available Tools") + "\n")
	toolList := []struct {
		name string
		desc string
	}{
		{"read_file", "Read file contents"},
		{"write_new_file", "Create new files"},
		{"edit", "Edit existing files"},
		{"shell", "Execute shell commands"},
		{"grep", "Search file contents"},
		{"glob", "Find files by pattern"},
		{"ls", "List directory contents"},
	}
	for _, t := range toolList {
		sb.WriteString("  " + cmdStyle.Render(t.name))
		padding := strings.Repeat(" ", 18-len(t.name))
		sb.WriteString(padding + descStyle.Render(t.desc) + "\n")
	}

	return sb.String()
}

// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package main is the entry point for the Gemini CLI.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/google-gemini/gemini-cli/internal/tui"
)

var version = "dev"

func main() {
	args := os.Args[1:]

	// Handle flags
	for _, arg := range args {
		switch arg {
		case "--version", "-v":
			fmt.Printf("Gemini CLI %s (Go)\n", version)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	// If arguments provided, run in non-interactive mode
	if len(args) > 0 {
		prompt := strings.Join(args, " ")
		if err := tui.RunNonInteractive(prompt); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		return
	}

	// Interactive mode
	if err := tui.Run(""); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`
  ✦ Gemini CLI (Go Edition)

  Usage:
    gemini                    Launch interactive TUI
    gemini <prompt>           Send a single prompt (non-interactive)
    gemini --version          Show version

  Environment Variables:
    GEMINI_API_KEY            Gemini API key (skip interactive auth)
    GOOGLE_GENAI_USE_GCA      Set to "true" to use Google OAuth

  Authentication:
    On first run, you'll be prompted to authenticate via:
    1. Login with Google (OAuth - recommended)
    2. API Key from aistudio.google.com

  Examples:
    gemini
    gemini "explain this codebase"
    GEMINI_API_KEY=xxx gemini "hello"`)
}

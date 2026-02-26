// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package main is the entry point for the Gemini CLI.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google-gemini/gemini-cli/internal/config"
	"github.com/google-gemini/gemini-cli/internal/tui"
)

var version = "dev"

func main() {
	// CLI flags
	modelFlag := flag.String("model", "", "Model to use (e.g., gemini-2.5-pro, gemini-2.5-flash)")
	nonInteractive := flag.Bool("non-interactive", false, "Run in non-interactive mode (requires prompt)")
	showVersion := flag.Bool("version", false, "Show version")
	showHelp := flag.Bool("help", false, "Show help")

	// Short flags
	flag.StringVar(modelFlag, "m", "", "Model to use (shorthand)")
	flag.BoolVar(showVersion, "v", false, "Show version (shorthand)")
	flag.BoolVar(showHelp, "h", false, "Show help (shorthand)")

	flag.Parse()

	if *showVersion {
		fmt.Printf("Gemini CLI %s (Go)\n", version)
		os.Exit(0)
	}

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Load settings
	settings := config.LoadSettings()

	// Determine model
	model := settings.DefaultModel
	if *modelFlag != "" {
		model = *modelFlag
	}

	// Non-flag arguments are the prompt
	prompt := strings.Join(flag.Args(), " ")

	// Non-interactive mode
	if *nonInteractive || prompt != "" {
		if prompt == "" {
			fmt.Fprintln(os.Stderr, "Error: prompt required in non-interactive mode")
			os.Exit(1)
		}
		if err := tui.RunNonInteractive(prompt); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		return
	}

	// Interactive mode
	if err := tui.Run("", model); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`
  ✦ Gemini CLI (Go Edition)

  Usage:
    gemini                        Launch interactive TUI
    gemini <prompt>               Send a single prompt (non-interactive)
    gemini -m <model> <prompt>    Use specific model
    gemini --version              Show version

  Flags:
    -m, --model <model>           Model to use (e.g., gemini-2.5-flash)
    -n, --non-interactive         Force non-interactive mode
    -v, --version                 Show version
    -h, --help                    Show this help

  Slash Commands (in interactive mode):
    /help                         Show help
    /model                        Switch model
    /clear                        Clear conversation
    /history                      Show past sessions
    /quit                         Exit

  Environment Variables:
    GEMINI_API_KEY                Gemini API key (skip interactive auth)

  Examples:
    gemini
    gemini "explain this codebase"
    gemini -m gemini-2.5-flash "hello"
    GEMINI_API_KEY=xxx gemini "hello"`)
}

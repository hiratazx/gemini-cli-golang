// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package main is the entry point for the Gemini CLI.
package main

import (
	"fmt"
	"os"

	"github.com/google-gemini/gemini-cli/internal/config"
)

func main() {
	fmt.Println("Gemini CLI - Go Edition")
	fmt.Printf("Default model: %s\n", config.DefaultGeminiModel)

	if len(os.Args) < 2 {
		fmt.Println("Usage: gemini <prompt>")
		fmt.Println("  Bring the power of Gemini directly into your terminal.")
		os.Exit(0)
	}
}

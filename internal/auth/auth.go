// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package auth provides authentication for Gemini CLI.
package auth

import (
	"fmt"
	"os"
)

// AuthType represents the method of authentication.
type AuthType string

const (
	// AuthTypeOAuth uses Google OAuth for authentication ("Login with Google").
	AuthTypeOAuth AuthType = "oauth-personal"
	// AuthTypeAPIKey uses a Gemini API key.
	AuthTypeAPIKey AuthType = "gemini-api-key"
	// AuthTypeVertexAI uses Vertex AI authentication.
	AuthTypeVertexAI AuthType = "vertex-ai"
)

// Credentials holds authentication credentials.
type Credentials struct {
	AuthType     AuthType
	APIKey       string
	AccessToken  string
	RefreshToken string
	Email        string
}

// DetectAuthType detects the best authentication type from environment variables.
// Priority: GOOGLE_GENAI_USE_GCA → GEMINI_API_KEY → prompt user.
func DetectAuthType() (AuthType, string) {
	if os.Getenv("GOOGLE_GENAI_USE_GCA") == "true" {
		return AuthTypeOAuth, ""
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return AuthTypeAPIKey, key
	}
	// Default: needs user interaction
	return "", ""
}

// Authenticate performs authentication based on the auth type.
func Authenticate(authType AuthType, apiKey string) (*Credentials, error) {
	switch authType {
	case AuthTypeAPIKey:
		if apiKey == "" {
			return nil, fmt.Errorf("API key is required")
		}
		return &Credentials{
			AuthType: AuthTypeAPIKey,
			APIKey:   apiKey,
		}, nil
	case AuthTypeOAuth:
		return authenticateOAuth()
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}
}

// GetAuthOptions returns the available authentication options for the user.
func GetAuthOptions() []AuthOption {
	return []AuthOption{
		{
			Type:        AuthTypeOAuth,
			Label:       "Login with Google",
			Description: "Sign in with your Google account via browser (recommended)",
		},
		{
			Type:        AuthTypeAPIKey,
			Label:       "Use API Key",
			Description: "Enter a Gemini API key from aistudio.google.com",
		},
	}
}

// AuthOption represents a selectable authentication method.
type AuthOption struct {
	Type        AuthType
	Label       string
	Description string
}

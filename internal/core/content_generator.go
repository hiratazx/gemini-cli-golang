// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"fmt"
)

// ContentGenerator defines the interface for generating content from the Gemini API.
type ContentGenerator interface {
	// GenerateContent sends a request and returns the full response.
	GenerateContent(ctx context.Context, model string, contents []Content, config *GenerateContentConfig, tools []Tool, systemInstruction string) (*GenerateContentResponse, error)

	// GenerateContentStream sends a request and returns a stream of response chunks.
	GenerateContentStream(ctx context.Context, model string, contents []Content, config *GenerateContentConfig, tools []Tool, systemInstruction string) (<-chan StreamChunk, error)

	// CountTokens estimates the token count for the given content.
	CountTokens(ctx context.Context, model string, contents []Content) (int, error)
}

// StreamChunk represents a chunk from a streaming response.
type StreamChunk struct {
	// Response is the partial response data.
	Response *GenerateContentResponse
	// Error is set if the chunk represents an error.
	Error error
	// Done indicates this is the final chunk.
	Done bool
}

// AuthType represents the authentication method.
type AuthType string

const (
	// AuthTypeAPIKey uses an API key for authentication.
	AuthTypeAPIKey AuthType = "api_key"
	// AuthTypeOAuth uses OAuth for authentication.
	AuthTypeOAuth AuthType = "oauth"
)

// ContentGeneratorConfig holds configuration for creating a ContentGenerator.
type ContentGeneratorConfig struct {
	// APIKey is the Gemini API key (for API key auth).
	APIKey string
	// AuthType specifies the authentication method.
	AuthType AuthType
	// ProjectID is the Google Cloud project ID (for OAuth/service account auth).
	ProjectID string
	// Location is the Google Cloud location (e.g., "us-central1").
	Location string
	// Proxy is an optional HTTP proxy URL.
	Proxy string
}

// NewContentGenerator creates a ContentGenerator based on the configuration.
// This is a factory function that will dispatch to the appropriate implementation.
func NewContentGenerator(cfg ContentGeneratorConfig) (ContentGenerator, error) {
	switch cfg.AuthType {
	case AuthTypeAPIKey:
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("API key is required for API key authentication")
		}
		return newAPIKeyContentGenerator(cfg)
	case AuthTypeOAuth:
		return newOAuthContentGenerator(cfg)
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", cfg.AuthType)
	}
}

// apiKeyContentGenerator implements ContentGenerator using an API key.
type apiKeyContentGenerator struct {
	apiKey string
	proxy  string
}

func newAPIKeyContentGenerator(cfg ContentGeneratorConfig) (*apiKeyContentGenerator, error) {
	return &apiKeyContentGenerator{
		apiKey: cfg.APIKey,
		proxy:  cfg.Proxy,
	}, nil
}

func (g *apiKeyContentGenerator) GenerateContent(ctx context.Context, model string, contents []Content, config *GenerateContentConfig, tools []Tool, systemInstruction string) (*GenerateContentResponse, error) {
	// TODO: Implement actual API call using net/http
	return nil, fmt.Errorf("GenerateContent not yet implemented")
}

func (g *apiKeyContentGenerator) GenerateContentStream(ctx context.Context, model string, contents []Content, config *GenerateContentConfig, tools []Tool, systemInstruction string) (<-chan StreamChunk, error) {
	// TODO: Implement actual streaming API call
	return nil, fmt.Errorf("GenerateContentStream not yet implemented")
}

func (g *apiKeyContentGenerator) CountTokens(ctx context.Context, model string, contents []Content) (int, error) {
	// TODO: Implement token counting API call
	return 0, fmt.Errorf("CountTokens not yet implemented")
}

// oauthContentGenerator implements ContentGenerator using OAuth.
type oauthContentGenerator struct {
	projectID string
	location  string
	proxy     string
}

func newOAuthContentGenerator(cfg ContentGeneratorConfig) (*oauthContentGenerator, error) {
	return &oauthContentGenerator{
		projectID: cfg.ProjectID,
		location:  cfg.Location,
		proxy:     cfg.Proxy,
	}, nil
}

func (g *oauthContentGenerator) GenerateContent(ctx context.Context, model string, contents []Content, config *GenerateContentConfig, tools []Tool, systemInstruction string) (*GenerateContentResponse, error) {
	// TODO: Implement actual API call with OAuth
	return nil, fmt.Errorf("GenerateContent not yet implemented")
}

func (g *oauthContentGenerator) GenerateContentStream(ctx context.Context, model string, contents []Content, config *GenerateContentConfig, tools []Tool, systemInstruction string) (<-chan StreamChunk, error) {
	// TODO: Implement actual streaming API call with OAuth
	return nil, fmt.Errorf("GenerateContentStream not yet implemented")
}

func (g *oauthContentGenerator) CountTokens(ctx context.Context, model string, contents []Content) (int, error) {
	// TODO: Implement token counting API call with OAuth
	return 0, fmt.Errorf("CountTokens not yet implemented")
}

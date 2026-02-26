// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package genai wraps the Google GenAI Go SDK for Gemini CLI usage.
package genai

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"google.golang.org/genai"

	"github.com/google-gemini/gemini-cli/internal/auth"
)

// Client wraps the Google GenAI SDK client.
type Client struct {
	client *genai.Client
	creds  *auth.Credentials
}

// NewClient creates a new GenAI client from credentials.
func NewClient(ctx context.Context, creds *auth.Credentials) (*Client, error) {
	var client *genai.Client
	var err error

	switch creds.AuthType {
	case auth.AuthTypeAPIKey:
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  creds.APIKey,
			Backend: genai.BackendGeminiAPI,
		})
	case auth.AuthTypeOAuth:
		// For OAuth, we use the access token via HTTP headers
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  creds.AccessToken,
			Backend: genai.BackendGeminiAPI,
			HTTPOptions: genai.HTTPOptions{
				Headers: map[string][]string{
					"Authorization": {"Bearer " + creds.AccessToken},
				},
			},
		})
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", creds.AuthType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GenAI client: %w", err)
	}

	return &Client{client: client, creds: creds}, nil
}

// StreamResponse holds a chunk of streaming response.
type StreamResponse struct {
	Text         string
	Thought      string
	Done         bool
	FinishReason string
	Error        error

	// Token usage
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// GenerateContentStream sends a message and streams the response.
func (c *Client) GenerateContentStream(ctx context.Context, model string, systemInstruction string, history []Message, message string) <-chan StreamResponse {
	ch := make(chan StreamResponse, 100)

	go func() {
		defer close(ch)

		contents := buildContents(history, message)

		config := &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
			},
		}

		result := c.client.Models.GenerateContentStream(ctx, model, contents, config)

		for resp, err := range result {
			if err != nil {
				ch <- StreamResponse{Error: err, Done: true}
				return
			}

			sr := StreamResponse{}

			if resp.UsageMetadata != nil {
				sr.InputTokens = int(resp.UsageMetadata.PromptTokenCount)
				sr.OutputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
				sr.TotalTokens = int(resp.UsageMetadata.TotalTokenCount)
			}

			if len(resp.Candidates) > 0 {
				candidate := resp.Candidates[0]
				if candidate.FinishReason != "" {
					sr.FinishReason = string(candidate.FinishReason)
				}

				if candidate.Content != nil {
					for _, part := range candidate.Content.Parts {
						if part.Text != "" {
							if part.Thought {
								sr.Thought = part.Text
							} else {
								sr.Text = part.Text
							}
						}
					}
				}
			}

			ch <- sr
		}

		ch <- StreamResponse{Done: true}
	}()

	return ch
}

// GenerateContent sends a message and returns the full response (non-streaming).
func (c *Client) GenerateContent(ctx context.Context, model string, systemInstruction string, history []Message, message string) (string, error) {
	contents := buildContents(history, message)

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
		},
	}

	resp, err := c.client.Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return "", err
	}

	return resp.Text(), nil
}

// Message represents a conversation message.
type Message struct {
	Role string // "user" or "model"
	Text string
}

// buildContents converts the message history + new message into genai Content.
func buildContents(history []Message, newMessage string) []*genai.Content {
	var contents []*genai.Content

	for _, msg := range history {
		contents = append(contents, &genai.Content{
			Role:  msg.Role,
			Parts: []*genai.Part{genai.NewPartFromText(msg.Text)},
		})
	}

	// Add the new user message
	contents = append(contents, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(newMessage)},
	})

	return contents
}

// GetDefaultSystemPrompt returns the default system prompt.
func GetDefaultSystemPrompt(cwd string) string {
	var sb strings.Builder
	sb.WriteString("You are Gemini CLI, an interactive AI assistant in the terminal.\n")
	sb.WriteString("You help users with coding, file operations, and general questions.\n\n")
	sb.WriteString("## Environment\n")
	sb.WriteString(fmt.Sprintf("- Working directory: %s\n", cwd))
	sb.WriteString(fmt.Sprintf("- OS: %s/%s\n", runtime.GOOS, runtime.GOARCH))
	sb.WriteString("\n")
	sb.WriteString("## Guidelines\n")
	sb.WriteString("- Be concise and helpful.\n")
	sb.WriteString("- Use markdown formatting in responses.\n")
	sb.WriteString("- When showing code, use fenced code blocks with the language.\n")
	return sb.String()
}

// ListModels returns available model names.
func ListModels() []ModelInfo {
	return []ModelInfo{
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", Description: "Best for complex tasks (default)", Default: true},
		{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", Description: "Fast and efficient"},
		{ID: "gemini-2.5-flash-lite", Name: "Gemini 2.5 Flash Lite", Description: "Lightest and fastest"},
		{ID: "gemini-3-pro-preview", Name: "Gemini 3 Pro (Preview)", Description: "Latest preview model"},
		{ID: "gemini-3-flash-preview", Name: "Gemini 3 Flash (Preview)", Description: "Latest fast preview model"},
	}
}

// ModelInfo holds information about a model.
type ModelInfo struct {
	ID          string
	Name        string
	Description string
	Default     bool
}

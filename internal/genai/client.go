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
	client     *genai.Client
	creds      *auth.Credentials
	toolBridge *ToolBridge
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

// SetToolBridge attaches a tool bridge for function calling.
func (c *Client) SetToolBridge(tb *ToolBridge) {
	c.toolBridge = tb
}

// FunctionCallInfo holds information about a function call from the model.
type FunctionCallInfo struct {
	Name string
	Args map[string]interface{}
	ID   string
}

// StreamResponse holds a chunk of streaming response.
type StreamResponse struct {
	Text         string
	Thought      string
	Done         bool
	FinishReason string
	Error        error

	// Tool calls
	FunctionCalls []FunctionCallInfo
	ToolStatus    string // status message for tool execution

	// Token usage
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// GenerateContentStream sends a message and streams the response.
// It handles function call loops: if the model calls tools, it executes them
// and sends results back automatically.
func (c *Client) GenerateContentStream(ctx context.Context, model string, systemInstruction string, history []Message, message string) <-chan StreamResponse {
	ch := make(chan StreamResponse, 100)

	go func() {
		defer close(ch)

		contents := buildContents(history, message)
		c.streamWithToolLoop(ctx, model, systemInstruction, contents, ch)
	}()

	return ch
}

// streamWithToolLoop handles the generate → tool call → generate loop.
func (c *Client) streamWithToolLoop(ctx context.Context, model string, systemInstruction string, contents []*genai.Content, ch chan<- StreamResponse) {
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
		},
	}

	// Add tool declarations if bridge is available
	if c.toolBridge != nil {
		config.Tools = c.toolBridge.GetFunctionDeclarations(model)
	}

	maxToolRounds := 10

	for round := 0; round < maxToolRounds; round++ {
		// Accumulate the full response to detect function calls
		var allText string
		var allThought string
		var functionCalls []FunctionCallInfo
		var lastTokenInfo StreamResponse
		var responseParts []*genai.Part

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
				lastTokenInfo = sr
			}

			if len(resp.Candidates) > 0 {
				candidate := resp.Candidates[0]
				if candidate.FinishReason != "" {
					sr.FinishReason = string(candidate.FinishReason)
				}

				if candidate.Content != nil {
					for _, part := range candidate.Content.Parts {
						responseParts = append(responseParts, part)

						if part.Text != "" {
							if part.Thought {
								sr.Thought = part.Text
								allThought = part.Text
							} else {
								sr.Text = part.Text
								allText += part.Text
							}
						}

						if part.FunctionCall != nil {
							fc := FunctionCallInfo{
								Name: part.FunctionCall.Name,
								ID:   part.FunctionCall.ID,
							}
							if part.FunctionCall.Args != nil {
								fc.Args = part.FunctionCall.Args
							}
							functionCalls = append(functionCalls, fc)
						}
					}
				}
			}

			// Send text/thought chunks as they arrive (but not Done yet)
			if sr.Text != "" || sr.Thought != "" {
				ch <- sr
			}
		}

		// No function calls → we're done
		if len(functionCalls) == 0 {
			ch <- StreamResponse{
				Done:         true,
				InputTokens:  lastTokenInfo.InputTokens,
				OutputTokens: lastTokenInfo.OutputTokens,
				TotalTokens:  lastTokenInfo.TotalTokens,
			}
			return
		}

		// Execute function calls
		if c.toolBridge == nil {
			ch <- StreamResponse{
				Text: "\n\n⚠️ Model requested tool calls but no tool bridge is configured.",
				Done: true,
			}
			return
		}

		// Add model's response (with function calls) to contents
		contents = append(contents, &genai.Content{
			Role:  "model",
			Parts: responseParts,
		})

		// Execute each function call and collect results
		var resultParts []*genai.Part
		for _, fc := range functionCalls {
			displayName := c.toolBridge.GetToolDisplayName(fc.Name)
			ch <- StreamResponse{
				ToolStatus: fmt.Sprintf("🔧 %s", displayName),
			}

			toolResult, err := c.toolBridge.ExecuteFunctionCall(ctx, fc.Name, fc.Args)

			var responseData map[string]interface{}
			if err != nil {
				responseData = map[string]interface{}{
					"error": err.Error(),
				}
			} else if toolResult.Error != nil {
				responseData = map[string]interface{}{
					"error": toolResult.Error.Message,
				}
			} else {
				responseData = map[string]interface{}{
					"result": toolResult.LLMContent,
				}
			}

			resultParts = append(resultParts, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					Name:     fc.Name,
					Response: responseData,
					ID:       fc.ID,
				},
			})

			// Show tool result summary
			summary := ""
			if toolResult != nil && toolResult.ReturnDisplay != "" {
				summary = toolResult.ReturnDisplay
			}
			ch <- StreamResponse{
				ToolStatus: fmt.Sprintf("✓ %s → %s", displayName, summary),
			}
		}

		// Add tool results to contents
		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: resultParts,
		})

		// Clear accumulated text for next round
		_ = allText
		_ = allThought
	}

	ch <- StreamResponse{
		Text: "\n\n⚠️ Maximum tool call rounds reached.",
		Done: true,
	}
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
	sb.WriteString("## Tools\n")
	sb.WriteString("You have access to tools for reading files, writing files, editing files, ")
	sb.WriteString("running shell commands, searching files with grep, finding files with glob, ")
	sb.WriteString("and listing directories. Use them when the user asks about files or code.\n\n")
	sb.WriteString("## Guidelines\n")
	sb.WriteString("- Be concise and helpful.\n")
	sb.WriteString("- Use markdown formatting in responses.\n")
	sb.WriteString("- When showing code, use fenced code blocks with the language.\n")
	sb.WriteString("- Use tools to read files before answering questions about code.\n")
	sb.WriteString("- Use the edit tool for targeted changes, write_new_file for new files.\n")
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

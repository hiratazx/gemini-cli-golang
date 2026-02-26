// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package genai

import (
	"context"
	"fmt"

	"github.com/google-gemini/gemini-cli/internal/config"
	"github.com/google-gemini/gemini-cli/internal/tools"
	"google.golang.org/genai"
)

// ToolBridge bridges the GenAI SDK function calling with the tool registry.
type ToolBridge struct {
	registry *tools.ToolRegistry
	cfg      *config.Config
}

// NewToolBridge creates a new ToolBridge.
func NewToolBridge(cfg *config.Config) *ToolBridge {
	reg := tools.NewToolRegistry()

	// Register built-in tools
	reg.RegisterTool(tools.NewReadFileTool(cfg))
	reg.RegisterTool(tools.NewWriteFileTool(cfg))
	reg.RegisterTool(tools.NewEditTool(cfg))
	reg.RegisterTool(tools.NewShellTool(cfg))
	reg.RegisterTool(tools.NewGrepTool(cfg))
	reg.RegisterTool(tools.NewGlobTool(cfg))
	reg.RegisterTool(tools.NewLSTool(cfg))

	return &ToolBridge{registry: reg, cfg: cfg}
}

// GetFunctionDeclarations returns GenAI SDK tool declarations.
func (tb *ToolBridge) GetFunctionDeclarations(modelID string) []*genai.Tool {
	decls := tb.registry.GetFunctionDeclarations(modelID)
	if len(decls) == 0 {
		return nil
	}

	var sdkDecls []*genai.FunctionDeclaration
	for _, d := range decls {
		sdkDecl := &genai.FunctionDeclaration{
			Name:        d.Name,
			Description: d.Description,
		}
		if d.ParametersJSONSchema != nil {
			sdkDecl.Parameters = convertToSchema(d.ParametersJSONSchema)
		}
		sdkDecls = append(sdkDecls, sdkDecl)
	}

	return []*genai.Tool{{FunctionDeclarations: sdkDecls}}
}

// ExecuteFunctionCall executes a function call from the model.
func (tb *ToolBridge) ExecuteFunctionCall(ctx context.Context, name string, args map[string]interface{}) (*tools.ToolResult, error) {
	tool := tb.registry.GetTool(name)
	if tool == nil {
		return tools.ErrorResult(tools.ToolErrorTypeGeneral,
			fmt.Sprintf("Unknown tool: %s", name)), nil
	}

	inv, err := tool.CreateInvocation(args)
	if err != nil {
		return tools.ErrorResult(tools.ToolErrorTypeGeneral, err.Error()), nil
	}

	return inv.Execute(ctx)
}

// GetToolDisplayName returns a display name for a tool call.
func (tb *ToolBridge) GetToolDisplayName(name string) string {
	tool := tb.registry.GetTool(name)
	if tool == nil {
		return name
	}
	return tool.DisplayName()
}

// convertToSchema converts a JSON schema map to a genai.Schema.
func convertToSchema(schema map[string]interface{}) *genai.Schema {
	s := &genai.Schema{}

	if t, ok := schema["type"].(string); ok {
		switch t {
		case "object":
			s.Type = genai.TypeObject
		case "string":
			s.Type = genai.TypeString
		case "integer":
			s.Type = genai.TypeInteger
		case "number":
			s.Type = genai.TypeNumber
		case "boolean":
			s.Type = genai.TypeBoolean
		case "array":
			s.Type = genai.TypeArray
		}
	}

	if desc, ok := schema["description"].(string); ok {
		s.Description = desc
	}

	if props, ok := schema["properties"].(map[string]interface{}); ok {
		s.Properties = make(map[string]*genai.Schema)
		for k, v := range props {
			if propMap, ok := v.(map[string]interface{}); ok {
				s.Properties[k] = convertToSchema(propMap)
			}
		}
	}

	if required, ok := schema["required"].([]string); ok {
		s.Required = required
	}
	// Handle []interface{} for required (from JSON)
	if required, ok := schema["required"].([]interface{}); ok {
		for _, r := range required {
			if rs, ok := r.(string); ok {
				s.Required = append(s.Required, rs)
			}
		}
	}

	if items, ok := schema["items"].(map[string]interface{}); ok {
		s.Items = convertToSchema(items)
	}

	return s
}

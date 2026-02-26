// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"sort"
	"strings"
	"sync"

	"github.com/google-gemini/gemini-cli/internal/core"
)

// ToolRegistry manages the registration and lookup of tools.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]DeclarativeTool
	order []string // maintains insertion order
}

// NewToolRegistry creates a new empty ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]DeclarativeTool),
	}
}

// RegisterTool registers a tool definition.
func (r *ToolRegistry) RegisterTool(tool DeclarativeTool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if _, exists := r.tools[name]; !exists {
		r.order = append(r.order, name)
	}
	r.tools[name] = tool
}

// UnregisterTool removes a tool by name.
func (r *ToolRegistry) UnregisterTool(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.tools, name)
	for i, n := range r.order {
		if n == name {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
}

// GetTool returns a tool by name, resolving legacy aliases.
func (r *ToolRegistry) GetTool(name string) DeclarativeTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if tool, ok := r.tools[name]; ok {
		return tool
	}

	// Try legacy alias
	if canonical, ok := ToolLegacyAliases[name]; ok {
		return r.tools[canonical]
	}

	return nil
}

// GetFunctionDeclarations returns all function declarations for the API.
func (r *ToolRegistry) GetFunctionDeclarations(modelID string) []core.FunctionDeclaration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	declarations := make([]core.FunctionDeclaration, 0, len(r.tools))
	for _, name := range r.order {
		if tool, ok := r.tools[name]; ok {
			declarations = append(declarations, tool.GetSchema(modelID))
		}
	}
	return declarations
}

// GetAllTools returns all registered tools.
func (r *ToolRegistry) GetAllTools() []DeclarativeTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]DeclarativeTool, 0, len(r.tools))
	for _, name := range r.order {
		if tool, ok := r.tools[name]; ok {
			tools = append(tools, tool)
		}
	}
	return tools
}

// GetToolNames returns names of all registered tools.
func (r *ToolRegistry) GetToolNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, len(r.order))
	copy(names, r.order)
	return names
}

// SortTools sorts tools: built-in first, then discovered, then MCP tools by server name.
func (r *ToolRegistry) SortTools() {
	r.mu.Lock()
	defer r.mu.Unlock()

	sort.SliceStable(r.order, func(i, j int) bool {
		pi := getToolPriority(r.order[i])
		pj := getToolPriority(r.order[j])
		return pi < pj
	})
}

func getToolPriority(name string) int {
	// Built-in tools
	for _, builtIn := range AllBuiltinToolNames {
		if name == builtIn {
			return 0
		}
	}
	// Discovered tools
	if strings.HasPrefix(name, DiscoveredToolPrefix) {
		return 1
	}
	// MCP tools
	return 2
}

// RemoveDiscoveredTools removes all dynamically discovered tools.
func (r *ToolRegistry) RemoveDiscoveredTools() {
	r.mu.Lock()
	defer r.mu.Unlock()

	var newOrder []string
	for _, name := range r.order {
		if strings.HasPrefix(name, DiscoveredToolPrefix) {
			delete(r.tools, name)
		} else {
			newOrder = append(newOrder, name)
		}
	}
	r.order = newOrder
}

// RemoveMCPToolsByServer removes all tools from a specific MCP server.
func (r *ToolRegistry) RemoveMCPToolsByServer(serverName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	prefix := serverName + "__"
	var newOrder []string
	for _, name := range r.order {
		if strings.HasPrefix(name, prefix) {
			delete(r.tools, name)
		} else {
			newOrder = append(newOrder, name)
		}
	}
	r.order = newOrder
}

// Count returns the number of registered tools.
func (r *ToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

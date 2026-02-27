// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

// Package mcp provides a basic MCP (Model Context Protocol) client.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

// MCPClient communicates with an MCP server over stdio.
type MCPClient struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  *bufio.Reader
	nextID  atomic.Int64
	pending map[int64]chan *JSONRPCResponse
	closed  bool
	name    string
}

// NewMCPClient creates a new MCP client for a stdio-based server.
func NewMCPClient(name string) *MCPClient {
	return &MCPClient{
		name:    name,
		pending: make(map[int64]chan *JSONRPCResponse),
	}
}

// Connect starts the MCP server process and performs initialization.
func (c *MCPClient) Connect(ctx context.Context, command string, args []string, env []string) error {
	c.cmd = exec.CommandContext(ctx, command, args...)
	if len(env) > 0 {
		c.cmd.Env = env
	}

	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	stdoutPipe, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	c.stdout = bufio.NewReader(stdoutPipe)

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("start MCP server %q: %w", command, err)
	}

	// Read responses in background
	go c.readLoop()

	// Send initialize request
	initResult, err := c.call(ctx, "initialize", map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "gemini-cli-go",
			"version": "1.0.0",
		},
	})
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	_ = initResult

	// Send initialized notification
	c.notify("notifications/initialized", nil)

	return nil
}

// ListTools returns the tools available from the MCP server.
func (c *MCPClient) ListTools(ctx context.Context) ([]MCPTool, error) {
	result, err := c.call(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}

	var toolList struct {
		Tools []MCPTool `json:"tools"`
	}

	data, _ := json.Marshal(result)
	if err := json.Unmarshal(data, &toolList); err != nil {
		return nil, fmt.Errorf("parse tools list: %w", err)
	}

	return toolList.Tools, nil
}

// CallTool invokes a tool on the MCP server.
func (c *MCPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (*MCPToolResult, error) {
	params := map[string]interface{}{
		"name": name,
	}
	if args != nil {
		params["arguments"] = args
	}

	result, err := c.call(ctx, "tools/call", params)
	if err != nil {
		return nil, err
	}

	var toolResult MCPToolResult
	data, _ := json.Marshal(result)
	if err := json.Unmarshal(data, &toolResult); err != nil {
		return nil, fmt.Errorf("parse tool result: %w", err)
	}

	return &toolResult, nil
}

// Close shuts down the MCP server.
func (c *MCPClient) Close() error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()

	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}
	return nil
}

// Name returns the server name.
func (c *MCPClient) Name() string {
	return c.name
}

// call sends a JSON-RPC request and waits for the response.
func (c *MCPClient) call(ctx context.Context, method string, params interface{}) (interface{}, error) {
	id := c.nextID.Add(1)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	respCh := make(chan *JSONRPCResponse, 1)

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client closed")
	}
	c.pending[id] = respCh
	c.mu.Unlock()

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	_, err = fmt.Fprintf(c.stdin, "%s\n", data)
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	select {
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	}
}

// notify sends a JSON-RPC notification (no response expected).
func (c *MCPClient) notify(method string, params interface{}) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	data, _ := json.Marshal(req)

	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Fprintf(c.stdin, "%s\n", data)
}

// readLoop reads JSON-RPC responses from stdout.
func (c *MCPClient) readLoop() {
	for {
		line, err := c.stdout.ReadBytes('\n')
		if err != nil {
			c.mu.Lock()
			for _, ch := range c.pending {
				close(ch)
			}
			c.pending = make(map[int64]chan *JSONRPCResponse)
			c.mu.Unlock()
			return
		}

		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		// Only dispatch responses with IDs (not notifications)
		if resp.ID > 0 {
			c.mu.Lock()
			if ch, ok := c.pending[resp.ID]; ok {
				ch <- &resp
				delete(c.pending, resp.ID)
			}
			c.mu.Unlock()
		}
	}
}

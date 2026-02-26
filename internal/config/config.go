// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ApprovalMode controls how tool executions are approved.
type ApprovalMode int

const (
	// ApprovalModeDefault requires confirmation for modifying tools.
	ApprovalModeDefault ApprovalMode = iota
	// ApprovalModeYolo auto-approves all tool executions.
	ApprovalModeYolo
	// ApprovalModePlan only allows read-only tools.
	ApprovalModePlan
)

// OutputFormat specifies the output format.
type OutputFormat string

const (
	OutputFormatText     OutputFormat = "text"
	OutputFormatJSON     OutputFormat = "json"
	OutputFormatMarkdown OutputFormat = "markdown"
)

// AuthType represents authentication types.
type AuthType string

const (
	AuthTypeAPIKey          AuthType = "api_key"
	AuthTypeOAuth           AuthType = "oauth"
	AuthTypeServiceAccount  AuthType = "service_account"
	AuthTypeGoogleCloud     AuthType = "google_cloud"
)

// DefaultTruncateToolOutputThreshold is the default threshold for tool output truncation.
const DefaultTruncateToolOutputThreshold = 40_000

// DefaultMaxAttempts is the default number of retry attempts.
const DefaultMaxAttempts = 5

// MCPServerConfig holds configuration for an MCP server connection.
type MCPServerConfig struct {
	// Stdio transport
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`

	// SSE transport
	URL string `json:"url,omitempty"`

	// Streamable HTTP transport
	HTTPURL string            `json:"httpUrl,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`

	// WebSocket transport
	TCP string `json:"tcp,omitempty"`

	// Transport type
	Type string `json:"type,omitempty"` // "sse" or "http"

	// Common
	Timeout     int    `json:"timeout,omitempty"`
	Trust       bool   `json:"trust,omitempty"`
	Description string `json:"description,omitempty"`

	// Tool filtering
	IncludeTools []string `json:"includeTools,omitempty"`
	ExcludeTools []string `json:"excludeTools,omitempty"`
}

// SandboxConfig configures the sandbox environment.
type SandboxConfig struct {
	Command string `json:"command"` // "docker", "podman", or "sandbox-exec"
	Image   string `json:"image"`
}

// TelemetrySettings configures telemetry behavior.
type TelemetrySettings struct {
	Enabled      bool   `json:"enabled,omitempty"`
	Target       string `json:"target,omitempty"`
	OTLPEndpoint string `json:"otlpEndpoint,omitempty"`
	OTLPProtocol string `json:"otlpProtocol,omitempty"` // "grpc" or "http"
	LogPrompts   bool   `json:"logPrompts,omitempty"`
	Outfile      string `json:"outfile,omitempty"`
	UseCollector bool   `json:"useCollector,omitempty"`
	UseCliAuth   bool   `json:"useCliAuth,omitempty"`
}

// AccessibilitySettings configures accessibility features.
type AccessibilitySettings struct {
	ScreenReader bool `json:"screenReader,omitempty"`
}

// BugCommandSettings configures the bug report command.
type BugCommandSettings struct {
	URLTemplate string `json:"urlTemplate"`
}

// ToolOutputMaskingConfig configures tool output masking behavior.
type ToolOutputMaskingConfig struct {
	Enabled                    bool    `json:"enabled"`
	ToolProtectionThreshold    float64 `json:"toolProtectionThreshold"`
	MinPrunableTokensThreshold int     `json:"minPrunableTokensThreshold"`
	ProtectLatestTurn          bool    `json:"protectLatestTurn"`
}

// ShellExecutionConfig configures shell execution behavior.
type ShellExecutionConfig struct {
	Shell   string `json:"shell,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// ConfigParameters holds all parameters for creating a Config instance.
type ConfigParameters struct {
	SessionID       string
	ClientVersion   string
	EmbeddingModel  string
	Sandbox         *SandboxConfig
	TargetDir       string
	DebugMode       bool
	Question        string
	Cwd             string
	Model           string
	Interactive     bool
	ApprovalMode    ApprovalMode
	NoBrowser       bool
	Proxy           string
	Checkpointing   bool

	// Tool configuration
	CoreTools            []string
	AllowedTools         []string
	ExcludeTools         []string
	ToolDiscoveryCommand string
	ToolCallCommand      string

	// MCP configuration
	MCPServerCommand string
	MCPServers       map[string]*MCPServerConfig
	MCPEnabled       bool
	AllowedMCPServers []string
	BlockedMCPServers []string

	// Memory
	UserMemory       interface{} // string or *HierarchicalMemory
	GeminiMdFileCount int
	GeminiMdFilePaths []string

	// File filtering
	FileFiltering *FileFilteringOptions

	// Include directories
	IncludeDirectories []string

	// Feature flags
	ShowMemoryUsage          bool
	FolderTrust              bool
	IDEMode                  bool
	UseRipgrep               bool
	EnableInteractiveShell   bool
	SkipNextSpeakerCheck     bool
	DisableLoopDetection     bool
	EnableHooks              bool
	EnableAgents             bool
	SkillsSupport            bool
	DirectWebFetch           bool
	ContinueOnFailedAPICall  bool
	RetryFetchErrors         bool
	EnableShellOutputEfficiency bool

	// Limits
	MaxSessionTurns              int
	MaxAttempts                  int
	TruncateToolOutputThreshold  int
	ShellToolInactivityTimeout   int
	CompressionThreshold         int

	// Telemetry
	Telemetry              *TelemetrySettings
	UsageStatisticsEnabled bool

	// Accessibility
	Accessibility *AccessibilitySettings

	// Output
	OutputFormat OutputFormat

	// Shell execution
	ShellExecutionConfig *ShellExecutionConfig

	// Bug report
	BugCommand *BugCommandSettings
}

// Config holds the runtime configuration for a Gemini CLI session.
type Config struct {
	mu sync.RWMutex

	sessionID     string
	clientVersion string
	model         string
	activeModel   string
	targetDir     string
	cwd           string
	debugMode     bool
	question      string
	interactive   bool
	approvalMode  ApprovalMode

	// Feature flags
	ideMode              bool
	noBrowser            bool
	folderTrust          bool
	useRipgrep           bool
	enableInteractiveShell bool
	skipNextSpeakerCheck bool
	disableLoopDetection bool
	enableHooks          bool
	enableAgents         bool
	skillsSupport        bool
	directWebFetch       bool
	continueOnFailedAPICall bool
	retryFetchErrors     bool
	enableShellOutputEfficiency bool
	checkpointing        bool

	// Memory
	userMemory        interface{}
	geminiMdFileCount int
	geminiMdFilePaths []string

	// MCP
	mcpServers        map[string]*MCPServerConfig
	mcpEnabled        bool
	allowedMCPServers []string
	blockedMCPServers []string

	// Tools
	coreTools            []string
	allowedTools         []string
	excludeTools         []string
	toolDiscoveryCommand string
	toolCallCommand      string

	// Limits
	maxSessionTurns             int
	maxAttempts                 int
	truncateToolOutputThreshold int
	shellToolInactivityTimeout  int
	compressionThreshold        int

	// File filtering
	fileFiltering FileFilteringOptions

	// Include directories
	includeDirectories []string

	// Embedding
	embeddingModel string

	// Sandbox
	sandbox *SandboxConfig

	// Proxy
	proxy string

	// Telemetry
	telemetrySettings      TelemetrySettings
	usageStatisticsEnabled bool

	// Accessibility
	accessibility AccessibilitySettings

	// Output
	outputFormat OutputFormat

	// Shell execution
	shellExecutionConfig ShellExecutionConfig

	// Bug report
	bugCommand *BugCommandSettings

	// Runtime state
	quotaErrorOccurred bool
	initialized        bool
}

// NewConfig creates a new Config from the given parameters.
func NewConfig(params ConfigParameters) *Config {
	targetDir := params.TargetDir
	if targetDir == "" {
		targetDir = params.Cwd
	}
	targetDir, _ = filepath.Abs(targetDir)

	c := &Config{
		sessionID:     params.SessionID,
		clientVersion: params.ClientVersion,
		model:         params.Model,
		activeModel:   params.Model,
		targetDir:     targetDir,
		cwd:           params.Cwd,
		debugMode:     params.DebugMode,
		question:      params.Question,
		interactive:   params.Interactive,
		approvalMode:  params.ApprovalMode,

		ideMode:              params.IDEMode,
		noBrowser:            params.NoBrowser,
		folderTrust:          params.FolderTrust,
		useRipgrep:           params.UseRipgrep,
		enableInteractiveShell: params.EnableInteractiveShell,
		skipNextSpeakerCheck: params.SkipNextSpeakerCheck,
		disableLoopDetection: params.DisableLoopDetection,
		enableHooks:          params.EnableHooks,
		enableAgents:         params.EnableAgents,
		skillsSupport:        params.SkillsSupport,
		directWebFetch:       params.DirectWebFetch,
		continueOnFailedAPICall: params.ContinueOnFailedAPICall,
		retryFetchErrors:     params.RetryFetchErrors,
		enableShellOutputEfficiency: params.EnableShellOutputEfficiency,
		checkpointing:        params.Checkpointing,

		userMemory:        params.UserMemory,
		geminiMdFileCount: params.GeminiMdFileCount,
		geminiMdFilePaths: params.GeminiMdFilePaths,

		mcpServers:        params.MCPServers,
		mcpEnabled:        params.MCPEnabled,
		allowedMCPServers: params.AllowedMCPServers,
		blockedMCPServers: params.BlockedMCPServers,

		coreTools:            params.CoreTools,
		allowedTools:         params.AllowedTools,
		excludeTools:         params.ExcludeTools,
		toolDiscoveryCommand: params.ToolDiscoveryCommand,
		toolCallCommand:      params.ToolCallCommand,

		maxSessionTurns:             params.MaxSessionTurns,
		maxAttempts:                 params.MaxAttempts,
		truncateToolOutputThreshold: params.TruncateToolOutputThreshold,
		shellToolInactivityTimeout:  params.ShellToolInactivityTimeout,
		compressionThreshold:        params.CompressionThreshold,

		includeDirectories: params.IncludeDirectories,
		embeddingModel:     params.EmbeddingModel,
		sandbox:            params.Sandbox,
		proxy:              params.Proxy,
		usageStatisticsEnabled: params.UsageStatisticsEnabled,
		outputFormat:       params.OutputFormat,
	}

	// Apply defaults
	if c.embeddingModel == "" {
		c.embeddingModel = DefaultGeminiEmbeddingModel
	}
	if c.maxAttempts == 0 {
		c.maxAttempts = DefaultMaxAttempts
	}
	if c.truncateToolOutputThreshold == 0 {
		c.truncateToolOutputThreshold = DefaultTruncateToolOutputThreshold
	}
	if c.clientVersion == "" {
		c.clientVersion = "unknown"
	}

	// File filtering
	if params.FileFiltering != nil {
		c.fileFiltering = *params.FileFiltering
	} else {
		c.fileFiltering = DefaultFileFilteringOptions
	}

	// Telemetry
	if params.Telemetry != nil {
		c.telemetrySettings = *params.Telemetry
	}

	// Accessibility
	if params.Accessibility != nil {
		c.accessibility = *params.Accessibility
	}

	// Shell execution
	if params.ShellExecutionConfig != nil {
		c.shellExecutionConfig = *params.ShellExecutionConfig
	}

	// Bug command
	c.bugCommand = params.BugCommand

	return c
}

// GetSessionID returns the current session ID.
func (c *Config) GetSessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}

// GetClientVersion returns the client version string.
func (c *Config) GetClientVersion() string {
	return c.clientVersion
}

// GetModel returns the configured model name.
func (c *Config) GetModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.model
}

// SetModel sets the active model.
func (c *Config) SetModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.model = model
	c.activeModel = model
}

// GetActiveModel returns the currently active model.
func (c *Config) GetActiveModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.activeModel
}

// GetTargetDir returns the target working directory.
func (c *Config) GetTargetDir() string {
	return c.targetDir
}

// GetCwd returns the current working directory.
func (c *Config) GetCwd() string {
	return c.cwd
}

// GetDebugMode returns whether debug mode is enabled.
func (c *Config) GetDebugMode() bool {
	return c.debugMode
}

// GetQuestion returns the initial question, if any.
func (c *Config) GetQuestion() string {
	return c.question
}

// GetInteractive returns whether interactive mode is enabled.
func (c *Config) GetInteractive() bool {
	return c.interactive
}

// GetApprovalMode returns the current approval mode.
func (c *Config) GetApprovalMode() ApprovalMode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.approvalMode
}

// SetApprovalMode sets the approval mode.
func (c *Config) SetApprovalMode(mode ApprovalMode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.approvalMode = mode
}

// GetIDEMode returns whether IDE mode is enabled.
func (c *Config) GetIDEMode() bool {
	return c.ideMode
}

// GetNoBrowser returns whether browser launch is disabled.
func (c *Config) GetNoBrowser() bool {
	return c.noBrowser
}

// GetFileFilteringOptions returns the file filtering configuration.
func (c *Config) GetFileFilteringOptions() FileFilteringOptions {
	return c.fileFiltering
}

// GetMaxSessionTurns returns the max session turns limit.
func (c *Config) GetMaxSessionTurns() int {
	return c.maxSessionTurns
}

// GetMCPServers returns the configured MCP servers.
func (c *Config) GetMCPServers() map[string]*MCPServerConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mcpServers
}

// GetUserMemory returns the user memory as a flattened string.
func (c *Config) GetUserMemory() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return FlattenMemory(c.userMemory)
}

// GetSkipNextSpeakerCheck returns whether the next speaker check is skipped.
func (c *Config) GetSkipNextSpeakerCheck() bool {
	return c.skipNextSpeakerCheck
}

// GetContinueOnFailedAPICall returns whether to continue on failed API calls.
func (c *Config) GetContinueOnFailedAPICall() bool {
	return c.continueOnFailedAPICall
}

// GetQuotaErrorOccurred returns whether a quota error has occurred.
func (c *Config) GetQuotaErrorOccurred() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.quotaErrorOccurred
}

// SetQuotaErrorOccurred marks that a quota error has occurred.
func (c *Config) SetQuotaErrorOccurred(occurred bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.quotaErrorOccurred = occurred
}

// GetProxy returns the configured proxy URL.
func (c *Config) GetProxy() string {
	return c.proxy
}

// GetEnableHooks returns whether hooks are enabled.
func (c *Config) GetEnableHooks() bool {
	return c.enableHooks
}

// ResetTurn resets per-turn state. Called at the start of each conversation turn.
func (c *Config) ResetTurn() {
	// Reset any per-turn state here
}

// ValidatePathAccess checks if a path is within the allowed workspace.
func (c *Config) ValidatePathAccess(path string, operation string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "Invalid path: " + err.Error()
	}

	// Check if path is within target directory or any include directory
	if isSubpath(c.targetDir, absPath) {
		return ""
	}

	for _, dir := range c.includeDirectories {
		if isSubpath(dir, absPath) {
			return ""
		}
	}

	return "Path " + absPath + " is not within the workspace. " +
		"Only paths within " + c.targetDir + " are allowed."
}

// Homedir returns the user's home directory.
func Homedir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// isSubpath checks if child is a subpath of parent.
func isSubpath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

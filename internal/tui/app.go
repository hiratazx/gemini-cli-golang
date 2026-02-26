// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/google-gemini/gemini-cli/internal/auth"
	"github.com/google-gemini/gemini-cli/internal/genai"
)

// AppState represents the current state of the application.
type AppState int

const (
	StateAuth AppState = iota
	StateModelSelect
	StateChat
	StateQuitting
)

// Model is the main Bubble Tea model.
type Model struct {
	state    AppState
	width    int
	height   int

	// Auth state
	authOptions  []auth.AuthOption
	authCursor   int
	authLoading  bool
	authSpinner  spinner.Model
	creds        *auth.Credentials
	apiKeyInput  textinput.Model
	enteringKey  bool
	authError    string

	// Model selection
	models      []genai.ModelInfo
	modelCursor int

	// Chat state
	client       *genai.Client
	chatInput    textinput.Model
	messages     []chatMessage
	streaming    bool
	streamText   string
	thinking     string
	selectedModel string
	tokenInfo    string
	chatSpinner  spinner.Model

	// Non-interactive mode
	initialPrompt string
}

type chatMessage struct {
	role string // "user" or "model"
	text string
}

// Message types
type authCompleteMsg struct {
	creds *auth.Credentials
	err   error
}

type streamChunkMsg struct {
	resp genai.StreamResponse
}

type streamDoneMsg struct{}

// NewModel creates a new TUI model.
func NewModel(initialPrompt string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	cs := spinner.New()
	cs.Spinner = spinner.Dot
	cs.Style = spinnerStyle

	ti := textinput.New()
	ti.Placeholder = "Enter API key..."
	ti.CharLimit = 100
	ti.Width = 60

	chatInput := textinput.New()
	chatInput.Placeholder = "Ask Gemini anything... (Ctrl+C to quit)"
	chatInput.CharLimit = 2000
	chatInput.Width = 80

	return Model{
		state:         StateAuth,
		authOptions:   auth.GetAuthOptions(),
		authSpinner:   s,
		apiKeyInput:   ti,
		models:        genai.ListModels(),
		chatInput:     chatInput,
		chatSpinner:   cs,
		initialPrompt: initialPrompt,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	// Check if already authenticated
	authType, apiKey := auth.DetectAuthType()
	if apiKey != "" {
		return func() tea.Msg {
			creds, err := auth.Authenticate(authType, apiKey)
			return authCompleteMsg{creds: creds, err: err}
		}
	}

	if auth.HasCachedCredentials() {
		return func() tea.Msg {
			creds, err := auth.Authenticate(auth.AuthTypeOAuth, "")
			return authCompleteMsg{creds: creds, err: err}
		}
	}

	return m.authSpinner.Tick
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.chatInput.Width = msg.Width - 6
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			m.state = StateQuitting
			return m, tea.Quit
		}
	}

	switch m.state {
	case StateAuth:
		return m.updateAuth(msg)
	case StateModelSelect:
		return m.updateModelSelect(msg)
	case StateChat:
		return m.updateChat(msg)
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.state == StateQuitting {
		return "\n  " + successStyle.Render("Goodbye! 👋") + "\n\n"
	}

	switch m.state {
	case StateAuth:
		return m.viewAuth()
	case StateModelSelect:
		return m.viewModelSelect()
	case StateChat:
		return m.viewChat()
	}

	return ""
}

// === Auth State ===

func (m Model) updateAuth(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.enteringKey {
			switch msg.String() {
			case "enter":
				key := m.apiKeyInput.Value()
				if key != "" {
					m.authLoading = true
					m.enteringKey = false
					return m, tea.Batch(m.authSpinner.Tick, func() tea.Msg {
						creds, err := auth.Authenticate(auth.AuthTypeAPIKey, key)
						if err == nil {
							auth.SaveAPIKey(key)
						}
						return authCompleteMsg{creds: creds, err: err}
					})
				}
				return m, nil
			case "esc":
				m.enteringKey = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "up", "k":
			if m.authCursor > 0 {
				m.authCursor--
			}
		case "down", "j":
			if m.authCursor < len(m.authOptions)-1 {
				m.authCursor++
			}
		case "enter":
			selected := m.authOptions[m.authCursor]
			switch selected.Type {
			case auth.AuthTypeOAuth:
				m.authLoading = true
				return m, tea.Batch(m.authSpinner.Tick, func() tea.Msg {
					creds, err := auth.Authenticate(auth.AuthTypeOAuth, "")
					return authCompleteMsg{creds: creds, err: err}
				})
			case auth.AuthTypeAPIKey:
				cachedKey := auth.LoadAPIKey()
				if cachedKey != "" {
					m.authLoading = true
					return m, tea.Batch(m.authSpinner.Tick, func() tea.Msg {
						creds, err := auth.Authenticate(auth.AuthTypeAPIKey, cachedKey)
						return authCompleteMsg{creds: creds, err: err}
					})
				}
				m.enteringKey = true
				m.apiKeyInput.Focus()
				return m, textinput.Blink
			}
		}

	case authCompleteMsg:
		m.authLoading = false
		if msg.err != nil {
			m.authError = msg.err.Error()
			return m, nil
		}
		m.creds = msg.creds
		m.state = StateModelSelect
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.authSpinner, cmd = m.authSpinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) viewAuth() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("  " + logoGemini + versionStyle.Render("v1.0.0 (Go)") + "\n")
	sb.WriteString("  " + divider(50) + "\n\n")

	if m.authLoading {
		sb.WriteString("  " + m.authSpinner.View() + " Authenticating...\n")
		return sb.String()
	}

	if m.authError != "" {
		sb.WriteString("  " + errorStyle.Render("⚠ "+m.authError) + "\n\n")
	}

	if m.enteringKey {
		sb.WriteString("  " + titleStyle.Render("Enter your Gemini API Key") + "\n\n")
		sb.WriteString("  " + m.apiKeyInput.View() + "\n\n")
		sb.WriteString("  " + helpDescStyle.Render("Get one at https://aistudio.google.com/apikey") + "\n")
		sb.WriteString("  " + helpDescStyle.Render("Press Enter to submit, Esc to go back") + "\n")
		return sb.String()
	}

	sb.WriteString("  " + titleStyle.Render("Choose Authentication Method") + "\n\n")

	for i, opt := range m.authOptions {
		cursor := "  "
		if i == m.authCursor {
			cursor = cursorStyle.Render("▸ ")
			sb.WriteString("  " + cursor + selectedStyle.Render(opt.Label) + "\n")
			sb.WriteString("    " + descriptionStyle.Render(opt.Description) + "\n\n")
		} else {
			sb.WriteString("  " + cursor + unselectedStyle.Render(opt.Label) + "\n\n")
		}
	}

	sb.WriteString("\n  " + helpDescStyle.Render("↑/↓ navigate • enter select • ctrl+c quit") + "\n")

	return sb.String()
}

// === Model Select State ===

func (m Model) updateModelSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.modelCursor > 0 {
				m.modelCursor--
			}
		case "down", "j":
			if m.modelCursor < len(m.models)-1 {
				m.modelCursor++
			}
		case "enter":
			m.selectedModel = m.models[m.modelCursor].ID
			m.state = StateChat
			m.chatInput.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m Model) viewModelSelect() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("  " + logoGemini + "\n")
	sb.WriteString("  " + divider(50) + "\n\n")

	if m.creds != nil && m.creds.Email != "" {
		sb.WriteString("  " + successStyle.Render("✓ Authenticated") + " as " + statusEmailStyle.Render(m.creds.Email) + "\n\n")
	} else {
		sb.WriteString("  " + successStyle.Render("✓ Authenticated") + "\n\n")
	}

	sb.WriteString("  " + titleStyle.Render("Select Model") + "\n\n")

	for i, model := range m.models {
		cursor := "  "
		if i == m.modelCursor {
			cursor = cursorStyle.Render("▸ ")
			label := model.Name
			if model.Default {
				label += " ★"
			}
			sb.WriteString("  " + cursor + selectedStyle.Render(label) + "\n")
			sb.WriteString("    " + descriptionStyle.Render(model.Description) + "\n\n")
		} else {
			label := model.Name
			if model.Default {
				label += " ★"
			}
			sb.WriteString("  " + cursor + unselectedStyle.Render(label) + "\n\n")
		}
	}

	sb.WriteString("\n  " + helpDescStyle.Render("↑/↓ navigate • enter select") + "\n")

	return sb.String()
}

// === Chat State ===

func (m Model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.streaming {
			return m, nil
		}

		if msg.String() == "enter" {
			input := m.chatInput.Value()
			if input == "" {
				return m, nil
			}

			m.chatInput.SetValue("")
			m.messages = append(m.messages, chatMessage{role: "user", text: input})
			m.streaming = true
			m.streamText = ""
			m.thinking = ""

			return m, tea.Batch(m.chatSpinner.Tick, m.sendMessage(input))
		}

		// Forward all other key presses to the text input
		var cmd tea.Cmd
		m.chatInput, cmd = m.chatInput.Update(msg)
		return m, cmd

	case streamChunkMsg:
		resp := msg.resp
		if resp.Error != nil {
			m.streaming = false
			m.streamText += "\n" + errorStyle.Render("Error: "+resp.Error.Error())
			m.messages = append(m.messages, chatMessage{role: "model", text: m.streamText})
			m.streamText = ""
			return m, nil
		}

		if resp.Text != "" {
			m.streamText += resp.Text
		}
		if resp.Thought != "" {
			m.thinking = resp.Thought
		}

		if resp.TotalTokens > 0 {
			m.tokenInfo = fmt.Sprintf("Tokens: %d in / %d out", resp.InputTokens, resp.OutputTokens)
		}

		if resp.Done && m.streamText != "" {
			m.streaming = false
			m.messages = append(m.messages, chatMessage{role: "model", text: m.streamText})
			m.streamText = ""
			m.thinking = ""
			m.chatInput.Focus()
			return m, textinput.Blink
		}

		return m, nil

	case streamDoneMsg:
		m.streaming = false
		if m.streamText != "" {
			m.messages = append(m.messages, chatMessage{role: "model", text: m.streamText})
			m.streamText = ""
		}
		m.thinking = ""
		m.chatInput.Focus()
		return m, textinput.Blink

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.chatSpinner, cmd = m.chatSpinner.Update(msg)
		return m, cmd
	}

	// Forward any other messages (e.g., Blink) to the chat input
	var cmd tea.Cmd
	m.chatInput, cmd = m.chatInput.Update(msg)
	return m, cmd
}

func (m Model) sendMessage(input string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, m.creds)
		if err != nil {
			return streamChunkMsg{resp: genai.StreamResponse{Error: err, Done: true}}
		}

		cwd, _ := os.Getwd()
		systemPrompt := genai.GetDefaultSystemPrompt(cwd)

		var history []genai.Message
		for _, msg := range m.messages {
			history = append(history, genai.Message{Role: msg.role, Text: msg.text})
		}

		ch := client.GenerateContentStream(ctx, m.selectedModel, systemPrompt, history, input)

		for resp := range ch {
			// We need to send each chunk back to the TUI
			// Since we're in a cmd, we'll collect and return the full response
			if resp.Error != nil {
				return streamChunkMsg{resp: resp}
			}
			if resp.Done {
				return streamDoneMsg{}
			}
			// For now, accumulate and return final
			return streamChunkMsg{resp: resp}
		}

		return streamDoneMsg{}
	}
}

func (m Model) viewChat() string {
	var sb strings.Builder

	// Header
	sb.WriteString("  " + logoGemini)
	sb.WriteString("  " + statusModelStyle.Render(m.selectedModel))
	if m.creds != nil && m.creds.Email != "" {
		sb.WriteString("  " + statusEmailStyle.Render(m.creds.Email))
	}
	sb.WriteString("\n")
	sb.WriteString("  " + divider(min(m.width-4, 80)) + "\n\n")

	// Messages
	maxShow := m.height - 10 // Leave room for input and status
	startIdx := 0
	if len(m.messages) > maxShow/3 {
		startIdx = len(m.messages) - maxShow/3
	}

	for _, msg := range m.messages[startIdx:] {
		if msg.role == "user" {
			sb.WriteString("  " + promptStyle.Render("❯ ") + userMessageStyle.Render(msg.text) + "\n\n")
		} else {
			sb.WriteString("  " + modelMessageStyle.Render(msg.text) + "\n\n")
		}
	}

	// Streaming content
	if m.streaming {
		if m.thinking != "" {
			sb.WriteString("  " + thinkingStyle.Render("💭 "+m.thinking) + "\n")
		}
		if m.streamText != "" {
			sb.WriteString("  " + modelMessageStyle.Render(m.streamText))
		} else {
			sb.WriteString("  " + m.chatSpinner.View() + " " + thinkingStyle.Render("Thinking..."))
		}
		sb.WriteString("\n\n")
	}

	// Input
	if !m.streaming {
		sb.WriteString("  " + promptStyle.Render("❯ ") + m.chatInput.View() + "\n")
	}

	// Status bar
	statusItems := []string{}
	if m.tokenInfo != "" {
		statusItems = append(statusItems, m.tokenInfo)
	}
	statusItems = append(statusItems, helpKeyStyle.Render("ctrl+c")+" "+helpDescStyle.Render("quit"))

	sb.WriteString("\n  " + statusInfoStyle.Render(strings.Join(statusItems, "  •  ")) + "\n")

	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Run starts the TUI application.
func Run(initialPrompt string) error {
	m := NewModel(initialPrompt)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

// RunNonInteractive sends a single prompt and prints the response.
func RunNonInteractive(prompt string) error {
	ctx := context.Background()

	authType, apiKey := auth.DetectAuthType()
	if authType == "" {
		// Check for cached credentials
		if auth.HasCachedCredentials() {
			authType = auth.AuthTypeOAuth
		} else if key := auth.LoadAPIKey(); key != "" {
			authType = auth.AuthTypeAPIKey
			apiKey = key
		} else {
			return fmt.Errorf("no authentication found. Run 'gemini' interactively first to authenticate, or set GEMINI_API_KEY")
		}
	}

	creds, err := auth.Authenticate(authType, apiKey)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	client, err := genai.NewClient(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	cwd, _ := os.Getwd()
	systemPrompt := genai.GetDefaultSystemPrompt(cwd)

	ch := client.GenerateContentStream(ctx, "gemini-2.5-pro", systemPrompt, nil, prompt)

	for resp := range ch {
		if resp.Error != nil {
			return resp.Error
		}
		if resp.Text != "" {
			fmt.Print(resp.Text)
		}
	}
	fmt.Println()

	return nil
}

// Suppress unused import
var _ = lipgloss.NewStyle

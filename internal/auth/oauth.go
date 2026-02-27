// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuth redirect URLs.
const (
	signInSuccessURL = "https://developers.google.com/gemini-code-assist/auth_success_gemini"
	signInFailureURL = "https://developers.google.com/gemini-code-assist/auth_failure_gemini"
)

// These are set at build time via -ldflags from GitHub Actions secrets:
//
//	go build -ldflags="-X github.com/google-gemini/gemini-cli/internal/auth.defaultOAuthClientID=..."
//
// For local development, set the GEMINI_OAUTH_CLIENT_ID / GEMINI_OAUTH_CLIENT_SECRET env vars.
var (
	defaultOAuthClientID     string // injected at build time
	defaultOAuthClientSecret string // injected at build time
)

// getOAuthClientID returns the OAuth client ID.
// Priority: env var → build-time ldflags value.
func getOAuthClientID() string {
	if v := os.Getenv("GEMINI_OAUTH_CLIENT_ID"); v != "" {
		return v
	}
	return defaultOAuthClientID
}

// getOAuthClientSecret returns the OAuth client secret.
// Priority: env var → build-time ldflags value.
func getOAuthClientSecret() string {
	if v := os.Getenv("GEMINI_OAUTH_CLIENT_SECRET"); v != "" {
		return v
	}
	return defaultOAuthClientSecret
}

var oauthScopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
}

// credsCachePath returns the path where OAuth credentials are cached.
func credsCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gemini", "oauth_creds.json")
}

// cachedOAuthToken represents the stored token on disk.
type cachedOAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}

// authenticateOAuth performs the full OAuth2 flow.
func authenticateOAuth() (*Credentials, error) {
	conf := &oauth2.Config{
		ClientID:     getOAuthClientID(),
		ClientSecret: getOAuthClientSecret(),
		Scopes:       oauthScopes,
		Endpoint:     google.Endpoint,
	}

	// Try loading cached credentials first
	if creds, err := loadCachedOAuth(conf); err == nil && creds != nil {
		return creds, nil
	}

	// Start the OAuth web flow
	return oauthWebFlow(conf)
}

// loadCachedOAuth tries to load and refresh cached OAuth tokens.
func loadCachedOAuth(conf *oauth2.Config) (*Credentials, error) {
	data, err := os.ReadFile(credsCachePath())
	if err != nil {
		return nil, err
	}

	var cached cachedOAuthToken
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  cached.AccessToken,
		RefreshToken: cached.RefreshToken,
		TokenType:    cached.TokenType,
		Expiry:       cached.Expiry,
	}

	// Use the token source to auto-refresh if needed
	tokenSource := conf.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}

	// Save refreshed token
	saveOAuthToken(newToken)

	email := fetchUserEmail(newToken.AccessToken)

	return &Credentials{
		AuthType:     AuthTypeOAuth,
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		Email:        email,
	}, nil
}

// oauthWebFlow performs the browser-based OAuth flow.
func oauthWebFlow(conf *oauth2.Config) (*Credentials, error) {
	// Get available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	conf.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/oauth2callback", port)

	state := generateState()

	authURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Channel to receive the result
	resultCh := make(chan *oauth2.Token, 1)
	errCh := make(chan error, 1)

	// Start temporary HTTP server for callback
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if errMsg := q.Get("error"); errMsg != "" {
			http.Redirect(w, r, signInFailureURL, http.StatusMovedPermanently)
			errCh <- fmt.Errorf("OAuth error: %s - %s", errMsg, q.Get("error_description"))
			return
		}

		if q.Get("state") != state {
			http.Error(w, "State mismatch", http.StatusBadRequest)
			errCh <- fmt.Errorf("OAuth state mismatch")
			return
		}

		code := q.Get("code")
		if code == "" {
			http.Redirect(w, r, signInFailureURL, http.StatusMovedPermanently)
			errCh <- fmt.Errorf("no authorization code received")
			return
		}

		token, err := conf.Exchange(context.Background(), code)
		if err != nil {
			http.Redirect(w, r, signInFailureURL, http.StatusMovedPermanently)
			errCh <- fmt.Errorf("failed to exchange code: %w", err)
			return
		}

		http.Redirect(w, r, signInSuccessURL, http.StatusMovedPermanently)
		resultCh <- token
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Open browser
	if err := browser.OpenURL(authURL); err != nil {
		// If browser can't be opened, print the URL
		fmt.Printf("\nOpen this URL in your browser:\n\n%s\n\n", authURL)
	}

	// Wait for callback with timeout
	select {
	case token := <-resultCh:
		saveOAuthToken(token)
		email := fetchUserEmail(token.AccessToken)
		return &Credentials{
			AuthType:     AuthTypeOAuth,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			Email:        email,
		}, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authentication timed out after 5 minutes")
	}
}

// saveOAuthToken saves the token to disk.
func saveOAuthToken(token *oauth2.Token) {
	cached := cachedOAuthToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return
	}

	dir := filepath.Dir(credsCachePath())
	os.MkdirAll(dir, 0o700)
	os.WriteFile(credsCachePath(), data, 0o600)
}

// fetchUserEmail retrieves the user's email from the Google userinfo API.
func fetchUserEmail(accessToken string) string {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var info struct {
		Email string `json:"email"`
	}
	json.Unmarshal(body, &info)
	return info.Email
}

// generateState generates a random state string for CSRF protection.
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ClearCachedCredentials removes cached OAuth credentials.
func ClearCachedCredentials() error {
	return os.Remove(credsCachePath())
}

// ClearAPIKey removes the cached API key.
func ClearAPIKey() error {
	home, _ := os.UserHomeDir()
	if home == "" {
		return nil
	}
	return os.Remove(filepath.Join(home, ".gemini", "api_key"))
}

// Logout clears all cached credentials (OAuth + API key).
func Logout() {
	ClearCachedCredentials()
	ClearAPIKey()
}

// HasCachedCredentials checks if cached OAuth credentials exist.
func HasCachedCredentials() bool {
	_, err := os.Stat(credsCachePath())
	return err == nil
}

// LoadAPIKey loads the API key from environment or cached storage.
func LoadAPIKey() string {
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return key
	}
	// Check cached API key
	home, _ := os.UserHomeDir()
	keyPath := filepath.Join(home, ".gemini", "api_key")
	data, err := os.ReadFile(keyPath)
	if err == nil {
		return string(data)
	}
	return ""
}

// SaveAPIKey saves an API key to cached storage.
func SaveAPIKey(key string) error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".gemini")
	os.MkdirAll(dir, 0o700)
	return os.WriteFile(filepath.Join(dir, "api_key"), []byte(key), 0o600)
}

// GetAuthURL returns the OAuth URL without starting a server (for display).
func GetAuthURL() string {
	conf := &oauth2.Config{
		ClientID:     getOAuthClientID(),
		ClientSecret: getOAuthClientSecret(),
		Scopes:       oauthScopes,
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}
	return conf.AuthCodeURL(generateState(), oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for tokens (manual flow).
func ExchangeCode(code string) (*Credentials, error) {
	conf := &oauth2.Config{
		ClientID:     getOAuthClientID(),
		ClientSecret: getOAuthClientSecret(),
		Scopes:       oauthScopes,
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}

	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	saveOAuthToken(token)
	email := fetchUserEmail(token.AccessToken)

	return &Credentials{
		AuthType:     AuthTypeOAuth,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Email:        email,
	}, nil
}

// OAuthFlowAsync starts the OAuth flow asynchronously and returns channels.
func OAuthFlowAsync() (authURL string, resultCh <-chan *Credentials, errCh <-chan error) {
	creds := make(chan *Credentials, 1)
	errs := make(chan error, 1)

	conf := &oauth2.Config{
		ClientID:     getOAuthClientID(),
		ClientSecret: getOAuthClientSecret(),
		Scopes:       oauthScopes,
		Endpoint:     google.Endpoint,
	}

	// Get available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		errs <- err
		return "", creds, errs
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	conf.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/oauth2callback", port)
	state := generateState()
	authURL = conf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if errMsg := q.Get("error"); errMsg != "" {
				http.Redirect(w, r, signInFailureURL, http.StatusMovedPermanently)
				errs <- fmt.Errorf("OAuth error: %s", errMsg)
				return
			}
			if q.Get("state") != state {
				http.Error(w, "State mismatch", http.StatusBadRequest)
				errs <- fmt.Errorf("state mismatch")
				return
			}
			code := q.Get("code")
			token, err := conf.Exchange(context.Background(), code)
			if err != nil {
				http.Redirect(w, r, signInFailureURL, http.StatusMovedPermanently)
				errs <- err
				return
			}
			http.Redirect(w, r, signInSuccessURL, http.StatusMovedPermanently)
			saveOAuthToken(token)
			email := fetchUserEmail(token.AccessToken)
			creds <- &Credentials{
				AuthType:     AuthTypeOAuth,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				Email:        email,
			}
		})

		server := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Handler: mux}
		go func() {
			select {
			case <-creds:
			case <-errs:
			case <-time.After(5 * time.Minute):
				errs <- fmt.Errorf("authentication timed out")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			server.Shutdown(ctx)
		}()
		server.ListenAndServe()
	}()

	return authURL, creds, errs
}

// Suppress unused import warning
var _ = url.Parse

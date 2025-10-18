package handler

import (
	"bytes"
	"cc-switch/internal/config"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Version should match the version in package.json
	clientVersion = "1.0.16"
	userAgent     = "claude-code/" + clientVersion
)

// APITester handles API connectivity testing for Claude Code configurations
type APITester struct {
	configManager *config.ConfigManager
	httpClient    *http.Client
}

// NewAPITester creates a new API tester instance
func NewAPITester(configManager *config.ConfigManager) *APITester {
	// Create HTTP client with reasonable timeouts and security settings
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return &APITester{
		configManager: configManager,
		httpClient:    client,
	}
}

// TestAPIConnectivity tests the API connectivity for a specific profile
func (t *APITester) TestAPIConnectivity(profileName string, options TestOptions) (*APITestResult, error) {
	if profileName == "" {
		return nil, fmt.Errorf("profile name cannot be empty")
	}

	// Handle special case for empty mode
	if profileName == "empty_mode" {
		return &APITestResult{
			ProfileName:   "empty_mode",
			IsConnectable: false,
			TestedAt:      time.Now(),
			Error:         "Empty mode has no configuration to test",
		}, nil
	}

	// Extract API credentials from configuration
	credentials, err := t.extractAPICredentials(profileName)
	if err != nil {
		return &APITestResult{
			ProfileName:   profileName,
			IsConnectable: false,
			TestedAt:      time.Now(),
			Error:         fmt.Sprintf("Failed to extract credentials: %v", err),
		}, nil
	}

	// Apply custom timeout if specified
	if options.Timeout > 0 {
		t.httpClient.Timeout = options.Timeout
	}

	result := &APITestResult{
		ProfileName: profileName,
		TestedAt:    time.Now(),
		Tests:       []EndpointTest{},
	}

	start := time.Now()

	// Run appropriate test suite based on options
	if options.Quick {
		result.Tests = append(result.Tests, t.testBasicConnectivity(credentials))
	} else {
		// Full test suite
		result.Tests = append(result.Tests,
			t.testAuthentication(credentials),
			t.testModelsEndpoint(credentials),
		)

		// Always test chat endpoint with real Claude CLI
		// This is the most reliable test, especially for third-party proxies
		// that may not implement all Anthropic API endpoints
		result.Tests = append(result.Tests, t.testChatEndpoint(profileName, credentials))
	}

	// Calculate total response time and connectivity status
	result.ResponseTime = time.Since(start)
	result.IsConnectable = t.aggregateResults(result.Tests)

	return result, nil
}

// TestAllConfigurations tests API connectivity for all available configurations
func (t *APITester) TestAllConfigurations(options TestOptions) ([]APITestResult, error) {
	profiles, err := t.configManager.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	results := make([]APITestResult, 0, len(profiles))

	for _, profile := range profiles {
		result, err := t.TestAPIConnectivity(profile.Name, options)
		if err != nil {
			// Create error result for this profile
			result = &APITestResult{
				ProfileName:   profile.Name,
				IsConnectable: false,
				TestedAt:      time.Now(),
				Error:         err.Error(),
			}
		}
		results = append(results, *result)
	}

	return results, nil
}

// TestCurrentConfiguration tests the currently active configuration
func (t *APITester) TestCurrentConfiguration(options TestOptions) (*APITestResult, error) {
	// Check if in empty mode
	if t.configManager.IsEmptyMode() {
		return t.TestAPIConnectivity("empty_mode", options)
	}

	currentProfile, err := t.configManager.GetCurrentProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get current profile: %w", err)
	}

	return t.TestAPIConnectivity(currentProfile, options)
}

// extractAPICredentials extracts API credentials from a configuration profile
func (t *APITester) extractAPICredentials(profileName string) (*APICredentials, error) {
	content, _, err := t.configManager.GetProfileContent(profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile content: %w", err)
	}

	credentials := &APICredentials{
		BaseURL: "https://api.anthropic.com",
		Version: "2023-06-01",
	}

	// Extract API key from env section
	if env, ok := content["env"].(map[string]interface{}); ok {
		if apiKey, ok := env["ANTHROPIC_AUTH_TOKEN"].(string); ok && apiKey != "" {
			credentials.APIKey = apiKey
		} else if apiKey, ok := env["ANTHROPIC_API_KEY"].(string); ok && apiKey != "" {
			credentials.APIKey = apiKey
		}

		// Extract base URL if provided
		if baseURL, ok := env["ANTHROPIC_BASE_URL"].(string); ok && baseURL != "" {
			credentials.BaseURL = baseURL
		}

		// Extract API version if provided
		if version, ok := env["ANTHROPIC_VERSION"].(string); ok && version != "" {
			credentials.Version = version
		}
	}

	if credentials.APIKey == "" {
		return nil, fmt.Errorf("no API key found in configuration")
	}

	return credentials, nil
}

// testBasicConnectivity performs a basic connectivity test to the API
func (t *APITester) testBasicConnectivity(credentials *APICredentials) EndpointTest {
	start := time.Now()

	req, err := http.NewRequest("HEAD", credentials.BaseURL, nil)
	if err != nil {
		return EndpointTest{
			Endpoint:     credentials.BaseURL,
			FullURL:      credentials.BaseURL,
			Method:       "HEAD",
			Status:       "failed",
			ResponseTime: time.Since(start),
			Error:        fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	resp, err := t.httpClient.Do(req)
	duration := time.Since(start)

	test := EndpointTest{
		Endpoint:     credentials.BaseURL,
		FullURL:      credentials.BaseURL,
		Method:       "HEAD",
		ResponseTime: duration,
	}

	if err != nil {
		test.Status = "failed"
		test.Error = err.Error()
		return test
	}
	defer resp.Body.Close()

	test.StatusCode = resp.StatusCode

	if resp.StatusCode < 500 {
		test.Status = "success"
		test.Details = "Basic connectivity established"
	} else {
		test.Status = "failed"
		test.Error = fmt.Sprintf("Server error: %d", resp.StatusCode)
	}

	return test
}

// testAuthentication tests API authentication
func (t *APITester) testAuthentication(credentials *APICredentials) EndpointTest {
	start := time.Now()

	endpoint := "/v1/models"
	url := strings.TrimSuffix(credentials.BaseURL, "/") + endpoint

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return EndpointTest{
			Endpoint:     endpoint,
			FullURL:      url,
			Method:       "GET",
			Status:       "failed",
			ResponseTime: time.Since(start),
			Error:        fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	req.Header.Set("Authorization", "Bearer "+credentials.APIKey)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("anthropic-version", credentials.Version)

	resp, err := t.httpClient.Do(req)
	duration := time.Since(start)

	test := EndpointTest{
		Endpoint:     endpoint,
		FullURL:      url,
		Method:       "GET",
		ResponseTime: duration,
	}

	if err != nil {
		test.Status = "failed"
		test.Error = err.Error()
		return test
	}
	defer resp.Body.Close()

	test.StatusCode = resp.StatusCode

	switch resp.StatusCode {
	case 200:
		test.Status = "success"
		test.Details = "Authentication successful"
	case 401:
		test.Status = "failed"
		test.Error = "Invalid API key"
	case 403:
		test.Status = "failed"
		test.Error = "API key lacks required permissions"
	case 429:
		test.Status = "failed"
		test.Error = "Rate limit exceeded"
	case 500, 502, 503, 504:
		test.Status = "failed"
		test.Error = fmt.Sprintf("Server error: %d", resp.StatusCode)
	default:
		test.Status = "failed"
		test.Error = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
	}

	return test
}

// testModelsEndpoint tests the models endpoint specifically
func (t *APITester) testModelsEndpoint(credentials *APICredentials) EndpointTest {
	start := time.Now()

	endpoint := "/v1/models"
	url := strings.TrimSuffix(credentials.BaseURL, "/") + endpoint

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return EndpointTest{
			Endpoint:     endpoint,
			FullURL:      url,
			Method:       "GET-MODELS", // Different method to distinguish from auth test
			Status:       "failed",
			ResponseTime: time.Since(start),
			Error:        fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	req.Header.Set("Authorization", "Bearer "+credentials.APIKey)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("anthropic-version", credentials.Version)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := t.httpClient.Do(req)
	duration := time.Since(start)

	test := EndpointTest{
		Endpoint:     endpoint,
		FullURL:      url,
		Method:       "GET-MODELS", // Different method to distinguish from auth test
		ResponseTime: duration,
	}

	if err != nil {
		test.Status = "failed"
		test.Error = err.Error()
		return test
	}
	defer resp.Body.Close()

	test.StatusCode = resp.StatusCode

	if resp.StatusCode == 200 {
		// Try to parse response to validate it's working properly
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			test.Status = "failed"
			test.Error = "Failed to read response body"
			return test
		}

		var modelsResp map[string]interface{}
		if err := json.Unmarshal(body, &modelsResp); err != nil {
			test.Status = "failed"
			test.Error = "Invalid JSON response"
			return test
		}

		test.Status = "success"
		if data, ok := modelsResp["data"].([]interface{}); ok {
			test.Details = fmt.Sprintf("Models endpoint functional (%d models available)", len(data))
		} else {
			test.Details = "Models endpoint functional"
		}
	} else {
		test.Status = "failed"
		body, _ := io.ReadAll(resp.Body)
		test.Error = fmt.Sprintf("Status %d: %s", resp.StatusCode, string(body))
	}

	return test
}

// testChatEndpoint tests the chat endpoint using real Claude Code CLI
func (t *APITester) testChatEndpoint(profileName string, credentials *APICredentials) EndpointTest {
	start := time.Now()

	endpoint := "/v1/messages"
	fullURL := strings.TrimSuffix(credentials.BaseURL, "/") + endpoint

	test := EndpointTest{
		Endpoint: endpoint,
		FullURL:  fullURL,
		Method:   "claude-cli",
	}

	// Check if claude command is available
	claudePath, err := t.findClaudeCommand()
	if err != nil {
		test.Status = "failed"
		test.Error = fmt.Sprintf("Claude CLI not found: %v", err)
		test.ResponseTime = time.Since(start)
		return test
	}

	// Get the actual configuration file path for the profile being tested
	configPath, err := t.getConfigFilePath(profileName)
	if err != nil {
		test.Status = "failed"
		test.Error = fmt.Sprintf("Failed to get config file path: %v", err)
		test.ResponseTime = time.Since(start)
		return test
	}

	// Execute claude command with 30s timeout using the actual config file
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, claudePath, "-p", "Hi", "--settings", configPath)

	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	test.ResponseTime = time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		test.Status = "timeout"
		test.Error = "Command timed out after 30 seconds"
		return test
	}

	if err != nil {
		test.Status = "failed"
		stderrStr := stderr.String()
		if stderrStr != "" {
			test.Error = fmt.Sprintf("Command failed: %v - %s", err, stderrStr)
		} else {
			test.Error = fmt.Sprintf("Command failed: %v", err)
		}
		return test
	}

	// Check if we got a response
	output := stdout.String()
	if output == "" {
		test.Status = "failed"
		test.Error = "No output from Claude CLI"
		return test
	}

	test.Status = "success"
	test.Details = "Chat endpoint functional via Claude CLI"
	return test
}

// findClaudeCommand locates the claude command in common locations
func (t *APITester) findClaudeCommand() (string, error) {
	// Try common locations for claude command
	locations := []string{
		"claude",                   // In PATH
		"/usr/local/bin/claude",    // Common install location
		"/opt/homebrew/bin/claude", // Homebrew on M1 Macs
		"/usr/bin/claude",          // System location
		"~/.local/bin/claude",      // User install
	}

	for _, location := range locations {
		// Expand home directory if needed
		if strings.HasPrefix(location, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			location = filepath.Join(home, location[2:])
		}

		// Check if command exists and is executable
		if _, err := exec.LookPath(location); err == nil {
			return location, nil
		}

		// Also try direct path check
		if _, err := os.Stat(location); err == nil {
			return location, nil
		}
	}

	return "", fmt.Errorf("claude command not found in common locations")
}

// getConfigFilePath returns the full path to the configuration file for the given profile
func (t *APITester) getConfigFilePath(profileName string) (string, error) {
	// Use GetProfileContent to verify the profile exists and get the path
	_, profile, err := t.configManager.GetProfileContent(profileName)
	if err != nil {
		return "", fmt.Errorf("failed to get profile content: %v", err)
	}

	return profile.Path, nil
}

// aggregateResults determines overall connectivity status from individual test results
func (t *APITester) aggregateResults(tests []EndpointTest) bool {
	if len(tests) == 0 {
		return false
	}

	// Count different types of results
	successCount := 0
	timeoutCount := 0
	failureCount := 0

	// Specific test status tracking
	authSuccess := false
	chatTestFound := false
	chatSuccess := false

	for _, test := range tests {
		switch test.Status {
		case "success":
			successCount++
			// Track authentication success specifically
			if test.Endpoint == "/v1/models" && test.Method == "GET" {
				authSuccess = true
			}
			// Track chat endpoint success (Claude CLI test)
			if test.Endpoint == "/v1/messages" && test.Method == "claude-cli" {
				chatTestFound = true
				chatSuccess = true
			}
		case "timeout":
			timeoutCount++
			// Chat endpoint timeout is critical
			if test.Endpoint == "/v1/messages" && test.Method == "claude-cli" {
				chatTestFound = true
				chatSuccess = false
			}
		case "failed":
			failureCount++
			// Chat endpoint failure
			if test.Endpoint == "/v1/messages" && test.Method == "claude-cli" {
				chatTestFound = true
				chatSuccess = false
			}
		}
	}

	// Priority 1: If Claude CLI test was performed, use its result as the primary indicator
	// This is the most reliable test since it uses the actual Claude Code CLI
	if chatTestFound {
		// Configuration is functional if Claude CLI test succeeded and no timeouts
		return chatSuccess && timeoutCount == 0
	}

	// Priority 2: Fall back to standard API tests if Claude CLI test wasn't performed
	// Configuration is functional only if:
	// 1. Authentication succeeded
	// 2. No timeouts occurred
	// 3. At least 50% of tests passed
	minSuccessRate := float64(successCount)/float64(len(tests)) >= 0.5

	return authSuccess && timeoutCount == 0 && minSuccessRate
}

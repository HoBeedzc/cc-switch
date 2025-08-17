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
	"strings"
	"time"
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

		// Only test chat endpoint if not quick mode and auth is working
		if len(result.Tests) > 0 && result.Tests[0].Status == "success" {
			result.Tests = append(result.Tests, t.testChatEndpoint(credentials))
		}
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
			Method:       "GET",
			Status:       "failed",
			ResponseTime: time.Since(start),
			Error:        fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	req.Header.Set("Authorization", "Bearer "+credentials.APIKey)
	req.Header.Set("User-Agent", "cc-switch-test/1.0")
	req.Header.Set("anthropic-version", credentials.Version)

	resp, err := t.httpClient.Do(req)
	duration := time.Since(start)

	test := EndpointTest{
		Endpoint:     endpoint,
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
			Method:       "GET",
			Status:       "failed",
			ResponseTime: time.Since(start),
			Error:        fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	req.Header.Set("Authorization", "Bearer "+credentials.APIKey)
	req.Header.Set("User-Agent", "cc-switch-test/1.0")
	req.Header.Set("anthropic-version", credentials.Version)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := t.httpClient.Do(req)
	duration := time.Since(start)

	test := EndpointTest{
		Endpoint:     endpoint,
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

// testChatEndpoint tests the chat/messages endpoint with a minimal request
func (t *APITester) testChatEndpoint(credentials *APICredentials) EndpointTest {
	start := time.Now()

	endpoint := "/v1/messages"
	url := strings.TrimSuffix(credentials.BaseURL, "/") + endpoint

	// Create minimal test payload
	payload := map[string]interface{}{
		"model": "claude-3-5-haiku-20241022", // Use fastest/cheapest model for testing
		"messages": []map[string]string{
			{"role": "user", "content": "Hi"},
		},
		"max_tokens": 1, // Minimal token usage
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return EndpointTest{
			Endpoint:     endpoint,
			Method:       "POST",
			Status:       "failed",
			ResponseTime: time.Since(start),
			Error:        fmt.Sprintf("Failed to create payload: %v", err),
		}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return EndpointTest{
			Endpoint:     endpoint,
			Method:       "POST",
			Status:       "failed",
			ResponseTime: time.Since(start),
			Error:        fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	req.Header.Set("Authorization", "Bearer "+credentials.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", credentials.Version)
	req.Header.Set("User-Agent", "cc-switch-test/1.0")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := t.httpClient.Do(req)
	duration := time.Since(start)

	test := EndpointTest{
		Endpoint:     endpoint,
		Method:       "POST",
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
		test.Status = "success"
		test.Details = "Chat endpoint functional"
	} else {
		test.Status = "failed"
		body, _ := io.ReadAll(resp.Body)
		test.Error = fmt.Sprintf("Status %d: %s", resp.StatusCode, string(body))
	}

	return test
}

// aggregateResults determines overall connectivity status from individual test results
func (t *APITester) aggregateResults(tests []EndpointTest) bool {
	if len(tests) == 0 {
		return false
	}

	// For basic connectivity, we just need one test to pass
	successCount := 0
	for _, test := range tests {
		if test.Status == "success" {
			successCount++
		}
	}

	// Consider configuration connectable if at least one critical test passes
	// (authentication or basic connectivity)
	return successCount > 0
}

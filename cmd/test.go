package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [profile-name]",
	Short: "Test configuration API connectivity",
	Long: `Test Claude Code configuration by making actual API requests
to verify authentication, connectivity, and service availability.

Modes:
- Interactive: cc-switch test (no arguments) or cc-switch test -i
- CLI: cc-switch test <profile-name>
- Current: cc-switch test -c or cc-switch test --current
- All: cc-switch test --all

The interactive mode allows you to browse and select configurations to test.

Examples:
  cc-switch test                    # Interactive mode - select configuration to test
  cc-switch test work-config        # Test specific configuration
  cc-switch test -c                 # Test current configuration
  cc-switch test --all             # Test all configurations
  cc-switch test --quick           # Quick connectivity test only
  cc-switch test --verbose         # Show detailed request/response info`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTest,
}

func init() {
	testCmd.Flags().BoolP("all", "a", false, "Test all configurations")
	testCmd.Flags().BoolP("current", "c", false, "Test current configuration")
	testCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
	testCmd.Flags().BoolP("verbose", "v", false, "Show detailed request/response information")
	testCmd.Flags().BoolP("quick", "q", false, "Quick test (basic connectivity only)")
	testCmd.Flags().String("endpoint", "", "Test specific endpoint (chat, models, auth)")
	testCmd.Flags().Duration("timeout", 30*time.Second, "Request timeout")
	testCmd.Flags().Bool("json", false, "Output results in JSON format")
}

func runTest(cmd *cobra.Command, args []string) error {
	// Check Claude config exists
	if err := checkClaudeConfig(); err != nil {
		return err
	}

	// Create config manager and handler
	configManager, err := config.NewConfigManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	configHandler := handler.NewConfigHandler(configManager)

	// Parse flags
	interactiveFlag, _ := cmd.Flags().GetBool("interactive")
	currentFlag, _ := cmd.Flags().GetBool("current")
	allFlag, _ := cmd.Flags().GetBool("all")

	// Validate flag combinations
	flagCount := 0
	if currentFlag {
		flagCount++
	}
	if allFlag {
		flagCount++
	}
	if len(args) > 0 {
		flagCount++
	}

	if flagCount > 1 {
		return fmt.Errorf("cannot use multiple operation flags together")
	}

	if (currentFlag || allFlag) && interactiveFlag {
		return fmt.Errorf("cannot use operation flags with -i/--interactive")
	}

	// Parse test options
	options := handler.TestOptions{
		Quick:      cmd.Flag("quick").Value.String() == "true",
		Verbose:    cmd.Flag("verbose").Value.String() == "true",
		JSONOutput: cmd.Flag("json").Value.String() == "true",
		Timeout:    parseDuration(cmd.Flag("timeout").Value.String()),
	}

	// Parse endpoint filter if provided
	if endpoint := cmd.Flag("endpoint").Value.String(); endpoint != "" {
		options.Endpoints = []string{endpoint}
	}

	// Create UI provider based on mode
	var uiProvider ui.UIProvider
	if !currentFlag && !allFlag && ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
		uiProvider = ui.NewInteractiveUI()
	} else {
		uiProvider = ui.NewCLIUI()
	}

	// Handle special operations
	if allFlag {
		return runTestAll(configHandler, uiProvider, options)
	}

	if currentFlag {
		return runTestCurrent(configHandler, uiProvider, options)
	}

	// Execute normal test operation
	return executeTest(configHandler, uiProvider, args, options)
}

// executeTest handles the test operation with the given dependencies
func executeTest(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, options handler.TestOptions) error {
	// Get all configurations for interactive mode
	profiles, err := configHandler.ListConfigs()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		uiProvider.ShowWarning("No configurations found.")
		fmt.Println("Use 'cc-switch new <name>' to create your first configuration.")
		return nil
	}

	var targetName string

	// Determine execution mode
	if len(args) == 0 {
		// Interactive mode - use enhanced selector
		if ui.NewInteractiveUI().DetectMode(false, args) == ui.Interactive {
			interactiveUI := uiProvider.(ui.InteractiveUI)
			selected, err := interactiveUI.SelectConfiguration(profiles, "test")
			if err != nil {
				return fmt.Errorf("selection cancelled: %w", err)
			}
			targetName = selected.Name
		} else {
			// Fallback to regular selection
			selected, err := uiProvider.SelectConfiguration(profiles, "test")
			if err != nil {
				return fmt.Errorf("selection cancelled: %w", err)
			}
			targetName = selected.Name
		}
	} else {
		// CLI mode
		targetName = args[0]
	}

	return runTestSingle(configHandler, uiProvider, targetName, options)
}

// runTestCurrent tests the current configuration
func runTestCurrent(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, options handler.TestOptions) error {
	if !options.JSONOutput {
		uiProvider.ShowInfo("Testing current configuration...")
	}

	result, err := configHandler.TestCurrentConfiguration(options)
	if err != nil {
		return fmt.Errorf("failed to test current configuration: %w", err)
	}

	return displaySingleResultWithUI(uiProvider, result, options)
}

func runTestSingle(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, profileName string, options handler.TestOptions) error {
	if !options.JSONOutput {
		uiProvider.ShowInfo("Testing configuration: %s", profileName)
	}

	result, err := configHandler.TestAPIConnectivity(profileName, options)
	if err != nil {
		return fmt.Errorf("failed to test configuration: %w", err)
	}

	return displaySingleResultWithUI(uiProvider, result, options)
}

func runTestAll(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, options handler.TestOptions) error {
	if !options.JSONOutput {
		uiProvider.ShowInfo("Testing all configurations...")
		fmt.Println()
	}

	results, err := configHandler.TestAllConfigurations(options)
	if err != nil {
		return fmt.Errorf("failed to test configurations: %w", err)
	}

	return displayAllResultsWithUI(uiProvider, results, options)
}

func displayJSONResult(result *handler.APITestResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON output: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func displayJSONResults(results []handler.APITestResult) error {
	output := map[string]interface{}{
		"tested_at": time.Now(),
		"results":   results,
		"summary": map[string]interface{}{
			"total_tested":  len(results),
			"valid_count":   countValidResults(results),
			"invalid_count": len(results) - countValidResults(results),
		},
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON output: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func getStatusSymbol(status string) string {
	switch status {
	case "success":
		return "✅"
	case "failed":
		return "❌"
	case "timeout":
		return "⏱️"
	default:
		return "❓"
	}
}

func formatTestDescription(test handler.EndpointTest) string {
	baseDesc := ""
	switch test.Endpoint {
	case "/v1/models":
		if test.Method == "GET" {
			baseDesc = "Authentication Test"
		} else if test.Method == "GET-MODELS" {
			baseDesc = "Models Endpoint"
		} else {
			baseDesc = "Models Endpoint"
		}
	case "/v1/messages":
		if test.Method == "claude-cli" {
			baseDesc = "Chat Endpoint (Claude CLI)"
		} else {
			baseDesc = "Chat Endpoint"
		}
	default:
		if test.Method == "HEAD" {
			baseDesc = "Basic Connectivity"
		} else if test.Method == "claude-cli" {
			baseDesc = "Claude CLI Test"
		} else {
			baseDesc = fmt.Sprintf("%s %s", test.Method, test.Endpoint)
		}
	}

	if test.ResponseTime > 0 {
		return fmt.Sprintf("%s (%s)", baseDesc, formatDuration(test.ResponseTime))
	}
	return baseDesc
}

func formatVerboseTestDetails(test handler.EndpointTest) string {
	var details []string

	// Show full URL in verbose mode if available, otherwise show endpoint
	if test.FullURL != "" {
		details = append(details, fmt.Sprintf("  Endpoint: %s", test.FullURL))
	} else {
		details = append(details, fmt.Sprintf("  Endpoint: %s", test.Endpoint))
	}

	details = append(details, fmt.Sprintf("  Method: %s", test.Method))
	if test.StatusCode > 0 {
		details = append(details, fmt.Sprintf("  Status Code: %d", test.StatusCode))
	}
	details = append(details, fmt.Sprintf("  Response Time: %s", formatDuration(test.ResponseTime)))

	if test.Details != "" {
		details = append(details, fmt.Sprintf("  Details: %s", test.Details))
	}
	if test.Error != "" {
		details = append(details, fmt.Sprintf("  Error: %s", test.Error))
	}

	return strings.Join(details, "\n")
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.1fμs", float64(d)/float64(time.Microsecond))
	}
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d)/float64(time.Millisecond))
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func countValidResults(results []handler.APITestResult) int {
	count := 0
	for _, result := range results {
		if result.IsConnectable {
			count++
		}
	}
	return count
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Second // default
	}
	return d
}

// UI-based display functions
func displaySingleResultWithUI(uiProvider ui.UIProvider, result *handler.APITestResult, options handler.TestOptions) error {
	if options.JSONOutput {
		return displayJSONResult(result)
	}

	// Display header and handle error case
	if result.Error != "" {
		uiProvider.ShowError(fmt.Errorf("❌ %s", result.Error))
		return nil
	}

	// Display test results
	for _, test := range result.Tests {
		symbol := getStatusSymbol(test.Status)
		message := fmt.Sprintf("%s %s", symbol, formatTestDescription(test))

		if test.Status == "success" {
			uiProvider.ShowSuccess(message)
		} else {
			uiProvider.ShowError(fmt.Errorf("%s", message))
		}

		// Show details in verbose mode
		if options.Verbose {
			fmt.Println(formatVerboseTestDetails(test))
		}
	}

	// Display summary
	if result.IsConnectable {
		uiProvider.ShowSuccess("✅ Result: Configuration is functional")
		fmt.Printf("   Total response time: %s\n", formatDuration(result.ResponseTime))
	} else {
		uiProvider.ShowError(fmt.Errorf("❌ Result: Configuration has connectivity issues"))
	}

	return nil
}

func displayAllResultsWithUI(uiProvider ui.UIProvider, results []handler.APITestResult, options handler.TestOptions) error {
	if options.JSONOutput {
		return displayJSONResults(results)
	}

	validCount := 0
	totalCount := len(results)

	for _, result := range results {
		symbol := "❌"
		status := "Invalid"
		details := ""

		if result.Error != "" {
			details = fmt.Sprintf(" (%s)", result.Error)
		} else if result.IsConnectable {
			symbol = "✅"
			status = "Valid"
			validCount++
			if !options.Quick {
				successCount := 0
				for _, test := range result.Tests {
					if test.Status == "success" {
						successCount++
					}
				}
				if successCount < len(result.Tests) {
					symbol = "⚠️"
					status = fmt.Sprintf("Valid with warnings (%d/%d tests passed)", successCount, len(result.Tests))
				}
			}
		} else {
			failedTests := 0
			for _, test := range result.Tests {
				if test.Status == "failed" {
					failedTests++
				}
			}
			if failedTests > 0 {
				details = fmt.Sprintf(" (%d errors)", failedTests)
			}
		}

		message := fmt.Sprintf("%-20s %s %s%s", result.ProfileName, symbol, status, details)

		if result.IsConnectable {
			uiProvider.ShowSuccess(message)
		} else {
			uiProvider.ShowError(fmt.Errorf("%s", message))
		}

		// Show individual test results in verbose mode
		if options.Verbose && len(result.Tests) > 0 {
			for _, test := range result.Tests {
				symbol := getStatusSymbol(test.Status)
				testMsg := fmt.Sprintf("  └─ %s %s", symbol, formatTestDescription(test))
				if test.Status == "success" {
					uiProvider.ShowSuccess(testMsg)
				} else {
					uiProvider.ShowError(fmt.Errorf("%s", testMsg))
				}
			}
		}
	}

	// Display summary
	summaryMsg := fmt.Sprintf("Summary: %d/%d configurations functional", validCount, totalCount)
	if validCount == totalCount {
		uiProvider.ShowSuccess("✅ %s", summaryMsg)
	} else if validCount > 0 {
		uiProvider.ShowWarning("⚠️  %s", summaryMsg)
	} else {
		uiProvider.ShowError(fmt.Errorf("❌ %s", summaryMsg))
	}

	return nil
}

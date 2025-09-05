package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"

	"github.com/fatih/color"
)

// cliUI implements basic CLI UI operations
type cliUI struct{}

// NewCLIUI creates a new CLI UI provider
func NewCLIUI() UIProvider {
	return &cliUI{}
}

// DetectMode for CLI always returns CLI mode
func (ui *cliUI) DetectMode(hasInteractiveFlag bool, args []string) ExecutionMode {
	if hasInteractiveFlag {
		return Interactive
	}
	return CLI
}

// SelectConfiguration is not supported in CLI mode
func (ui *cliUI) SelectConfiguration(configs []config.Profile, action string) (*config.Profile, error) {
	return nil, fmt.Errorf("configuration selection not supported in CLI mode")
}

// SelectConfigurationWithEmptyMode is not supported in CLI mode
func (ui *cliUI) SelectConfigurationWithEmptyMode(configs []config.Profile, action string, isEmptyMode bool) (*SpecialSelection, error) {
	return nil, fmt.Errorf("configuration selection not supported in CLI mode")
}

// SelectAction is not supported in CLI mode
func (ui *cliUI) SelectAction(selectedConfig *config.Profile) (string, error) {
	return "", fmt.Errorf("action selection not supported in CLI mode")
}

// ConfirmAction provides basic CLI confirmation
func (ui *cliUI) ConfirmAction(message string, defaultValue bool) bool {
	var defaultStr string
	if defaultValue {
		defaultStr = "Y/n"
	} else {
		defaultStr = "y/N"
	}

	fmt.Printf("%s (%s): ", message, defaultStr)
	var response string
	fmt.Scanln(&response)

	if response == "" {
		return defaultValue
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// GetInput prompts for user input with a default value
func (ui *cliUI) GetInput(prompt string, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	var input string
	fmt.Scanln(&input)

	input = strings.TrimSpace(input)
	if input == "" && defaultValue != "" {
		return defaultValue, nil
	}

	if input == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	return input, nil
}

// Init-specific operations

// GetInitInput prompts for initialization input with special handling for empty values
func (ui *cliUI) GetInitInput(fieldName, description string) (string, error) {
	fmt.Printf("? %s: ", description)

	var input string
	fmt.Scanln(&input)

	// For init, empty values are allowed
	return strings.TrimSpace(input), nil
}

// ShowInitWelcome displays welcome message for initialization
func (ui *cliUI) ShowInitWelcome() {
	color.Cyan("🚀 Welcome to Claude Code configuration setup!")
	fmt.Println()
	color.Yellow("This will create your initial Claude Code configuration.")
	fmt.Println("You can leave fields empty if you don't have the information yet.")
	fmt.Println()
}

// ShowInitSuccess displays success message after initialization
func (ui *cliUI) ShowInitSuccess() {
	fmt.Println()
	color.Green("✓ Configuration created successfully")
	color.Green("✓ Default profile 'default' created")
	color.Green("✓ cc-switch directory structure initialized")
	fmt.Println()
	color.Cyan("🎉 Your Claude Code configuration is ready!")
	color.White("You can now use 'cc-switch' commands to manage your configurations.")
}

// ShowAlreadyInitialized displays message when configuration already exists
func (ui *cliUI) ShowAlreadyInitialized() {
	color.Yellow("⚠ Claude Code configuration already exists")
	fmt.Println()
	fmt.Println("If you want to reconfigure, you can:")
	color.Cyan("  • Use 'cc-switch edit default' to modify existing configuration")
	color.Cyan("  • Use 'cc-switch new <name>' to create additional configurations")
	color.Cyan("  • Manually backup and remove settings.json to reinitialize")
}

// ShowError displays error messages
func (ui *cliUI) ShowError(err error) {
	color.Red("Error: %v", err)
}

// ShowSuccess displays success messages
func (ui *cliUI) ShowSuccess(message string, args ...interface{}) {
	color.Green("✓ "+message, args...)
}

// ShowWarning displays warning messages
func (ui *cliUI) ShowWarning(message string, args ...interface{}) {
	color.Yellow("⚠ "+message, args...)
}

// ShowInfo displays informational messages
func (ui *cliUI) ShowInfo(message string, args ...interface{}) {
	color.Cyan("ℹ "+message, args...)
}

// DisplayConfiguration displays configuration content
func (ui *cliUI) DisplayConfiguration(view *handler.ConfigView, raw bool) error {
	if raw {
		// Raw JSON output
		jsonData, err := json.MarshalIndent(view.Content, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		// Formatted output
		color.Blue("Configuration: %s", view.Name)
		if view.IsCurrent {
			color.Green("Status: Current")
		} else {
			color.White("Status: Available")
		}
		fmt.Printf("Path: %s\n\n", view.Path)

		color.Yellow("Content:")
		jsonData, err := json.MarshalIndent(view.Content, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	}

	return nil
}

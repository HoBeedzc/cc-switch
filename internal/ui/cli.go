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
			fmt.Println("Status: Available")
		}
		fmt.Printf("Path: %s\n", view.Path)
		fmt.Println()

		color.Yellow("Content:")
		jsonData, err := json.MarshalIndent(view.Content, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	}

	return nil
}

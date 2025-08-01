package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// interactiveUI implements the InteractiveUI interface
type interactiveUI struct{}

// NewInteractiveUI creates a new interactive UI provider
func NewInteractiveUI() InteractiveUI {
	return &interactiveUI{}
}

// DetectMode determines the execution mode based on flags and arguments
func (ui *interactiveUI) DetectMode(hasInteractiveFlag bool, args []string) ExecutionMode {
	// Explicit interactive flag
	if hasInteractiveFlag {
		return Interactive
	}

	// No arguments automatically enters interactive mode
	if len(args) == 0 {
		return Interactive
	}

	return CLI
}

// SelectConfiguration shows an interactive configuration selector
func (ui *interactiveUI) SelectConfiguration(configs []config.Profile, action string) (*config.Profile, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no configurations available")
	}

	// Custom templates for better visual experience
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "▶ {{ .Name | cyan }}{{ if .IsCurrent }} {{ \"(current)\" | green }}{{ end }}",
		Inactive: "  {{ .Name }}{{ if .IsCurrent }} {{ \"(current)\" | faint }}{{ end }}",
		Selected: "✓ {{ .Name | green }}{{ if .IsCurrent }} {{ \"(current)\" | faint }}{{ end }}",
		Details: `
--------- Configuration Details ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Status:" | faint }}	{{ if .IsCurrent }}{{ "Current" | green }}{{ else }}{{ "Available" | yellow }}{{ end }}
{{ "Path:" | faint }}	{{ .Path }}`,
	}

	prompt := promptui.Select{
		Label:        fmt.Sprintf("Select configuration to %s", action),
		Items:        configs,
		Templates:    templates,
		Size:         10,
		HideSelected: false,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &configs[i], nil
}

// SelectWithPreview provides configuration selection with preview
func (ui *interactiveUI) SelectWithPreview(configs []config.Profile, action string) (*config.Profile, error) {
	// For now, use the same implementation as SelectConfiguration
	// This can be enhanced later with preview functionality
	return ui.SelectConfiguration(configs, action)
}

// SelectAction shows action selection menu (legacy support)
func (ui *interactiveUI) SelectAction(selectedConfig *config.Profile) (string, error) {
	return ui.ShowActionMenu(selectedConfig)
}

// ShowActionMenu displays an action selection menu
func (ui *interactiveUI) ShowActionMenu(selectedConfig *config.Profile) (string, error) {
	actions := []string{"View", "Edit", "Use", "Delete", "Cancel"}

	// If it's the current configuration, cannot delete
	if selectedConfig.IsCurrent {
		actions = []string{"View", "Edit", "Cancel"}
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "▶ {{ . | cyan }}",
		Inactive: "  {{ . }}",
		Selected: "✓ {{ . | green }}",
	}

	prompt := promptui.Select{
		Label:     fmt.Sprintf("What do you want to do with '%s'?", selectedConfig.Name),
		Items:     actions,
		Templates: templates,
		Size:      len(actions),
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return strings.ToLower(result), nil
}

// ConfirmAction asks for user confirmation
func (ui *interactiveUI) ConfirmAction(message string, defaultValue bool) bool {
	var defaultStr string
	if defaultValue {
		defaultStr = "Y/n"
	} else {
		defaultStr = "y/N"
	}

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("%s (%s)", message, defaultStr),
		IsConfirm: true,
		Default:   fmt.Sprintf("%t", defaultValue),
	}

	result, err := prompt.Run()
	if err != nil {
		return false
	}

	result = strings.ToLower(strings.TrimSpace(result))
	if result == "" {
		return defaultValue
	}

	return result == "y" || result == "yes" || result == "true"
}

// GetUserInput prompts for user input
func (ui *interactiveUI) GetUserInput(prompt string) (string, error) {
	promptUI := promptui.Prompt{
		Label: prompt,
	}

	result, err := promptUI.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}

// GetFieldInput prompts for field-specific input
func (ui *interactiveUI) GetFieldInput(fieldName string, currentValue interface{}) (interface{}, error) {
	// Display current value
	fmt.Printf("Current value of '%s': ", fieldName)
	if currentValue != nil {
		currentValueJson, _ := json.Marshal(currentValue)
		fmt.Println(string(currentValueJson))
	} else {
		fmt.Println("<not set>")
	}

	// Get new value
	input, err := ui.GetUserInput("Enter new value (JSON format)")
	if err != nil {
		return nil, err
	}

	if input == "" {
		return nil, fmt.Errorf("no value provided")
	}

	// Parse JSON input
	var newValue interface{}
	if err := json.Unmarshal([]byte(input), &newValue); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	return newValue, nil
}

// ShowError displays error messages
func (ui *interactiveUI) ShowError(err error) {
	color.Red("Error: %v", err)
}

// ShowSuccess displays success messages
func (ui *interactiveUI) ShowSuccess(message string, args ...interface{}) {
	color.Green("✓ "+message, args...)
}

// ShowWarning displays warning messages
func (ui *interactiveUI) ShowWarning(message string, args ...interface{}) {
	color.Yellow("⚠ "+message, args...)
}

// ShowInfo displays informational messages
func (ui *interactiveUI) ShowInfo(message string, args ...interface{}) {
	color.Cyan("ℹ "+message, args...)
}

// DisplayConfiguration displays configuration content
func (ui *interactiveUI) DisplayConfiguration(view *handler.ConfigView, raw bool) error {
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

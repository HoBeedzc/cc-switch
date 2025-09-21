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

// SelectConfigurationWithEmptyMode shows an interactive selector including empty mode options
func (ui *interactiveUI) SelectConfigurationWithEmptyMode(configs []config.Profile, action string, isEmptyMode bool) (*SpecialSelection, error) {
	// Create special items for the selector
	type SelectItem struct {
		Name        string
		Type        string
		IsCurrent   bool
		IsSpecial   bool
		Profile     *config.Profile
		Description string
	}

	var items []SelectItem

	// Add special options based on mode
	if isEmptyMode {
		items = append(items, SelectItem{
			Name:        "<Restore Previous>",
			Type:        "restore",
			IsSpecial:   true,
			Description: "Restore from empty mode to previous configuration",
		})
	} else {
		items = append(items, SelectItem{
			Name:        "<Empty Mode>",
			Type:        "empty_mode",
			IsSpecial:   true,
			Description: "Disable all configurations temporarily",
		})
	}

	// Add regular configurations
	for i := range configs {
		items = append(items, SelectItem{
			Name:        configs[i].Name,
			Type:        "profile",
			IsCurrent:   configs[i].IsCurrent,
			IsSpecial:   false,
			Profile:     &configs[i],
			Description: fmt.Sprintf("Switch to %s configuration", configs[i].Name),
		})
	}

	// Custom templates
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "▶ {{ if .IsSpecial }}{{ .Name | yellow }}{{ else }}{{ .Name | cyan }}{{ end }}{{ if .IsCurrent }} {{ \"(current)\" | green }}{{ end }}",
		Inactive: "  {{ if .IsSpecial }}{{ .Name | faint }}{{ else }}{{ .Name }}{{ end }}{{ if .IsCurrent }} {{ \"(current)\" | faint }}{{ end }}",
		Selected: "✓ {{ if .IsSpecial }}{{ .Name | yellow }}{{ else }}{{ .Name | green }}{{ end }}{{ if .IsCurrent }} {{ \"(current)\" | faint }}{{ end }}",
		Details: `
--------- Selection Details ----------
{{ "Option:" | faint }}	{{ .Name }}
{{ "Type:" | faint }}	{{ if .IsSpecial }}{{ "Special Action" | yellow }}{{ else }}{{ "Configuration" | green }}{{ end }}
{{ "Description:" | faint }}	{{ .Description }}{{ if .Profile }}
{{ "Path:" | faint }}	{{ .Profile.Path }}{{ end }}`,
	}

	prompt := promptui.Select{
		Label:        fmt.Sprintf("Select configuration to %s", action),
		Items:        items,
		Templates:    templates,
		Size:         10,
		HideSelected: false,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	selected := &items[i]
	return &SpecialSelection{
		Type:    selected.Type,
		Profile: selected.Profile,
		Action:  selected.Type,
	}, nil
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

// GetInput prompts for user input with a default value
func (ui *interactiveUI) GetInput(prompt string, defaultValue string) (string, error) {
	promptUI := promptui.Prompt{
		Label:   prompt,
		Default: defaultValue,
	}

	result, err := promptUI.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}

// Template field input operations

// GetTemplateFieldInput prompts for template field input using promptui
func (ui *interactiveUI) GetTemplateFieldInput(field config.TemplateField) (string, error) {
	// Build label with description and required indicator
	label := field.Description
	if field.Required {
		label += " (required)"
	}

	promptUI := promptui.Prompt{
		Label: label,
	}

	// Add validation for required fields and field-specific validation
	promptUI.Validate = func(input string) error {
		input = strings.TrimSpace(input)

		// Check required fields
		if field.Required && input == "" {
			return fmt.Errorf("this field is required and cannot be empty")
		}

		// Apply field-specific validation
		return validateFieldValueUI(field.Name, input)
	}

	result, err := promptUI.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get input for field '%s': %w", field.Name, err)
	}

	return strings.TrimSpace(result), nil
}

// validateFieldValueUI performs client-side validation
func validateFieldValueUI(fieldName, value string) error {
	if value == "" {
		return nil // Empty values handled by required field check
	}

	switch fieldName {
	case "ANTHROPIC_AUTH_TOKEN", "OPENAI_API_KEY", "API_KEY", "TOKEN":
		if len(value) < 10 {
			return fmt.Errorf("API token appears to be too short (minimum 10 characters)")
		}
		if strings.Contains(value, " ") {
			return fmt.Errorf("API token should not contain spaces")
		}
	case "ANTHROPIC_BASE_URL", "BASE_URL", "ENDPOINT":
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return fmt.Errorf("URL must start with http:// or https://")
		}
		if strings.Contains(value, " ") {
			return fmt.Errorf("URL should not contain spaces")
		}
	}

	return nil
}

// ConfirmTemplateCreation asks for confirmation with enhanced display
func (ui *interactiveUI) ConfirmTemplateCreation(fields []config.TemplateField) bool {
	prompt := promptui.Prompt{
		Label:     "Continue with interactive template field input?",
		IsConfirm: true,
		Default:   "n",
	}

	result, err := prompt.Run()
	if err != nil {
		return false
	}

	result = strings.ToLower(strings.TrimSpace(result))
	if result == "" {
		return false // 默认值：不继续（退出）
	}
	return result == "y" || result == "yes" || result == "true"
}

// ShowTemplateFieldSummary shows an enhanced summary using colors and formatting
func (ui *interactiveUI) ShowTemplateFieldSummary(fields []config.TemplateField) {
	if len(fields) == 0 {
		return
	}

	color.Cyan("📝 Template Configuration Summary")
	fmt.Printf("Template has %d empty field(s) that need to be filled:\n\n", len(fields))

	for _, field := range fields {
		if field.Required {
			color.Yellow("  📋 %s", field.Name)
			color.Red("      • Required field")
		} else {
			color.White("  📋 %s", field.Name)
			color.Green("      • Optional field")
		}
		color.Cyan("      • %s", field.Description)
		fmt.Println()
	}
	fmt.Println()
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

// DisplayTemplate displays template content
func (ui *interactiveUI) DisplayTemplate(view *handler.TemplateView, raw bool) error {
	if raw {
		// Raw JSON output
		jsonData, err := json.MarshalIndent(view.Content, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		// Formatted output
		color.Blue("Template: %s", view.Name)
		if view.Name == "default" {
			color.Green("Type: System Default")
		} else {
			fmt.Println("Type: Custom Template")
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

// Init-specific operations (reuse CLI implementation)

// GetInitInput prompts for initialization input using promptui
func (ui *interactiveUI) GetInitInput(fieldName, description string) (string, error) {
	prompt := promptui.Prompt{
		Label:   description,
		Default: "",
	}

	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get input: %w", err)
	}

	return strings.TrimSpace(result), nil
}

// ShowInitWelcome displays welcome message for initialization
func (ui *interactiveUI) ShowInitWelcome() {
	color.Cyan("🚀 Welcome to Claude Code configuration setup!")
	fmt.Println()
	color.Yellow("This will create your initial Claude Code configuration.")
	fmt.Println("You can leave fields empty if you don't have the information yet.")
	fmt.Println()
}

// ShowInitSuccess displays success message after initialization
func (ui *interactiveUI) ShowInitSuccess() {
	fmt.Println()
	color.Green("✓ Configuration created successfully")
	color.Green("✓ Default profile 'default' created")
	color.Green("✓ cc-switch directory structure initialized")
	fmt.Println()
	color.Cyan("🎉 Your Claude Code configuration is ready!")
	color.White("You can now use 'cc-switch' commands to manage your configurations.")
}

// ShowAlreadyInitialized displays message when configuration already exists
func (ui *interactiveUI) ShowAlreadyInitialized() {
	color.Yellow("⚠ Claude Code configuration already exists")
	fmt.Println()
	fmt.Println("If you want to reconfigure, you can:")
	color.Cyan("  • Use 'cc-switch edit default' to modify existing configuration")
	color.Cyan("  • Use 'cc-switch new <name>' to create additional configurations")
	color.Cyan("  • Manually backup and remove settings.json to reinitialize")
}

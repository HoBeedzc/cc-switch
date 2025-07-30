package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cc-switch/internal/config"
)

// configHandler implements the ConfigHandler interface
type configHandler struct {
	configManager *config.ConfigManager
}

// NewConfigHandler creates a new config handler instance
func NewConfigHandler(cm *config.ConfigManager) ConfigHandler {
	return &configHandler{
		configManager: cm,
	}
}

// ListConfigs returns all available configurations
func (h *configHandler) ListConfigs() ([]config.Profile, error) {
	return h.configManager.ListProfiles()
}

// DeleteConfig deletes a configuration with optional force flag
func (h *configHandler) DeleteConfig(name string, force bool) error {
	// Validate configuration exists
	if err := h.ValidateConfigExists(name); err != nil {
		return err
	}

	// Check if it's the current configuration
	if h.IsCurrentConfig(name) {
		return fmt.Errorf("cannot delete current configuration '%s'. Switch to another configuration first", name)
	}

	// Delete the configuration
	return h.configManager.DeleteProfile(name)
}

// UseConfig switches to the specified configuration
func (h *configHandler) UseConfig(name string) error {
	// Validate configuration exists
	if err := h.ValidateConfigExists(name); err != nil {
		return err
	}

	// Check if already current
	if h.IsCurrentConfig(name) {
		return fmt.Errorf("configuration '%s' is already active", name)
	}

	// Switch configuration
	return h.configManager.UseProfile(name)
}

// ViewConfig returns the configuration view
func (h *configHandler) ViewConfig(name string, raw bool) (*ConfigView, error) {
	// Validate configuration exists
	if err := h.ValidateConfigExists(name); err != nil {
		return nil, err
	}

	// Get configuration content
	content, metadata, err := h.configManager.GetProfileContent(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %w", err)
	}

	return &ConfigView{
		Name:      metadata.Name,
		IsCurrent: metadata.IsCurrent,
		Path:      metadata.Path,
		Content:   content,
	}, nil
}

// EditConfig edits a configuration
func (h *configHandler) EditConfig(name string, field string, useNano bool) error {
	// Validate configuration exists
	if err := h.ValidateConfigExists(name); err != nil {
		return err
	}

	if field != "" {
		// Field editing mode
		return h.editProfileField(name, field)
	} else {
		// Editor mode
		return h.editProfileWithEditor(name, useNano)
	}
}

// ValidateConfigExists checks if a configuration exists
func (h *configHandler) ValidateConfigExists(name string) error {
	if !h.configManager.ProfileExists(name) {
		return fmt.Errorf("configuration '%s' does not exist", name)
	}
	return nil
}

// GetCurrentConfig returns the current configuration name
func (h *configHandler) GetCurrentConfig() (string, error) {
	return h.configManager.GetCurrentProfile()
}

// IsCurrentConfig checks if the given configuration is current
func (h *configHandler) IsCurrentConfig(name string) bool {
	current, err := h.GetCurrentConfig()
	return err == nil && current == name
}

// editProfileField edits a specific field in the configuration
func (h *configHandler) editProfileField(name, field string) error {
	content, _, err := h.configManager.GetProfileContent(name)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	// Parse field path (supports nested fields, like "env.ANTHROPIC_API_KEY")
	fieldParts := strings.Split(field, ".")
	
	// Display current value
	currentValue := h.getNestedValue(content, fieldParts)
	fmt.Printf("Current value of '%s': ", field)
	if currentValue != nil {
		currentValueJson, _ := json.Marshal(currentValue)
		fmt.Println(string(currentValueJson))
	} else {
		fmt.Println("<not set>")
	}

	// Get new value
	fmt.Print("Enter new value (JSON format): ")
	var newValueStr string
	fmt.Scanln(&newValueStr)

	if newValueStr == "" {
		return fmt.Errorf("no value provided")
	}

	// Parse new value
	var newValue interface{}
	if err := json.Unmarshal([]byte(newValueStr), &newValue); err != nil {
		return fmt.Errorf("invalid JSON format for new value: %w", err)
	}

	// Set new value
	if err := h.setNestedValue(content, fieldParts, newValue); err != nil {
		return fmt.Errorf("failed to set field value: %w", err)
	}

	// Save changes
	if err := h.configManager.UpdateProfile(name, content); err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	return nil
}

// editProfileWithEditor uses system editor to edit configuration
func (h *configHandler) editProfileWithEditor(name string, useNano bool) error {
	// Get current configuration content
	content, _, err := h.configManager.GetProfileContent(name)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("cc-switch-%s-*.json", name))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write current content to temporary file
	jsonData, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	if _, err := tmpFile.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}
	tmpFile.Close()

	// Get file modification time (to detect changes)
	stat, err := os.Stat(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}
	originalModTime := stat.ModTime()

	// Determine editor
	editor := h.getEditor(useNano)

	// Launch editor
	fmt.Printf("Opening configuration '%s' in %s...\n", name, editor)
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// Check for changes
	stat, err = os.Stat(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to get file stats after editing: %w", err)
	}

	if stat.ModTime().Equal(originalModTime) {
		return fmt.Errorf("no changes detected")
	}

	// Read edited content
	editedData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// Validate JSON format
	var editedContent map[string]interface{}
	if err := json.Unmarshal(editedData, &editedContent); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Save changes
	if err := h.configManager.UpdateProfile(name, editedContent); err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	return nil
}

// getNestedValue retrieves a nested field value
func (h *configHandler) getNestedValue(data map[string]interface{}, fieldParts []string) interface{} {
	current := data
	for i, part := range fieldParts {
		if i == len(fieldParts)-1 {
			return current[part]
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	return nil
}

// setNestedValue sets a nested field value
func (h *configHandler) setNestedValue(data map[string]interface{}, fieldParts []string, value interface{}) error {
	current := data
	for i, part := range fieldParts {
		if i == len(fieldParts)-1 {
			current[part] = value
			return nil
		}
		if _, ok := current[part]; !ok {
			current[part] = make(map[string]interface{})
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return fmt.Errorf("cannot set nested field: intermediate field '%s' is not an object", part)
		}
	}
	return nil
}

// getEditor determines which editor to use based on priority
func (h *configHandler) getEditor(useNano bool) string {
	// 1. --nano flag has highest priority
	if useNano {
		return "nano"
	}
	
	// 2. EDITOR environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	
	// 3. Default to vim
	return "vim"
}
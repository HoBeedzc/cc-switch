package ui

import (
	"cc-switch/internal/config"
	"cc-switch/internal/handler"
)

// ExecutionMode defines the mode of execution
type ExecutionMode int

const (
	CLI ExecutionMode = iota
	Interactive
)

// UIProvider defines the interface for user interaction
type UIProvider interface {
	// Mode detection
	DetectMode(hasInteractiveFlag bool, args []string) ExecutionMode

	// Selection operations
	SelectConfiguration(configs []config.Profile, action string) (*config.Profile, error)
	SelectAction(selectedConfig *config.Profile) (string, error)

	// Confirmation operations
	ConfirmAction(message string, defaultValue bool) bool

	// Input operations
	GetInput(prompt string, defaultValue string) (string, error)

	// Display operations
	ShowError(err error)
	ShowSuccess(message string, args ...interface{})
	ShowWarning(message string, args ...interface{})
	ShowInfo(message string, args ...interface{})

	// Configuration display
	DisplayConfiguration(view *handler.ConfigView, raw bool) error
}

// InteractiveUI defines additional interactive-specific operations
type InteractiveUI interface {
	UIProvider

	// Advanced selection with preview
	SelectWithPreview(configs []config.Profile, action string) (*config.Profile, error)

	// Multi-step workflows
	ShowActionMenu(selectedConfig *config.Profile) (string, error)

	// Input operations
	GetUserInput(prompt string) (string, error)
	GetFieldInput(fieldName string, currentValue interface{}) (interface{}, error)
}

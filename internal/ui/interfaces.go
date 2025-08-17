package ui

import (
	"cc-switch/internal/config"
	"cc-switch/internal/handler"
)

// SpecialSelection represents a selection that could be a profile or a special action
type SpecialSelection struct {
	Type    string         // "profile", "empty_mode", "restore"
	Profile *config.Profile // nil for special actions
	Action  string         // action name for special selections
}

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
	SelectConfigurationWithEmptyMode(configs []config.Profile, action string, isEmptyMode bool) (*SpecialSelection, error)
	SelectAction(selectedConfig *config.Profile) (string, error)

	// Confirmation operations
	ConfirmAction(message string, defaultValue bool) bool

	// Input operations
	GetInput(prompt string, defaultValue string) (string, error)

	// Init-specific operations
	GetInitInput(fieldName, description string) (string, error)
	ShowInitWelcome()
	ShowInitSuccess()
	ShowAlreadyInitialized()

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

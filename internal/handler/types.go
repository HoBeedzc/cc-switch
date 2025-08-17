package handler

import (
	"cc-switch/internal/config"
)

// ConfigHandler defines the business logic interface for configuration operations
type ConfigHandler interface {
	// Configuration management operations
	ListConfigs() ([]config.Profile, error)
	DeleteConfig(name string, force bool) error
	UseConfig(name string) error
	ViewConfig(name string, raw bool) (*ConfigView, error)
	EditConfig(name string, field string, useNano bool) error

	// New configuration operations
	MoveConfig(oldName, newName string) error
	CopyConfig(sourceName, destName string) error

	// Template management operations
	ListTemplates() ([]string, error)
	CreateTemplate(name string) error
	EditTemplate(name string, field string, useNano bool) error
	DeleteTemplate(name string) error
	ValidateTemplateExists(name string) error

	// Init operations
	InitializeConfig(authToken, baseURL string) error
	IsConfigInitialized() bool

	// Helper operations
	ValidateConfigExists(name string) error
	GetCurrentConfig() (string, error)
	IsCurrentConfig(name string) bool
	GetPreviousConfig() (string, error)

	// Empty mode operations
	UseEmptyMode() error
	RestoreFromEmptyMode() error
	RestoreToPreviousFromEmptyMode() error
	IsEmptyMode() bool
	GetEmptyModeStatus() (*EmptyModeStatus, error)
}

// ConfigView represents the view of a configuration
type ConfigView struct {
	Name      string                 `json:"name"`
	IsCurrent bool                   `json:"is_current"`
	Path      string                 `json:"path"`
	Content   map[string]interface{} `json:"content"`
}

// DeleteResult represents the result of a delete operation
type DeleteResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UseResult represents the result of a use operation
type UseResult struct {
	Name     string `json:"name"`
	Previous string `json:"previous,omitempty"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

// EditResult represents the result of an edit operation
type EditResult struct {
	Name    string `json:"name"`
	Field   string `json:"field,omitempty"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// MoveResult represents the result of a move operation
type MoveResult struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CopyResult represents the result of a copy operation
type CopyResult struct {
	SourceName string `json:"source_name"`
	DestName   string `json:"dest_name"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

// TemplateView represents the view of a template
type TemplateView struct {
	Name    string                 `json:"name"`
	Path    string                 `json:"path"`
	Content map[string]interface{} `json:"content"`
}

// TemplateResult represents the result of a template operation
type TemplateResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// EmptyModeStatus represents the current empty mode status
type EmptyModeStatus struct {
	Enabled         bool   `json:"enabled"`
	PreviousProfile string `json:"previous_profile,omitempty"`
	CanRestore      bool   `json:"can_restore"`
	Timestamp       string `json:"timestamp,omitempty"`
}

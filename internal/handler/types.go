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

	// Helper operations
	ValidateConfigExists(name string) error
	GetCurrentConfig() (string, error)
	IsCurrentConfig(name string) bool
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

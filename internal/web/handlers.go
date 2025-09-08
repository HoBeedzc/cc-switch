package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"cc-switch/internal/handler"
)

// APIHandler handles API requests
type APIHandler struct {
	handler handler.ConfigHandler
}

// validateTemplateName validates template names to prevent path traversal attacks
func validateTemplateName(name string) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if len(name) > 255 {
		return fmt.Errorf("template name must be 255 characters or less")
	}

	// Check for path traversal attempts
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("template name contains forbidden characters")
	}

	// Check for invalid characters
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("template name can only contain letters, numbers, hyphens, and underscores")
	}

	return nil
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// HandleProfiles handles /api/profiles requests
func (api *APIHandler) HandleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.listProfiles(w, r)
	case http.MethodPost:
		api.createProfile(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleProfile handles /api/profiles/{name} requests
func (api *APIHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	// Extract path after /api/profiles/
	path := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		api.sendError(w, "Profile name is required", http.StatusBadRequest)
		return
	}

	profileName := parts[0]

	if len(parts) == 1 {
		// Simple profile operations: /api/profiles/{name}
		api.handleSingleProfile(w, r, profileName)
	} else if len(parts) == 2 {
		// Profile sub-operations: /api/profiles/{name}/{operation}
		operation := parts[1]
		api.handleProfileOperation(w, r, profileName, operation)
	} else {
		api.sendError(w, "Invalid profile path", http.StatusBadRequest)
	}
}

// handleSingleProfile handles basic CRUD operations on profiles
func (api *APIHandler) handleSingleProfile(w http.ResponseWriter, r *http.Request, profileName string) {
	switch r.Method {
	case http.MethodGet:
		api.getProfile(w, r, profileName)
	case http.MethodPut:
		api.updateProfile(w, r, profileName)
	case http.MethodDelete:
		api.deleteProfile(w, r, profileName)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleProfileOperation handles profile sub-operations like move, copy etc.
func (api *APIHandler) handleProfileOperation(w http.ResponseWriter, r *http.Request, profileName, operation string) {
	switch operation {
	case "move":
		api.moveProfile(w, r, profileName)
	default:
		api.sendError(w, fmt.Sprintf("Unknown operation: %s", operation), http.StatusBadRequest)
	}
}

// HandleCurrent handles /api/current requests
func (api *APIHandler) HandleCurrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentProfile, err := api.handler.GetCurrentConfig()
	if err != nil {
		api.sendError(w, fmt.Sprintf("Failed to get current config: %v", err), http.StatusInternalServerError)
		return
	}

	isEmptyMode := api.handler.IsEmptyMode()

	response := map[string]interface{}{
		"current":    currentProfile,
		"empty_mode": isEmptyMode,
	}

	if isEmptyMode {
		if status, err := api.handler.GetEmptyModeStatus(); err == nil {
			response["empty_mode_status"] = status
		}
	}

	api.sendSuccess(w, response)
}

// HandleSwitch handles /api/switch requests
func (api *APIHandler) HandleSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Profile string `json:"profile"`
		Restore bool   `json:"restore"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	var err error
	var message string

	if request.Restore {
		err = api.handler.RestoreFromEmptyMode()
		message = "Configuration restored from empty mode"
	} else if request.Profile == "" {
		err = api.handler.UseEmptyMode()
		message = "Switched to empty mode"
	} else {
		err = api.handler.UseConfig(request.Profile)
		message = fmt.Sprintf("Switched to configuration: %s", request.Profile)
	}

	if err != nil {
		api.sendError(w, fmt.Sprintf("Failed to switch configuration: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"message": message,
		"profile": request.Profile,
	})
}

// HandleTest handles /api/test requests
func (api *APIHandler) HandleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Profile string `json:"profile"`
		Quick   bool   `json:"quick"`
		Timeout int    `json:"timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	options := handler.TestOptions{
		Quick:   request.Quick,
		Timeout: time.Duration(request.Timeout) * time.Second,
	}

	if options.Timeout == 0 {
		options.Timeout = 10 * time.Second
	}

	var result *handler.APITestResult
	var err error

	if request.Profile == "" {
		result, err = api.handler.TestCurrentConfiguration(options)
	} else {
		result, err = api.handler.TestAPIConnectivity(request.Profile, options)
	}

	if err != nil {
		api.sendError(w, fmt.Sprintf("Test failed: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, result)
}

// HandleTemplates handles /api/templates requests
func (api *APIHandler) HandleTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.listTemplates(w, r)
	case http.MethodPost:
		api.createTemplate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleTemplateRoutes handles all /api/templates/{path} requests with routing logic
func (api *APIHandler) HandleTemplateRoutes(w http.ResponseWriter, r *http.Request) {
	// Extract path after /api/templates/
	path := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		api.sendError(w, "Template name is required", http.StatusBadRequest)
		return
	}

	templateName := parts[0]

	// Validate template name to prevent path traversal
	if err := validateTemplateName(templateName); err != nil {
		api.sendError(w, fmt.Sprintf("Invalid template name: %v", err), http.StatusBadRequest)
		return
	}

	if len(parts) == 1 {
		// Simple template operations: /api/templates/{name}
		api.handleSingleTemplate(w, r, templateName)
	} else if len(parts) == 2 {
		// Template operations: /api/templates/{name}/{operation}
		operation := parts[1]
		api.handleTemplateOperation(w, r, templateName, operation)
	} else {
		api.sendError(w, "Invalid template path", http.StatusBadRequest)
	}
}

func (api *APIHandler) handleSingleTemplate(w http.ResponseWriter, r *http.Request, templateName string) {
	switch r.Method {
	case http.MethodGet:
		api.getTemplate(w, r, templateName)
	case http.MethodPut:
		api.updateTemplate(w, r, templateName)
	case http.MethodDelete:
		api.deleteTemplate(w, r, templateName)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *APIHandler) handleTemplateOperation(w http.ResponseWriter, r *http.Request, templateName string, operation string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch operation {
	case "copy":
		api.copyTemplate(w, r, templateName)
	case "move":
		api.moveTemplate(w, r, templateName)
	default:
		api.sendError(w, fmt.Sprintf("Unknown operation: %s", operation), http.StatusBadRequest)
	}
}

// HandleHealth handles /api/health requests
func (api *APIHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	isInitialized := api.handler.IsConfigInitialized()

	health := map[string]interface{}{
		"status":      "ok",
		"initialized": isInitialized,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"version":     "1.0.0",
	}

	api.sendSuccess(w, health)
}

// Helper methods

func (api *APIHandler) listProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := api.handler.ListConfigs()
	if err != nil {
		api.sendError(w, fmt.Sprintf("Failed to list profiles: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"profiles": profiles,
	})
}

func (api *APIHandler) createProfile(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name     string `json:"name"`
		Template string `json:"template,omitempty"`
		Content  struct {
			Env         map[string]string `json:"env"`
			Permissions struct {
				Allow []string `json:"allow"`
				Deny  []string `json:"deny"`
			} `json:"permissions"`
			StatusLine map[string]interface{} `json:"statusLine"`
		} `json:"content,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		api.sendError(w, "Profile name is required", http.StatusBadRequest)
		return
	}

	// Check if profile already exists
	if err := api.handler.ValidateConfigExists(request.Name); err == nil {
		api.sendError(w, fmt.Sprintf("Profile '%s' already exists", request.Name), http.StatusConflict)
		return
	}

	var err error
	var message string

	// Check if custom content is provided
	if request.Content.Env != nil || len(request.Content.Permissions.Allow) > 0 || len(request.Content.Permissions.Deny) > 0 || request.Content.StatusLine != nil {
		// Create with custom content
		content := map[string]interface{}{
			"env": request.Content.Env,
			"permissions": map[string]interface{}{
				"allow": request.Content.Permissions.Allow,
				"deny":  request.Content.Permissions.Deny,
			},
			"statusLine": request.Content.StatusLine,
		}

		err = api.handler.CreateConfigWithContent(request.Name, content)
		if err != nil {
			api.sendError(w, fmt.Sprintf("Failed to create profile: %v", err), http.StatusInternalServerError)
			return
		}

		message = fmt.Sprintf("Profile '%s' created successfully with custom content", request.Name)
	} else {
		// Create from template
		template := request.Template
		if template == "" {
			template = "default"
		}

		// Check if template exists
		templates, err := api.handler.ListTemplates()
		if err != nil {
			api.sendError(w, "Failed to list templates", http.StatusInternalServerError)
			return
		}

		templateExists := false
		for _, t := range templates {
			if t == template {
				templateExists = true
				break
			}
		}

		if !templateExists && template != "default" {
			// Fallback to default template
			template = "default"
		}

		err = api.handler.CreateConfig(request.Name, template)
		if err != nil {
			api.sendError(w, fmt.Sprintf("Failed to create profile: %v", err), http.StatusInternalServerError)
			return
		}

		message = fmt.Sprintf("Profile '%s' created successfully from template '%s'", request.Name, template)
	}

	api.sendSuccess(w, map[string]interface{}{
		"message": message,
		"name":    request.Name,
	})
}

func (api *APIHandler) getProfile(w http.ResponseWriter, r *http.Request, profileName string) {
	view, err := api.handler.ViewConfig(profileName, false)
	if err != nil {
		api.sendError(w, fmt.Sprintf("Failed to get profile: %v", err), http.StatusNotFound)
		return
	}

	api.sendSuccess(w, view)
}

func (api *APIHandler) updateProfile(w http.ResponseWriter, r *http.Request, profileName string) {
	// For Raw JSON mode, we accept the entire configuration object
	// For Form mode, we accept the structured request format

	// First, try to decode as complete JSON configuration
	var completeConfig map[string]interface{}
	body := r.Body

	// Read the body first
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		api.sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Try to decode as complete configuration first
	if err := json.Unmarshal(bodyBytes, &completeConfig); err != nil {
		api.sendError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Check if this is a structured form request (has env, permissions, statusLine at top level)
	if env, hasEnv := completeConfig["env"]; hasEnv {
		if permissions, hasPerms := completeConfig["permissions"]; hasPerms {
			if statusLine, hasStatus := completeConfig["statusLine"]; hasStatus {
				// This looks like a form request, use it directly
				updatedConfig := map[string]interface{}{
					"env":         env,
					"permissions": permissions,
					"statusLine":  statusLine,
				}

				// Call the handler to update the configuration
				if err := api.handler.UpdateConfig(profileName, updatedConfig); err != nil {
					api.sendError(w, fmt.Sprintf("Failed to update profile: %v", err), http.StatusInternalServerError)
					return
				}

				api.sendSuccess(w, map[string]interface{}{
					"message": fmt.Sprintf("Profile '%s' updated successfully", profileName),
					"name":    profileName,
				})
				return
			}
		}
	}

	// Otherwise, treat it as raw JSON configuration - use the entire object
	if err := api.handler.UpdateConfig(profileName, completeConfig); err != nil {
		api.sendError(w, fmt.Sprintf("Failed to update profile: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"message": fmt.Sprintf("Profile '%s' updated successfully", profileName),
		"name":    profileName,
	})
}

func (api *APIHandler) deleteProfile(w http.ResponseWriter, r *http.Request, profileName string) {
	var request struct {
		Force bool `json:"force"`
	}

	json.NewDecoder(r.Body).Decode(&request) // Ignore errors for optional body

	err := api.handler.DeleteConfig(profileName, request.Force)
	if err != nil {
		api.sendError(w, fmt.Sprintf("Failed to delete profile: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"message": fmt.Sprintf("Profile '%s' deleted successfully", profileName),
		"name":    profileName,
	})
}

func (api *APIHandler) sendSuccess(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
	}
	api.sendJSON(w, response, http.StatusOK)
}

func (api *APIHandler) sendError(w http.ResponseWriter, message string, statusCode int) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	api.sendJSON(w, response, statusCode)
}

func (api *APIHandler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Template helper methods

func (api *APIHandler) listTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := api.handler.ListTemplates()
	if err != nil {
		api.sendError(w, fmt.Sprintf("Failed to list templates: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"templates": templates,
	})
}

func (api *APIHandler) createTemplate(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate template name to prevent path traversal
	if err := validateTemplateName(request.Name); err != nil {
		api.sendError(w, fmt.Sprintf("Invalid template name: %v", err), http.StatusBadRequest)
		return
	}

	// Prevent creation of default template
	if request.Name == "default" {
		api.sendError(w, "Cannot create template with reserved name 'default'", http.StatusBadRequest)
		return
	}

	// Check if template already exists
	if err := api.handler.ValidateTemplateExists(request.Name); err == nil {
		api.sendError(w, fmt.Sprintf("Template '%s' already exists", request.Name), http.StatusConflict)
		return
	}

	if err := api.handler.CreateTemplate(request.Name); err != nil {
		api.sendError(w, fmt.Sprintf("Failed to create template: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"message": fmt.Sprintf("Template '%s' created successfully", request.Name),
		"name":    request.Name,
	})
}

func (api *APIHandler) getTemplate(w http.ResponseWriter, r *http.Request, templateName string) {
	view, err := api.handler.ViewTemplate(templateName, false)
	if err != nil {
		api.sendError(w, fmt.Sprintf("Failed to get template: %v", err), http.StatusNotFound)
		return
	}

	api.sendSuccess(w, view)
}

func (api *APIHandler) updateTemplate(w http.ResponseWriter, r *http.Request, templateName string) {
	// Prevent modification of default template
	if templateName == "default" {
		api.sendError(w, "Cannot modify the default template", http.StatusForbidden)
		return
	}

	// Limit request body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var completeConfig map[string]interface{}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		api.sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(bodyBytes, &completeConfig); err != nil {
		api.sendError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate template exists before updating
	if err := api.handler.ValidateTemplateExists(templateName); err != nil {
		api.sendError(w, fmt.Sprintf("Template '%s' does not exist", templateName), http.StatusNotFound)
		return
	}

	// Update template using the handler
	if err := api.handler.UpdateTemplate(templateName, completeConfig); err != nil {
		api.sendError(w, fmt.Sprintf("Failed to update template: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"message": fmt.Sprintf("Template '%s' updated successfully", templateName),
		"name":    templateName,
	})
}

func (api *APIHandler) deleteTemplate(w http.ResponseWriter, r *http.Request, templateName string) {
	// Prevent deletion of default template
	if templateName == "default" {
		api.sendError(w, "Cannot delete the default template", http.StatusForbidden)
		return
	}

	var request struct {
		Force bool `json:"force"`
	}

	json.NewDecoder(r.Body).Decode(&request) // Ignore errors for optional body

	if err := api.handler.DeleteTemplate(templateName); err != nil {
		api.sendError(w, fmt.Sprintf("Failed to delete template: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"message": fmt.Sprintf("Template '%s' deleted successfully", templateName),
		"name":    templateName,
	})
}

func (api *APIHandler) copyTemplate(w http.ResponseWriter, r *http.Request, sourceName string) {
	var request struct {
		DestName string `json:"dest_name"`
		ToConfig bool   `json:"to_config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if request.DestName == "" {
		api.sendError(w, "Destination name is required", http.StatusBadRequest)
		return
	}

	if request.ToConfig {
		// Create configuration from template
		if err := api.handler.CreateConfig(request.DestName, sourceName); err != nil {
			api.sendError(w, fmt.Sprintf("Failed to create configuration from template: %v", err), http.StatusInternalServerError)
			return
		}

		api.sendSuccess(w, map[string]interface{}{
			"message": fmt.Sprintf("Configuration '%s' created from template '%s' successfully", request.DestName, sourceName),
			"name":    request.DestName,
			"type":    "configuration",
		})
	} else {
		// Copy template to template
		if request.DestName == "default" {
			api.sendError(w, "Cannot create template with reserved name 'default'", http.StatusBadRequest)
			return
		}

		if err := api.handler.CopyTemplate(sourceName, request.DestName); err != nil {
			api.sendError(w, fmt.Sprintf("Failed to copy template: %v", err), http.StatusInternalServerError)
			return
		}

		api.sendSuccess(w, map[string]interface{}{
			"message": fmt.Sprintf("Template '%s' copied to '%s' successfully", sourceName, request.DestName),
			"name":    request.DestName,
			"type":    "template",
		})
	}
}

func (api *APIHandler) moveTemplate(w http.ResponseWriter, r *http.Request, oldName string) {
	// Prevent moving of default template
	if oldName == "default" {
		api.sendError(w, "Cannot move/rename the default template", http.StatusForbidden)
		return
	}

	var request struct {
		NewName string `json:"new_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if request.NewName == "" {
		api.sendError(w, "New template name is required", http.StatusBadRequest)
		return
	}

	if request.NewName == "default" {
		api.sendError(w, "Cannot rename template to reserved name 'default'", http.StatusBadRequest)
		return
	}

	if err := api.handler.MoveTemplate(oldName, request.NewName); err != nil {
		api.sendError(w, fmt.Sprintf("Failed to move template: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"message":  fmt.Sprintf("Template moved from '%s' to '%s' successfully", oldName, request.NewName),
		"old_name": oldName,
		"new_name": request.NewName,
	})
}

// moveProfile handles POST /api/profiles/{name}/move
func (api *APIHandler) moveProfile(w http.ResponseWriter, r *http.Request, oldName string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		NewName string `json:"new_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if request.NewName == "" {
		api.sendError(w, "New profile name is required", http.StatusBadRequest)
		return
	}

	// Validate new profile name (similar to template validation)
	if len(request.NewName) > 255 {
		api.sendError(w, "Profile name must be 255 characters or less", http.StatusBadRequest)
		return
	}

	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(request.NewName) {
		api.sendError(w, "Profile name can only contain letters, numbers, hyphens, and underscores", http.StatusBadRequest)
		return
	}

	// Call the handler to move the profile
	if err := api.handler.MoveConfig(oldName, request.NewName); err != nil {
		api.sendError(w, fmt.Sprintf("Failed to move profile: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"message":  fmt.Sprintf("Profile moved from '%s' to '%s' successfully", oldName, request.NewName),
		"old_name": oldName,
		"new_name": request.NewName,
	})
}

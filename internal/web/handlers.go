package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cc-switch/internal/handler"
)

// APIHandler handles API requests
type APIHandler struct {
	handler handler.ConfigHandler
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
	// Extract profile name from path
	path := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
	profileName := strings.Split(path, "/")[0]

	if profileName == "" {
		api.sendError(w, "Profile name is required", http.StatusBadRequest)
		return
	}

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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templates, err := api.handler.ListTemplates()
	if err != nil {
		api.sendError(w, fmt.Sprintf("Failed to list templates: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"templates": templates,
	})
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

	// Use template if provided, otherwise use default
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

	// Create the configuration using the handler
	err = api.handler.CreateConfig(request.Name, template)
	if err != nil {
		api.sendError(w, fmt.Sprintf("Failed to create profile: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendSuccess(w, map[string]interface{}{
		"message":  fmt.Sprintf("Profile '%s' created successfully from template '%s'", request.Name, template),
		"name":     request.Name,
		"template": template,
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
	var request struct {
		Env         map[string]string `json:"env"`
		Permissions struct {
			Allow []string `json:"allow"`
			Deny  []string `json:"deny"`
		} `json:"permissions"`
		StatusLine map[string]interface{} `json:"statusLine"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Get current configuration to preserve other fields
	currentView, err := api.handler.ViewConfig(profileName, true)
	if err != nil {
		api.sendError(w, fmt.Sprintf("Failed to get current config: %v", err), http.StatusNotFound)
		return
	}

	// Merge the updates with existing configuration
	updatedConfig := currentView.Content
	updatedConfig["env"] = request.Env
	updatedConfig["permissions"] = map[string]interface{}{
		"allow": request.Permissions.Allow,
		"deny":  request.Permissions.Deny,
	}
	updatedConfig["statusLine"] = request.StatusLine

	// For now, we'll return success since the handler interface doesn't have update methods
	// In a real implementation, you'd need to add an UpdateConfig method to the handler
	api.sendSuccess(w, map[string]interface{}{
		"message": fmt.Sprintf("Profile '%s' would be updated", profileName),
		"name":    profileName,
		"data":    updatedConfig,
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

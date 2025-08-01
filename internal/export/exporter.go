package export

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cc-switch/internal/config"
)

// Exporter interface defines export operations
type Exporter interface {
	ExportProfile(name string, password string, outputPath string) error
	ExportAll(password string, outputPath string) error
	ExportCurrent(password string, outputPath string) error
}

// ExporterImpl implements the Exporter interface
type ExporterImpl struct {
	configManager *config.ConfigManager
	ccxHandler    *CCXHandler
}

// NewExporter creates a new exporter instance
func NewExporter(configManager *config.ConfigManager) *ExporterImpl {
	return &ExporterImpl{
		configManager: configManager,
		ccxHandler:    NewCCXHandler(),
	}
}

// ExportProfile exports a single profile
func (e *ExporterImpl) ExportProfile(name string, password string, outputPath string) error {
	// Validate profile exists
	if !e.configManager.ProfileExists(name) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// Get profile content
	content, metadata, err := e.configManager.GetProfileContent(name)
	if err != nil {
		return fmt.Errorf("failed to read profile '%s': %w", name, err)
	}

	// Create export data
	exportData := &ExportData{
		Profiles: []ProfileData{
			{
				Name:      metadata.Name,
				IsCurrent: metadata.IsCurrent,
				Content:   content,
				Metadata: ProfileMetadata{
					CreatedAt:  time.Now().UTC().Format(time.RFC3339),
					ModifiedAt: time.Now().UTC().Format(time.RFC3339),
				},
			},
		},
	}

	return e.writeExportFile(exportData, password, outputPath)
}

// ExportAll exports all profiles
func (e *ExporterImpl) ExportAll(password string, outputPath string) error {
	// Get all profiles
	profiles, err := e.configManager.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		return fmt.Errorf("no profiles found to export")
	}

	// Create export data
	exportData := &ExportData{
		Profiles: make([]ProfileData, 0, len(profiles)),
	}

	// Collect all profile data
	for _, profile := range profiles {
		content, _, err := e.configManager.GetProfileContent(profile.Name)
		if err != nil {
			return fmt.Errorf("failed to read profile '%s': %w", profile.Name, err)
		}

		profileData := ProfileData{
			Name:      profile.Name,
			IsCurrent: profile.IsCurrent,
			Content:   content,
			Metadata: ProfileMetadata{
				CreatedAt:  time.Now().UTC().Format(time.RFC3339),
				ModifiedAt: time.Now().UTC().Format(time.RFC3339),
			},
		}

		exportData.Profiles = append(exportData.Profiles, profileData)
	}

	return e.writeExportFile(exportData, password, outputPath)
}

// ExportCurrent exports the current active profile
func (e *ExporterImpl) ExportCurrent(password string, outputPath string) error {
	// Get current profile name
	currentProfile, err := e.configManager.GetCurrentProfile()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	if currentProfile == "" {
		return fmt.Errorf("no current profile set")
	}

	return e.ExportProfile(currentProfile, password, outputPath)
}

// writeExportFile writes export data to file
func (e *ExporterImpl) writeExportFile(data *ExportData, password string, outputPath string) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Set restrictive permissions for the export file
	if err := file.Chmod(0600); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Write data using CCX format
	if err := e.ccxHandler.Write(data, file, password); err != nil {
		// Clean up the file on error
		os.Remove(outputPath)
		return fmt.Errorf("failed to write export data: %w", err)
	}

	return nil
}
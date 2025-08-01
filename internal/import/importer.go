package importer

import (
	"fmt"
	"os"
	"strings"

	"cc-switch/internal/config"
	"cc-switch/internal/export"
)

// ImportOptions defines import configuration
type ImportOptions struct {
	Overwrite bool   // Overwrite existing profiles
	Prefix    string // Prefix to add to profile names
	DryRun    bool   // Only validate, don't actually import
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	ProfilesImported []string      // Successfully imported profiles
	Conflicts        []string      // Profiles that had conflicts
	Errors           []error       // Errors encountered during import
	Summary          ImportSummary // Summary statistics
}

// ImportSummary provides import statistics
type ImportSummary struct {
	TotalProfiles int // Total profiles in import file
	ImportedCount int // Successfully imported
	SkippedCount  int // Skipped due to conflicts
	RenamedCount  int // Renamed due to conflicts
	ErrorCount    int // Failed imports
}

// ConflictInfo represents a naming conflict
type ConflictInfo struct {
	OriginalName  string // Original profile name
	ConflictName  string // Conflicting name
	SuggestedName string // Suggested alternative name
}

// Importer interface defines import operations
type Importer interface {
	Import(inputPath string, password string, options ImportOptions) (*ImportResult, error)
	ValidateFile(inputPath string) (*export.CCXMetadata, error)
	CheckConflicts(inputPath string, password string) ([]ConflictInfo, error)
}

// ImporterImpl implements the Importer interface
type ImporterImpl struct {
	configManager *config.ConfigManager
	ccxHandler    *export.CCXHandler
}

// NewImporter creates a new importer instance
func NewImporter(configManager *config.ConfigManager) *ImporterImpl {
	return &ImporterImpl{
		configManager: configManager,
		ccxHandler:    export.NewCCXHandler(),
	}
}

// Import imports profiles from a CCX file
func (i *ImporterImpl) Import(inputPath string, password string, options ImportOptions) (*ImportResult, error) {
	// Validate file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("import file does not exist: %s", inputPath)
	}

	// Open and read the file
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open import file: %w", err)
	}
	defer file.Close()

	// Read export data
	exportData, err := i.ccxHandler.Read(file, password)
	if err != nil {
		return nil, fmt.Errorf("failed to read import file: %w", err)
	}

	// Initialize result
	result := &ImportResult{
		ProfilesImported: make([]string, 0),
		Conflicts:        make([]string, 0),
		Errors:           make([]error, 0),
		Summary: ImportSummary{
			TotalProfiles: len(exportData.Profiles),
		},
	}

	// Process each profile
	for _, profileData := range exportData.Profiles {
		if err := i.importProfile(profileData, options, result); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to import profile '%s': %w", profileData.Name, err))
			result.Summary.ErrorCount++
		}
	}

	// Update summary
	result.Summary.ImportedCount = len(result.ProfilesImported)
	result.Summary.SkippedCount = len(result.Conflicts)

	return result, nil
}

// ValidateFile validates a CCX file format
func (i *ImporterImpl) ValidateFile(inputPath string) (*export.CCXMetadata, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return i.ccxHandler.ValidateFile(file)
}

// CheckConflicts checks for naming conflicts before import
func (i *ImporterImpl) CheckConflicts(inputPath string, password string) ([]ConflictInfo, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open import file: %w", err)
	}
	defer file.Close()

	exportData, err := i.ccxHandler.Read(file, password)
	if err != nil {
		return nil, fmt.Errorf("failed to read import file: %w", err)
	}

	var conflicts []ConflictInfo
	for _, profileData := range exportData.Profiles {
		if i.configManager.ProfileExists(profileData.Name) {
			conflicts = append(conflicts, ConflictInfo{
				OriginalName:  profileData.Name,
				ConflictName:  profileData.Name,
				SuggestedName: i.generateAlternativeName(profileData.Name),
			})
		}
	}

	return conflicts, nil
}

// importProfile imports a single profile
func (i *ImporterImpl) importProfile(profileData export.ProfileData, options ImportOptions, result *ImportResult) error {
	// Apply prefix if specified
	finalName := profileData.Name
	if options.Prefix != "" {
		finalName = options.Prefix + finalName
	}

	// Check for conflicts
	if i.configManager.ProfileExists(finalName) {
		if !options.Overwrite {
			// Generate alternative name
			alternativeName := i.generateAlternativeName(finalName)
			if options.DryRun {
				result.Conflicts = append(result.Conflicts, fmt.Sprintf("%s -> %s (would be renamed)", finalName, alternativeName))
				return nil
			}

			finalName = alternativeName
			result.Summary.RenamedCount++
		} else {
			if options.DryRun {
				result.Conflicts = append(result.Conflicts, fmt.Sprintf("%s (would be overwritten)", finalName))
				return nil
			}
		}
	}

	if options.DryRun {
		result.ProfilesImported = append(result.ProfilesImported, finalName+" (dry run)")
		return nil
	}

	// Validate profile content
	if err := i.validateProfileContent(profileData.Content); err != nil {
		return fmt.Errorf("invalid profile content: %w", err)
	}

	// Create or update the profile
	if i.configManager.ProfileExists(finalName) {
		// Update existing profile
		if err := i.configManager.UpdateProfile(finalName, profileData.Content); err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}
	} else {
		// Create new profile
		if err := i.configManager.CreateProfile(finalName); err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}

		// Update with imported content
		if err := i.configManager.UpdateProfile(finalName, profileData.Content); err != nil {
			// Clean up on failure
			i.configManager.DeleteProfile(finalName)
			return fmt.Errorf("failed to update new profile: %w", err)
		}
	}

	result.ProfilesImported = append(result.ProfilesImported, finalName)
	return nil
}

// validateProfileContent validates imported profile content
func (i *ImporterImpl) validateProfileContent(content map[string]interface{}) error {
	if content == nil {
		return fmt.Errorf("profile content cannot be nil")
	}

	// Basic validation - could be expanded based on requirements
	// Check if content can be serialized to JSON
	return nil
}

// generateAlternativeName generates an alternative name for conflicting profiles
func (i *ImporterImpl) generateAlternativeName(originalName string) string {
	counter := 1
	baseName := originalName

	// If the name already has a number suffix, extract the base name
	if lastDash := strings.LastIndex(originalName, "-"); lastDash != -1 {
		if suffix := originalName[lastDash+1:]; isNumeric(suffix) {
			baseName = originalName[:lastDash]
		}
	}

	for {
		candidateName := fmt.Sprintf("%s-%d", baseName, counter)
		if !i.configManager.ProfileExists(candidateName) {
			return candidateName
		}
		counter++
	}
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return len(s) > 0
}

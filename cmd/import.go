package cmd

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"cc-switch/internal/config"
	"cc-switch/internal/export"
	importpkg "cc-switch/internal/import"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	importPassword  string
	importOverwrite bool
	importPrefix    string
	importDryRun    bool
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import configurations from a backup file",
	Long: `Import Claude Code configurations from an encrypted backup file.

Examples:
  # Import from backup file
  cc-switch import backup.ccx -p mypassword

  # Import with overwrite existing profiles
  cc-switch import backup.ccx --overwrite

  # Import with prefix to avoid conflicts
  cc-switch import team-configs.ccx --prefix team-

  # Dry run to see what would be imported
  cc-switch import backup.ccx --dry-run

  # Interactive password input (recommended for security)
  cc-switch import backup.ccx`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkClaudeConfig(); err != nil {
			return err
		}

		inputFile := args[0]

		// Validate input file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("import file does not exist: %s", inputFile)
		}

		// Initialize config manager
		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		// Create importer
		importer := importpkg.NewImporter(cm)

		// Validate file format first
		color.Cyan("🔍 Validating file format...")
		metadata, err := importer.ValidateFile(inputFile)
		if err != nil {
			return fmt.Errorf("invalid import file: %w", err)
		}

		// Show file information
		showFileInfo(metadata)

		// Get password if not provided
		password := importPassword
		isEncrypted := strings.Contains(metadata.Encryption, "aes")

		if isEncrypted && password == "" {
			password, err = promptForDecryptionPassword()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		}

		// Check for conflicts if not in dry-run mode
		if !importDryRun {
			color.Cyan("🔍 Checking for conflicts...")
			conflicts, err := importer.CheckConflicts(inputFile, password)
			if err != nil {
				return fmt.Errorf("failed to check conflicts: %w", err)
			}

			if len(conflicts) > 0 && !importOverwrite && importPrefix == "" {
				showConflicts(conflicts)
				if !confirmProceed() {
					color.Yellow("Import cancelled by user")
					return nil
				}
			}
		}

		// Prepare import options
		options := importpkg.ImportOptions{
			Overwrite: importOverwrite,
			Prefix:    importPrefix,
			DryRun:    importDryRun,
		}

		// Perform import
		if importDryRun {
			color.Cyan("🔍 Performing dry run...")
		} else {
			color.Cyan("📥 Importing configurations...")
		}

		result, err := importer.Import(inputFile, password, options)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}

		// Show results
		showImportResults(result, importDryRun)

		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importPassword, "password", "p", "", "Decryption password (prompt if not provided)")
	importCmd.Flags().BoolVar(&importOverwrite, "overwrite", false, "Overwrite existing profiles")
	importCmd.Flags().StringVar(&importPrefix, "prefix", "", "Add prefix to imported profile names")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Show what would be imported without making changes")
}

func promptForDecryptionPassword() (string, error) {
	fmt.Print("Enter password for decryption: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	if err != nil {
		return "", err
	}
	return string(password), nil
}

func showFileInfo(metadata *export.CCXMetadata) {
	color.Blue("📋 Import File Information:")
	color.Blue("   Version: %s", metadata.Version)
	color.Blue("   Exported: %s", metadata.ExportedAt)
	color.Blue("   Tool: %s", metadata.ToolVersion)
	color.Blue("   Profiles: %d", metadata.ProfilesCount)
	color.Blue("   Type: %s", metadata.ExportType)

	if metadata.Encryption != "" {
		color.Blue("   Encryption: %s", metadata.Encryption)
	}
	if metadata.Compression != "" {
		color.Blue("   Compression: %s", metadata.Compression)
	}
	fmt.Println()
}

func showConflicts(conflicts []importpkg.ConflictInfo) {
	color.Yellow("⚠️  Naming conflicts detected:")
	for _, conflict := range conflicts {
		color.Yellow("   • '%s' already exists (suggested: '%s')",
			conflict.ConflictName, conflict.SuggestedName)
	}
	fmt.Println()
	color.Yellow("Options to resolve conflicts:")
	color.Yellow("   • Use --overwrite to replace existing profiles")
	color.Yellow("   • Use --prefix to add a prefix to imported names")
	color.Yellow("   • Rename existing profiles manually")
	fmt.Println()
}

func confirmProceed() bool {
	fmt.Print("Continue with import? Conflicting profiles will be renamed automatically [y/N]: ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func showImportResults(result *importpkg.ImportResult, isDryRun bool) {
	summary := result.Summary

	if isDryRun {
		color.Cyan("🔍 Dry Run Results:")
	} else {
		if summary.ErrorCount == 0 {
			color.Green("✅ Import completed successfully!")
		} else {
			color.Yellow("⚠️  Import completed with some errors")
		}
	}

	fmt.Println()
	color.Blue("📊 Summary:")
	color.Blue("   Total profiles: %d", summary.TotalProfiles)

	if isDryRun {
		color.Blue("   Would import: %d", summary.ImportedCount)
		if summary.SkippedCount > 0 {
			color.Blue("   Would skip: %d", summary.SkippedCount)
		}
		if summary.RenamedCount > 0 {
			color.Blue("   Would rename: %d", summary.RenamedCount)
		}
	} else {
		color.Blue("   Imported: %d", summary.ImportedCount)
		if summary.SkippedCount > 0 {
			color.Blue("   Skipped: %d", summary.SkippedCount)
		}
		if summary.RenamedCount > 0 {
			color.Blue("   Renamed: %d", summary.RenamedCount)
		}
		if summary.ErrorCount > 0 {
			color.Red("   Errors: %d", summary.ErrorCount)
		}
	}

	// Show imported profiles
	if len(result.ProfilesImported) > 0 {
		fmt.Println()
		if isDryRun {
			color.Cyan("Profiles that would be imported:")
		} else {
			color.Green("Successfully imported profiles:")
		}
		for _, profile := range result.ProfilesImported {
			if isDryRun {
				color.Cyan("   • %s", profile)
			} else {
				color.Green("   • %s", profile)
			}
		}
	}

	// Show conflicts
	if len(result.Conflicts) > 0 {
		fmt.Println()
		color.Yellow("Conflicts handled:")
		for _, conflict := range result.Conflicts {
			color.Yellow("   • %s", conflict)
		}
	}

	// Show errors
	if len(result.Errors) > 0 {
		fmt.Println()
		color.Red("Errors encountered:")
		for _, err := range result.Errors {
			color.Red("   • %s", err.Error())
		}
	}

	if !isDryRun && summary.ImportedCount > 0 {
		fmt.Println()
		color.Blue("💡 Use 'cc-switch list' to see all available profiles")
		color.Blue("💡 Use 'cc-switch use <profile>' to switch to an imported profile")
	}
}

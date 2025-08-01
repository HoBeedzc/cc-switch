package cmd

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"cc-switch/internal/config"
	"cc-switch/internal/export"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	exportOutput   string
	exportPassword string
	exportAll      bool
	exportCurrent  bool
)

var exportCmd = &cobra.Command{
	Use:   "export [profile-name]",
	Short: "Export configurations to a backup file",
	Long: `Export Claude Code configurations to an encrypted backup file.

Examples:
  # Export a specific profile
  cc-switch export default -o backup.ccx -p mypassword

  # Export all profiles
  cc-switch export --all -o all-configs.ccx

  # Export current profile
  cc-switch export --current -o current-config.ccx

  # Interactive password input (recommended for security)
  cc-switch export default -o backup.ccx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkClaudeConfig(); err != nil {
			return err
		}

		// Validate flags
		if err := validateExportFlags(args); err != nil {
			return err
		}

		// Initialize config manager
		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		// Create exporter
		exporter := export.NewExporter(cm)

		// Get password if not provided
		password := exportPassword
		if password == "" {
			password, err = promptForPassword("Enter password for encryption (leave empty for no encryption): ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		}

		// Show security recommendations if using password
		if password != "" {
			showSecurityRecommendations()
		}

		// Ensure output file has .ccx extension
		outputPath := ensureCCXExtension(exportOutput)

		var exportErr error
		profileCount := 0

		color.Cyan("📦 Preparing export...")

		if exportAll {
			// Export all profiles
			profiles, err := cm.ListProfiles()
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}
			profileCount = len(profiles)
			
			if profileCount == 0 {
				return fmt.Errorf("no profiles found to export")
			}
			
			color.Cyan("📦 Collecting profiles... (%d found)", profileCount)
			exportErr = exporter.ExportAll(password, outputPath)
		} else if exportCurrent {
			// Export current profile
			current, err := cm.GetCurrentProfile()
			if err != nil {
				return fmt.Errorf("failed to get current profile: %w", err)
			}
			if current == "" {
				return fmt.Errorf("no current profile set")
			}
			
			profileCount = 1
			color.Cyan("📦 Exporting current profile '%s'...", current)
			exportErr = exporter.ExportCurrent(password, outputPath)
		} else {
			// Export specific profile
			profileName := args[0]
			if !cm.ProfileExists(profileName) {
				return fmt.Errorf("profile '%s' does not exist", profileName)
			}
			
			profileCount = 1
			color.Cyan("📦 Exporting profile '%s'...", profileName)
			exportErr = exporter.ExportProfile(profileName, password, outputPath)
		}

		if exportErr != nil {
			return fmt.Errorf("export failed: %w", exportErr)
		}

		// Show success message with file size
		fileInfo, err := os.Stat(outputPath)
		if err == nil {
			size := formatFileSize(fileInfo.Size())
			color.Green("✅ Export completed (%d profiles, %s)", profileCount, size)
			color.Blue("📁 Saved to: %s", outputPath)
		} else {
			color.Green("✅ Export completed (%d profiles)", profileCount)
		}

		if password != "" {
			color.Yellow("🔒 File is encrypted and protected")
		}

		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (required)")
	exportCmd.Flags().StringVarP(&exportPassword, "password", "p", "", "Encryption password (prompt if not provided)")
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "Export all profiles")
	exportCmd.Flags().BoolVar(&exportCurrent, "current", false, "Export current profile")
	
	exportCmd.MarkFlagRequired("output")
}

func validateExportFlags(args []string) error {
	flagCount := 0
	if exportAll {
		flagCount++
	}
	if exportCurrent {
		flagCount++
	}
	if len(args) > 0 {
		flagCount++
	}

	if flagCount == 0 {
		return fmt.Errorf("must specify either a profile name, --all, or --current")
	}

	if flagCount > 1 {
		return fmt.Errorf("cannot use --all, --current, and profile name together")
	}

	if exportOutput == "" {
		return fmt.Errorf("output file path is required (-o/--output)")
	}

	return nil
}

func promptForPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	if err != nil {
		return "", err
	}

	passwordStr := string(password)
	
	if passwordStr != "" {
		fmt.Print("Confirm password: ")
		confirm, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return "", err
		}

		if passwordStr != string(confirm) {
			return "", fmt.Errorf("passwords do not match")
		}
	}

	return passwordStr, nil
}

func showSecurityRecommendations() {
	color.Yellow("⚠️  Security Recommendations:")
	color.Yellow("   • Use a strong, unique password")
	color.Yellow("   • Store the backup file securely")
	color.Yellow("   • Consider using a password manager")
	fmt.Println()
}

func ensureCCXExtension(filename string) string {
	if !strings.HasSuffix(strings.ToLower(filename), ".ccx") {
		return filename + ".ccx"
	}
	return filename
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
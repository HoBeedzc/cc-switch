package cmd

import (
	"fmt"
	"strings"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Switch to a configuration",
	Long: `Switch to the specified configuration. This will replace the current Claude Code settings.

Modes:
- Interactive: cc-switch use (no arguments) or cc-switch use -i
- CLI: cc-switch use <name>
- Previous: cc-switch use -p or cc-switch use --previous
- Empty Mode: cc-switch use -e or cc-switch use --empty
- Restore: cc-switch use --restore

The interactive mode allows you to browse and select configurations with arrow keys.
The previous mode switches to the last used configuration.
The empty mode temporarily removes all configurations.
The restore mode restores from empty mode to the previous configuration.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkClaudeConfig(); err != nil {
			return err
		}

		// Initialize dependencies
		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		configHandler := handler.NewConfigHandler(cm)
		interactiveFlag, _ := cmd.Flags().GetBool("interactive")
		previousFlag, _ := cmd.Flags().GetBool("previous")
		emptyFlag, _ := cmd.Flags().GetBool("empty")
		restoreFlag, _ := cmd.Flags().GetBool("restore")

		// Validate flag combinations
		flagCount := 0
		if previousFlag {
			flagCount++
		}
		if emptyFlag {
			flagCount++
		}
		if restoreFlag {
			flagCount++
		}
		if len(args) > 0 {
			flagCount++
		}

		if flagCount > 1 {
			return fmt.Errorf("cannot use multiple operation flags together")
		}

		if (previousFlag || emptyFlag || restoreFlag) && interactiveFlag {
			return fmt.Errorf("cannot use operation flags with -i/--interactive")
		}

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if !previousFlag && !emptyFlag && !restoreFlag && ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Handle special operations
		if emptyFlag {
			return handleEmptyMode(configHandler, uiProvider)
		}

		if restoreFlag {
			return handleRestoreMode(configHandler, uiProvider)
		}

		if previousFlag {
			return handlePreviousConfig(configHandler, uiProvider)
		}

		// Execute normal use operation
		return executeUse(configHandler, uiProvider, args)
	},
}

// executeUse handles the use operation with the given dependencies
func executeUse(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string) error {
	// Check if currently in empty mode - if so, any use command should restore first
	if configHandler.IsEmptyMode() {
		uiProvider.ShowInfo("Currently in empty mode. Restoring settings first...")
		if err := configHandler.RestoreFromEmptyMode(); err != nil {
			uiProvider.ShowError(fmt.Errorf("failed to restore from empty mode: %w", err))
			return err
		}
		uiProvider.ShowInfo("Settings restored from empty mode.")
	}

	// Get all configurations
	profiles, err := configHandler.ListConfigs()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		uiProvider.ShowWarning("No configurations found.")
		fmt.Println("Use 'cc-switch new <name>' to create your first configuration.")
		return nil
	}

	var targetName string

	// Determine execution mode
	if len(args) == 0 {
		// Interactive mode - use enhanced selector with empty mode support
		if ui.NewInteractiveUI().DetectMode(false, args) == ui.Interactive {
			interactiveUI := uiProvider.(ui.InteractiveUI)
			selection, err := interactiveUI.SelectConfigurationWithEmptyMode(profiles, "use", configHandler.IsEmptyMode())
			if err != nil {
				return fmt.Errorf("selection cancelled: %w", err)
			}

			// Handle special selections
			switch selection.Type {
			case "empty_mode":
				return handleEmptyMode(configHandler, uiProvider)
			case "restore":
				return handleRestoreMode(configHandler, uiProvider)
			case "profile":
				// Check if already current
				if selection.Profile.IsCurrent && !configHandler.IsEmptyMode() {
					uiProvider.ShowWarning("Configuration '%s' is already active", selection.Profile.Name)
					return nil
				}
				targetName = selection.Profile.Name
			default:
				return fmt.Errorf("unknown selection type: %s", selection.Type)
			}
		} else {
			// Fallback to regular selection
			selected, err := uiProvider.SelectConfiguration(profiles, "use")
			if err != nil {
				return fmt.Errorf("selection cancelled: %w", err)
			}

			// Check if already current
			if selected.IsCurrent && !configHandler.IsEmptyMode() {
				uiProvider.ShowWarning("Configuration '%s' is already active", selected.Name)
				return nil
			}

			targetName = selected.Name
		}
	} else {
		// CLI mode
		targetName = args[0]
	}

	// Execute switch
	if err := configHandler.UseConfig(targetName); err != nil {
		// Handle specific error messages
		if err.Error() == fmt.Sprintf("configuration '%s' is already active", targetName) {
			uiProvider.ShowWarning("Configuration '%s' is already active", targetName)
			return nil
		}
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("Switched to configuration '%s'", targetName)
	return nil
}

// handlePreviousConfig handles switching to the previous configuration
func handlePreviousConfig(configHandler handler.ConfigHandler, uiProvider ui.UIProvider) error {
	// Get previous configuration
	previousName, err := configHandler.GetPreviousConfig()
	if err != nil {
		if err.Error() == "no previous configuration available" {
			uiProvider.ShowWarning("No previous configuration available")
			fmt.Println("💡 Use 'cc-switch use <name>' to switch configurations first")
			return nil
		}
		if strings.Contains(err.Error(), "no longer exists") {
			uiProvider.ShowWarning(err.Error())
			fmt.Println("💡 The previous configuration has been deleted")
			return nil
		}
		uiProvider.ShowError(err)
		return err
	}

	// Get current configuration for display
	currentName, _ := configHandler.GetCurrentConfig()

	// Check if previous is same as current (edge case)
	if previousName == currentName {
		uiProvider.ShowWarning("Previous configuration '%s' is the same as current", previousName)
		return nil
	}

	// Execute switch
	if err := configHandler.UseConfig(previousName); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	// Show success message with context
	if currentName != "" {
		uiProvider.ShowSuccess("Switched to configuration '%s' (previous: '%s')", previousName, currentName)
	} else {
		uiProvider.ShowSuccess("Switched to configuration '%s'", previousName)
	}

	return nil
}

// handleEmptyMode handles enabling empty mode
func handleEmptyMode(configHandler handler.ConfigHandler, uiProvider ui.UIProvider) error {
	// Check if already in empty mode
	if configHandler.IsEmptyMode() {
		uiProvider.ShowWarning("Already in empty mode")
		fmt.Println("💡 Use 'cc-switch use <profile>' to restore a configuration")
		fmt.Println("💡 Use 'cc-switch use --restore' to restore the previous configuration")
		return nil
	}

	// Get current configuration for display
	currentName, _ := configHandler.GetCurrentConfig()

	// Enable empty mode
	if err := configHandler.UseEmptyMode(); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	// Show success message
	uiProvider.ShowSuccess("Empty mode enabled. Settings temporarily removed.")
	if currentName != "" {
		fmt.Printf("💡 Previous configuration: %s\n", currentName)
	}
	fmt.Println("💡 Use 'cc-switch use <profile>' to restore a configuration")
	fmt.Println("💡 Use 'cc-switch use --restore' to restore the previous configuration")

	return nil
}

// handleRestoreMode handles restoring from empty mode
func handleRestoreMode(configHandler handler.ConfigHandler, uiProvider ui.UIProvider) error {
	// Check if in empty mode
	if !configHandler.IsEmptyMode() {
		uiProvider.ShowWarning("Not in empty mode")
		return nil
	}

	// Get empty mode status
	status, err := configHandler.GetEmptyModeStatus()
	if err != nil {
		uiProvider.ShowError(fmt.Errorf("failed to get empty mode status: %w", err))
		return err
	}

	// Check if we can restore to previous
	if !status.CanRestore {
		uiProvider.ShowWarning("No previous configuration to restore to")
		fmt.Println("💡 Use 'cc-switch use <profile>' to switch to a specific configuration")
		return nil
	}

	// Restore to previous configuration
	if err := configHandler.RestoreToPreviousFromEmptyMode(); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("Restored to previous configuration '%s'", status.PreviousProfile)
	return nil
}

func init() {
	useCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
	useCmd.Flags().BoolP("previous", "p", false, "Switch to previous configuration")
	useCmd.Flags().BoolP("empty", "e", false, "Enable empty mode (remove settings)")
	useCmd.Flags().Bool("restore", false, "Restore from empty mode to previous configuration")
}

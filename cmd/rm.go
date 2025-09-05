package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a configuration",
	Long: `Remove the specified configuration with enhanced deletion options.

Modes:
- Interactive: cc-switch rm (no arguments) or cc-switch rm -i
- CLI: cc-switch rm <name>

Enhanced Options:
- -a, --all: Delete ALL configurations (requires manual confirmation, cannot use with -f/-y)
- -c, --current: Delete current configuration and enter EMPTY MODE
- -y, --yes: Skip confirmation prompts (cannot use with --all)
- -f, --force: Legacy force flag (same as --yes, cannot use with --all)

The interactive mode allows you to browse and select configurations with arrow keys.`,
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

		// Get flags
		interactiveFlag, _ := cmd.Flags().GetBool("interactive")
		force, _ := cmd.Flags().GetBool("force")
		all, _ := cmd.Flags().GetBool("all")
		current, _ := cmd.Flags().GetBool("current")
		yes, _ := cmd.Flags().GetBool("yes")

		// Validate flag combinations
		if err := validateRemoveFlags(all, force, yes, args); err != nil {
			return err
		}

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute remove operation with enhanced logic
		return executeEnhancedRemove(configHandler, uiProvider, args, all, current, force || yes)
	},
}

// validateRemoveFlags validates flag combinations for the rm command
func validateRemoveFlags(all, force, yes bool, args []string) error {
	// --all cannot be combined with -f/--force or -y/--yes
	if all && (force || yes) {
		return fmt.Errorf("--all flag cannot be combined with --force (-f) or --yes (-y) flags")
	}

	// --all cannot have arguments
	if all && len(args) > 0 {
		return fmt.Errorf("--all flag cannot be used with specific configuration names")
	}

	return nil
}

// executeEnhancedRemove handles the enhanced remove operation with new flags
func executeEnhancedRemove(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, all, current, skipConfirm bool) error {
	// Handle --all flag (delete all configurations)
	if all {
		return executeRemoveAll(configHandler, uiProvider)
	}

	// Handle --current flag (delete current configuration)
	if current {
		return executeRemoveCurrent(configHandler, uiProvider, skipConfirm)
	}

	// Fall back to original remove logic for specific configuration
	return executeRemove(configHandler, uiProvider, args, skipConfirm)
}

// executeRemoveAll handles deleting all configurations
func executeRemoveAll(configHandler handler.ConfigHandler, uiProvider ui.UIProvider) error {
	// Get all configurations
	profiles, err := configHandler.ListConfigs()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		uiProvider.ShowWarning("No configurations found to remove.")
		return nil
	}

	// Show warning and require manual confirmation (cannot be bypassed)
	uiProvider.ShowWarning("This will delete ALL %d configuration(s):", len(profiles))
	for _, profile := range profiles {
		if profile.IsCurrent {
			fmt.Printf("  - %s (current)\n", profile.Name)
		} else {
			fmt.Printf("  - %s\n", profile.Name)
		}
	}
	fmt.Println()

	// Mandatory confirmation - cannot be bypassed with any flag
	confirmMsg := "Type 'DELETE ALL' to confirm deletion of all configurations"
	fmt.Printf("%s: ", confirmMsg)
	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "DELETE ALL" {
		uiProvider.ShowInfo("Operation cancelled - confirmation text did not match")
		return nil
	}

	// Delete all configurations and enter empty mode
	if err := configHandler.DeleteAllConfigs(); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("All configurations deleted successfully. Entering EMPTY MODE.")
	return nil
}

// executeRemoveCurrent handles deleting the current configuration
func executeRemoveCurrent(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, skipConfirm bool) error {
	// Get current configuration
	currentName, err := configHandler.GetCurrentConfig()
	if err != nil {
		return fmt.Errorf("failed to get current configuration: %w", err)
	}

	// Check if we're already in empty mode
	if configHandler.IsEmptyMode() {
		uiProvider.ShowWarning("Already in EMPTY MODE - no current configuration to remove.")
		return nil
	}

	// Confirm deletion if not skipping
	if !skipConfirm {
		confirmMsg := fmt.Sprintf("Delete current configuration '%s' and enter EMPTY MODE?", currentName)
		if !uiProvider.ConfirmAction(confirmMsg, false) {
			uiProvider.ShowInfo("Operation cancelled")
			return nil
		}
	}

	// Delete current configuration and enter empty mode
	if err := configHandler.DeleteCurrentConfig(); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("Current configuration '%s' deleted. Entering EMPTY MODE.", currentName)
	return nil
}

// executeRemove handles the remove operation with the given dependencies
// This function reuses the original logic for specific configuration removal
func executeRemove(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, force bool) error {
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
		// Interactive mode - filter out current config (cannot remove current)
		var removableProfiles []config.Profile
		for _, profile := range profiles {
			if !profile.IsCurrent {
				removableProfiles = append(removableProfiles, profile)
			}
		}

		if len(removableProfiles) == 0 {
			uiProvider.ShowWarning("No configurations available for removal.")
			fmt.Println("The current configuration cannot be removed. Switch to another configuration first.")
			return nil
		}

		// Select configuration interactively
		selected, err := uiProvider.SelectConfiguration(removableProfiles, "remove")
		if err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}
		targetName = selected.Name
	} else {
		// CLI mode
		targetName = args[0]
	}

	// Confirm removal if not forced
	if !force {
		confirmMsg := fmt.Sprintf("Are you sure you want to remove configuration '%s'?", targetName)
		if !uiProvider.ConfirmAction(confirmMsg, false) {
			uiProvider.ShowInfo("Operation cancelled")
			return nil
		}
	}

	// Execute removal
	if err := configHandler.DeleteConfig(targetName, force); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("Configuration '%s' removed successfully", targetName)
	return nil
}

func init() {
	rmCmd.Flags().BoolP("force", "f", false, "Force remove without confirmation (cannot use with --all)")
	rmCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
	rmCmd.Flags().BoolP("all", "a", false, "Delete ALL configurations (requires manual confirmation)")
	rmCmd.Flags().BoolP("current", "c", false, "Delete current configuration and enter EMPTY MODE")
	rmCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts (cannot use with --all)")
}

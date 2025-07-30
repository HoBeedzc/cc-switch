package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a configuration",
	Long: `Delete the specified configuration. You cannot delete the currently active configuration.

Modes:
- Interactive: cc-switch delete (no arguments) or cc-switch delete -i
- CLI: cc-switch delete <name>

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
		interactiveFlag, _ := cmd.Flags().GetBool("interactive")
		force, _ := cmd.Flags().GetBool("force")

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute delete operation
		return executeDelete(configHandler, uiProvider, args, force)
	},
}

// executeDelete handles the delete operation with the given dependencies
func executeDelete(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, force bool) error {
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
		// Interactive mode - filter out current config (cannot delete current)
		var deletableProfiles []config.Profile
		for _, profile := range profiles {
			if !profile.IsCurrent {
				deletableProfiles = append(deletableProfiles, profile)
			}
		}

		if len(deletableProfiles) == 0 {
			uiProvider.ShowWarning("No configurations available for deletion.")
			fmt.Println("The current configuration cannot be deleted. Switch to another configuration first.")
			return nil
		}

		// Select configuration interactively
		selected, err := uiProvider.SelectConfiguration(deletableProfiles, "delete")
		if err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}
		targetName = selected.Name
	} else {
		// CLI mode
		targetName = args[0]
	}

	// Confirm deletion if not forced
	if !force {
		confirmMsg := fmt.Sprintf("Are you sure you want to delete configuration '%s'?", targetName)
		if !uiProvider.ConfirmAction(confirmMsg, false) {
			uiProvider.ShowInfo("Operation cancelled")
			return nil
		}
	}

	// Execute deletion
	if err := configHandler.DeleteConfig(targetName, force); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("Configuration '%s' deleted successfully", targetName)
	return nil
}

func init() {
	deleteCmd.Flags().BoolP("force", "f", false, "Force delete without confirmation")
	deleteCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}

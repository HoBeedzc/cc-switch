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
	Long: `Remove the specified configuration. You cannot remove the currently active configuration.

Modes:
- Interactive: cc-switch rm (no arguments) or cc-switch rm -i
- CLI: cc-switch rm <name>

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

		// Execute remove operation (reuse delete logic)
		return executeRemove(configHandler, uiProvider, args, force)
	},
}

// executeRemove handles the remove operation with the given dependencies
// This function reuses the logic from delete.go with appropriate naming changes
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
	rmCmd.Flags().BoolP("force", "f", false, "Force remove without confirmation")
	rmCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}
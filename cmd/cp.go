package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var cpCmd = &cobra.Command{
	Use:   "cp <source> <destination>",
	Short: "Copy a configuration",
	Long: `Copy the specified configuration to a new name. The original configuration 
remains unchanged. This is useful for creating backups or variations of configurations.

Modes:
- Interactive: cc-switch cp (no arguments) or cc-switch cp -i
- CLI: cc-switch cp <source> <destination>

The interactive mode allows you to browse and select configurations with arrow keys.`,
	Args: cobra.MaximumNArgs(2),
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

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute copy operation
		return executeCopy(configHandler, uiProvider, args)
	},
}

// executeCopy handles the copy operation with the given dependencies
func executeCopy(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string) error {
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

	var sourceName, destName string

	// Determine execution mode
	if len(args) == 0 {
		// Interactive mode - select source configuration
		selected, err := uiProvider.SelectConfiguration(profiles, "copy")
		if err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}
		sourceName = selected.Name

		// Get destination name interactively
		destName, err = uiProvider.GetInput(fmt.Sprintf("Enter destination name for copying '%s'", sourceName), "")
		if err != nil {
			return fmt.Errorf("input cancelled: %w", err)
		}
	} else if len(args) == 1 {
		// Partial CLI mode - source name provided, get destination name interactively
		sourceName = args[0]
		var err error
		// Create temporary UI for input
		tempUI := ui.NewCLIUI()
		destName, err = tempUI.GetInput(fmt.Sprintf("Enter destination name for copying '%s'", sourceName), "")
		if err != nil {
			return fmt.Errorf("input cancelled: %w", err)
		}
	} else {
		// Full CLI mode
		sourceName = args[0]
		destName = args[1]
	}

	// Validate input
	if sourceName == "" {
		return fmt.Errorf("source configuration name cannot be empty")
	}
	if destName == "" {
		return fmt.Errorf("destination configuration name cannot be empty")
	}

	// Execute copy
	if err := configHandler.CopyConfig(sourceName, destName); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("Configuration copied from '%s' to '%s' successfully", sourceName, destName)
	return nil
}

func init() {
	cpCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}

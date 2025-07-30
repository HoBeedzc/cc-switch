package cmd

import (
	"fmt"

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

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute use operation
		return executeUse(configHandler, uiProvider, args)
	},
}

// executeUse handles the use operation with the given dependencies
func executeUse(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string) error {
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
		// Interactive mode
		selected, err := uiProvider.SelectConfiguration(profiles, "use")
		if err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}

		// Check if already current
		if selected.IsCurrent {
			uiProvider.ShowWarning("Configuration '%s' is already active", selected.Name)
			return nil
		}

		targetName = selected.Name
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

func init() {
	useCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}
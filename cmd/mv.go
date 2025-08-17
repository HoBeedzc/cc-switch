package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:   "mv <old-name> <new-name>",
	Short: "Move (rename) a configuration",
	Long: `Move (rename) the specified configuration. If the configuration is currently active, 
the current configuration marker will be updated automatically.

Modes:
- Interactive: cc-switch mv (no arguments) or cc-switch mv -i
- CLI: cc-switch mv <old-name> <new-name>

The interactive mode allows you to browse and select configurations with arrow keys.`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkClaudeConfig(); err != nil {
			return err
		}

		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		configHandler := handler.NewConfigHandler(cm)
		interactiveFlag, _ := cmd.Flags().GetBool("interactive")

		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		return executeMove(configHandler, uiProvider, args)
	},
}

// executeMove handles the move operation with the given dependencies
func executeMove(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string) error {
	profiles, err := configHandler.ListConfigs()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		uiProvider.ShowWarning("No configurations found.")
		fmt.Println("Use 'cc-switch new <name>' to create your first configuration.")
		return nil
	}

	var oldName, newName string

	if len(args) == 0 {
		selected, err := uiProvider.SelectConfiguration(profiles, "move")
		if err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}
		oldName = selected.Name

		// Get new name interactively
		newName, err = uiProvider.GetInput(fmt.Sprintf("Enter new name for configuration '%s'", oldName), "")
		if err != nil {
			return fmt.Errorf("input cancelled: %w", err)
		}
	} else if len(args) == 1 {
		// Partial CLI mode - old name provided, get new name interactively
		oldName = args[0]
		var err error
		// Create temporary UI for input
		tempUI := ui.NewCLIUI()
		newName, err = tempUI.GetInput(fmt.Sprintf("Enter new name for configuration '%s'", oldName), "")
		if err != nil {
			return fmt.Errorf("input cancelled: %w", err)
		}
	} else {
		// Full CLI mode
		oldName = args[0]
		newName = args[1]
	}

	// Validate input
	if oldName == "" {
		return fmt.Errorf("old configuration name cannot be empty")
	}
	if newName == "" {
		return fmt.Errorf("new configuration name cannot be empty")
	}

	// Execute move
	if err := configHandler.MoveConfig(oldName, newName); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("Configuration moved from '%s' to '%s' successfully", oldName, newName)
	return nil
}

func init() {
	mvCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}

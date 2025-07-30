package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var (
	editField string
	useNano   bool
)

var editCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit configuration content",
	Long: `Edit the content of a specified configuration using your system editor.

Modes:
- Interactive: cc-switch edit (no arguments) or cc-switch edit -i
- CLI: cc-switch edit <name>

The interactive mode allows you to browse and select configurations with arrow keys.

The editor is determined by priority:
1. --nano flag uses nano editor
2. EDITOR environment variable 
3. Default to vim editor

Changes are validated for JSON syntax before saving.`,
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
		field, _ := cmd.Flags().GetString("field")
		nano, _ := cmd.Flags().GetBool("nano")

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute edit operation
		return executeEdit(configHandler, uiProvider, args, field, nano)
	},
}

// executeEdit handles the edit operation with the given dependencies
func executeEdit(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, field string, useNano bool) error {
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
		selected, err := uiProvider.SelectConfiguration(profiles, "edit")
		if err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}
		targetName = selected.Name
	} else {
		// CLI mode
		targetName = args[0]
	}

	// Execute edit
	if err := configHandler.EditConfig(targetName, field, useNano); err != nil {
		// Handle specific error messages
		if err.Error() == "no changes detected" {
			uiProvider.ShowInfo("No changes detected")
			return nil
		}
		if err.Error() == "no value provided" {
			uiProvider.ShowInfo("No changes made")
			return nil
		}
		uiProvider.ShowError(err)
		return err
	}

	if field != "" {
		uiProvider.ShowSuccess("Field '%s' updated successfully in configuration '%s'", field, targetName)
	} else {
		uiProvider.ShowSuccess("Configuration '%s' updated successfully", targetName)
	}
	return nil
}

func init() {
	editCmd.Flags().StringVar(&editField, "field", "", "Edit a specific field (e.g., 'env.ANTHROPIC_API_KEY')")
	editCmd.Flags().BoolVar(&useNano, "nano", false, "Use nano editor instead of default")
	editCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}
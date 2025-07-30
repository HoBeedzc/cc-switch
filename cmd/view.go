package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var (
	rawOutput bool
)

var viewCmd = &cobra.Command{
	Use:   "view [name]",
	Short: "View configuration content",
	Long: `Display the content and metadata of a specified configuration.

Modes:
- Interactive: cc-switch view (no arguments) or cc-switch view -i
- CLI: cc-switch view <name>

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
		raw, _ := cmd.Flags().GetBool("raw")

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute view operation
		return executeView(configHandler, uiProvider, args, raw)
	},
}

// executeView handles the view operation with the given dependencies
func executeView(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, raw bool) error {
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
		selected, err := uiProvider.SelectConfiguration(profiles, "view")
		if err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}
		targetName = selected.Name
	} else {
		// CLI mode
		targetName = args[0]
	}

	// Get configuration view
	view, err := configHandler.ViewConfig(targetName, raw)
	if err != nil {
		uiProvider.ShowError(err)
		return err
	}

	// Display configuration
	return uiProvider.DisplayConfiguration(view, raw)
}

func init() {
	viewCmd.Flags().BoolVar(&rawOutput, "raw", false, "Output raw JSON without metadata")
	viewCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}

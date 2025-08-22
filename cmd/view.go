package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var (
	rawOutput   bool
	currentFlag bool
)

var viewCmd = &cobra.Command{
	Use:   "view [name]",
	Short: "View configuration content",
	Long: `Display the content and metadata of a specified configuration.

Modes:
- Interactive: cc-switch view (no arguments) or cc-switch view -i
- CLI: cc-switch view <name>
- Current: cc-switch view --current or cc-switch view -c

The interactive mode allows you to browse and select configurations with arrow keys.
The --current flag displays the currently active configuration.`,
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
		current, _ := cmd.Flags().GetBool("current")

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive && !current {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute view operation
		return executeView(configHandler, uiProvider, args, raw, current)
	},
}

// executeView handles the view operation with the given dependencies
func executeView(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, raw bool, useCurrent bool) error {
	var targetName string

	// Priority: explicit profile name > --current flag > interactive mode
	if len(args) > 0 {
		// Has explicit profile name, ignore --current flag
		targetName = args[0]
	} else if useCurrent {
		// Use --current flag
		currentProfile, err := configHandler.GetCurrentConfigurationForOperation()
		if err != nil {
			return handleCurrentConfigError(err, uiProvider)
		}
		targetName = currentProfile
	} else {
		// Enter interactive mode
		profiles, err := configHandler.ListConfigs()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		if len(profiles) == 0 {
			uiProvider.ShowWarning("No configurations found.")
			fmt.Println("Use 'cc-switch new <name>' to create your first configuration.")
			return nil
		}

		selected, err := uiProvider.SelectConfiguration(profiles, "view")
		if err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}
		targetName = selected.Name
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
	viewCmd.Flags().BoolP("current", "c", false, "View current active configuration")
}

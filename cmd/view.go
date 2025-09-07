package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var rawOutput bool

var viewCmd = &cobra.Command{
	Use:   "view [name]",
	Short: "View configuration or template content",
	Long: `Display the content and metadata of a specified configuration or template.

Configuration Modes:
- Interactive: cc-switch view (no arguments) or cc-switch view -i
- CLI: cc-switch view <name>
- Current: cc-switch view --current or cc-switch view -c

Template Modes:
- Interactive: cc-switch view -t (no arguments) or cc-switch view -t -i
- CLI: cc-switch view -t <template-name>

The interactive mode allows you to browse and select configurations/templates with arrow keys.
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
		templateFlag, _ := cmd.Flags().GetBool("template")

		// Validate flag combinations
		if current && templateFlag {
			return fmt.Errorf("--current cannot be used with --template (-t)")
		}

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive && !current {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute view operation based on mode
		if templateFlag {
			return executeViewTemplate(configHandler, uiProvider, args, raw)
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

// executeViewTemplate handles the template view operation
func executeViewTemplate(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, raw bool) error {
	var targetName string

	// Determine execution mode
	if len(args) > 0 {
		// CLI mode - template name provided
		targetName = args[0]
	} else {
		// Interactive mode - select template
		templates, err := configHandler.ListTemplates()
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}

		if len(templates) == 0 {
			uiProvider.ShowWarning("No templates found.")
			fmt.Println("Use 'cc-switch edit -t <template-name>' to create your first template.")
			return nil
		}

		// Simple CLI selection for templates
		fmt.Println("Available templates:")
		for i, template := range templates {
			if template == "default" {
				fmt.Printf("  %d) %s (system default)\n", i+1, template)
			} else {
				fmt.Printf("  %d) %s\n", i+1, template)
			}
		}

		fmt.Printf("Select template to view (1-%d): ", len(templates))
		var selection int
		if _, err := fmt.Scanln(&selection); err != nil || selection < 1 || selection > len(templates) {
			return fmt.Errorf("invalid selection")
		}
		targetName = templates[selection-1]
	}

	// Validate input
	if targetName == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	// Get template view
	view, err := configHandler.ViewTemplate(targetName, raw)
	if err != nil {
		uiProvider.ShowError(err)
		return err
	}

	// Display template
	return uiProvider.DisplayTemplate(view, raw)
}

func init() {
	viewCmd.Flags().BoolVar(&rawOutput, "raw", false, "Output raw JSON without metadata")
	viewCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
	viewCmd.Flags().BoolP("current", "c", false, "View current active configuration")
	viewCmd.Flags().BoolP("template", "t", false, "View template instead of configuration")
}

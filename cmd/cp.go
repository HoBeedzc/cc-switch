package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var cpCmd = &cobra.Command{
	Use:     "cp <source> <destination>",
	Aliases: []string{"copy"},
	Short:   "Copy a configuration or template",
	Long: `Copy the specified configuration or template to a new name. The original remains unchanged.

Configuration Modes:
- Interactive: cc-switch cp (no arguments) or cc-switch cp -i
- CLI: cc-switch cp <source> <destination>

Template Modes:
- Interactive: cc-switch cp -t (no arguments) or cc-switch cp -t -i
- CLI: cc-switch cp -t <source-template> <destination-template>
- Create from template: cc-switch cp -t <template> <config-name> --to-config

The interactive mode allows you to browse and select configurations/templates with arrow keys.`,
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
		templateFlag, _ := cmd.Flags().GetBool("template")
		toConfigFlag, _ := cmd.Flags().GetBool("to-config")

		// Validate flag combinations
		if toConfigFlag && !templateFlag {
			return fmt.Errorf("--to-config can only be used with --template (-t)")
		}

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute copy operation based on mode
		if templateFlag {
			return executeCopyTemplate(configHandler, uiProvider, args, toConfigFlag)
		}

		// Execute regular copy operation
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

// executeCopyTemplate handles template copy operations
func executeCopyTemplate(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, toConfig bool) error {
	// Get all templates
	templates, err := configHandler.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templates) == 0 {
		uiProvider.ShowWarning("No templates found.")
		fmt.Println("Use 'cc-switch edit -t <template-name>' to create your first template.")
		return nil
	}

	var sourceName, destName string

	// Determine execution mode
	if len(args) == 0 {
		// Interactive mode - select source template
		fmt.Println("Available templates:")
		for i, template := range templates {
			if template == "default" {
				fmt.Printf("  %d) %s (system default)\n", i+1, template)
			} else {
				fmt.Printf("  %d) %s\n", i+1, template)
			}
		}

		fmt.Printf("Select template to copy (1-%d): ", len(templates))
		var selection int
		if _, err := fmt.Scanln(&selection); err != nil || selection < 1 || selection > len(templates) {
			return fmt.Errorf("invalid selection")
		}
		sourceName = templates[selection-1]

		// Get destination name interactively
		if toConfig {
			destName, err = uiProvider.GetInput(fmt.Sprintf("Enter configuration name to create from template '%s'", sourceName), "")
		} else {
			destName, err = uiProvider.GetInput(fmt.Sprintf("Enter destination template name for copying '%s'", sourceName), "")
		}
		if err != nil {
			return fmt.Errorf("input cancelled: %w", err)
		}
	} else if len(args) == 1 {
		// Partial CLI mode - source template provided, get destination name interactively
		sourceName = args[0]
		var err error
		// Create temporary UI for input
		tempUI := ui.NewCLIUI()
		if toConfig {
			destName, err = tempUI.GetInput(fmt.Sprintf("Enter configuration name to create from template '%s'", sourceName), "")
		} else {
			destName, err = tempUI.GetInput(fmt.Sprintf("Enter destination template name for copying '%s'", sourceName), "")
		}
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
		return fmt.Errorf("source template name cannot be empty")
	}
	if destName == "" {
		if toConfig {
			return fmt.Errorf("configuration name cannot be empty")
		}
		return fmt.Errorf("destination template name cannot be empty")
	}

	// Validate source template exists
	if err := configHandler.ValidateTemplateExists(sourceName); err != nil {
		return fmt.Errorf("source template '%s' does not exist", sourceName)
	}

	// Execute operation based on mode
	if toConfig {
		// Create configuration from template
		if err := configHandler.CreateConfig(destName, sourceName); err != nil {
			uiProvider.ShowError(err)
			return err
		}
		uiProvider.ShowSuccess("Configuration '%s' created from template '%s' successfully", destName, sourceName)
	} else {
		// Copy template to template
		if err := configHandler.CopyTemplate(sourceName, destName); err != nil {
			uiProvider.ShowError(err)
			return err
		}
		uiProvider.ShowSuccess("Template copied from '%s' to '%s' successfully", sourceName, destName)
	}

	return nil
}

func init() {
	cpCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
	cpCmd.Flags().BoolP("template", "t", false, "Copy template instead of configuration")
	cpCmd.Flags().Bool("to-config", false, "Create configuration from template (use with -t)")
}

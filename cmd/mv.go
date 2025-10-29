package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:     "mv <old-name> <new-name>",
	Aliases: []string{"rename", "move"},
	Short:   "Move (rename) a configuration or template",
	Long: `Move (rename) the specified configuration or template. If the configuration is currently active, 
the current configuration marker will be updated automatically.

Configuration Modes:
- Interactive: cc-switch mv (no arguments) or cc-switch mv -i
- CLI: cc-switch mv <old-name> <new-name>

Template Modes:
- Interactive: cc-switch mv -t (no arguments) or cc-switch mv -t -i
- CLI: cc-switch mv -t <old-template> <new-template>

The interactive mode allows you to browse and select configurations/templates with arrow keys.`,
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
		templateFlag, _ := cmd.Flags().GetBool("template")

		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute move operation based on mode
		if templateFlag {
			return executeMoveTemplate(configHandler, uiProvider, args)
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

// executeMoveTemplate handles template move operations
func executeMoveTemplate(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string) error {
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

	var oldName, newName string

	// Determine execution mode
	if len(args) == 0 {
		// Interactive mode - select template to move
		// Filter out default template (cannot be moved)
		var movableTemplates []string
		for _, template := range templates {
			if template != "default" {
				movableTemplates = append(movableTemplates, template)
			}
		}

		if len(movableTemplates) == 0 {
			uiProvider.ShowWarning("No templates available for renaming (default template cannot be moved).")
			return nil
		}

		fmt.Println("Available templates for renaming:")
		for i, template := range movableTemplates {
			fmt.Printf("  %d) %s\n", i+1, template)
		}

		fmt.Printf("Select template to rename (1-%d): ", len(movableTemplates))
		var selection int
		if _, err := fmt.Scanln(&selection); err != nil || selection < 1 || selection > len(movableTemplates) {
			return fmt.Errorf("invalid selection")
		}
		oldName = movableTemplates[selection-1]

		// Get new name interactively
		newName, err = uiProvider.GetInput(fmt.Sprintf("Enter new name for template '%s'", oldName), "")
		if err != nil {
			return fmt.Errorf("input cancelled: %w", err)
		}
	} else if len(args) == 1 {
		// Partial CLI mode - old name provided, get new name interactively
		oldName = args[0]
		var err error
		// Create temporary UI for input
		tempUI := ui.NewCLIUI()
		newName, err = tempUI.GetInput(fmt.Sprintf("Enter new name for template '%s'", oldName), "")
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
		return fmt.Errorf("old template name cannot be empty")
	}
	if newName == "" {
		return fmt.Errorf("new template name cannot be empty")
	}

	// Safety check: prevent moving of default template
	if oldName == "default" {
		return fmt.Errorf("default template cannot be moved/renamed")
	}

	// Validate source template exists
	if err := configHandler.ValidateTemplateExists(oldName); err != nil {
		return fmt.Errorf("template '%s' does not exist", oldName)
	}

	// Execute move
	if err := configHandler.MoveTemplate(oldName, newName); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("Template moved from '%s' to '%s' successfully", oldName, newName)
	return nil
}

func init() {
	mvCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
	mvCmd.Flags().BoolP("template", "t", false, "Move template instead of configuration")
}

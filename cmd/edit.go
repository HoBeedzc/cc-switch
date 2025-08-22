package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var (
	editField     string
	useNano       bool
	templateName  string
	listTemplates bool
)

var editCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit configuration or template content",
	Long: `Edit the content of a configuration or template using your system editor.

Configuration Mode:
- Interactive: cc-switch edit (no arguments) or cc-switch edit -i
- CLI: cc-switch edit <name>
- Current: cc-switch edit --current or cc-switch edit -c

Template Mode:
- Edit template: cc-switch edit -t <template-name> or cc-switch edit --template <template-name>
- List templates: cc-switch edit -t --list or cc-switch edit --template --list

The interactive mode allows you to browse and select configurations with arrow keys.
The --current flag edits the currently active configuration.

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
		templateName, _ := cmd.Flags().GetString("template")
		listTemplates, _ := cmd.Flags().GetBool("list")
		field, _ := cmd.Flags().GetString("field")
		nano, _ := cmd.Flags().GetBool("nano")
		current, _ := cmd.Flags().GetBool("current")

		// Template mode handling
		if listTemplates {
			// If just --list without template name, list templates
			return executeListTemplates(configHandler)
		}

		if templateName != "" {
			return executeEditTemplate(configHandler, templateName, field, nano)
		}

		// Regular configuration editing mode
		interactiveFlag, _ := cmd.Flags().GetBool("interactive")

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive && !current {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Execute edit operation
		return executeEdit(configHandler, uiProvider, args, field, nano, current)
	},
}

// executeEdit handles the edit operation with the given dependencies
func executeEdit(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, field string, useNano bool, useCurrent bool) error {
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

		selected, err := uiProvider.SelectConfiguration(profiles, "edit")
		if err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}
		targetName = selected.Name
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

// executeListTemplates handles listing templates
func executeListTemplates(configHandler handler.ConfigHandler) error {
	templates, err := configHandler.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templates) == 0 {
		fmt.Println("No templates found.")
		return nil
	}

	fmt.Println("Available templates:")
	for _, template := range templates {
		fmt.Printf("  %s\n", template)
	}

	return nil
}

// executeEditTemplate handles template editing
func executeEditTemplate(configHandler handler.ConfigHandler, templateName string, field string, useNano bool) error {
	if templateName == "" {
		return fmt.Errorf("template name is required")
	}

	// Check if template exists, create if it doesn't
	if err := configHandler.ValidateTemplateExists(templateName); err != nil {
		// Template doesn't exist, create it
		fmt.Printf("Template '%s' does not exist. Creating...\n", templateName)
		if err := configHandler.CreateTemplate(templateName); err != nil {
			return fmt.Errorf("failed to create template: %w", err)
		}
		fmt.Printf("Template '%s' created successfully.\n", templateName)
	}

	// Execute edit
	if err := configHandler.EditTemplate(templateName, field, useNano); err != nil {
		// Handle specific error messages
		if err.Error() == "no changes detected" {
			fmt.Println("No changes detected")
			return nil
		}
		if err.Error() == "no value provided" {
			fmt.Println("No changes made")
			return nil
		}
		return err
	}

	if field != "" {
		fmt.Printf("Field '%s' updated successfully in template '%s'\n", field, templateName)
	} else {
		fmt.Printf("Template '%s' updated successfully\n", templateName)
	}

	return nil
}

func init() {
	editCmd.Flags().StringVar(&editField, "field", "", "Edit a specific field (e.g., 'env.ANTHROPIC_API_KEY')")
	editCmd.Flags().BoolVar(&useNano, "nano", false, "Use nano editor instead of default")
	editCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
	editCmd.Flags().StringVarP(&templateName, "template", "t", "", "Edit template instead of configuration")
	editCmd.Flags().BoolVar(&listTemplates, "list", false, "List available templates (use with --template)")
	editCmd.Flags().BoolP("current", "c", false, "Edit current active configuration")
}

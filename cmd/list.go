package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available configurations or templates",
	Long: `Display all available Claude Code configurations or templates.

Modes:
- Configurations: cc-switch list (default)
- Templates: cc-switch list -t or cc-switch list --template

The current configuration is highlighted when listing configurations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkClaudeConfig(); err != nil {
			return err
		}

		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		configHandler := handler.NewConfigHandler(cm)

		// Check for template flag
		template, _ := cmd.Flags().GetBool("template")

		// Handle template listing
		if template {
			return executeListTemplates(configHandler)
		}

		// Check if in empty mode first
		if configHandler.IsEmptyMode() {
			color.Yellow("⚠️  Empty mode active (no configuration active)")
			fmt.Println()
		}

		profiles, err := cm.ListProfiles()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		if len(profiles) == 0 {
			fmt.Println("No configurations found. Use 'cc-switch new <name>' to create your first configuration.")
			return nil
		}

		fmt.Println("Available configurations:")
		for _, profile := range profiles {
			if profile.IsCurrent && !configHandler.IsEmptyMode() {
				color.Green("  * %s (current)", profile.Name)
			} else {
				fmt.Printf("    %s\n", profile.Name)
			}
		}

		// Show helpful tips if in empty mode
		if configHandler.IsEmptyMode() {
			fmt.Println("\n💡 Use 'cc-switch use <name>' to activate a configuration or 'cc-switch use --restore' to restore previous")
		}

		return nil
	},
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
		if template == "default" {
			fmt.Printf("  %s (system default)\n", template)
		} else {
			fmt.Printf("  %s\n", template)
		}
	}

	return nil
}

func init() {
	listCmd.Flags().BoolP("template", "t", false, "List templates instead of configurations")
}

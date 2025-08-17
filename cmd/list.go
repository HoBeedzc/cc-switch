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
	Short: "List all available configurations",
	Long:  `Display all available Claude Code configurations with the current one highlighted.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkClaudeConfig(); err != nil {
			return err
		}

		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		configHandler := handler.NewConfigHandler(cm)

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
			fmt.Println("No configurations found.")
			fmt.Println("Use 'cc-switch new <name>' to create your first configuration.")
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
			fmt.Println()
			fmt.Println("💡 Use 'cc-switch use <name>' to activate a configuration")
			fmt.Println("💡 Use 'cc-switch use --restore' to restore previous configuration")
		}

		return nil
	},
}

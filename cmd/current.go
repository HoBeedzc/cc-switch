package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current configuration",
	Long:  `Display the name of the currently active configuration.`,
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
			status, err := configHandler.GetEmptyModeStatus()
			if err != nil {
				return fmt.Errorf("failed to get empty mode status: %w", err)
			}

			color.Yellow("Empty mode (no configuration active)")
			if status.PreviousProfile != "" {
				fmt.Printf("💡 Previous configuration: %s\n", status.PreviousProfile)
				fmt.Printf("💡 Enabled at: %s\n", status.Timestamp)
				fmt.Println("💡 Use 'cc-switch use --restore' to restore previous configuration")
			} else {
				fmt.Println("💡 Use 'cc-switch use <profile>' to activate a configuration")
			}
			return nil
		}

		current, err := cm.GetCurrentProfile()
		if err != nil {
			return fmt.Errorf("failed to get current profile: %w", err)
		}

		if current == "" {
			color.Yellow("No current configuration set")
		} else {
			color.Green("Current configuration: %s", current)
		}

		return nil
	},
}

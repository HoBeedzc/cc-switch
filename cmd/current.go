package cmd

import (
	"fmt"

	"cc-switch/internal/config"

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
package cmd

import (
	"fmt"

	"cc-switch/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch to a configuration",
	Long:  `Switch to the specified configuration. This will replace the current Claude Code settings.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf(`Missing required argument: configuration name

Usage: cc-switch use <name>

Example: cc-switch use production

Use 'cc-switch list' to see available configurations.
Use 'cc-switch use --help' for more information.`)
		}
		if len(args) > 1 {
			return fmt.Errorf(`Too many arguments provided

Usage: cc-switch use <name>

Example: cc-switch use production

Use 'cc-switch use --help' for more information.`)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := checkClaudeConfig(); err != nil {
			return err
		}

		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		// 检查配置是否存在
		if !cm.ProfileExists(name) {
			return fmt.Errorf("configuration '%s' does not exist. Use 'cc-switch list' to see available configurations", name)
		}

		// 检查是否已经是当前配置
		currentProfile, err := cm.GetCurrentProfile()
		if err == nil && currentProfile == name {
			color.Yellow("Configuration '%s' is already active", name)
			return nil
		}

		// 切换配置
		if err := cm.UseProfile(name); err != nil {
			return err
		}

		color.Green("✓ Switched to configuration '%s'", name)
		return nil
	},
}
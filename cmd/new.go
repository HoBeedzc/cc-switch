package cmd

import (
	"fmt"

	"cc-switch/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new configuration",
	Long:  `Create a new configuration with empty template structure ready for customization.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf(`Missing required argument: configuration name

Usage: cc-switch new <name>

Example: cc-switch new production

Use 'cc-switch new --help' for more information.`)
		}
		if len(args) > 1 {
			return fmt.Errorf(`Too many arguments provided

Usage: cc-switch new <name>

Example: cc-switch new production

Use 'cc-switch new --help' for more information.`)
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

		// 检查配置是否已存在
		if cm.ProfileExists(name) {
			return fmt.Errorf("configuration '%s' already exists", name)
		}

		// 创建新配置
		if err := cm.CreateProfile(name); err != nil {
			return err
		}

		color.Green("✓ Configuration '%s' created successfully", name)
		fmt.Printf("Use 'cc-switch edit %s' to customize the configuration.\n", name)
		fmt.Printf("Use 'cc-switch use %s' to switch to this configuration.\n", name)

		return nil
	},
}
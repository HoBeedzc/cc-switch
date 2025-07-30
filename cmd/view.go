package cmd

import (
	"encoding/json"
	"fmt"

	"cc-switch/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	rawOutput bool
)

var viewCmd = &cobra.Command{
	Use:   "view <name>",
	Short: "View configuration content",
	Long:  `Display the content and metadata of a specified configuration.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf(`Missing required argument: configuration name

Usage: cc-switch view <name>

Example: cc-switch view production

Use 'cc-switch list' to see available configurations.
Use 'cc-switch view --help' for more information.`)
		}
		if len(args) > 1 {
			return fmt.Errorf(`Too many arguments provided

Usage: cc-switch view <name>

Example: cc-switch view production

Use 'cc-switch view --help' for more information.`)
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
			return fmt.Errorf("configuration '%s' does not exist", name)
		}

		// 获取配置内容
		content, metadata, err := cm.GetProfileContent(name)
		if err != nil {
			return fmt.Errorf("failed to read configuration: %w", err)
		}

		if rawOutput {
			// 原始JSON输出
			jsonData, err := json.MarshalIndent(content, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(jsonData))
		} else {
			// 格式化输出
			color.Blue("Configuration: %s", name)
			if metadata.IsCurrent {
				color.Green("Status: Current")
			} else {
				fmt.Println("Status: Available")
			}
			fmt.Printf("Path: %s\n", metadata.Path)
			fmt.Println()

			color.Yellow("Content:")
			jsonData, err := json.MarshalIndent(content, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(jsonData))
		}

		return nil
	},
}

func init() {
	viewCmd.Flags().BoolVar(&rawOutput, "raw", false, "Output raw JSON without metadata")
}
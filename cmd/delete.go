package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"cc-switch/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a configuration",
	Long:  `Delete the specified configuration. You cannot delete the currently active configuration.`,
	Args:  cobra.ExactArgs(1),
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

		// 获取确认标志
		force, _ := cmd.Flags().GetBool("force")

		// 如果没有force标志，请求确认
		if !force {
			fmt.Printf("Are you sure you want to delete configuration '%s'? (y/N): ", name)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read user input: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				color.Yellow("Operation cancelled")
				return nil
			}
		}

		// 删除配置
		if err := cm.DeleteProfile(name); err != nil {
			return err
		}

		color.Green("✓ Configuration '%s' deleted successfully", name)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolP("force", "f", false, "Force delete without confirmation")
}
package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/interactive"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a configuration",
	Long: `Delete the specified configuration. You cannot delete the currently active configuration.

Modes:
- Interactive: cc-switch delete (no arguments) or cc-switch delete -i
- CLI: cc-switch delete <name>

The interactive mode allows you to browse and select configurations with arrow keys.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkClaudeConfig(); err != nil {
			return err
		}

		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		// 检测执行模式
		interactiveFlag, _ := cmd.Flags().GetBool("interactive")
		mode := interactive.DetectMode(interactiveFlag, args)

		switch mode {
		case interactive.Interactive:
			return handleInteractiveDelete(cm)
		case interactive.CLI:
			force, _ := cmd.Flags().GetBool("force")
			return handleCLIDelete(cm, args[0], force)
		}

		return nil
	},
}

// handleInteractiveDelete 处理交互式删除
func handleInteractiveDelete(cm *config.ConfigManager) error {
	// 获取所有配置
	profiles, err := cm.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		interactive.ShowWarning("No configurations found.")
		fmt.Println("Use 'cc-switch new <name>' to create your first configuration.")
		return nil
	}

	// 过滤掉当前配置（不能删除当前配置）
	var deletableProfiles []config.Profile
	for _, profile := range profiles {
		if !profile.IsCurrent {
			deletableProfiles = append(deletableProfiles, profile)
		}
	}

	if len(deletableProfiles) == 0 {
		interactive.ShowWarning("No configurations available for deletion.")
		fmt.Println("The current configuration cannot be deleted. Switch to another configuration first.")
		return nil
	}

	// 显示选择界面
	selector := interactive.NewInteractiveSelector(deletableProfiles, "delete")
	selected, err := selector.Show()
	if err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	// 确认删除
	confirmMsg := fmt.Sprintf("Are you sure you want to delete configuration '%s'?", selected.Name)
	if !interactive.ConfirmAction(confirmMsg, false) {
		interactive.ShowInfo("Operation cancelled")
		return nil
	}

	// 执行删除
	if err := cm.DeleteProfile(selected.Name); err != nil {
		return fmt.Errorf("failed to delete configuration: %w", err)
	}

	interactive.ShowSuccess("Configuration '%s' deleted successfully", selected.Name)
	return nil
}

// handleCLIDelete 处理命令行删除
func handleCLIDelete(cm *config.ConfigManager, name string, force bool) error {
	// 检查配置是否存在
	if !cm.ProfileExists(name) {
		return fmt.Errorf("configuration '%s' does not exist", name)
	}

	// 如果没有force标志，请求确认
	if !force {
		confirmMsg := fmt.Sprintf("Are you sure you want to delete configuration '%s'?", name)
		if !interactive.ConfirmAction(confirmMsg, false) {
			interactive.ShowInfo("Operation cancelled")
			return nil
		}
	}

	// 删除配置
	if err := cm.DeleteProfile(name); err != nil {
		return err
	}

	interactive.ShowSuccess("Configuration '%s' deleted successfully", name)
	return nil
}

func init() {
	deleteCmd.Flags().BoolP("force", "f", false, "Force delete without confirmation")
	deleteCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}
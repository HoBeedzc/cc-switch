package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/interactive"
)

// handleInteractiveUse 处理交互式切换
func handleInteractiveUse(cm *config.ConfigManager) error {
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

	// 显示选择界面
	selector := interactive.NewInteractiveSelector(profiles, "use")
	selected, err := selector.Show()
	if err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	// 检查是否已经是当前配置
	if selected.IsCurrent {
		interactive.ShowWarning("Configuration '%s' is already active", selected.Name)
		return nil
	}

	// 切换配置
	if err := cm.UseProfile(selected.Name); err != nil {
		return fmt.Errorf("failed to switch configuration: %w", err)
	}

	interactive.ShowSuccess("Switched to configuration '%s'", selected.Name)
	return nil
}

// handleCLIUse 处理命令行切换
func handleCLIUse(cm *config.ConfigManager, name string) error {
	// 检查配置是否存在
	if !cm.ProfileExists(name) {
		return fmt.Errorf("configuration '%s' does not exist. Use 'cc-switch list' to see available configurations", name)
	}

	// 检查是否已经是当前配置
	currentProfile, err := cm.GetCurrentProfile()
	if err == nil && currentProfile == name {
		interactive.ShowWarning("Configuration '%s' is already active", name)
		return nil
	}

	// 切换配置
	if err := cm.UseProfile(name); err != nil {
		return err
	}

	interactive.ShowSuccess("Switched to configuration '%s'", name)
	return nil
}
package cmd

import (
	"cc-switch/internal/config"
	"cc-switch/internal/interactive"
	"fmt"
)

// handleInteractiveEdit 处理交互式编辑
func handleInteractiveEdit(cm *config.ConfigManager) error {
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
	selector := interactive.NewInteractiveSelector(profiles, "edit")
	selected, err := selector.Show()
	if err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	// 执行编辑
	return performEdit(cm, selected.Name)
}

// handleCLIEdit 处理命令行编辑
func handleCLIEdit(cm *config.ConfigManager, name string) error {
	// 检查配置是否存在
	if !cm.ProfileExists(name) {
		return fmt.Errorf("configuration '%s' does not exist", name)
	}

	// 执行编辑
	return performEdit(cm, name)
}

// performEdit 执行编辑操作
func performEdit(cm *config.ConfigManager, name string) error {
	if editField != "" {
		// 字段编辑模式
		return editProfileField(cm, name, editField)
	} else {
		// 编辑器模式
		return editProfileWithEditor(cm, name)
	}
}

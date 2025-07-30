package cmd

import (
	"encoding/json"
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/interactive"

	"github.com/fatih/color"
)

// handleInteractiveView 处理交互式查看
func handleInteractiveView(cm *config.ConfigManager, rawOutput bool) error {
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
	selector := interactive.NewInteractiveSelector(profiles, "view")
	selected, err := selector.Show()
	if err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	// 显示配置内容
	return displayConfiguration(cm, selected.Name, rawOutput)
}

// handleCLIView 处理命令行查看
func handleCLIView(cm *config.ConfigManager, name string, rawOutput bool) error {
	// 检查配置是否存在
	if !cm.ProfileExists(name) {
		return fmt.Errorf("configuration '%s' does not exist", name)
	}

	// 显示配置内容
	return displayConfiguration(cm, name, rawOutput)
}

// displayConfiguration 显示配置内容
func displayConfiguration(cm *config.ConfigManager, name string, rawOutput bool) error {
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
}

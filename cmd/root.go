package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cc-switch",
	Short: "Claude Code configuration switcher",
	Long: `A command-line tool for managing and switching between different Claude Code configurations.

This tool allows you to:
- List all available configurations
- Create new configurations  
- Switch between configurations
- Move (rename) configurations
- Copy configurations
- Remove configurations
- Export configurations to backup files
- Import configurations from backup files`,
	SilenceUsage: true,
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(mvCmd)
	rootCmd.AddCommand(cpCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(currentCmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(initCmd)
}

// 检查Claude配置是否存在的助手函数
func checkClaudeConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	settingsPath := filepath.Join(claudeDir, "settings.json")
	emptyModeFile := filepath.Join(claudeDir, ".empty_mode")

	// Check if in empty mode - if so, allow the operation
	if _, err := os.Stat(emptyModeFile); err == nil {
		return nil // Empty mode is valid
	}

	// Check for profiles directory (indicates cc-switch is initialized)
	profilesDir := filepath.Join(claudeDir, "profiles")
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		return fmt.Errorf(`Claude configuration not found at %s

To initialize your Claude Code configuration, run:
  cc-switch init

This will guide you through the initial setup process.`, settingsPath)
	}

	// If not in empty mode, settings.json should exist
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return fmt.Errorf(`Claude configuration not found at %s

To initialize your Claude Code configuration, run:
  cc-switch init

This will guide you through the initial setup process.`, settingsPath)
	}

	return nil
}

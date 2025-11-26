package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"cc-switch/internal/common"

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
	Version:      common.Version,
}

// skipUpdateNotice determines if update notice should be skipped for certain commands
var skipUpdateNotice bool

// Execute 执行根命令
func Execute() error {
	// Start background update check if needed
	if common.ShouldCheckUpdate() {
		common.CheckUpdateBackground(nil)
	}

	// Execute the command
	err := rootCmd.Execute()

	// Show update notice after command execution (if cached)
	// Skip for update command (it handles its own update logic)
	if !skipUpdateNotice {
		common.PrintUpdateNotice()
	}

	return err
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
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(uninstallCmd)
}

// 检查Claude配置是否存在的助手函数
func checkClaudeConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	settingsPath := filepath.Join(claudeDir, "settings.json")
	profilesDir := filepath.Join(claudeDir, "profiles")
	emptyModeFile := filepath.Join(profilesDir, ".empty_mode")

	// Check if in empty mode - if so, allow the operation
	if _, err := os.Stat(emptyModeFile); err == nil {
		return nil // Empty mode is valid
	}

	// Check for profiles directory (indicates cc-switch is initialized)
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		return fmt.Errorf("claude configuration not found at %s", settingsPath)
	}

	// If not in empty mode, settings.json should exist
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return fmt.Errorf("claude configuration not found at %s", settingsPath)
	}

	return nil
}

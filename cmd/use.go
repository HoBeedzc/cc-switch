package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Switch to a configuration",
	Long: `Switch to the specified configuration. This will replace the current Claude Code settings.

Modes:
- Interactive: cc-switch use (no arguments) or cc-switch use -i
- CLI: cc-switch use <name>
- Previous: cc-switch use -p or cc-switch use --previous
- Empty Mode: cc-switch use -e or cc-switch use --empty
- Restore: cc-switch use -r or cc-switch use --restore

Options:
- Launch Claude Code: Add -l or --launch to automatically launch Claude Code CLI after switching

The interactive mode allows you to browse and select configurations with arrow keys.
The previous mode switches to the last used configuration.
The empty mode temporarily removes all configurations.
The restore mode restores from empty mode to the previous configuration.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkClaudeConfig(); err != nil {
			return err
		}

		// Initialize dependencies
		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		configHandler := handler.NewConfigHandler(cm)
		interactiveFlag, _ := cmd.Flags().GetBool("interactive")
		previousFlag, _ := cmd.Flags().GetBool("previous")
		emptyFlag, _ := cmd.Flags().GetBool("empty")
		restoreFlag, _ := cmd.Flags().GetBool("restore")
		launchFlag, _ := cmd.Flags().GetBool("launch")

		// Validate flag combinations
		flagCount := 0
		if previousFlag {
			flagCount++
		}
		if emptyFlag {
			flagCount++
		}
		if restoreFlag {
			flagCount++
		}
		if len(args) > 0 {
			flagCount++
		}

		if flagCount > 1 {
			return fmt.Errorf("cannot use multiple operation flags together")
		}

		if (previousFlag || emptyFlag || restoreFlag) && interactiveFlag {
			return fmt.Errorf("cannot use operation flags with -i/--interactive")
		}

		// Create UI provider based on mode
		var uiProvider ui.UIProvider
		if !previousFlag && !emptyFlag && !restoreFlag && ui.NewInteractiveUI().DetectMode(interactiveFlag, args) == ui.Interactive {
			uiProvider = ui.NewInteractiveUI()
		} else {
			uiProvider = ui.NewCLIUI()
		}

		// Handle special operations
		if emptyFlag {
			return handleEmptyMode(configHandler, uiProvider)
		}

		if restoreFlag {
			return handleRestoreMode(configHandler, uiProvider, launchFlag)
		}

		if previousFlag {
			return handlePreviousConfig(configHandler, uiProvider, launchFlag)
		}

		// Execute normal use operation
		return executeUse(configHandler, uiProvider, args, launchFlag)
	},
}

// executeUse handles the use operation with the given dependencies
func executeUse(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, args []string, launchCode bool) error {
	// Check if currently in empty mode - if so, any use command should restore first
	if configHandler.IsEmptyMode() {
		uiProvider.ShowInfo("Currently in empty mode. Restoring settings first...")
		if err := configHandler.RestoreFromEmptyMode(); err != nil {
			uiProvider.ShowError(fmt.Errorf("failed to restore from empty mode: %w", err))
			return err
		}
		uiProvider.ShowInfo("Settings restored from empty mode.")
	}

	// Get all configurations
	profiles, err := configHandler.ListConfigs()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		uiProvider.ShowWarning("No configurations found. Use 'cc-switch new <name>' to create your first configuration.")
		return nil
	}

	var targetName string

	// Determine execution mode
	if len(args) == 0 {
		// Interactive mode - use enhanced selector with empty mode support
		if ui.NewInteractiveUI().DetectMode(false, args) == ui.Interactive {
			interactiveUI := uiProvider.(ui.InteractiveUI)
			selection, err := interactiveUI.SelectConfigurationWithEmptyMode(profiles, "use", configHandler.IsEmptyMode())
			if err != nil {
				return fmt.Errorf("selection cancelled: %w", err)
			}

			// Handle special selections
			switch selection.Type {
			case "empty_mode":
				return handleEmptyMode(configHandler, uiProvider)
			case "restore":
				return handleRestoreMode(configHandler, uiProvider, launchCode)
			case "profile":
				// Check if already current
				if selection.Profile.IsCurrent && !configHandler.IsEmptyMode() {
					uiProvider.ShowWarning("Configuration '%s' is already active", selection.Profile.Name)
					return nil
				}
				targetName = selection.Profile.Name
			default:
				return fmt.Errorf("unknown selection type: %s", selection.Type)
			}
		} else {
			// Fallback to regular selection
			selected, err := uiProvider.SelectConfiguration(profiles, "use")
			if err != nil {
				return fmt.Errorf("selection cancelled: %w", err)
			}

			// Check if already current
			if selected.IsCurrent && !configHandler.IsEmptyMode() {
				uiProvider.ShowWarning("Configuration '%s' is already active", selected.Name)
				return nil
			}

			targetName = selected.Name
		}
	} else {
		// CLI mode
		targetName = args[0]
	}

	// Execute switch
	if err := configHandler.UseConfig(targetName); err != nil {
		// Handle specific error messages
		if err.Error() == fmt.Sprintf("configuration '%s' is already active", targetName) {
			uiProvider.ShowWarning("Configuration '%s' is already active", targetName)
			return nil
		}
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("Switched to configuration '%s'", targetName)

	// Launch Claude Code if requested
	if launchCode {
		if err := launchClaudeCode(uiProvider); err != nil {
			uiProvider.ShowWarning("Failed to launch Claude Code: %v. Launch manually with: claude", err)
		}
	}

	return nil
}

// handlePreviousConfig handles switching to the previous configuration
func handlePreviousConfig(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, launchCode bool) error {
	// Special handling for empty mode: -p should behave like -r
	if configHandler.IsEmptyMode() {
		uiProvider.ShowInfo("In empty mode: using previous (-p) will restore from empty mode")
		return handleRestoreMode(configHandler, uiProvider, launchCode)
	}

	// Get previous configuration
	previousName, err := configHandler.GetPreviousConfig()
	if err != nil {
		if err.Error() == "no previous configuration available" {
			uiProvider.ShowWarning("No previous configuration available. Use 'cc-switch use <name>' to switch configurations first")
			return nil
		}
		if strings.Contains(err.Error(), "no longer exists") {
			uiProvider.ShowWarning(err.Error() + ". The previous configuration has been deleted")
			return nil
		}
		uiProvider.ShowError(err)
		return err
	}

	// Special case: if previous is "empty_mode", enter empty mode
	if previousName == "empty_mode" {
		uiProvider.ShowInfo("Previous state was empty mode. Entering empty mode...")
		return handleEmptyMode(configHandler, uiProvider)
	}

	// Get current configuration for display
	currentName, _ := configHandler.GetCurrentConfig()

	// Check if previous is same as current (edge case)
	if previousName == currentName {
		uiProvider.ShowWarning("Previous configuration '%s' is the same as current", previousName)
		return nil
	}

	// Execute switch
	if err := configHandler.UseConfig(previousName); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	// Show success message with context
	if currentName != "" {
		uiProvider.ShowSuccess("Switched to configuration '%s' (previous: '%s')", previousName, currentName)
	} else {
		uiProvider.ShowSuccess("Switched to configuration '%s'", previousName)
	}

	// Launch Claude Code if requested
	if launchCode {
		if err := launchClaudeCode(uiProvider); err != nil {
			uiProvider.ShowWarning("Failed to launch Claude Code: %v. Launch manually with: claude", err)
		}
	}

	return nil
}

// handleEmptyMode handles enabling empty mode
func handleEmptyMode(configHandler handler.ConfigHandler, uiProvider ui.UIProvider) error {
	// Check if already in empty mode
	if configHandler.IsEmptyMode() {
		uiProvider.ShowWarning("Already in empty mode. Use 'cc-switch use <profile>' to restore or 'cc-switch use --restore' for previous configuration")
		return nil
	}

	// Get current configuration for display
	currentName, _ := configHandler.GetCurrentConfig()

	// Enable empty mode
	if err := configHandler.UseEmptyMode(); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	// Show success message
	if currentName != "" {
		uiProvider.ShowSuccess("Empty mode enabled. Previous: %s. Use 'cc-switch use <profile>' to restore or '--restore' for previous", currentName)
	} else {
		uiProvider.ShowSuccess("Empty mode enabled. Use 'cc-switch use <profile>' to restore a configuration")
	}

	return nil
}

// handleRestoreMode handles restoring from empty mode
func handleRestoreMode(configHandler handler.ConfigHandler, uiProvider ui.UIProvider, launchCode bool) error {
	// Check if in empty mode
	if !configHandler.IsEmptyMode() {
		uiProvider.ShowWarning("Not in empty mode")
		return nil
	}

	// Get empty mode status
	status, err := configHandler.GetEmptyModeStatus()
	if err != nil {
		uiProvider.ShowError(fmt.Errorf("failed to get empty mode status: %w", err))
		return err
	}

	// Check if we can restore to previous
	if !status.CanRestore {
		uiProvider.ShowWarning("No previous configuration to restore to. Use 'cc-switch use <profile>' to switch to a specific configuration")
		return nil
	}

	// Restore to previous configuration
	if err := configHandler.RestoreToPreviousFromEmptyMode(); err != nil {
		uiProvider.ShowError(err)
		return err
	}

	uiProvider.ShowSuccess("Restored to previous configuration '%s'", status.PreviousProfile)

	// Launch Claude Code if requested
	if launchCode {
		if err := launchClaudeCode(uiProvider); err != nil {
			uiProvider.ShowWarning("Failed to launch Claude Code: %v. Launch manually with: claude", err)
		}
	}

	return nil
}

// launchClaudeCode launches Claude Code CLI with appropriate error handling
func launchClaudeCode(uiProvider ui.UIProvider) error {
	// Try to find Claude Code CLI executable
	claudePath, err := findClaudeCodeExecutable()
	if err != nil {
		return fmt.Errorf("claude Code CLI not found: %w", err)
	}

	uiProvider.ShowInfo("Starting Claude Code CLI in current terminal... (Press Ctrl+C or type 'exit' to return)")
	fmt.Println("") // Visual separation

	// Create the command with proper terminal inheritance
	cmd := exec.Command(claudePath)

	// Inherit the current terminal's stdin, stdout, and stderr
	// This allows Claude Code to run interactively in the current terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the process and wait for it to complete
	// This will make Claude Code take over the current terminal
	if err := cmd.Run(); err != nil {
		// Don't treat non-zero exit codes as errors for interactive programs
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			// Exit code 1 might be normal for Claude Code CLI (user exit, etc.)
			if exitCode == 1 {
				uiProvider.ShowInfo("Claude Code session ended")
				return nil
			}
		}
		return fmt.Errorf("claude Code exited with error: %w", err)
	}

	// This line will execute after Claude Code exits normally
	uiProvider.ShowInfo("Claude Code session ended")
	return nil
}

// findClaudeCodeExecutable attempts to find the Claude Code CLI executable
func findClaudeCodeExecutable() (string, error) {
	// List of possible Claude Code CLI commands to try
	possibleCommands := []string{
		"claude",      // Most common installation
		"claude-code", // Alternative installation
		"code",        // Some setups might alias this way
	}

	for _, cmd := range possibleCommands {
		// Check if command exists and is executable
		if path, err := exec.LookPath(cmd); err == nil {
			// Verify it's actually Claude Code CLI by checking version output
			if isClaudeCodeCLI(path) {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("claude Code CLI executable not found in PATH. Please install Claude Code CLI or ensure 'claude' command is available")
}

// isClaudeCodeCLI verifies that the given executable is actually Claude Code CLI
func isClaudeCodeCLI(execPath string) bool {
	// Try to run the command with --version flag to verify it's Claude Code CLI
	cmd := exec.Command(execPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	outputStr := strings.ToLower(string(output))
	// Look for indicators that this is Claude Code CLI
	claudeIndicators := []string{
		"claude",
		"anthropic",
		"claude code",
	}

	for _, indicator := range claudeIndicators {
		if strings.Contains(outputStr, indicator) {
			return true
		}
	}

	return false
}

func init() {
	useCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
	useCmd.Flags().BoolP("previous", "p", false, "Switch to previous configuration")
	useCmd.Flags().BoolP("empty", "e", false, "Enable empty mode (remove settings)")
	useCmd.Flags().BoolP("restore", "r", false, "Restore from empty mode to previous configuration")
	useCmd.Flags().BoolP("launch", "l", false, "Launch Claude Code CLI after switching")
}

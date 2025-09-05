package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Claude Code configuration",
	Long: `Initialize Claude Code configuration with interactive setup.

This command will:
- Check if settings.json already exists
- Create initial configuration with user input
- Set up directory structure for cc-switch
- Create default profile from initial configuration

You will be prompted to enter:
- ANTHROPIC_AUTH_TOKEN (your Claude API token)
- ANTHROPIC_BASE_URL (optional custom API endpoint)

Both fields can be left empty and configured later.`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	// Create config manager without auto-initialization
	configManager, err := config.NewConfigManagerNoInit()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Create handler and UI provider
	configHandler := handler.NewConfigHandler(configManager)
	uiProvider := ui.NewCLIUI()

	// Check if in empty mode and warn user FIRST
	if configManager.IsEmptyMode() {
		uiProvider.ShowWarning("Currently in empty mode, which may cause unexpected errors. Recommend using 'cc-switch use --restore' or 'cc-switch use <profile>' to exit empty mode first")
		fmt.Println()

		if !uiProvider.ConfirmAction("Continue with initialization?", false) {
			fmt.Println("Initialization cancelled")
			return nil
		}

		// If user chooses to continue in empty mode, don't show "already initialized"
		// message even if profiles exist, because we're in a special state
	} else {
		// Only check for existing initialization if NOT in empty mode

		// Check if already initialized (check for profiles directory)
		homeDir, _ := os.UserHomeDir()
		profilesDir := filepath.Join(homeDir, ".claude", "profiles")
		if _, err := os.Stat(profilesDir); err == nil {
			// Profiles directory exists, which means cc-switch has been set up before
			uiProvider.ShowAlreadyInitialized()
			return nil
		}

		// Check if Claude settings exist (normal initialization check)
		if configHandler.IsConfigInitialized() {
			uiProvider.ShowAlreadyInitialized()
			return nil
		}
	}

	// Show welcome message
	uiProvider.ShowInitWelcome()

	// Get user input for authentication token
	authToken, err := uiProvider.GetInitInput(
		"ANTHROPIC_AUTH_TOKEN",
		"Enter your Anthropic API token (leave empty if not available)",
	)
	if err != nil {
		return fmt.Errorf("failed to get auth token input: %w", err)
	}

	// Get user input for base URL
	baseURL, err := uiProvider.GetInitInput(
		"ANTHROPIC_BASE_URL",
		"Enter custom API base URL (leave empty for default)",
	)
	if err != nil {
		return fmt.Errorf("failed to get base URL input: %w", err)
	}

	// Perform initialization
	fmt.Println("\nCreating configuration...")

	if err := configHandler.InitializeConfig(authToken, baseURL); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Show success message
	uiProvider.ShowInitSuccess()

	return nil
}

package cmd

import (
	"fmt"

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

	// Check if already initialized
	if configHandler.IsConfigInitialized() {
		uiProvider.ShowAlreadyInitialized()
		return nil
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
	fmt.Println()
	fmt.Println("Creating configuration...")

	if err := configHandler.InitializeConfig(authToken, baseURL); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Show success message
	uiProvider.ShowInitSuccess()

	return nil
}
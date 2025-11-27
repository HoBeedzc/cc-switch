package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	uninstallFull   bool
	uninstallYes    bool
	uninstallDryRun bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall cc-switch",
	Long: `Uninstall cc-switch and optionally remove all configurations.

By default, this command will:
  - Remove cc-switch binary
  - Remove internal state files (.current, .history, etc.)
  - Remove templates directory
  - Keep your configuration profiles (~/.claude/profiles/*.json)
  - Keep current settings.json (Claude Code will continue working)

Use --full to also remove all configuration profiles (but still keeps settings.json).`,
	RunE: runUninstall,
}

func init() {
	uninstallCmd.Flags().BoolVarP(&uninstallFull, "full", "f", false, "Remove all configurations (profiles)")
	uninstallCmd.Flags().BoolVarP(&uninstallYes, "yes", "y", false, "Skip confirmation prompt")
	uninstallCmd.Flags().BoolVar(&uninstallDryRun, "dry-run", false, "Preview what would be removed without actually removing anything")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	// Skip update notice for uninstall command
	skipUpdateNotice = true

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	ccSwitchBinDir := filepath.Join(claudeDir, "cc-switch")
	profilesDir := filepath.Join(claudeDir, "profiles")

	// Detect installation method
	installMethod := detectInstallMethod()

	// Show confirmation
	if !uninstallYes || uninstallDryRun {
		fmt.Println()
		if uninstallDryRun {
			fmt.Println("🔍 [DRY RUN] Uninstall CC-Switch Preview")
		} else {
			fmt.Println("🗑️  Uninstall CC-Switch")
		}
		fmt.Println()
		fmt.Println("This will:")
		fmt.Printf("  ✓ Remove cc-switch binary (%s)\n", ccSwitchBinDir)
		fmt.Println("  ✓ Remove internal state files (.current, .history, etc.)")
		fmt.Println("  ✓ Remove templates directory")

		if uninstallFull {
			fmt.Printf("  ✓ Remove all configuration profiles (%s/*.json)\n", profilesDir)
		} else {
			fmt.Printf("  ✗ Keep your configuration profiles (%s/*.json)\n", profilesDir)
		}

		fmt.Println("  ✗ Keep current settings.json (Claude Code will continue working)")
		fmt.Println()

		if installMethod == "npm" {
			fmt.Println("Detected: npm installation")
			fmt.Println("Will run: npm uninstall -g @hobeeliu/cc-switch")
			fmt.Println()
		}

		// In dry run mode, skip confirmation and just show preview
		if uninstallDryRun {
			fmt.Println("🔍 [DRY RUN] No changes will be made.")
			return nil
		}

		fmt.Print("Are you sure? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Uninstall cancelled.")
			return nil
		}
	}

	fmt.Println()
	fmt.Println("🧹 Cleaning up...")

	// Step 1: Clean up configuration files
	if err := cleanupConfigFiles(profilesDir, uninstallFull); err != nil {
		fmt.Printf("  ⚠ Warning: %v\n", err)
	}

	// Step 2: Uninstall based on installation method
	if installMethod == "npm" {
		// For npm installation, run npm uninstall
		// This will trigger preuninstall.js which handles binary cleanup
		fmt.Println("  ✓ Detected npm installation")
		fmt.Println("  → Running: npm uninstall -g @hobeeliu/cc-switch")
		fmt.Println()

		if err := uninstallNpm(); err != nil {
			return fmt.Errorf("npm uninstall failed: %w", err)
		}
	} else {
		// For direct binary installation, delete the binary
		if err := cleanupBinary(ccSwitchBinDir); err != nil {
			fmt.Printf("  ⚠ Warning: %v\n", err)
		}
	}

	fmt.Println()
	fmt.Println("✨ CC-Switch has been uninstalled!")
	fmt.Println()

	// Show remaining data info
	if !uninstallFull {
		if _, err := os.Stat(profilesDir); err == nil {
			// Check if there are any remaining profile files
			entries, _ := os.ReadDir(profilesDir)
			var profiles []string
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
					profiles = append(profiles, entry.Name())
				}
			}

			if len(profiles) > 0 {
				fmt.Println("Your Claude Code configurations are preserved at:")
				fmt.Printf("  %s\n", profilesDir)
				fmt.Printf("  Files: %s\n", strings.Join(profiles, ", "))
				fmt.Println()
				fmt.Println("To completely remove all data:")
				if runtime.GOOS == "windows" {
					fmt.Printf("  rmdir /s /q \"%s\"\n", profilesDir)
				} else {
					fmt.Printf("  rm -rf \"%s\"\n", profilesDir)
				}
			}
		}
	}

	return nil
}

// detectInstallMethod detects whether cc-switch was installed via npm or directly
func detectInstallMethod() string {
	// Method 1: Check current executable path
	execPath, err := os.Executable()
	if err == nil {
		// Resolve symlinks
		execPath, _ = filepath.EvalSymlinks(execPath)

		// If path contains node_modules, it's npm installation
		if strings.Contains(execPath, "node_modules") {
			return "npm"
		}
	}

	// Method 2: Check if npm package is installed globally
	cmd := exec.Command("npm", "list", "-g", "@hobeeliu/cc-switch", "--depth=0")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "@hobeeliu/cc-switch") {
		return "npm"
	}

	return "binary"
}

// cleanupConfigFiles removes cc-switch internal files and optionally all profiles
func cleanupConfigFiles(profilesDir string, full bool) error {
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		fmt.Println("  ✓ No profiles directory found")
		return nil
	}

	if full {
		// Full cleanup: remove entire profiles directory
		if err := os.RemoveAll(profilesDir); err != nil {
			return fmt.Errorf("failed to remove profiles directory: %w", err)
		}
		fmt.Printf("  ✓ Removed profiles directory: %s\n", profilesDir)
	} else {
		// Partial cleanup: only remove internal files
		internalFiles := []string{
			".current",
			".history",
			".empty_mode",
			".empty_backup_settings.json",
			".update_check",
		}

		for _, file := range internalFiles {
			filePath := filepath.Join(profilesDir, file)
			if err := os.Remove(filePath); err == nil {
				fmt.Printf("  ✓ Removed: %s\n", file)
			}
		}

		// Remove templates directory
		templatesDir := filepath.Join(profilesDir, "templates")
		if _, err := os.Stat(templatesDir); err == nil {
			if err := os.RemoveAll(templatesDir); err == nil {
				fmt.Println("  ✓ Removed templates directory")
			}
		}

		// Try to remove profiles directory if empty
		entries, _ := os.ReadDir(profilesDir)
		if len(entries) == 0 {
			os.Remove(profilesDir)
			fmt.Println("  ✓ Removed empty profiles directory")
		}
	}

	return nil
}

// cleanupBinary removes the cc-switch binary
func cleanupBinary(binDir string) error {
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		fmt.Println("  ✓ Binary directory not found (already removed)")
		return nil
	}

	if runtime.GOOS == "windows" {
		// Windows: Create a delayed cleanup script
		return createWindowsCleanupScript(binDir)
	}

	// macOS/Linux: Can delete directly
	if err := os.RemoveAll(binDir); err != nil {
		return fmt.Errorf("failed to remove binary directory: %w", err)
	}

	fmt.Printf("  ✓ Removed binary directory: %s\n", binDir)
	return nil
}

// createWindowsCleanupScript creates a batch script to delete files after process exits
func createWindowsCleanupScript(binDir string) error {
	binaryPath := filepath.Join(binDir, "cc-switch.exe")

	// Create cleanup script in temp directory
	tempDir := os.TempDir()
	scriptPath := filepath.Join(tempDir, "cc-switch-cleanup.bat")

	script := fmt.Sprintf(`@echo off
:loop
timeout /t 1 /nobreak >nul
del "%s" 2>nul
if exist "%s" goto loop
rmdir /s /q "%s" 2>nul
del "%%~f0"
`, binaryPath, binaryPath, binDir)

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to create cleanup script: %w", err)
	}

	// Run the script in background
	cmd := exec.Command("cmd", "/c", "start", "/b", scriptPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start cleanup script: %w", err)
	}

	fmt.Printf("  ✓ Scheduled binary removal: %s\n", binDir)
	fmt.Println("    (will complete after this process exits)")

	return nil
}

// uninstallNpm runs npm uninstall -g @hobeeliu/cc-switch
func uninstallNpm() error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "npm", "uninstall", "-g", "@hobeeliu/cc-switch")
	} else {
		cmd = exec.Command("npm", "uninstall", "-g", "@hobeeliu/cc-switch")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}


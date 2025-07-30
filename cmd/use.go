package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/interactive"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Switch to a configuration",
	Long: `Switch to the specified configuration. This will replace the current Claude Code settings.

Modes:
- Interactive: cc-switch use (no arguments) or cc-switch use -i
- CLI: cc-switch use <name>

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
			return handleInteractiveUse(cm)
		case interactive.CLI:
			return handleCLIUse(cm, args[0])
		}

		return nil
	},
}

func init() {
	useCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}
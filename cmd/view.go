package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/interactive"

	"github.com/spf13/cobra"
)

var (
	rawOutput bool
)

var viewCmd = &cobra.Command{
	Use:   "view [name]",
	Short: "View configuration content",
	Long: `Display the content and metadata of a specified configuration.

Modes:
- Interactive: cc-switch view (no arguments) or cc-switch view -i
- CLI: cc-switch view <name>

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
			raw, _ := cmd.Flags().GetBool("raw")
			return handleInteractiveView(cm, raw)
		case interactive.CLI:
			raw, _ := cmd.Flags().GetBool("raw")
			return handleCLIView(cm, args[0], raw)
		}

		return nil
	},
}

func init() {
	viewCmd.Flags().BoolVar(&rawOutput, "raw", false, "Output raw JSON without metadata")
	viewCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}

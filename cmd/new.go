package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/ui"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	newTemplate     string
	newInteractive  bool
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new configuration",
	Long: `Create a new configuration with template structure ready for customization.

You can specify a template to use when creating the configuration:
- Default template: cc-switch new <name>
- Specific template: cc-switch new <name> -t <template> or cc-switch new <name> --template <template>
- Interactive mode: cc-switch new <name> -i or cc-switch new <name> --interactive

In interactive mode, cc-switch will prompt you to fill in any empty fields in the template.
If the specified template does not exist, the default template will be used.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing required argument: configuration name")
		}
		if len(args) > 1 {
			return fmt.Errorf("too many arguments provided")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := checkClaudeConfig(); err != nil {
			return err
		}

		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		// 检查配置是否已存在
		if cm.ProfileExists(name) {
			return fmt.Errorf("configuration '%s' already exists", name)
		}

		// 获取模板名称
		templateName, _ := cmd.Flags().GetString("template")
		if templateName == "" {
			templateName = "default"
		}

		// 检查模板是否存在
		if !cm.TemplateExists(templateName) {
			if templateName != "default" {
				color.Yellow("Warning: template '%s' not found, using default template", templateName)
				templateName = "default"
			}
		}

		// 根据是否启用交互模式选择创建方法
		if newInteractive {
			// 初始化UI提供者
			var uiProvider ui.UIProvider
			if isInteractiveMode() {
				uiProvider = ui.NewInteractiveUI()
			} else {
				uiProvider = ui.NewCLIUI()
			}

			// 使用交互式创建
			if err := cm.CreateProfileFromTemplateInteractive(name, templateName, uiProvider); err != nil {
				return err
			}
		} else {
			// 使用传统创建方法
			if err := cm.CreateProfileFromTemplate(name, templateName); err != nil {
				return err
			}
		}

		color.Green("✓ Configuration '%s' created successfully from template '%s'", name, templateName)
		if newInteractive {
			fmt.Printf("All template fields have been filled with your input.\n")
		} else {
			fmt.Printf("Use 'cc-switch edit %s' to customize the configuration.\n", name)
		}
		fmt.Printf("Use 'cc-switch use %s' to switch to this configuration.\n", name)

		return nil
	},
}

// isInteractiveMode checks if we should use interactive UI
func isInteractiveMode() bool {
	// Check if we're in a TTY and interactive flag is set
	return newInteractive
}

func init() {
	newCmd.Flags().StringVarP(&newTemplate, "template", "t", "", "Template to use for new configuration (default: default)")
	newCmd.Flags().BoolVarP(&newInteractive, "interactive", "i", false, "Interactive template field input mode")
}

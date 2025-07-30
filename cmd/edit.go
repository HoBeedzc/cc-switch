package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cc-switch/internal/config"
	"cc-switch/internal/interactive"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	editField string
	useNano   bool
)

var editCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit configuration content",
	Long: `Edit the content of a specified configuration using your system editor.

Modes:
- Interactive: cc-switch edit (no arguments) or cc-switch edit -i
- CLI: cc-switch edit <name>

The interactive mode allows you to browse and select configurations with arrow keys.

The editor is determined by priority:
1. --nano flag uses nano editor
2. EDITOR environment variable 
3. Default to vim editor

Changes are validated for JSON syntax before saving.`,
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
			return handleInteractiveEdit(cm)
		case interactive.CLI:
			return handleCLIEdit(cm, args[0])
		}

		return nil
	},
}

// editProfileWithEditor 使用系统编辑器编辑配置
func editProfileWithEditor(cm *config.ConfigManager, name string) error {
	// 获取当前配置内容
	content, _, err := cm.GetProfileContent(name)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("cc-switch-%s-*.json", name))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 写入当前内容到临时文件
	jsonData, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	if _, err := tmpFile.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}
	tmpFile.Close()

	// 获取文件修改时间（用于检测是否有更改）
	stat, err := os.Stat(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}
	originalModTime := stat.ModTime()

	// 确定编辑器
	editor := getEditor(useNano)

	// 启动编辑器
	fmt.Printf("Opening configuration '%s' in %s...\n", name, editor)
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// 检测是否有更改
	stat, err = os.Stat(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to get file stats after editing: %w", err)
	}

	if stat.ModTime().Equal(originalModTime) {
		fmt.Println("No changes detected.")
		return nil
	}

	// 读取编辑后的内容
	editedData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// 验证JSON格式
	var editedContent map[string]interface{}
	if err := json.Unmarshal(editedData, &editedContent); err != nil {
		color.Red("Error: Invalid JSON format")
		fmt.Printf("JSON validation error: %v\n", err)
		
		// 询问用户是否要重新编辑
		fmt.Print("Do you want to re-edit the file? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			return editProfileWithEditor(cm, name)
		}
		return fmt.Errorf("configuration not saved due to invalid JSON")
	}

	// 保存更改
	if err := cm.UpdateProfile(name, editedContent); err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	interactive.ShowSuccess("Configuration '%s' updated successfully", name)
	return nil
}

// editProfileField 编辑配置的特定字段
func editProfileField(cm *config.ConfigManager, name, field string) error {
	content, _, err := cm.GetProfileContent(name)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	// 解析字段路径（支持嵌套字段，如 "env.ANTHROPIC_API_KEY"）
	fieldParts := strings.Split(field, ".")
	
	// 显示当前值
	currentValue := getNestedValue(content, fieldParts)
	fmt.Printf("Current value of '%s': ", field)
	if currentValue != nil {
		currentValueJson, _ := json.Marshal(currentValue)
		fmt.Println(string(currentValueJson))
	} else {
		fmt.Println("<not set>")
	}

	// 获取新值
	fmt.Print("Enter new value (JSON format): ")
	var newValueStr string
	fmt.Scanln(&newValueStr)

	if newValueStr == "" {
		fmt.Println("No changes made.")
		return nil
	}

	// 解析新值
	var newValue interface{}
	if err := json.Unmarshal([]byte(newValueStr), &newValue); err != nil {
		return fmt.Errorf("invalid JSON format for new value: %w", err)
	}

	// 设置新值
	if err := setNestedValue(content, fieldParts, newValue); err != nil {
		return fmt.Errorf("failed to set field value: %w", err)
	}

	// 保存更改
	if err := cm.UpdateProfile(name, content); err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	interactive.ShowSuccess("Field '%s' updated successfully in configuration '%s'", field, name)
	return nil
}

// getNestedValue 获取嵌套字段的值
func getNestedValue(data map[string]interface{}, fieldParts []string) interface{} {
	current := data
	for i, part := range fieldParts {
		if i == len(fieldParts)-1 {
			return current[part]
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	return nil
}

// setNestedValue 设置嵌套字段的值
func setNestedValue(data map[string]interface{}, fieldParts []string, value interface{}) error {
	current := data
	for i, part := range fieldParts {
		if i == len(fieldParts)-1 {
			current[part] = value
			return nil
		}
		if _, ok := current[part]; !ok {
			current[part] = make(map[string]interface{})
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return fmt.Errorf("cannot set nested field: intermediate field '%s' is not an object", part)
		}
	}
	return nil
}

func init() {
	editCmd.Flags().StringVar(&editField, "field", "", "Edit a specific field (e.g., 'env.ANTHROPIC_API_KEY')")
	editCmd.Flags().BoolVar(&useNano, "nano", false, "Use nano editor instead of default")
	editCmd.Flags().BoolP("interactive", "i", false, "Enter interactive mode")
}

// getEditor 根据优先级确定使用的编辑器
func getEditor(useNano bool) string {
	// 1. --nano 标志优先级最高
	if useNano {
		return "nano"
	}
	
	// 2. EDITOR 环境变量
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	
	// 3. 默认使用 vim
	return "vim"
}
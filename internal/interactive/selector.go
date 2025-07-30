package interactive

import (
	"fmt"
	"strings"

	"cc-switch/internal/config"

	"github.com/manifoldco/promptui"
	"github.com/fatih/color"
)

// ExecutionMode 定义执行模式
type ExecutionMode int

const (
	CLI ExecutionMode = iota
	Interactive
)

// InteractiveSelector 交互式选择器
type InteractiveSelector struct {
	configs []config.Profile
	action  string // "delete", "edit", "view", "use"
}

// NewInteractiveSelector 创建新的交互式选择器
func NewInteractiveSelector(configs []config.Profile, action string) *InteractiveSelector {
	return &InteractiveSelector{
		configs: configs,
		action:  action,
	}
}

// Show 显示选择界面
func (s *InteractiveSelector) Show() (*config.Profile, error) {
	if len(s.configs) == 0 {
		return nil, fmt.Errorf("no configurations available")
	}

	// 自定义模板
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "▶ {{ .Name | cyan }}{{ if .IsCurrent }} {{ \"(current)\" | green }}{{ end }}",
		Inactive: "  {{ .Name }}{{ if .IsCurrent }} {{ \"(current)\" | faint }}{{ end }}",
		Selected: "✓ {{ .Name | green }}{{ if .IsCurrent }} {{ \"(current)\" | faint }}{{ end }}",
		Details: `
--------- Configuration Details ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Status:" | faint }}	{{ if .IsCurrent }}{{ "Current" | green }}{{ else }}{{ "Available" | yellow }}{{ end }}
{{ "Path:" | faint }}	{{ .Path }}`,
	}

	prompt := promptui.Select{
		Label:        fmt.Sprintf("Select configuration to %s", s.action),
		Items:        s.configs,
		Templates:    templates,
		Size:         10,
		HideSelected: false,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &s.configs[i], nil
}

// ActionMenu 操作菜单
type ActionMenu struct {
	selectedConfig *config.Profile
}

// NewActionMenu 创建新的操作菜单
func NewActionMenu(selectedConfig *config.Profile) *ActionMenu {
	return &ActionMenu{selectedConfig: selectedConfig}
}

// ShowActions 显示操作选择菜单
func (m *ActionMenu) ShowActions() (string, error) {
	actions := []string{"View", "Edit", "Use", "Delete", "Cancel"}
	
	// 如果是当前配置，不能删除
	if m.selectedConfig.IsCurrent {
		actions = []string{"View", "Edit", "Cancel"}
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "▶ {{ . | cyan }}",
		Inactive: "  {{ . }}",
		Selected: "✓ {{ . | green }}",
	}

	prompt := promptui.Select{
		Label:     fmt.Sprintf("What do you want to do with '%s'?", m.selectedConfig.Name),
		Items:     actions,
		Templates: templates,
		Size:      len(actions),
	}
	
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	
	return strings.ToLower(result), nil
}

// DetectMode 检测执行模式
func DetectMode(hasInteractiveFlag bool, args []string) ExecutionMode {
	// 显式交互标志
	if hasInteractiveFlag {
		return Interactive
	}
	
	// 无参数自动进入交互模式
	if len(args) == 0 {
		return Interactive
	}
	
	return CLI
}

// ConfirmAction 确认操作
func ConfirmAction(message string, defaultValue bool) bool {
	var defaultStr string
	if defaultValue {
		defaultStr = "Y/n"
	} else {
		defaultStr = "y/N"
	}

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("%s (%s)", message, defaultStr),
		IsConfirm: true,
		Default:   fmt.Sprintf("%t", defaultValue),
	}

	result, err := prompt.Run()
	if err != nil {
		return false
	}

	result = strings.ToLower(strings.TrimSpace(result))
	if result == "" {
		return defaultValue
	}
	
	return result == "y" || result == "yes" || result == "true"
}

// ShowError 显示错误信息
func ShowError(err error) {
	color.Red("Error: %v", err)
}

// ShowSuccess 显示成功信息
func ShowSuccess(message string, args ...interface{}) {
	color.Green("✓ "+message, args...)
}

// ShowWarning 显示警告信息
func ShowWarning(message string, args ...interface{}) {
	color.Yellow("⚠ "+message, args...)
}

// ShowInfo 显示信息
func ShowInfo(message string, args ...interface{}) {
	color.Cyan("ℹ "+message, args...)
}
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConfigManager 管理Claude配置切换
type ConfigManager struct {
	claudeDir       string
	profilesDir     string
	templatesDir    string
	currentFile     string
	settingsFile    string
	historyFile     string
	emptyModeFile   string
}

// Profile 配置文件信息
type Profile struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"is_current"`
	Path      string `json:"path"`
}

// ConfigHistory 配置历史记录
type ConfigHistory struct {
	Current   string    `json:"current"`
	Previous  string    `json:"previous"`
	History   []string  `json:"history"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EmptyModeInfo 空配置模式信息
type EmptyModeInfo struct {
	Enabled         bool      `json:"enabled"`
	BackupPath      string    `json:"backup_path"`
	PreviousProfile string    `json:"previous_profile"`
	Timestamp       time.Time `json:"timestamp"`
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager() (*ConfigManager, error) {
	cm, err := NewConfigManagerNoInit()
	if err != nil {
		return nil, err
	}

	// 初始化检查
	if err := cm.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}

	return cm, nil
}

// NewConfigManagerNoInit 创建配置管理器但不执行初始化（用于init命令）
func NewConfigManagerNoInit() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	profilesDir := filepath.Join(claudeDir, "profiles")
	templatesDir := filepath.Join(profilesDir, "templates")
	currentFile := filepath.Join(claudeDir, ".current")
	settingsFile := filepath.Join(claudeDir, "settings.json")
	historyFile := filepath.Join(profilesDir, ".history")
	emptyModeFile := filepath.Join(claudeDir, ".empty_mode")

	cm := &ConfigManager{
		claudeDir:       claudeDir,
		profilesDir:     profilesDir,
		templatesDir:    templatesDir,
		currentFile:     currentFile,
		settingsFile:    settingsFile,
		historyFile:     historyFile,
		emptyModeFile:   emptyModeFile,
	}

	return cm, nil
}

// Initialize 初始化配置目录和默认配置
func (cm *ConfigManager) Initialize() error {
	// 创建profiles目录
	if err := os.MkdirAll(cm.profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// 创建templates目录
	if err := os.MkdirAll(cm.templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	// 初始化默认模板
	if err := cm.initializeDefaultTemplate(); err != nil {
		return fmt.Errorf("failed to initialize default template: %w", err)
	}

	// 检查settings.json是否存在
	if _, err := os.Stat(cm.settingsFile); err == nil {
		// 存在settings.json，检查是否已经有default配置
		defaultProfilePath := filepath.Join(cm.profilesDir, "default.json")
		if _, err := os.Stat(defaultProfilePath); os.IsNotExist(err) {
			// 创建default配置
			if err := cm.copyFile(cm.settingsFile, defaultProfilePath); err != nil {
				return fmt.Errorf("failed to create default profile: %w", err)
			}

			// 设置权限
			if err := os.Chmod(defaultProfilePath, 0600); err != nil {
				return fmt.Errorf("failed to set profile permissions: %w", err)
			}
		}

		// 设置当前配置为default（如果.current文件不存在）
		if _, err := os.Stat(cm.currentFile); os.IsNotExist(err) {
			if err := cm.setCurrentProfile("default"); err != nil {
				return fmt.Errorf("failed to set current profile: %w", err)
			}
		}
	}

	return nil
}

// ListProfiles 列出所有配置
func (cm *ConfigManager) ListProfiles() ([]Profile, error) {
	entries, err := os.ReadDir(cm.profilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	currentProfile, _ := cm.getCurrentProfile()
	var profiles []Profile

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".json")
		profiles = append(profiles, Profile{
			Name:      name,
			IsCurrent: name == currentProfile,
			Path:      filepath.Join(cm.profilesDir, entry.Name()),
		})
	}

	return profiles, nil
}

// CreateProfile 创建新配置（从模板）
func (cm *ConfigManager) CreateProfile(name string) error {
	return cm.CreateProfileFromTemplate(name, "default")
}

// CreateProfileFromTemplate 从指定模板创建新配置
func (cm *ConfigManager) CreateProfileFromTemplate(name, templateName string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// 检查配置是否已存在
	profilePath := filepath.Join(cm.profilesDir, name+".json")
	if _, err := os.Stat(profilePath); err == nil {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	// 检查模板是否存在
	templatePath := filepath.Join(cm.templatesDir, templateName+".json")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' does not exist", templateName)
	}

	// 从模板复制创建配置
	if err := cm.copyFile(templatePath, profilePath); err != nil {
		return fmt.Errorf("failed to create profile from template: %w", err)
	}

	return nil
}

// UseProfile 切换到指定配置
func (cm *ConfigManager) UseProfile(name string) error {
	profilePath := filepath.Join(cm.profilesDir, name+".json")

	// 检查配置是否存在
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// 备份当前配置到profiles中（如果有的话）
	currentProfile, err := cm.getCurrentProfile()
	if err == nil && currentProfile != "" {
		currentProfilePath := filepath.Join(cm.profilesDir, currentProfile+".json")
		if err := cm.copyFile(cm.settingsFile, currentProfilePath); err != nil {
			return fmt.Errorf("failed to backup current profile: %w", err)
		}
	}

	// 原子性操作：使用临时文件
	tempFile := cm.settingsFile + ".tmp"
	if err := cm.copyFile(profilePath, tempFile); err != nil {
		return fmt.Errorf("failed to prepare new settings: %w", err)
	}

	// 原子性重命名
	if err := os.Rename(tempFile, cm.settingsFile); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to switch profile: %w", err)
	}

	// 更新当前配置标记
	if err := cm.setCurrentProfile(name); err != nil {
		return fmt.Errorf("failed to update current profile marker: %w", err)
	}

	// 更新历史记录
	if err := cm.updateHistory(name); err != nil {
		// 历史记录更新失败不应该阻止配置切换，只记录错误
		fmt.Fprintf(os.Stderr, "Warning: failed to update history: %v\n", err)
	}

	return nil
}

// DeleteProfile 删除配置
func (cm *ConfigManager) DeleteProfile(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// 检查是否为当前配置
	currentProfile, _ := cm.getCurrentProfile()
	if name == currentProfile {
		return fmt.Errorf("cannot delete current profile '%s'. Switch to another profile first", name)
	}

	profilePath := filepath.Join(cm.profilesDir, name+".json")

	// 检查配置是否存在
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// 删除配置文件
	if err := os.Remove(profilePath); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}

// GetCurrentProfile 获取当前配置名
func (cm *ConfigManager) GetCurrentProfile() (string, error) {
	return cm.getCurrentProfile()
}

// 私有方法

// getCurrentProfile 读取当前配置名
func (cm *ConfigManager) getCurrentProfile() (string, error) {
	data, err := os.ReadFile(cm.currentFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// setCurrentProfile 设置当前配置名
func (cm *ConfigManager) setCurrentProfile(name string) error {
	return os.WriteFile(cm.currentFile, []byte(name), 0644)
}

// copyFile 复制文件
func (cm *ConfigManager) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// 验证JSON格式
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid JSON format in source file: %w", err)
	}

	return os.WriteFile(dst, data, 0600)
}

// ProfileExists 检查配置是否存在
func (cm *ConfigManager) ProfileExists(name string) bool {
	profilePath := filepath.Join(cm.profilesDir, name+".json")
	_, err := os.Stat(profilePath)
	return err == nil
}

// GetProfileContent 获取配置内容和元数据
func (cm *ConfigManager) GetProfileContent(name string) (map[string]interface{}, Profile, error) {
	profilePath := filepath.Join(cm.profilesDir, name+".json")

	// 检查配置是否存在
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return nil, Profile{}, fmt.Errorf("profile '%s' does not exist", name)
	}

	// 读取配置文件
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, Profile{}, fmt.Errorf("failed to read profile file: %w", err)
	}

	// 解析JSON内容
	var content map[string]interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, Profile{}, fmt.Errorf("failed to parse JSON content: %w", err)
	}

	// 创建元数据
	currentProfile, _ := cm.getCurrentProfile()
	metadata := Profile{
		Name:      name,
		IsCurrent: name == currentProfile,
		Path:      profilePath,
	}

	return content, metadata, nil
}

// UpdateProfile 更新配置内容
func (cm *ConfigManager) UpdateProfile(name string, content map[string]interface{}) error {
	profilePath := filepath.Join(cm.profilesDir, name+".json")

	// 检查配置是否存在
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// 验证JSON内容
	if err := cm.validateProfileContent(content); err != nil {
		return fmt.Errorf("invalid profile content: %w", err)
	}

	// 创建备份
	backupPath := profilePath + ".backup"
	if err := cm.copyFile(profilePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// 序列化新内容
	jsonData, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize JSON: %w", err)
	}

	// 原子性写入
	tempFile := profilePath + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// 原子性重命名
	if err := os.Rename(tempFile, profilePath); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to update profile: %w", err)
	}

	// 如果是当前配置，同时更新settings.json
	currentProfile, _ := cm.getCurrentProfile()
	if name == currentProfile {
		if err := os.WriteFile(cm.settingsFile, jsonData, 0600); err != nil {
			return fmt.Errorf("failed to sync current settings: %w", err)
		}
	}

	// 清理备份文件（更新成功后）
	os.Remove(backupPath)

	return nil
}

// validateProfileContent 验证配置内容
func (cm *ConfigManager) validateProfileContent(content map[string]interface{}) error {
	// 基本JSON格式验证（通过能够unmarshal已经验证）

	// 检查必要字段（可根据需要扩展）
	if content == nil {
		return fmt.Errorf("content cannot be nil")
	}

	// 验证是否可以序列化
	if _, err := json.Marshal(content); err != nil {
		return fmt.Errorf("content cannot be serialized to JSON: %w", err)
	}

	return nil
}

// RenameProfile 重命名配置文件
func (cm *ConfigManager) RenameProfile(oldName, newName string) error {
	if oldName == "" || newName == "" {
		return fmt.Errorf("profile names cannot be empty")
	}

	if oldName == newName {
		return fmt.Errorf("old and new names cannot be the same")
	}

	oldPath := filepath.Join(cm.profilesDir, oldName+".json")
	newPath := filepath.Join(cm.profilesDir, newName+".json")

	// 检查源配置是否存在
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", oldName)
	}

	// 检查目标名称是否已存在
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("profile '%s' already exists", newName)
	}

	// 执行重命名
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename profile: %w", err)
	}

	// 如果重命名的是当前配置，更新当前配置指向
	currentProfile, _ := cm.getCurrentProfile()
	if oldName == currentProfile {
		if err := cm.setCurrentProfile(newName); err != nil {
			// 如果更新当前配置失败，尝试回滚重命名操作
			os.Rename(newPath, oldPath)
			return fmt.Errorf("failed to update current profile marker: %w", err)
		}
	}

	return nil
}

// CopyProfile 复制配置文件
func (cm *ConfigManager) CopyProfile(sourceName, destName string) error {
	if sourceName == "" || destName == "" {
		return fmt.Errorf("profile names cannot be empty")
	}

	if sourceName == destName {
		return fmt.Errorf("source and destination names cannot be the same")
	}

	sourcePath := filepath.Join(cm.profilesDir, sourceName+".json")
	destPath := filepath.Join(cm.profilesDir, destName+".json")

	// 检查源配置是否存在
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", sourceName)
	}

	// 检查目标名称是否已存在
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("profile '%s' already exists", destName)
	}

	// 执行复制
	if err := cm.copyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy profile: %w", err)
	}

	return nil
}

// SetCurrentProfile 公开设置当前配置的方法
func (cm *ConfigManager) SetCurrentProfile(name string) error {
	// 检查配置是否存在
	profilePath := filepath.Join(cm.profilesDir, name+".json")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	return cm.setCurrentProfile(name)
}

// GetPreviousProfile 获取上一个配置名称
func (cm *ConfigManager) GetPreviousProfile() (string, error) {
	history, err := cm.loadHistory()
	if err != nil {
		return "", fmt.Errorf("failed to load history: %w", err)
	}

	if history.Previous == "" {
		return "", fmt.Errorf("no previous configuration available")
	}

	// Special case: "empty_mode" is a virtual state, not a real profile file
	if history.Previous == "empty_mode" {
		return history.Previous, nil
	}

	// 检查上一个配置是否仍然存在
	if !cm.ProfileExists(history.Previous) {
		// 清理无效的历史记录
		cm.cleanupHistory()
		return "", fmt.Errorf("previous configuration '%s' no longer exists", history.Previous)
	}

	return history.Previous, nil
}

// loadHistory 加载配置历史记录
func (cm *ConfigManager) loadHistory() (*ConfigHistory, error) {
	// 如果历史文件不存在，返回空历史记录
	if _, err := os.Stat(cm.historyFile); os.IsNotExist(err) {
		return &ConfigHistory{
			History:   make([]string, 0),
			UpdatedAt: time.Now(),
		}, nil
	}

	data, err := os.ReadFile(cm.historyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var history ConfigHistory
	if err := json.Unmarshal(data, &history); err != nil {
		// 如果解析失败，返回新的空历史记录
		return &ConfigHistory{
			History:   make([]string, 0),
			UpdatedAt: time.Now(),
		}, nil
	}

	return &history, nil
}

// saveHistory 保存配置历史记录
func (cm *ConfigManager) saveHistory(history *ConfigHistory) error {
	history.UpdatedAt = time.Now()

	jsonData, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// 原子性写入
	tempFile := cm.historyFile + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write temporary history file: %w", err)
	}

	if err := os.Rename(tempFile, cm.historyFile); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to save history file: %w", err)
	}

	return nil
}

// updateHistory 更新配置历史记录
func (cm *ConfigManager) updateHistory(newProfile string) error {
	history, err := cm.loadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	// 如果当前配置不为空且与新配置不同，将其设为previous
	if history.Current != "" && history.Current != newProfile {
		history.Previous = history.Current

		// 更新历史列表，保持最近5个记录
		history.History = cm.addToHistory(history.History, history.Current, 5)
	}

	history.Current = newProfile

	return cm.saveHistory(history)
}

// addToHistory 添加配置到历史列表，保持指定数量的最新记录
func (cm *ConfigManager) addToHistory(history []string, profile string, maxSize int) []string {
	// 移除重复项
	var newHistory []string
	for _, p := range history {
		if p != profile {
			newHistory = append(newHistory, p)
		}
	}

	// 添加到开头
	newHistory = append([]string{profile}, newHistory...)

	// 限制大小
	if len(newHistory) > maxSize {
		newHistory = newHistory[:maxSize]
	}

	return newHistory
}

// cleanupHistory 清理历史记录中不存在的配置
func (cm *ConfigManager) cleanupHistory() error {
	history, err := cm.loadHistory()
	if err != nil {
		return err
	}

	// 清理previous配置
	if history.Previous != "" && !cm.ProfileExists(history.Previous) {
		history.Previous = ""
	}

	// 清理历史列表
	var validHistory []string
	for _, profile := range history.History {
		if cm.ProfileExists(profile) {
			validHistory = append(validHistory, profile)
		}
	}
	history.History = validHistory

	return cm.saveHistory(history)
}

// Template Management Methods

// initializeDefaultTemplate 初始化默认模板
func (cm *ConfigManager) initializeDefaultTemplate() error {
	defaultTemplatePath := filepath.Join(cm.templatesDir, "default.json")
	
	// 如果默认模板已存在，直接返回
	if _, err := os.Stat(defaultTemplatePath); err == nil {
		return nil
	}

	// 创建默认模板内容
	template := map[string]interface{}{
		"env": map[string]interface{}{
			"ANTHROPIC_AUTH_TOKEN": "",
			"ANTHROPIC_BASE_URL":   "",
		},
		"permissions": map[string]interface{}{
			"allow": []interface{}{},
			"deny":  []interface{}{},
		},
	}

	// 序列化模板
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create template data: %w", err)
	}

	// 写入默认模板文件
	if err := os.WriteFile(defaultTemplatePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to create default template: %w", err)
	}

	return nil
}

// ListTemplates 列出所有模板
func (cm *ConfigManager) ListTemplates() ([]string, error) {
	entries, err := os.ReadDir(cm.templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		templates = append(templates, name)
	}

	return templates, nil
}

// TemplateExists 检查模板是否存在
func (cm *ConfigManager) TemplateExists(name string) bool {
	templatePath := filepath.Join(cm.templatesDir, name+".json")
	_, err := os.Stat(templatePath)
	return err == nil
}

// CreateTemplate 创建新模板
func (cm *ConfigManager) CreateTemplate(name string) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	templatePath := filepath.Join(cm.templatesDir, name+".json")
	
	// 检查模板是否已存在
	if _, err := os.Stat(templatePath); err == nil {
		return fmt.Errorf("template '%s' already exists", name)
	}

	// 创建空模板内容（基于默认模板）
	template := map[string]interface{}{
		"env": map[string]interface{}{
			"ANTHROPIC_AUTH_TOKEN": "",
			"ANTHROPIC_BASE_URL":   "",
		},
		"permissions": map[string]interface{}{
			"allow": []interface{}{},
			"deny":  []interface{}{},
		},
	}

	// 序列化模板
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(templatePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

// GetTemplateContent 获取模板内容
func (cm *ConfigManager) GetTemplateContent(name string) (map[string]interface{}, error) {
	templatePath := filepath.Join(cm.templatesDir, name+".json")

	// 检查模板是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template '%s' does not exist", name)
	}

	// 读取模板文件
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	// 解析JSON内容
	var content map[string]interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("failed to parse JSON content: %w", err)
	}

	return content, nil
}

// UpdateTemplate 更新模板内容
func (cm *ConfigManager) UpdateTemplate(name string, content map[string]interface{}) error {
	templatePath := filepath.Join(cm.templatesDir, name+".json")

	// 检查模板是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' does not exist", name)
	}

	// 验证JSON内容
	if err := cm.validateTemplateContent(content); err != nil {
		return fmt.Errorf("invalid template content: %w", err)
	}

	// 创建备份
	backupPath := templatePath + ".backup"
	if err := cm.copyFile(templatePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// 序列化新内容
	jsonData, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize JSON: %w", err)
	}

	// 原子性写入
	tempFile := templatePath + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// 原子性重命名
	if err := os.Rename(tempFile, templatePath); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to update template: %w", err)
	}

	// 清理备份文件（更新成功后）
	os.Remove(backupPath)

	return nil
}

// DeleteTemplate 删除模板
func (cm *ConfigManager) DeleteTemplate(name string) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if name == "default" {
		return fmt.Errorf("cannot delete default template")
	}

	templatePath := filepath.Join(cm.templatesDir, name+".json")

	// 检查模板是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' does not exist", name)
	}

	// 删除模板文件
	if err := os.Remove(templatePath); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// validateTemplateContent 验证模板内容
func (cm *ConfigManager) validateTemplateContent(content map[string]interface{}) error {
	// 基本JSON格式验证（通过能够unmarshal已经验证）
	if content == nil {
		return fmt.Errorf("content cannot be nil")
	}

	// 验证是否可以序列化
	if _, err := json.Marshal(content); err != nil {
		return fmt.Errorf("content cannot be serialized to JSON: %w", err)
	}

	return nil
}

// Init Command Support Methods

// IsInitialized 检查Claude配置是否已初始化
func (cm *ConfigManager) IsInitialized() bool {
	_, err := os.Stat(cm.settingsFile)
	return err == nil
}

// InitializeFromScratch 从零开始初始化Claude配置
func (cm *ConfigManager) InitializeFromScratch(authToken, baseURL string) error {
	// 检查是否已初始化
	if cm.IsInitialized() {
		return fmt.Errorf("configuration already exists at %s", cm.settingsFile)
	}

	// 创建初始配置内容
	initialConfig := map[string]interface{}{
		"env": map[string]interface{}{
			"ANTHROPIC_AUTH_TOKEN": authToken,
		},
		"permissions": map[string]interface{}{
			"allow": []interface{}{},
			"deny":  []interface{}{},
		},
	}

	// 如果提供了baseURL，添加到配置中
	if baseURL != "" {
		initialConfig["env"].(map[string]interface{})["ANTHROPIC_BASE_URL"] = baseURL
	}

	// 确保目录存在
	if err := os.MkdirAll(cm.claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create claude directory: %w", err)
	}

	// 创建settings.json
	if err := cm.writeConfigFile(cm.settingsFile, initialConfig); err != nil {
		return fmt.Errorf("failed to create settings file: %w", err)
	}

	// 初始化cc-switch结构
	if err := cm.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize cc-switch: %w", err)
	}

	return nil
}

// writeConfigFile 写入配置文件的辅助方法
func (cm *ConfigManager) writeConfigFile(filePath string, content map[string]interface{}) error {
	// 序列化配置
	jsonData, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize JSON: %w", err)
	}

	// 原子性写入
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// 原子性重命名
	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to create configuration file: %w", err)
	}

	return nil
}

// Empty Mode Methods

// IsEmptyMode 检查是否处于空配置模式
func (cm *ConfigManager) IsEmptyMode() bool {
	_, err := os.Stat(cm.emptyModeFile)
	return err == nil
}

// GetEmptyModeInfo 获取空配置模式信息
func (cm *ConfigManager) GetEmptyModeInfo() (*EmptyModeInfo, error) {
	if !cm.IsEmptyMode() {
		return nil, fmt.Errorf("not in empty mode")
	}

	data, err := os.ReadFile(cm.emptyModeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read empty mode file: %w", err)
	}

	var info EmptyModeInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse empty mode info: %w", err)
	}

	return &info, nil
}

// EnableEmptyMode 启用空配置模式
func (cm *ConfigManager) EnableEmptyMode() error {
	// 检查当前状态
	if cm.IsEmptyMode() {
		return fmt.Errorf("already in empty mode")
	}

	// 检查settings.json是否存在
	if _, err := os.Stat(cm.settingsFile); os.IsNotExist(err) {
		return fmt.Errorf("no settings.json found to backup")
	}

	// 创建备份路径
	backupPath := filepath.Join(cm.profilesDir, ".empty_backup_settings.json")

	// 获取当前配置名
	currentProfile, _ := cm.getCurrentProfile()

	// 步骤1: 原子性备份 settings.json
	if err := cm.copyFile(cm.settingsFile, backupPath); err != nil {
		return fmt.Errorf("failed to backup settings: %w", err)
	}

	// 步骤2: 创建状态标记
	emptyInfo := &EmptyModeInfo{
		Enabled:         true,
		BackupPath:      backupPath,
		PreviousProfile: currentProfile,
		Timestamp:       time.Now(),
	}

	// 步骤3: 保存状态（原子性）
	if err := cm.saveEmptyModeInfo(emptyInfo); err != nil {
		os.Remove(backupPath) // 清理备份
		return fmt.Errorf("failed to save empty mode info: %w", err)
	}

	// 步骤4: 更新历史记录，将进入empty mode记录为历史
	if err := cm.updateHistory("empty_mode"); err != nil {
		// 历史记录更新失败不应该阻止empty mode启用，只记录错误
		fmt.Fprintf(os.Stderr, "Warning: failed to update history: %v\n", err)
	}

	// 步骤5: 移除 settings.json（最后步骤）
	if err := os.Remove(cm.settingsFile); err != nil {
		// 回滚操作
		cm.removeEmptyModeInfo()
		os.Remove(backupPath)
		return fmt.Errorf("failed to remove settings file: %w", err)
	}

	return nil
}

// DisableEmptyMode 禁用空配置模式
func (cm *ConfigManager) DisableEmptyMode() error {
	// 检查当前状态
	if !cm.IsEmptyMode() {
		return fmt.Errorf("not in empty mode")
	}

	// 获取空配置模式信息
	emptyInfo, err := cm.GetEmptyModeInfo()
	if err != nil {
		return fmt.Errorf("failed to get empty mode info: %w", err)
	}

	// 检查备份文件是否存在
	if _, err := os.Stat(emptyInfo.BackupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", emptyInfo.BackupPath)
	}

	// 步骤1: 恢复 settings.json（原子性）
	tempFile := cm.settingsFile + ".tmp"
	if err := cm.copyFile(emptyInfo.BackupPath, tempFile); err != nil {
		return fmt.Errorf("failed to prepare settings restoration: %w", err)
	}

	// 原子性重命名
	if err := os.Rename(tempFile, cm.settingsFile); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to restore settings file: %w", err)
	}

	// 步骤2: 更新当前配置标记（如果有之前的配置）
	if emptyInfo.PreviousProfile != "" {
		if err := cm.setCurrentProfile(emptyInfo.PreviousProfile); err != nil {
			// 不是致命错误，记录警告但继续
			fmt.Fprintf(os.Stderr, "Warning: failed to set current profile marker: %v\n", err)
		}

		// 步骤3: 更新历史记录，恢复到之前的配置
		if err := cm.updateHistory(emptyInfo.PreviousProfile); err != nil {
			// 历史记录更新失败不应该阻止配置恢复，只记录错误
			fmt.Fprintf(os.Stderr, "Warning: failed to update history: %v\n", err)
		}
	}

	// 步骤4: 清理空配置模式文件
	if err := cm.removeEmptyModeInfo(); err != nil {
		return fmt.Errorf("failed to remove empty mode info: %w", err)
	}

	// 步骤5: 清理备份文件
	os.Remove(emptyInfo.BackupPath)

	return nil
}

// RestoreToPreviousProfile 恢复到之前的配置
func (cm *ConfigManager) RestoreToPreviousProfile() error {
	if !cm.IsEmptyMode() {
		return fmt.Errorf("not in empty mode")
	}

	emptyInfo, err := cm.GetEmptyModeInfo()
	if err != nil {
		return fmt.Errorf("failed to get empty mode info: %w", err)
	}

	if emptyInfo.PreviousProfile == "" {
		return fmt.Errorf("no previous profile to restore to")
	}

	// 先禁用空配置模式
	if err := cm.DisableEmptyMode(); err != nil {
		return fmt.Errorf("failed to disable empty mode: %w", err)
	}

	// 然后切换到之前的配置（如果不是已经激活的话）
	currentProfile, _ := cm.getCurrentProfile()
	if currentProfile != emptyInfo.PreviousProfile {
		if err := cm.UseProfile(emptyInfo.PreviousProfile); err != nil {
			return fmt.Errorf("failed to switch to previous profile '%s': %w", emptyInfo.PreviousProfile, err)
		}
	}

	return nil
}

// saveEmptyModeInfo 保存空配置模式信息
func (cm *ConfigManager) saveEmptyModeInfo(info *EmptyModeInfo) error {
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal empty mode info: %w", err)
	}

	// 原子性写入
	tempFile := cm.emptyModeFile + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write temporary empty mode file: %w", err)
	}

	if err := os.Rename(tempFile, cm.emptyModeFile); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to save empty mode file: %w", err)
	}

	return nil
}

// removeEmptyModeInfo 移除空配置模式信息
func (cm *ConfigManager) removeEmptyModeInfo() error {
	if err := os.Remove(cm.emptyModeFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove empty mode file: %w", err)
	}
	return nil
}

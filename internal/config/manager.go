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
	claudeDir    string
	profilesDir  string
	currentFile  string
	settingsFile string
	historyFile  string
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

// NewConfigManager 创建新的配置管理器
func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	profilesDir := filepath.Join(claudeDir, "profiles")
	currentFile := filepath.Join(claudeDir, ".current")
	settingsFile := filepath.Join(claudeDir, "settings.json")
	historyFile := filepath.Join(profilesDir, ".history")

	cm := &ConfigManager{
		claudeDir:    claudeDir,
		profilesDir:  profilesDir,
		currentFile:  currentFile,
		settingsFile: settingsFile,
		historyFile:  historyFile,
	}

	// 初始化检查
	if err := cm.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}

	return cm, nil
}

// Initialize 初始化配置目录和默认配置
func (cm *ConfigManager) Initialize() error {
	// 创建profiles目录
	if err := os.MkdirAll(cm.profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
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
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// 检查配置是否已存在
	profilePath := filepath.Join(cm.profilesDir, name+".json")
	if _, err := os.Stat(profilePath); err == nil {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	// 创建配置模板
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
	if err := os.WriteFile(profilePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
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

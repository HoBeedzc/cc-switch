package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigManager 管理Claude配置切换
type ConfigManager struct {
	claudeDir   string
	profilesDir string
	currentFile string
	settingsFile string
}

// Profile 配置文件信息
type Profile struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"is_current"`
	Path      string `json:"path"`
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

	cm := &ConfigManager{
		claudeDir:   claudeDir,
		profilesDir: profilesDir,
		currentFile: currentFile,
		settingsFile: settingsFile,
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

// CreateProfile 创建新配置
func (cm *ConfigManager) CreateProfile(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// 检查配置是否已存在
	profilePath := filepath.Join(cm.profilesDir, name+".json")
	if _, err := os.Stat(profilePath); err == nil {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	// 复制当前settings.json到新配置
	if err := cm.copyFile(cm.settingsFile, profilePath); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	// 设置文件权限
	if err := os.Chmod(profilePath, 0600); err != nil {
		return fmt.Errorf("failed to set profile permissions: %w", err)
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
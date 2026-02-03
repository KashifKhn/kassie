package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrFileNotFound   = errors.New("config file not found")
	ErrInvalidJSON    = errors.New("invalid JSON format")
	ErrFileReadError  = errors.New("failed to read config file")
	ErrFileWriteError = errors.New("failed to write config file")
	ErrInvalidPath    = errors.New("invalid config path")
)

type Loader struct {
	primaryPath  string
	fallbackPath string
	explicitPath string
}

func NewLoader() *Loader {
	homeDir, _ := os.UserHomeDir()
	primaryPath := filepath.Join(homeDir, ".config", "kassie", "config.json")
	fallbackPath := filepath.Join(".", "kassie.config.json")

	return &Loader{
		primaryPath:  primaryPath,
		fallbackPath: fallbackPath,
	}
}

func NewLoaderWithPath(path string) *Loader {
	loader := NewLoader()
	loader.explicitPath = path
	return loader
}

func (l *Loader) GetConfigPath() (string, error) {
	if l.explicitPath != "" {
		if fileExists(l.explicitPath) {
			return l.explicitPath, nil
		}
		return "", fmt.Errorf("%w: %s", ErrFileNotFound, l.explicitPath)
	}

	if fileExists(l.primaryPath) {
		return l.primaryPath, nil
	}

	if fileExists(l.fallbackPath) {
		return l.fallbackPath, nil
	}

	return "", ErrFileNotFound
}

func (l *Loader) Load() (*Config, error) {
	configPath, err := l.GetConfigPath()
	if err != nil {
		return nil, err
	}

	return l.LoadFromPath(configPath)
}

func (l *Loader) LoadFromPath(path string) (*Config, error) {
	if path == "" {
		return nil, ErrInvalidPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrFileNotFound, path)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("%w: permission denied", ErrFileReadError)
		}
		return nil, fmt.Errorf("%w: %v", ErrFileReadError, err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("%w: file is empty", ErrInvalidJSON)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	config.SetDefaults()

	if err := InterpolateConfig(&config); err != nil {
		return nil, fmt.Errorf("failed to interpolate environment variables: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func (l *Loader) Save(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	path := l.explicitPath
	if path == "" {
		path = l.primaryPath
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%w: %v", ErrFileWriteError, err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("%w: permission denied", ErrFileWriteError)
		}
		return fmt.Errorf("%w: %v", ErrFileWriteError, err)
	}

	return nil
}

func (l *Loader) Exists() bool {
	_, err := l.GetConfigPath()
	return err == nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func LoadConfig() (*Config, error) {
	loader := NewLoader()
	return loader.Load()
}

func LoadConfigFromPath(path string) (*Config, error) {
	loader := NewLoaderWithPath(path)
	return loader.Load()
}

func SaveConfig(config *Config) error {
	loader := NewLoader()
	return loader.Save(config)
}

func SaveConfigToPath(config *Config, path string) error {
	loader := NewLoaderWithPath(path)
	return loader.Save(config)
}

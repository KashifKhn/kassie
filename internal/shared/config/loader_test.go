package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()

	if loader.primaryPath == "" {
		t.Error("primaryPath should not be empty")
	}
	if loader.fallbackPath == "" {
		t.Error("fallbackPath should not be empty")
	}
	if loader.explicitPath != "" {
		t.Error("explicitPath should be empty")
	}
}

func TestNewLoaderWithPath(t *testing.T) {
	path := "/custom/path/config.json"
	loader := NewLoaderWithPath(path)

	if loader.explicitPath != path {
		t.Errorf("explicitPath = %v, want %v", loader.explicitPath, path)
	}
}

func TestLoaderGetConfigPath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		setupFiles func() *Loader
		wantErr    error
		checkPath  func(string) bool
	}{
		{
			name: "explicit path exists",
			setupFiles: func() *Loader {
				path := filepath.Join(tmpDir, "explicit.json")
				os.WriteFile(path, []byte("{}"), 0644)
				return NewLoaderWithPath(path)
			},
			wantErr: nil,
			checkPath: func(path string) bool {
				return filepath.Base(path) == "explicit.json"
			},
		},
		{
			name: "explicit path does not exist",
			setupFiles: func() *Loader {
				path := filepath.Join(tmpDir, "nonexistent.json")
				return NewLoaderWithPath(path)
			},
			wantErr: ErrFileNotFound,
		},
		{
			name: "primary path exists",
			setupFiles: func() *Loader {
				loader := NewLoader()
				loader.primaryPath = filepath.Join(tmpDir, "primary.json")
				os.WriteFile(loader.primaryPath, []byte("{}"), 0644)
				return loader
			},
			wantErr: nil,
			checkPath: func(path string) bool {
				return filepath.Base(path) == "primary.json"
			},
		},
		{
			name: "fallback path exists",
			setupFiles: func() *Loader {
				loader := NewLoader()
				loader.primaryPath = filepath.Join(tmpDir, "nonexistent_primary.json")
				loader.fallbackPath = filepath.Join(tmpDir, "fallback.json")
				os.WriteFile(loader.fallbackPath, []byte("{}"), 0644)
				return loader
			},
			wantErr: nil,
			checkPath: func(path string) bool {
				return filepath.Base(path) == "fallback.json"
			},
		},
		{
			name: "no config file exists",
			setupFiles: func() *Loader {
				loader := NewLoader()
				loader.primaryPath = filepath.Join(tmpDir, "nonexistent1.json")
				loader.fallbackPath = filepath.Join(tmpDir, "nonexistent2.json")
				return loader
			},
			wantErr: ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := tt.setupFiles()
			path, err := loader.GetConfigPath()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("GetConfigPath() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetConfigPath() unexpected error = %v", err)
				return
			}

			if tt.checkPath != nil && !tt.checkPath(path) {
				t.Errorf("GetConfigPath() path check failed for %v", path)
			}
		})
	}
}

func TestLoaderLoadFromPath(t *testing.T) {
	tmpDir := t.TempDir()

	validConfig := Config{
		Version: "1.0",
		Profiles: []Profile{
			{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
		},
		Defaults: DefaultConfig{
			PageSize:  100,
			TimeoutMs: 5000,
		},
		Clients: ClientConfig{
			Web: WebConfig{DefaultPort: 8080},
			TUI: TUIConfig{Theme: "default"},
		},
	}

	tests := []struct {
		name       string
		setupFile  func() string
		wantErr    error
		checkError func(error) bool
	}{
		{
			name: "valid config file",
			setupFile: func() string {
				path := filepath.Join(tmpDir, "valid.json")
				data, _ := json.Marshal(validConfig)
				os.WriteFile(path, data, 0644)
				return path
			},
			wantErr: nil,
		},
		{
			name: "file does not exist",
			setupFile: func() string {
				return filepath.Join(tmpDir, "nonexistent.json")
			},
			wantErr: ErrFileNotFound,
		},
		{
			name: "empty path",
			setupFile: func() string {
				return ""
			},
			wantErr: ErrInvalidPath,
		},
		{
			name: "empty file",
			setupFile: func() string {
				path := filepath.Join(tmpDir, "empty.json")
				os.WriteFile(path, []byte(""), 0644)
				return path
			},
			wantErr: ErrInvalidJSON,
		},
		{
			name: "invalid json",
			setupFile: func() string {
				path := filepath.Join(tmpDir, "invalid.json")
				os.WriteFile(path, []byte("{invalid json}"), 0644)
				return path
			},
			wantErr: ErrInvalidJSON,
		},
		{
			name: "valid json but invalid config structure",
			setupFile: func() string {
				path := filepath.Join(tmpDir, "invalid_structure.json")
				os.WriteFile(path, []byte(`{"version": "1.0"}`), 0644)
				return path
			},
			checkError: func(err error) bool {
				return err != nil
			},
		},
		{
			name: "config with validation error",
			setupFile: func() string {
				invalidConfig := Config{
					Version: "1.0",
					Profiles: []Profile{
						{Name: "test", Hosts: []string{}, Port: 9042},
					},
				}
				path := filepath.Join(tmpDir, "validation_error.json")
				data, _ := json.Marshal(invalidConfig)
				os.WriteFile(path, data, 0644)
				return path
			},
			checkError: func(err error) bool {
				return err != nil && errors.Is(err, ErrNoHosts)
			},
		},
		{
			name: "config with environment variable interpolation",
			setupFile: func() string {
				configWithEnv := Config{
					Version: "1.0",
					Profiles: []Profile{
						{
							Name:  "test",
							Hosts: []string{"localhost"},
							Port:  9042,
							Auth: &AuthConfig{
								Username: "admin",
								Password: "${TEST_PASSWORD_LOADER}",
							},
						},
					},
					Defaults: DefaultConfig{
						PageSize:  100,
						TimeoutMs: 5000,
					},
					Clients: ClientConfig{
						Web: WebConfig{DefaultPort: 8080},
						TUI: TUIConfig{Theme: "default"},
					},
				}
				path := filepath.Join(tmpDir, "with_env.json")
				data, _ := json.Marshal(configWithEnv)
				os.WriteFile(path, data, 0644)
				return path
			},
			wantErr: nil,
		},
		{
			name: "config with missing environment variable",
			setupFile: func() string {
				configWithEnv := Config{
					Version: "1.0",
					Profiles: []Profile{
						{
							Name:  "test",
							Hosts: []string{"localhost"},
							Port:  9042,
							Auth: &AuthConfig{
								Password: "${MISSING_VAR}",
							},
						},
					},
					Defaults: DefaultConfig{
						PageSize:  100,
						TimeoutMs: 5000,
					},
					Clients: ClientConfig{
						Web: WebConfig{DefaultPort: 8080},
					},
				}
				path := filepath.Join(tmpDir, "missing_env.json")
				data, _ := json.Marshal(configWithEnv)
				os.WriteFile(path, data, 0644)
				return path
			},
			checkError: func(err error) bool {
				return err != nil && errors.Is(err, ErrVarNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "config with environment variable interpolation" {
				os.Setenv("TEST_PASSWORD_LOADER", "secret123")
				defer os.Unsetenv("TEST_PASSWORD_LOADER")
			}

			loader := NewLoader()
			path := tt.setupFile()
			config, err := loader.LoadFromPath(path)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("LoadFromPath() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("LoadFromPath() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if tt.checkError != nil {
				if !tt.checkError(err) {
					t.Errorf("LoadFromPath() error check failed: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("LoadFromPath() unexpected error = %v", err)
				return
			}

			if config == nil {
				t.Error("LoadFromPath() returned nil config")
			}
		})
	}
}

func TestLoaderSave(t *testing.T) {
	tmpDir := t.TempDir()

	validConfig := &Config{
		Version: "1.0",
		Profiles: []Profile{
			{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
		},
		Defaults: DefaultConfig{
			PageSize:  100,
			TimeoutMs: 5000,
		},
		Clients: ClientConfig{
			Web: WebConfig{DefaultPort: 8080},
			TUI: TUIConfig{Theme: "default"},
		},
	}

	tests := []struct {
		name      string
		setup     func() (*Loader, *Config)
		wantErr   bool
		checkFile func(string) bool
	}{
		{
			name: "save valid config to explicit path",
			setup: func() (*Loader, *Config) {
				path := filepath.Join(tmpDir, "save_explicit.json")
				loader := NewLoaderWithPath(path)
				return loader, validConfig
			},
			wantErr: false,
			checkFile: func(path string) bool {
				_, err := os.Stat(path)
				return err == nil
			},
		},
		{
			name: "save valid config to primary path",
			setup: func() (*Loader, *Config) {
				loader := NewLoader()
				loader.primaryPath = filepath.Join(tmpDir, "save_primary.json")
				return loader, validConfig
			},
			wantErr: false,
		},
		{
			name: "save creates parent directories",
			setup: func() (*Loader, *Config) {
				path := filepath.Join(tmpDir, "nested", "dir", "config.json")
				loader := NewLoaderWithPath(path)
				return loader, validConfig
			},
			wantErr: false,
			checkFile: func(path string) bool {
				_, err := os.Stat(path)
				return err == nil
			},
		},
		{
			name: "save nil config",
			setup: func() (*Loader, *Config) {
				path := filepath.Join(tmpDir, "nil_config.json")
				loader := NewLoaderWithPath(path)
				return loader, nil
			},
			wantErr: true,
		},
		{
			name: "save invalid config",
			setup: func() (*Loader, *Config) {
				path := filepath.Join(tmpDir, "invalid_config.json")
				loader := NewLoaderWithPath(path)
				invalidConfig := &Config{
					Profiles: []Profile{},
				}
				return loader, invalidConfig
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader, config := tt.setup()
			err := loader.Save(config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFile != nil {
				path := loader.explicitPath
				if path == "" {
					path = loader.primaryPath
				}
				if !tt.checkFile(path) {
					t.Errorf("Save() file check failed for %v", path)
				}

				data, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("Failed to read saved file: %v", err)
					return
				}

				var loaded Config
				if err := json.Unmarshal(data, &loaded); err != nil {
					t.Errorf("Failed to unmarshal saved config: %v", err)
				}
			}
		})
	}
}

func TestLoaderExists(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		setup func() *Loader
		want  bool
	}{
		{
			name: "explicit path exists",
			setup: func() *Loader {
				path := filepath.Join(tmpDir, "exists.json")
				os.WriteFile(path, []byte("{}"), 0644)
				return NewLoaderWithPath(path)
			},
			want: true,
		},
		{
			name: "explicit path does not exist",
			setup: func() *Loader {
				path := filepath.Join(tmpDir, "nonexistent.json")
				return NewLoaderWithPath(path)
			},
			want: false,
		},
		{
			name: "primary path exists",
			setup: func() *Loader {
				loader := NewLoader()
				loader.primaryPath = filepath.Join(tmpDir, "primary.json")
				os.WriteFile(loader.primaryPath, []byte("{}"), 0644)
				return loader
			},
			want: true,
		},
		{
			name: "fallback path exists",
			setup: func() *Loader {
				loader := NewLoader()
				loader.primaryPath = filepath.Join(tmpDir, "nonexistent_primary.json")
				loader.fallbackPath = filepath.Join(tmpDir, "fallback.json")
				os.WriteFile(loader.fallbackPath, []byte("{}"), 0644)
				return loader
			},
			want: true,
		},
		{
			name: "no path exists",
			setup: func() *Loader {
				loader := NewLoader()
				loader.primaryPath = filepath.Join(tmpDir, "none1.json")
				loader.fallbackPath = filepath.Join(tmpDir, "none2.json")
				return loader
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := tt.setup()
			got := loader.Exists()

			if got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	validConfig := Config{
		Version: "1.0",
		Profiles: []Profile{
			{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
		},
		Defaults: DefaultConfig{
			PageSize:  100,
			TimeoutMs: 5000,
		},
		Clients: ClientConfig{
			Web: WebConfig{DefaultPort: 8080},
			TUI: TUIConfig{Theme: "default"},
		},
	}

	configDir := filepath.Join(tmpDir, ".config", "kassie")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")
	data, _ := json.Marshal(validConfig)
	os.WriteFile(configPath, data, 0644)

	os.Setenv("HOME", tmpDir)

	config, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() unexpected error = %v", err)
	}
	if config == nil {
		t.Error("LoadConfig() returned nil config")
	}
}

func TestLoadConfigFromPath(t *testing.T) {
	tmpDir := t.TempDir()

	validConfig := Config{
		Version: "1.0",
		Profiles: []Profile{
			{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
		},
		Defaults: DefaultConfig{
			PageSize:  100,
			TimeoutMs: 5000,
		},
		Clients: ClientConfig{
			Web: WebConfig{DefaultPort: 8080},
			TUI: TUIConfig{Theme: "default"},
		},
	}

	path := filepath.Join(tmpDir, "custom.json")
	data, _ := json.Marshal(validConfig)
	os.WriteFile(path, data, 0644)

	config, err := LoadConfigFromPath(path)
	if err != nil {
		t.Errorf("LoadConfigFromPath() unexpected error = %v", err)
	}
	if config == nil {
		t.Error("LoadConfigFromPath() returned nil config")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	os.Setenv("HOME", tmpDir)

	validConfig := &Config{
		Version: "1.0",
		Profiles: []Profile{
			{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
		},
		Defaults: DefaultConfig{
			PageSize:  100,
			TimeoutMs: 5000,
		},
		Clients: ClientConfig{
			Web: WebConfig{DefaultPort: 8080},
			TUI: TUIConfig{Theme: "default"},
		},
	}

	err := SaveConfig(validConfig)
	if err != nil {
		t.Errorf("SaveConfig() unexpected error = %v", err)
	}
}

func TestSaveConfigToPath(t *testing.T) {
	tmpDir := t.TempDir()

	validConfig := &Config{
		Version: "1.0",
		Profiles: []Profile{
			{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
		},
		Defaults: DefaultConfig{
			PageSize:  100,
			TimeoutMs: 5000,
		},
		Clients: ClientConfig{
			Web: WebConfig{DefaultPort: 8080},
			TUI: TUIConfig{Theme: "default"},
		},
	}

	path := filepath.Join(tmpDir, "save_to_path.json")
	err := SaveConfigToPath(validConfig, path)
	if err != nil {
		t.Errorf("SaveConfigToPath() unexpected error = %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("Config file was not created at %v", path)
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		setup func() string
		want  bool
	}{
		{
			name: "file exists",
			setup: func() string {
				path := filepath.Join(tmpDir, "exists.txt")
				os.WriteFile(path, []byte("content"), 0644)
				return path
			},
			want: true,
		},
		{
			name: "file does not exist",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.txt")
			},
			want: false,
		},
		{
			name: "path is directory",
			setup: func() string {
				dir := filepath.Join(tmpDir, "directory")
				os.Mkdir(dir, 0755)
				return dir
			},
			want: false,
		},
		{
			name: "empty path",
			setup: func() string {
				return ""
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			got := fileExists(path)

			if got != tt.want {
				t.Errorf("fileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

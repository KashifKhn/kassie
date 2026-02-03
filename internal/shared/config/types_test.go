package config

import (
	"testing"
)

func TestProfileValidate(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr error
	}{
		{
			name: "valid profile",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			wantErr: nil,
		},
		{
			name: "valid profile with auth",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: nil,
		},
		{
			name: "valid profile with ssl",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled: true,
				},
			},
			wantErr: nil,
		},
		{
			name: "valid profile with keyspace",
			profile: Profile{
				Name:     "test",
				Hosts:    []string{"localhost"},
				Port:     9042,
				Keyspace: "system",
			},
			wantErr: nil,
		},
		{
			name: "multiple hosts",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"host1", "host2", "host3"},
				Port:  9042,
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			profile: Profile{
				Name:  "",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "no hosts",
			profile: Profile{
				Name:  "test",
				Hosts: []string{},
				Port:  9042,
			},
			wantErr: ErrNoHosts,
		},
		{
			name: "nil hosts",
			profile: Profile{
				Name:  "test",
				Hosts: nil,
				Port:  9042,
			},
			wantErr: ErrNoHosts,
		},
		{
			name: "port too low",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  0,
			},
			wantErr: ErrInvalidPort,
		},
		{
			name: "port negative",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  -1,
			},
			wantErr: ErrInvalidPort,
		},
		{
			name: "port too high",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  65536,
			},
			wantErr: ErrInvalidPort,
		},
		{
			name: "minimum valid port",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  1,
			},
			wantErr: nil,
		},
		{
			name: "maximum valid port",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  65535,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if err != tt.wantErr {
				t.Errorf("Profile.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	validProfile := Profile{
		Name:  "test",
		Hosts: []string{"localhost"},
		Port:  9042,
	}

	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "valid config",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					DefaultProfile: "test",
					PageSize:       100,
					TimeoutMs:      5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "no profiles",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{},
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: ErrNoProfiles,
		},
		{
			name: "nil profiles",
			config: Config{
				Version:  "1.0",
				Profiles: nil,
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: ErrNoProfiles,
		},
		{
			name: "duplicate profile names",
			config: Config{
				Version: "1.0",
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
					{Name: "test", Hosts: []string{"localhost"}, Port: 9043},
				},
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: ErrDuplicateProfile,
		},
		{
			name: "invalid profile in list",
			config: Config{
				Version: "1.0",
				Profiles: []Profile{
					{Name: "test", Hosts: []string{}, Port: 9042},
				},
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: ErrNoHosts,
		},
		{
			name: "page size too small",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					PageSize:  0,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: ErrInvalidPageSize,
		},
		{
			name: "page size too large",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					PageSize:  10001,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: ErrInvalidPageSize,
		},
		{
			name: "minimum valid page size",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					PageSize:  1,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "maximum valid page size",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					PageSize:  10000,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "timeout too small",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 99,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: ErrInvalidTimeout,
		},
		{
			name: "timeout too large",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 300001,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: ErrInvalidTimeout,
		},
		{
			name: "minimum valid timeout",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 100,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "maximum valid timeout",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 300000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "web port invalid",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 0,
					},
				},
			},
			wantErr: ErrInvalidPort,
		},
		{
			name: "web port too high",
			config: Config{
				Version:  "1.0",
				Profiles: []Profile{validProfile},
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					Web: WebConfig{
						DefaultPort: 65536,
					},
				},
			},
			wantErr: ErrInvalidPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigGetProfile(t *testing.T) {
	config := Config{
		Profiles: []Profile{
			{Name: "local", Hosts: []string{"localhost"}, Port: 9042},
			{Name: "prod", Hosts: []string{"prod.example.com"}, Port: 9042},
			{Name: "dev", Hosts: []string{"dev.example.com"}, Port: 9042},
		},
	}

	tests := []struct {
		name        string
		profileName string
		wantErr     error
		wantProfile *Profile
	}{
		{
			name:        "existing profile",
			profileName: "local",
			wantErr:     nil,
			wantProfile: &Profile{Name: "local", Hosts: []string{"localhost"}, Port: 9042},
		},
		{
			name:        "another existing profile",
			profileName: "prod",
			wantErr:     nil,
			wantProfile: &Profile{Name: "prod", Hosts: []string{"prod.example.com"}, Port: 9042},
		},
		{
			name:        "non-existent profile",
			profileName: "nonexistent",
			wantErr:     ErrProfileNotFound,
			wantProfile: nil,
		},
		{
			name:        "empty profile name",
			profileName: "",
			wantErr:     ErrProfileNotFound,
			wantProfile: nil,
		},
		{
			name:        "case sensitive",
			profileName: "Local",
			wantErr:     ErrProfileNotFound,
			wantProfile: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.GetProfile(tt.profileName)
			if err != tt.wantErr {
				t.Errorf("Config.GetProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantProfile != nil && got != nil {
				if got.Name != tt.wantProfile.Name {
					t.Errorf("Config.GetProfile() = %v, want %v", got.Name, tt.wantProfile.Name)
				}
			}
		})
	}
}

func TestConfigGetDefaultProfile(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		wantErr     error
		wantProfile string
	}{
		{
			name: "explicit default profile",
			config: Config{
				Profiles: []Profile{
					{Name: "local", Hosts: []string{"localhost"}, Port: 9042},
					{Name: "prod", Hosts: []string{"prod.example.com"}, Port: 9042},
				},
				Defaults: DefaultConfig{
					DefaultProfile: "prod",
				},
			},
			wantErr:     nil,
			wantProfile: "prod",
		},
		{
			name: "no default profile set returns first",
			config: Config{
				Profiles: []Profile{
					{Name: "local", Hosts: []string{"localhost"}, Port: 9042},
					{Name: "prod", Hosts: []string{"prod.example.com"}, Port: 9042},
				},
				Defaults: DefaultConfig{},
			},
			wantErr:     nil,
			wantProfile: "local",
		},
		{
			name: "default profile does not exist",
			config: Config{
				Profiles: []Profile{
					{Name: "local", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{
					DefaultProfile: "nonexistent",
				},
			},
			wantErr:     ErrProfileNotFound,
			wantProfile: "",
		},
		{
			name: "empty profiles list",
			config: Config{
				Profiles: []Profile{},
				Defaults: DefaultConfig{
					DefaultProfile: "test",
				},
			},
			wantErr:     ErrProfileNotFound,
			wantProfile: "",
		},
		{
			name: "single profile",
			config: Config{
				Profiles: []Profile{
					{Name: "only", Hosts: []string{"localhost"}, Port: 9042},
				},
			},
			wantErr:     nil,
			wantProfile: "only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.GetDefaultProfile()
			if err != tt.wantErr {
				t.Errorf("Config.GetDefaultProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantProfile != "" && got != nil {
				if got.Name != tt.wantProfile {
					t.Errorf("Config.GetDefaultProfile() = %v, want %v", got.Name, tt.wantProfile)
				}
			}
		})
	}
}

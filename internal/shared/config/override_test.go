package config

import (
	"testing"
)

func TestApplyOverrides(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		override *Override
		wantErr  bool
		check    func(*Config) bool
	}{
		{
			name: "nil override does nothing",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: nil,
			wantErr:  false,
			check: func(c *Config) bool {
				return c.Defaults.PageSize == 100
			},
		},
		{
			name:     "nil config returns error",
			config:   nil,
			override: &Override{PageSize: 200},
			wantErr:  true,
		},
		{
			name: "override page size",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Override{PageSize: 200},
			wantErr:  false,
			check: func(c *Config) bool {
				return c.Defaults.PageSize == 200
			},
		},
		{
			name: "override timeout",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Override{TimeoutMs: 10000},
			wantErr:  false,
			check: func(c *Config) bool {
				return c.Defaults.TimeoutMs == 10000
			},
		},
		{
			name: "override web port",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Override{WebPort: 3000},
			wantErr:  false,
			check: func(c *Config) bool {
				return c.Clients.Web.DefaultPort == 3000
			},
		},
		{
			name: "override theme",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}, TUI: TUIConfig{Theme: "default"}},
			},
			override: &Override{Theme: "dracula"},
			wantErr:  false,
			check: func(c *Config) bool {
				return c.Clients.TUI.Theme == "dracula"
			},
		},
		{
			name: "override vim mode",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}, TUI: TUIConfig{VimMode: false}},
			},
			override: &Override{VimMode: boolPtr(true)},
			wantErr:  false,
			check: func(c *Config) bool {
				return c.Clients.TUI.VimMode == true
			},
		},
		{
			name: "override profile hosts",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Override{
				ProfileName: "test",
				Hosts:       []string{"host1", "host2"},
			},
			wantErr: false,
			check: func(c *Config) bool {
				profile, _ := c.GetProfile("test")
				return len(profile.Hosts) == 2 && profile.Hosts[0] == "host1"
			},
		},
		{
			name: "override profile port",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Override{
				ProfileName: "test",
				Port:        9043,
			},
			wantErr: false,
			check: func(c *Config) bool {
				profile, _ := c.GetProfile("test")
				return profile.Port == 9043
			},
		},
		{
			name: "override profile keyspace",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Override{
				ProfileName: "test",
				Keyspace:    "mykeyspace",
			},
			wantErr: false,
			check: func(c *Config) bool {
				profile, _ := c.GetProfile("test")
				return profile.Keyspace == "mykeyspace"
			},
		},
		{
			name: "override profile auth",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Override{
				ProfileName: "test",
				Username:    "admin",
				Password:    "secret",
			},
			wantErr: false,
			check: func(c *Config) bool {
				profile, _ := c.GetProfile("test")
				return profile.Auth != nil && profile.Auth.Username == "admin" && profile.Auth.Password == "secret"
			},
		},
		{
			name: "override profile ssl",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Override{
				ProfileName: "test",
				SSLEnabled:  boolPtr(true),
			},
			wantErr: false,
			check: func(c *Config) bool {
				profile, _ := c.GetProfile("test")
				return profile.SSL != nil && profile.SSL.Enabled
			},
		},
		{
			name: "override nonexistent profile",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Override{
				ProfileName: "nonexistent",
				Port:        9043,
			},
			wantErr: true,
		},
		{
			name: "override causes validation error",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Override{PageSize: 20000},
			wantErr:  true,
		},
		{
			name: "multiple overrides",
			config: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}, TUI: TUIConfig{Theme: "default"}},
			},
			override: &Override{
				ProfileName: "test",
				Port:        9043,
				PageSize:    200,
				TimeoutMs:   10000,
				WebPort:     3000,
				Theme:       "nord",
			},
			wantErr: false,
			check: func(c *Config) bool {
				profile, _ := c.GetProfile("test")
				return profile.Port == 9043 &&
					c.Defaults.PageSize == 200 &&
					c.Defaults.TimeoutMs == 10000 &&
					c.Clients.Web.DefaultPort == 3000 &&
					c.Clients.TUI.Theme == "nord"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyOverrides(tt.config, tt.override)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyOverrides() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				if !tt.check(tt.config) {
					t.Errorf("ApplyOverrides() check failed")
				}
			}
		})
	}
}

func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		name     string
		base     *Config
		override *Config
		wantErr  bool
		check    func(*Config) bool
	}{
		{
			name: "nil base returns error",
			base: nil,
			override: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
			},
			wantErr: true,
		},
		{
			name: "nil override returns base clone",
			base: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: nil,
			wantErr:  false,
			check: func(c *Config) bool {
				return c.Defaults.PageSize == 100
			},
		},
		{
			name: "merge new profile",
			base: &Config{
				Profiles: []Profile{
					{Name: "test1", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Config{
				Profiles: []Profile{
					{Name: "test2", Hosts: []string{"host2"}, Port: 9043},
				},
			},
			wantErr: false,
			check: func(c *Config) bool {
				return len(c.Profiles) == 2
			},
		},
		{
			name: "merge existing profile",
			base: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"newhost"}, Port: 9043},
				},
			},
			wantErr: false,
			check: func(c *Config) bool {
				profile, _ := c.GetProfile("test")
				return len(c.Profiles) == 1 && profile.Port == 9043 && profile.Hosts[0] == "newhost"
			},
		},
		{
			name: "merge defaults",
			base: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000, DefaultProfile: "test"},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Config{
				Defaults: DefaultConfig{PageSize: 200, TimeoutMs: 10000},
			},
			wantErr: false,
			check: func(c *Config) bool {
				return c.Defaults.PageSize == 200 && c.Defaults.TimeoutMs == 10000 && c.Defaults.DefaultProfile == "test"
			},
		},
		{
			name: "merge client config",
			base: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080, AutoOpenBrowser: false}, TUI: TUIConfig{Theme: "default", VimMode: false}},
			},
			override: &Config{
				Clients: ClientConfig{Web: WebConfig{DefaultPort: 3000, AutoOpenBrowser: true}, TUI: TUIConfig{Theme: "dracula", VimMode: true}},
			},
			wantErr: false,
			check: func(c *Config) bool {
				return c.Clients.Web.DefaultPort == 3000 &&
					c.Clients.Web.AutoOpenBrowser == true &&
					c.Clients.TUI.Theme == "dracula" &&
					c.Clients.TUI.VimMode == true
			},
		},
		{
			name: "merge causes validation error",
			base: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Config{
				Defaults: DefaultConfig{PageSize: 20000},
			},
			wantErr: true,
		},
		{
			name: "merge invalid profile",
			base: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Config{
				Profiles: []Profile{
					{Name: "invalid", Hosts: []string{}, Port: 9042},
				},
			},
			wantErr: true,
		},
		{
			name: "merge does not modify base",
			base: &Config{
				Profiles: []Profile{
					{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
				},
				Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
				Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}},
			},
			override: &Config{
				Defaults: DefaultConfig{PageSize: 200},
			},
			wantErr: false,
			check: func(c *Config) bool {
				return c.Defaults.PageSize == 200
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalPageSize := 0
			if tt.base != nil {
				originalPageSize = tt.base.Defaults.PageSize
			}

			merged, err := MergeConfigs(tt.base, tt.override)

			if (err != nil) != tt.wantErr {
				t.Errorf("MergeConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if merged == nil {
					t.Error("MergeConfigs() returned nil")
					return
				}

				if tt.check != nil && !tt.check(merged) {
					t.Error("MergeConfigs() check failed")
				}

				if tt.base != nil && tt.name == "merge does not modify base" {
					if tt.base.Defaults.PageSize != originalPageSize {
						t.Error("MergeConfigs() modified base config")
					}
				}
			}
		})
	}
}

func TestApplyProfileOverride(t *testing.T) {
	tests := []struct {
		name     string
		profile  *Profile
		override *Override
		wantErr  bool
		check    func(*Profile) bool
	}{
		{
			name:     "nil profile returns error",
			profile:  nil,
			override: &Override{Port: 9043},
			wantErr:  true,
		},
		{
			name: "nil override does nothing",
			profile: &Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: nil,
			wantErr:  false,
			check: func(p *Profile) bool {
				return p.Port == 9042
			},
		},
		{
			name: "override port",
			profile: &Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: &Override{Port: 9043},
			wantErr:  false,
			check: func(p *Profile) bool {
				return p.Port == 9043
			},
		},
		{
			name: "override hosts",
			profile: &Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: &Override{Hosts: []string{"host1", "host2"}},
			wantErr:  false,
			check: func(p *Profile) bool {
				return len(p.Hosts) == 2 && p.Hosts[0] == "host1"
			},
		},
		{
			name: "override auth",
			profile: &Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: &Override{Username: "admin", Password: "secret"},
			wantErr:  false,
			check: func(p *Profile) bool {
				return p.Auth != nil && p.Auth.Username == "admin"
			},
		},
		{
			name: "override causes validation error",
			profile: &Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: &Override{Port: 70000},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyProfileOverride(tt.profile, tt.override)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyProfileOverride() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				if !tt.check(tt.profile) {
					t.Error("ApplyProfileOverride() check failed")
				}
			}
		})
	}
}

func TestConfigClone(t *testing.T) {
	original := &Config{
		Version: "1.0",
		Profiles: []Profile{
			{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
		},
		Defaults: DefaultConfig{PageSize: 100, TimeoutMs: 5000},
		Clients:  ClientConfig{Web: WebConfig{DefaultPort: 8080}, TUI: TUIConfig{Theme: "default"}},
	}

	clone := original.clone()

	if clone.Version != original.Version {
		t.Errorf("Version = %v, want %v", clone.Version, original.Version)
	}

	if len(clone.Profiles) != len(original.Profiles) {
		t.Errorf("Profiles length = %v, want %v", len(clone.Profiles), len(original.Profiles))
	}

	clone.Profiles[0].Port = 9043
	if original.Profiles[0].Port == 9043 {
		t.Error("Modifying clone affected original")
	}

	clone.Defaults.PageSize = 200
	if original.Defaults.PageSize == 200 {
		t.Error("Modifying clone defaults affected original")
	}
}

func TestBuildProfileOverride(t *testing.T) {
	tests := []struct {
		name     string
		override *Override
		check    func(*Profile) bool
	}{
		{
			name:     "empty override",
			override: &Override{},
			check: func(p *Profile) bool {
				return len(p.Hosts) == 0 && p.Port == 0
			},
		},
		{
			name:     "hosts only",
			override: &Override{Hosts: []string{"host1", "host2"}},
			check: func(p *Profile) bool {
				return len(p.Hosts) == 2 && p.Hosts[0] == "host1"
			},
		},
		{
			name:     "port only",
			override: &Override{Port: 9043},
			check: func(p *Profile) bool {
				return p.Port == 9043
			},
		},
		{
			name:     "keyspace only",
			override: &Override{Keyspace: "mykeyspace"},
			check: func(p *Profile) bool {
				return p.Keyspace == "mykeyspace"
			},
		},
		{
			name:     "username only",
			override: &Override{Username: "admin"},
			check: func(p *Profile) bool {
				return p.Auth != nil && p.Auth.Username == "admin"
			},
		},
		{
			name:     "password only",
			override: &Override{Password: "secret"},
			check: func(p *Profile) bool {
				return p.Auth != nil && p.Auth.Password == "secret"
			},
		},
		{
			name:     "both username and password",
			override: &Override{Username: "admin", Password: "secret"},
			check: func(p *Profile) bool {
				return p.Auth != nil && p.Auth.Username == "admin" && p.Auth.Password == "secret"
			},
		},
		{
			name:     "ssl enabled",
			override: &Override{SSLEnabled: boolPtr(true)},
			check: func(p *Profile) bool {
				return p.SSL != nil && p.SSL.Enabled
			},
		},
		{
			name:     "ssl disabled",
			override: &Override{SSLEnabled: boolPtr(false)},
			check: func(p *Profile) bool {
				return p.SSL != nil && !p.SSL.Enabled
			},
		},
		{
			name: "all fields",
			override: &Override{
				Hosts:      []string{"host1"},
				Port:       9043,
				Keyspace:   "ks",
				Username:   "admin",
				Password:   "secret",
				SSLEnabled: boolPtr(true),
			},
			check: func(p *Profile) bool {
				return len(p.Hosts) == 1 &&
					p.Port == 9043 &&
					p.Keyspace == "ks" &&
					p.Auth != nil &&
					p.SSL != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := buildProfileOverride(tt.override)

			if tt.check != nil && !tt.check(profile) {
				t.Error("buildProfileOverride() check failed")
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

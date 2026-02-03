package config

import (
	"errors"
	"testing"
)

func TestProfileClone(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
	}{
		{
			name: "simple profile",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
		},
		{
			name: "profile with auth",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "admin",
					Password: "secret",
				},
			},
		},
		{
			name: "profile with ssl",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled:            true,
					CertPath:           "/path/to/cert",
					KeyPath:            "/path/to/key",
					CAPath:             "/path/to/ca",
					InsecureSkipVerify: true,
				},
			},
		},
		{
			name: "profile with everything",
			profile: Profile{
				Name:     "test",
				Hosts:    []string{"host1", "host2", "host3"},
				Port:     9042,
				Keyspace: "mykeyspace",
				Auth: &AuthConfig{
					Username: "admin",
					Password: "secret",
				},
				SSL: &SSLConfig{
					Enabled:  true,
					CertPath: "/cert",
				},
			},
		},
		{
			name: "empty hosts slice",
			profile: Profile{
				Name:  "test",
				Hosts: []string{},
				Port:  9042,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.profile.Clone()

			if clone.Name != tt.profile.Name {
				t.Errorf("Name = %v, want %v", clone.Name, tt.profile.Name)
			}
			if clone.Port != tt.profile.Port {
				t.Errorf("Port = %v, want %v", clone.Port, tt.profile.Port)
			}
			if clone.Keyspace != tt.profile.Keyspace {
				t.Errorf("Keyspace = %v, want %v", clone.Keyspace, tt.profile.Keyspace)
			}

			if len(clone.Hosts) != len(tt.profile.Hosts) {
				t.Fatalf("Hosts length = %v, want %v", len(clone.Hosts), len(tt.profile.Hosts))
			}
			for i := range clone.Hosts {
				if clone.Hosts[i] != tt.profile.Hosts[i] {
					t.Errorf("Hosts[%d] = %v, want %v", i, clone.Hosts[i], tt.profile.Hosts[i])
				}
			}

			if tt.profile.Auth != nil {
				if clone.Auth == nil {
					t.Fatal("Auth is nil, want non-nil")
				}
				if clone.Auth.Username != tt.profile.Auth.Username {
					t.Errorf("Auth.Username = %v, want %v", clone.Auth.Username, tt.profile.Auth.Username)
				}
				if clone.Auth.Password != tt.profile.Auth.Password {
					t.Errorf("Auth.Password = %v, want %v", clone.Auth.Password, tt.profile.Auth.Password)
				}
				if clone.Auth == tt.profile.Auth {
					t.Error("Auth pointer is same, should be different")
				}
			}

			if tt.profile.SSL != nil {
				if clone.SSL == nil {
					t.Fatal("SSL is nil, want non-nil")
				}
				if clone.SSL.Enabled != tt.profile.SSL.Enabled {
					t.Errorf("SSL.Enabled = %v, want %v", clone.SSL.Enabled, tt.profile.SSL.Enabled)
				}
				if clone.SSL == tt.profile.SSL {
					t.Error("SSL pointer is same, should be different")
				}
			}

			if len(tt.profile.Hosts) > 0 {
				clone.Hosts[0] = "modified"
				if tt.profile.Hosts[0] == "modified" {
					t.Error("Modifying clone affected original hosts")
				}
			}
		})
	}
}

func TestProfileMergeWith(t *testing.T) {
	tests := []struct {
		name     string
		base     Profile
		override Profile
		want     Profile
		wantErr  bool
	}{
		{
			name: "override hosts",
			base: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: Profile{
				Name:  "test",
				Hosts: []string{"host1", "host2"},
			},
			want: Profile{
				Name:  "test",
				Hosts: []string{"host1", "host2"},
				Port:  9042,
			},
			wantErr: false,
		},
		{
			name: "override port",
			base: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: Profile{
				Port: 9043,
			},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9043,
			},
			wantErr: false,
		},
		{
			name: "override keyspace",
			base: Profile{
				Name:     "test",
				Hosts:    []string{"localhost"},
				Port:     9042,
				Keyspace: "old",
			},
			override: Profile{
				Keyspace: "new",
			},
			want: Profile{
				Name:     "test",
				Hosts:    []string{"localhost"},
				Port:     9042,
				Keyspace: "new",
			},
			wantErr: false,
		},
		{
			name: "add auth to profile without auth",
			base: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: Profile{
				Auth: &AuthConfig{
					Username: "admin",
					Password: "secret",
				},
			},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "admin",
					Password: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "override auth username only",
			base: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "olduser",
					Password: "oldpass",
				},
			},
			override: Profile{
				Auth: &AuthConfig{
					Username: "newuser",
				},
			},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "newuser",
					Password: "oldpass",
				},
			},
			wantErr: false,
		},
		{
			name: "override auth password only",
			base: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "user",
					Password: "oldpass",
				},
			},
			override: Profile{
				Auth: &AuthConfig{
					Password: "newpass",
				},
			},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "user",
					Password: "newpass",
				},
			},
			wantErr: false,
		},
		{
			name: "add ssl to profile without ssl",
			base: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: Profile{
				SSL: &SSLConfig{
					Enabled:  true,
					CertPath: "/cert",
				},
			},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled:  true,
					CertPath: "/cert",
				},
			},
			wantErr: false,
		},
		{
			name: "override ssl cert path",
			base: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled:  true,
					CertPath: "/old/cert",
					KeyPath:  "/old/key",
				},
			},
			override: Profile{
				SSL: &SSLConfig{
					Enabled:  true,
					CertPath: "/new/cert",
				},
			},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled:  true,
					CertPath: "/new/cert",
					KeyPath:  "/old/key",
				},
			},
			wantErr: false,
		},
		{
			name: "override insecure skip verify",
			base: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled:            true,
					InsecureSkipVerify: false,
				},
			},
			override: Profile{
				SSL: &SSLConfig{
					Enabled:            true,
					InsecureSkipVerify: true,
				},
			},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled:            true,
					InsecureSkipVerify: true,
				},
			},
			wantErr: false,
		},
		{
			name: "empty override does nothing",
			base: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: Profile{},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			wantErr: false,
		},
		{
			name: "different profile names",
			base: Profile{
				Name:  "test1",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: Profile{
				Name: "test2",
			},
			wantErr: true,
		},
		{
			name: "override with empty name allowed",
			base: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			override: Profile{
				Name: "",
				Port: 9043,
			},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9043,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := tt.base.Clone()
			err := profile.MergeWith(&tt.override)

			if (err != nil) != tt.wantErr {
				t.Errorf("MergeWith() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if profile.Name != tt.want.Name {
					t.Errorf("Name = %v, want %v", profile.Name, tt.want.Name)
				}
				if profile.Port != tt.want.Port {
					t.Errorf("Port = %v, want %v", profile.Port, tt.want.Port)
				}
				if profile.Keyspace != tt.want.Keyspace {
					t.Errorf("Keyspace = %v, want %v", profile.Keyspace, tt.want.Keyspace)
				}
				if len(profile.Hosts) != len(tt.want.Hosts) {
					t.Errorf("Hosts length = %v, want %v", len(profile.Hosts), len(tt.want.Hosts))
				}
				if tt.want.Auth != nil && profile.Auth != nil {
					if profile.Auth.Username != tt.want.Auth.Username {
						t.Errorf("Auth.Username = %v, want %v", profile.Auth.Username, tt.want.Auth.Username)
					}
					if profile.Auth.Password != tt.want.Auth.Password {
						t.Errorf("Auth.Password = %v, want %v", profile.Auth.Password, tt.want.Auth.Password)
					}
				}
				if tt.want.SSL != nil && profile.SSL != nil {
					if profile.SSL.Enabled != tt.want.SSL.Enabled {
						t.Errorf("SSL.Enabled = %v, want %v", profile.SSL.Enabled, tt.want.SSL.Enabled)
					}
					if profile.SSL.InsecureSkipVerify != tt.want.SSL.InsecureSkipVerify {
						t.Errorf("SSL.InsecureSkipVerify = %v, want %v", profile.SSL.InsecureSkipVerify, tt.want.SSL.InsecureSkipVerify)
					}
				}
			}
		})
	}
}

func TestConfigFindProfile(t *testing.T) {
	config := Config{
		Profiles: []Profile{
			{Name: "first", Hosts: []string{"localhost"}, Port: 9042},
			{Name: "second", Hosts: []string{"localhost"}, Port: 9043},
			{Name: "third", Hosts: []string{"localhost"}, Port: 9044},
		},
	}

	tests := []struct {
		name      string
		findName  string
		wantIndex int
		wantFound bool
	}{
		{
			name:      "find first",
			findName:  "first",
			wantIndex: 0,
			wantFound: true,
		},
		{
			name:      "find middle",
			findName:  "second",
			wantIndex: 1,
			wantFound: true,
		},
		{
			name:      "find last",
			findName:  "third",
			wantIndex: 2,
			wantFound: true,
		},
		{
			name:      "not found",
			findName:  "nonexistent",
			wantIndex: -1,
			wantFound: false,
		},
		{
			name:      "empty name",
			findName:  "",
			wantIndex: -1,
			wantFound: false,
		},
		{
			name:      "case sensitive",
			findName:  "First",
			wantIndex: -1,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, profile := config.FindProfile(tt.findName)

			if idx != tt.wantIndex {
				t.Errorf("FindProfile() index = %v, want %v", idx, tt.wantIndex)
			}

			if (profile != nil) != tt.wantFound {
				t.Errorf("FindProfile() found = %v, want %v", profile != nil, tt.wantFound)
			}

			if tt.wantFound && profile.Name != tt.findName {
				t.Errorf("FindProfile() profile name = %v, want %v", profile.Name, tt.findName)
			}
		})
	}
}

func TestConfigAddProfile(t *testing.T) {
	tests := []struct {
		name       string
		initial    []Profile
		addProfile Profile
		wantErr    error
		wantLen    int
	}{
		{
			name:    "add to empty",
			initial: []Profile{},
			addProfile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			wantErr: nil,
			wantLen: 1,
		},
		{
			name: "add to existing",
			initial: []Profile{
				{Name: "existing", Hosts: []string{"localhost"}, Port: 9042},
			},
			addProfile: Profile{
				Name:  "new",
				Hosts: []string{"localhost"},
				Port:  9043,
			},
			wantErr: nil,
			wantLen: 2,
		},
		{
			name: "add duplicate",
			initial: []Profile{
				{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
			},
			addProfile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9043,
			},
			wantErr: ErrDuplicateProfile,
			wantLen: 1,
		},
		{
			name:    "add invalid profile",
			initial: []Profile{},
			addProfile: Profile{
				Name:  "invalid",
				Hosts: []string{},
				Port:  9042,
			},
			wantErr: ErrNoHosts,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Profiles: make([]Profile, len(tt.initial))}
			copy(config.Profiles, tt.initial)

			err := config.AddProfile(tt.addProfile)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("AddProfile() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("AddProfile() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("AddProfile() unexpected error = %v", err)
				return
			}

			if len(config.Profiles) != tt.wantLen {
				t.Errorf("Profiles length = %v, want %v", len(config.Profiles), tt.wantLen)
			}
		})
	}
}

func TestConfigRemoveProfile(t *testing.T) {
	tests := []struct {
		name       string
		initial    []Profile
		removeName string
		wantErr    error
		wantLen    int
	}{
		{
			name: "remove existing",
			initial: []Profile{
				{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
			},
			removeName: "test",
			wantErr:    nil,
			wantLen:    0,
		},
		{
			name: "remove from multiple",
			initial: []Profile{
				{Name: "first", Hosts: []string{"localhost"}, Port: 9042},
				{Name: "second", Hosts: []string{"localhost"}, Port: 9043},
				{Name: "third", Hosts: []string{"localhost"}, Port: 9044},
			},
			removeName: "second",
			wantErr:    nil,
			wantLen:    2,
		},
		{
			name: "remove first",
			initial: []Profile{
				{Name: "first", Hosts: []string{"localhost"}, Port: 9042},
				{Name: "second", Hosts: []string{"localhost"}, Port: 9043},
			},
			removeName: "first",
			wantErr:    nil,
			wantLen:    1,
		},
		{
			name: "remove last",
			initial: []Profile{
				{Name: "first", Hosts: []string{"localhost"}, Port: 9042},
				{Name: "second", Hosts: []string{"localhost"}, Port: 9043},
			},
			removeName: "second",
			wantErr:    nil,
			wantLen:    1,
		},
		{
			name: "remove nonexistent",
			initial: []Profile{
				{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
			},
			removeName: "nonexistent",
			wantErr:    ErrProfileNotFound,
			wantLen:    1,
		},
		{
			name:       "remove from empty",
			initial:    []Profile{},
			removeName: "test",
			wantErr:    ErrProfileNotFound,
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Profiles: make([]Profile, len(tt.initial))}
			copy(config.Profiles, tt.initial)

			err := config.RemoveProfile(tt.removeName)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("RemoveProfile() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("RemoveProfile() unexpected error = %v", err)
			}

			if len(config.Profiles) != tt.wantLen {
				t.Errorf("Profiles length = %v, want %v", len(config.Profiles), tt.wantLen)
			}

			if tt.wantErr == nil {
				for _, p := range config.Profiles {
					if p.Name == tt.removeName {
						t.Errorf("Profile %s still exists after removal", tt.removeName)
					}
				}
			}
		})
	}
}

func TestConfigUpdateProfile(t *testing.T) {
	tests := []struct {
		name          string
		initial       []Profile
		updateProfile Profile
		wantErr       error
	}{
		{
			name: "update existing",
			initial: []Profile{
				{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
			},
			updateProfile: Profile{
				Name:  "test",
				Hosts: []string{"newhost"},
				Port:  9043,
			},
			wantErr: nil,
		},
		{
			name: "update nonexistent",
			initial: []Profile{
				{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
			},
			updateProfile: Profile{
				Name:  "nonexistent",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			wantErr: ErrProfileNotFound,
		},
		{
			name:    "update in empty config",
			initial: []Profile{},
			updateProfile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			wantErr: ErrProfileNotFound,
		},
		{
			name: "update with invalid profile",
			initial: []Profile{
				{Name: "test", Hosts: []string{"localhost"}, Port: 9042},
			},
			updateProfile: Profile{
				Name:  "test",
				Hosts: []string{},
				Port:  9042,
			},
			wantErr: ErrNoHosts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Profiles: make([]Profile, len(tt.initial))}
			copy(config.Profiles, tt.initial)

			err := config.UpdateProfile(tt.updateProfile)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UpdateProfile() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UpdateProfile() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("UpdateProfile() unexpected error = %v", err)
				return
			}

			if tt.wantErr == nil {
				_, profile := config.FindProfile(tt.updateProfile.Name)
				if profile == nil {
					t.Fatal("Profile not found after update")
				}
				if profile.Port != tt.updateProfile.Port {
					t.Errorf("Port = %v, want %v", profile.Port, tt.updateProfile.Port)
				}
			}
		})
	}
}

func TestConfigSetDefaults(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   Config
	}{
		{
			name:   "all defaults empty",
			config: Config{},
			want: Config{
				Defaults: DefaultConfig{
					PageSize:  100,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					TUI: TUIConfig{
						Theme: "default",
					},
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
		},
		{
			name: "partial defaults",
			config: Config{
				Defaults: DefaultConfig{
					PageSize: 200,
				},
				Clients: ClientConfig{
					TUI: TUIConfig{
						Theme: "dracula",
					},
				},
			},
			want: Config{
				Defaults: DefaultConfig{
					PageSize:  200,
					TimeoutMs: 5000,
				},
				Clients: ClientConfig{
					TUI: TUIConfig{
						Theme: "dracula",
					},
					Web: WebConfig{
						DefaultPort: 8080,
					},
				},
			},
		},
		{
			name: "all defaults set",
			config: Config{
				Defaults: DefaultConfig{
					PageSize:  500,
					TimeoutMs: 10000,
				},
				Clients: ClientConfig{
					TUI: TUIConfig{
						Theme: "nord",
					},
					Web: WebConfig{
						DefaultPort: 3000,
					},
				},
			},
			want: Config{
				Defaults: DefaultConfig{
					PageSize:  500,
					TimeoutMs: 10000,
				},
				Clients: ClientConfig{
					TUI: TUIConfig{
						Theme: "nord",
					},
					Web: WebConfig{
						DefaultPort: 3000,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config
			config.SetDefaults()

			if config.Defaults.PageSize != tt.want.Defaults.PageSize {
				t.Errorf("PageSize = %v, want %v", config.Defaults.PageSize, tt.want.Defaults.PageSize)
			}
			if config.Defaults.TimeoutMs != tt.want.Defaults.TimeoutMs {
				t.Errorf("TimeoutMs = %v, want %v", config.Defaults.TimeoutMs, tt.want.Defaults.TimeoutMs)
			}
			if config.Clients.TUI.Theme != tt.want.Clients.TUI.Theme {
				t.Errorf("Theme = %v, want %v", config.Clients.TUI.Theme, tt.want.Clients.TUI.Theme)
			}
			if config.Clients.Web.DefaultPort != tt.want.Clients.Web.DefaultPort {
				t.Errorf("DefaultPort = %v, want %v", config.Clients.Web.DefaultPort, tt.want.Clients.Web.DefaultPort)
			}
		})
	}
}

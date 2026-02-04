package db

import (
	"testing"
	"time"

	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/gocql/gocql"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ConnectionConfig
		wantErr bool
		check   func(*ConnectionConfig) bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "no hosts",
			cfg: &ConnectionConfig{
				Port: 9042,
			},
			wantErr: true,
		},
		{
			name: "invalid port too low",
			cfg: &ConnectionConfig{
				Hosts: []string{"localhost"},
				Port:  0,
			},
			wantErr: true,
		},
		{
			name: "invalid port too high",
			cfg: &ConnectionConfig{
				Hosts: []string{"localhost"},
				Port:  70000,
			},
			wantErr: true,
		},
		{
			name: "valid config with defaults",
			cfg: &ConnectionConfig{
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			wantErr: false,
			check: func(cfg *ConnectionConfig) bool {
				return cfg.Timeout == 10*time.Second &&
					cfg.PoolSize == 5 &&
					cfg.Consistency == gocql.Quorum
			},
		},
		{
			name: "valid config with custom values",
			cfg: &ConnectionConfig{
				Hosts:       []string{"localhost"},
				Port:        9042,
				Timeout:     5 * time.Second,
				PoolSize:    10,
				Consistency: gocql.One,
			},
			wantErr: false,
			check: func(cfg *ConnectionConfig) bool {
				return cfg.Timeout == 5*time.Second &&
					cfg.PoolSize == 10 &&
					cfg.Consistency == gocql.One
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				if !tt.check(tt.cfg) {
					t.Error("validateConfig() check failed")
				}
			}
		})
	}
}

func TestProfileToConnectionConfig(t *testing.T) {
	tests := []struct {
		name    string
		profile *config.Profile
		check   func(*ConnectionConfig) bool
	}{
		{
			name: "basic profile",
			profile: &config.Profile{
				Name:  "test",
				Hosts: []string{"localhost", "127.0.0.1"},
				Port:  9042,
			},
			check: func(cfg *ConnectionConfig) bool {
				return len(cfg.Hosts) == 2 &&
					cfg.Hosts[0] == "localhost" &&
					cfg.Port == 9042 &&
					cfg.Consistency == gocql.Quorum &&
					cfg.Timeout == 10*time.Second &&
					cfg.PoolSize == 5
			},
		},
		{
			name: "profile with auth",
			profile: &config.Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &config.AuthConfig{
					Username: "admin",
					Password: "secret",
				},
			},
			check: func(cfg *ConnectionConfig) bool {
				return cfg.Username == "admin" && cfg.Password == "secret"
			},
		},
		{
			name: "profile with ssl",
			profile: &config.Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &config.SSLConfig{
					Enabled:            true,
					CertPath:           "/path/to/cert",
					KeyPath:            "/path/to/key",
					CAPath:             "/path/to/ca",
					InsecureSkipVerify: true,
				},
			},
			check: func(cfg *ConnectionConfig) bool {
				return cfg.SSLEnabled &&
					cfg.SSLCertPath == "/path/to/cert" &&
					cfg.SSLKeyPath == "/path/to/key" &&
					cfg.SSLCAPath == "/path/to/ca" &&
					cfg.SSLSkipVerify
			},
		},
		{
			name: "profile with keyspace",
			profile: &config.Profile{
				Name:     "test",
				Hosts:    []string{"localhost"},
				Port:     9042,
				Keyspace: "mykeyspace",
			},
			check: func(cfg *ConnectionConfig) bool {
				return cfg.Keyspace == "mykeyspace"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ProfileToConnectionConfig(tt.profile)

			if cfg == nil {
				t.Fatal("ProfileToConnectionConfig() returned nil")
			}

			if tt.check != nil && !tt.check(cfg) {
				t.Error("ProfileToConnectionConfig() check failed")
			}
		})
	}
}

func TestNewPool(t *testing.T) {
	pool := NewPool()

	if pool == nil {
		t.Fatal("NewPool() returned nil")
	}

	if pool.connections == nil {
		t.Error("connections map not initialized")
	}

	if pool.configs == nil {
		t.Error("configs map not initialized")
	}

	if pool.closed {
		t.Error("pool should not be closed initially")
	}
}

func TestPoolListProfiles(t *testing.T) {
	pool := NewPool()

	profiles := pool.ListProfiles()
	if len(profiles) != 0 {
		t.Errorf("ListProfiles() = %v, want empty slice", profiles)
	}
}

func TestPoolClosedState(t *testing.T) {
	pool := NewPool()
	pool.CloseAll()

	if !pool.closed {
		t.Error("pool should be closed after CloseAll()")
	}

	cfg := &ConnectionConfig{
		Hosts: []string{"localhost"},
		Port:  9042,
	}

	_, err := pool.GetOrCreate("test", cfg)
	if err != ErrPoolClosed {
		t.Errorf("GetOrCreate() on closed pool error = %v, want %v", err, ErrPoolClosed)
	}

	_, err = pool.Get("test")
	if err != ErrPoolClosed {
		t.Errorf("Get() on closed pool error = %v, want %v", err, ErrPoolClosed)
	}
}

func TestPoolClose(t *testing.T) {
	pool := NewPool()

	err := pool.Close("nonexistent")
	if err != nil {
		t.Errorf("Close() on nonexistent profile error = %v, want nil", err)
	}
}

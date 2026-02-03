package config

import (
	"errors"
	"os"
	"testing"
)

func TestInterpolateEnvVars(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		envVars map[string]string
		want    string
		wantErr error
	}{
		{
			name:    "no interpolation needed",
			input:   "plain text",
			envVars: map[string]string{},
			want:    "plain text",
			wantErr: nil,
		},
		{
			name:    "empty string",
			input:   "",
			envVars: map[string]string{},
			want:    "",
			wantErr: nil,
		},
		{
			name:    "single variable",
			input:   "${DB_PASSWORD}",
			envVars: map[string]string{"DB_PASSWORD": "secret123"},
			want:    "secret123",
			wantErr: nil,
		},
		{
			name:    "variable in middle",
			input:   "prefix_${DB_NAME}_suffix",
			envVars: map[string]string{"DB_NAME": "mydb"},
			want:    "prefix_mydb_suffix",
			wantErr: nil,
		},
		{
			name:    "multiple variables",
			input:   "${USER}:${PASSWORD}@${HOST}",
			envVars: map[string]string{"USER": "admin", "PASSWORD": "pass", "HOST": "localhost"},
			want:    "admin:pass@localhost",
			wantErr: nil,
		},
		{
			name:    "repeated variable",
			input:   "${VAR}_${VAR}",
			envVars: map[string]string{"VAR": "test"},
			want:    "test_test",
			wantErr: nil,
		},
		{
			name:    "variable not found",
			input:   "${NONEXISTENT}",
			envVars: map[string]string{},
			want:    "",
			wantErr: ErrVarNotFound,
		},
		{
			name:    "empty variable value",
			input:   "${EMPTY_VAR}",
			envVars: map[string]string{"EMPTY_VAR": ""},
			want:    "",
			wantErr: nil,
		},
		{
			name:    "variable with underscores",
			input:   "${MY_LONG_VAR_NAME}",
			envVars: map[string]string{"MY_LONG_VAR_NAME": "value"},
			want:    "value",
			wantErr: nil,
		},
		{
			name:    "variable with numbers",
			input:   "${VAR_123}",
			envVars: map[string]string{"VAR_123": "numbered"},
			want:    "numbered",
			wantErr: nil,
		},
		{
			name:    "nested interpolation",
			input:   "${VAR1}",
			envVars: map[string]string{"VAR1": "${VAR2}", "VAR2": "final"},
			want:    "final",
			wantErr: nil,
		},
		{
			name:    "nested multiple levels",
			input:   "${VAR1}",
			envVars: map[string]string{"VAR1": "${VAR2}", "VAR2": "${VAR3}", "VAR3": "deep"},
			want:    "deep",
			wantErr: nil,
		},
		{
			name:    "circular reference",
			input:   "${VAR1}",
			envVars: map[string]string{"VAR1": "${VAR2}", "VAR2": "${VAR1}"},
			want:    "",
			wantErr: ErrCircularRef,
		},
		{
			name:  "deep nesting limit",
			input: "${VAR1}",
			envVars: map[string]string{
				"VAR1":  "${VAR2}",
				"VAR2":  "${VAR3}",
				"VAR3":  "${VAR4}",
				"VAR4":  "${VAR5}",
				"VAR5":  "${VAR6}",
				"VAR6":  "${VAR7}",
				"VAR7":  "${VAR8}",
				"VAR8":  "${VAR9}",
				"VAR9":  "${VAR10}",
				"VAR10": "${VAR11}",
				"VAR11": "too deep",
			},
			want:    "",
			wantErr: ErrCircularRef,
		},
		{
			name:    "partial match not interpolated",
			input:   "$VAR",
			envVars: map[string]string{"VAR": "value"},
			want:    "$VAR",
			wantErr: nil,
		},
		{
			name:    "incomplete syntax not interpolated",
			input:   "${VAR",
			envVars: map[string]string{"VAR": "value"},
			want:    "${VAR",
			wantErr: nil,
		},
		{
			name:    "incomplete syntax 2",
			input:   "${}",
			envVars: map[string]string{},
			want:    "${}",
			wantErr: nil,
		},
		{
			name:    "special characters in value",
			input:   "${SPECIAL}",
			envVars: map[string]string{"SPECIAL": "!@#$%^&*()"},
			want:    "!@#$%^&*()",
			wantErr: nil,
		},
		{
			name:    "spaces in value",
			input:   "${SPACED}",
			envVars: map[string]string{"SPACED": "hello world"},
			want:    "hello world",
			wantErr: nil,
		},
		{
			name:    "newlines in value",
			input:   "${MULTILINE}",
			envVars: map[string]string{"MULTILINE": "line1\nline2"},
			want:    "line1\nline2",
			wantErr: nil,
		},
		{
			name:    "unicode in value",
			input:   "${UNICODE}",
			envVars: map[string]string{"UNICODE": "こんにちは"},
			want:    "こんにちは",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			got, err := InterpolateEnvVars(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("InterpolateEnvVars() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("InterpolateEnvVars() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("InterpolateEnvVars() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("InterpolateEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterpolateProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		envVars map[string]string
		want    Profile
		wantErr bool
	}{
		{
			name: "interpolate password",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "admin",
					Password: "${DB_PASSWORD}",
				},
			},
			envVars: map[string]string{"DB_PASSWORD": "secret123"},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "admin",
					Password: "secret123",
				},
			},
			wantErr: false,
		},
		{
			name: "interpolate username",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "${DB_USER}",
					Password: "pass",
				},
			},
			envVars: map[string]string{"DB_USER": "admin"},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "admin",
					Password: "pass",
				},
			},
			wantErr: false,
		},
		{
			name: "interpolate both username and password",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "${DB_USER}",
					Password: "${DB_PASSWORD}",
				},
			},
			envVars: map[string]string{"DB_USER": "admin", "DB_PASSWORD": "secret"},
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
			name: "no auth config",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			envVars: map[string]string{},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			wantErr: false,
		},
		{
			name: "no interpolation needed",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "admin",
					Password: "plaintext",
				},
			},
			envVars: map[string]string{},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "admin",
					Password: "plaintext",
				},
			},
			wantErr: false,
		},
		{
			name: "missing password variable",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				Auth: &AuthConfig{
					Username: "admin",
					Password: "${MISSING_PASSWORD}",
				},
			},
			envVars: map[string]string{},
			wantErr: true,
		},
		{
			name: "interpolate ssl cert path",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled:  true,
					CertPath: "${CERT_PATH}",
				},
			},
			envVars: map[string]string{"CERT_PATH": "/path/to/cert.pem"},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled:  true,
					CertPath: "/path/to/cert.pem",
				},
			},
			wantErr: false,
		},
		{
			name: "interpolate all ssl paths",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled:  true,
					CertPath: "${CERT_PATH}",
					KeyPath:  "${KEY_PATH}",
					CAPath:   "${CA_PATH}",
				},
			},
			envVars: map[string]string{
				"CERT_PATH": "/path/to/cert.pem",
				"KEY_PATH":  "/path/to/key.pem",
				"CA_PATH":   "/path/to/ca.pem",
			},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
				SSL: &SSLConfig{
					Enabled:  true,
					CertPath: "/path/to/cert.pem",
					KeyPath:  "/path/to/key.pem",
					CAPath:   "/path/to/ca.pem",
				},
			},
			wantErr: false,
		},
		{
			name: "no ssl config",
			profile: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			envVars: map[string]string{},
			want: Profile{
				Name:  "test",
				Hosts: []string{"localhost"},
				Port:  9042,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			profile := tt.profile
			err := InterpolateProfile(&profile)

			if (err != nil) != tt.wantErr {
				t.Errorf("InterpolateProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if profile.Auth != nil && tt.want.Auth != nil {
					if profile.Auth.Username != tt.want.Auth.Username {
						t.Errorf("Username = %v, want %v", profile.Auth.Username, tt.want.Auth.Username)
					}
					if profile.Auth.Password != tt.want.Auth.Password {
						t.Errorf("Password = %v, want %v", profile.Auth.Password, tt.want.Auth.Password)
					}
				}
				if profile.SSL != nil && tt.want.SSL != nil {
					if profile.SSL.CertPath != tt.want.SSL.CertPath {
						t.Errorf("CertPath = %v, want %v", profile.SSL.CertPath, tt.want.SSL.CertPath)
					}
					if profile.SSL.KeyPath != tt.want.SSL.KeyPath {
						t.Errorf("KeyPath = %v, want %v", profile.SSL.KeyPath, tt.want.SSL.KeyPath)
					}
					if profile.SSL.CAPath != tt.want.SSL.CAPath {
						t.Errorf("CAPath = %v, want %v", profile.SSL.CAPath, tt.want.SSL.CAPath)
					}
				}
			}
		})
	}
}

func TestInterpolateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "interpolate multiple profiles",
			config: Config{
				Profiles: []Profile{
					{
						Name:  "profile1",
						Hosts: []string{"localhost"},
						Port:  9042,
						Auth: &AuthConfig{
							Username: "user1",
							Password: "${PASSWORD1}",
						},
					},
					{
						Name:  "profile2",
						Hosts: []string{"localhost"},
						Port:  9042,
						Auth: &AuthConfig{
							Username: "user2",
							Password: "${PASSWORD2}",
						},
					},
				},
			},
			envVars: map[string]string{
				"PASSWORD1": "secret1",
				"PASSWORD2": "secret2",
			},
			wantErr: false,
		},
		{
			name: "one profile fails interpolation",
			config: Config{
				Profiles: []Profile{
					{
						Name:  "profile1",
						Hosts: []string{"localhost"},
						Port:  9042,
						Auth: &AuthConfig{
							Password: "${PASSWORD1}",
						},
					},
					{
						Name:  "profile2",
						Hosts: []string{"localhost"},
						Port:  9042,
						Auth: &AuthConfig{
							Password: "${MISSING}",
						},
					},
				},
			},
			envVars: map[string]string{"PASSWORD1": "secret1"},
			wantErr: true,
		},
		{
			name: "empty profiles",
			config: Config{
				Profiles: []Profile{},
			},
			envVars: map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			err := InterpolateConfig(&tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("InterpolateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

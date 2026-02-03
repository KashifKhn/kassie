package config

import "errors"

var (
	ErrProfileNotFound  = errors.New("profile not found")
	ErrInvalidPort      = errors.New("invalid port number")
	ErrNoHosts          = errors.New("no hosts specified")
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrDuplicateProfile = errors.New("duplicate profile name")
	ErrNoProfiles       = errors.New("no profiles defined")
	ErrInvalidPageSize  = errors.New("invalid page size")
	ErrInvalidTimeout   = errors.New("invalid timeout")
)

type Config struct {
	Version  string        `json:"version"`
	Profiles []Profile     `json:"profiles"`
	Defaults DefaultConfig `json:"defaults"`
	Clients  ClientConfig  `json:"clients"`
}

type Profile struct {
	Name     string      `json:"name"`
	Hosts    []string    `json:"hosts"`
	Port     int         `json:"port"`
	Keyspace string      `json:"keyspace,omitempty"`
	Auth     *AuthConfig `json:"auth,omitempty"`
	SSL      *SSLConfig  `json:"ssl,omitempty"`
}

type AuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SSLConfig struct {
	Enabled            bool   `json:"enabled"`
	CertPath           string `json:"cert_path,omitempty"`
	KeyPath            string `json:"key_path,omitempty"`
	CAPath             string `json:"ca_path,omitempty"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify,omitempty"`
}

type DefaultConfig struct {
	DefaultProfile string `json:"default_profile"`
	PageSize       int    `json:"page_size"`
	TimeoutMs      int    `json:"timeout_ms"`
}

type ClientConfig struct {
	TUI TUIConfig `json:"tui"`
	Web WebConfig `json:"web"`
}

type TUIConfig struct {
	Theme   string `json:"theme"`
	VimMode bool   `json:"vim_mode"`
}

type WebConfig struct {
	AutoOpenBrowser bool `json:"auto_open_browser"`
	DefaultPort     int  `json:"default_port"`
}

func (p *Profile) Validate() error {
	if p.Name == "" {
		return ErrInvalidConfig
	}
	if len(p.Hosts) == 0 {
		return ErrNoHosts
	}
	if p.Port < 1 || p.Port > 65535 {
		return ErrInvalidPort
	}
	return nil
}

func (c *Config) Validate() error {
	if len(c.Profiles) == 0 {
		return ErrNoProfiles
	}

	profileNames := make(map[string]bool)
	for _, p := range c.Profiles {
		if profileNames[p.Name] {
			return ErrDuplicateProfile
		}
		profileNames[p.Name] = true

		if err := p.Validate(); err != nil {
			return err
		}
	}

	if c.Defaults.PageSize < 1 || c.Defaults.PageSize > 10000 {
		return ErrInvalidPageSize
	}

	if c.Defaults.TimeoutMs < 100 || c.Defaults.TimeoutMs > 300000 {
		return ErrInvalidTimeout
	}

	if c.Clients.Web.DefaultPort < 1 || c.Clients.Web.DefaultPort > 65535 {
		return ErrInvalidPort
	}

	return nil
}

func (c *Config) GetProfile(name string) (*Profile, error) {
	for _, p := range c.Profiles {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, ErrProfileNotFound
}

func (c *Config) GetDefaultProfile() (*Profile, error) {
	if c.Defaults.DefaultProfile != "" {
		return c.GetProfile(c.Defaults.DefaultProfile)
	}
	if len(c.Profiles) > 0 {
		return &c.Profiles[0], nil
	}
	return nil, ErrProfileNotFound
}

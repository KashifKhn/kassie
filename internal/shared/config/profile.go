package config

import "fmt"

func (p *Profile) Clone() *Profile {
	clone := &Profile{
		Name:     p.Name,
		Hosts:    make([]string, len(p.Hosts)),
		Port:     p.Port,
		Keyspace: p.Keyspace,
	}

	copy(clone.Hosts, p.Hosts)

	if p.Auth != nil {
		clone.Auth = &AuthConfig{
			Username: p.Auth.Username,
			Password: p.Auth.Password,
		}
	}

	if p.SSL != nil {
		clone.SSL = &SSLConfig{
			Enabled:            p.SSL.Enabled,
			CertPath:           p.SSL.CertPath,
			KeyPath:            p.SSL.KeyPath,
			CAPath:             p.SSL.CAPath,
			InsecureSkipVerify: p.SSL.InsecureSkipVerify,
		}
	}

	return clone
}

func (p *Profile) MergeWith(override *Profile) error {
	if override.Name != "" && override.Name != p.Name {
		return fmt.Errorf("cannot merge profiles with different names: %s != %s", p.Name, override.Name)
	}

	if len(override.Hosts) > 0 {
		p.Hosts = make([]string, len(override.Hosts))
		copy(p.Hosts, override.Hosts)
	}

	if override.Port != 0 {
		p.Port = override.Port
	}

	if override.Keyspace != "" {
		p.Keyspace = override.Keyspace
	}

	if override.Auth != nil {
		if p.Auth == nil {
			p.Auth = &AuthConfig{}
		}
		if override.Auth.Username != "" {
			p.Auth.Username = override.Auth.Username
		}
		if override.Auth.Password != "" {
			p.Auth.Password = override.Auth.Password
		}
	}

	if override.SSL != nil {
		if p.SSL == nil {
			p.SSL = &SSLConfig{}
		}
		p.SSL.Enabled = override.SSL.Enabled
		if override.SSL.CertPath != "" {
			p.SSL.CertPath = override.SSL.CertPath
		}
		if override.SSL.KeyPath != "" {
			p.SSL.KeyPath = override.SSL.KeyPath
		}
		if override.SSL.CAPath != "" {
			p.SSL.CAPath = override.SSL.CAPath
		}
		if override.SSL.InsecureSkipVerify {
			p.SSL.InsecureSkipVerify = override.SSL.InsecureSkipVerify
		}
	}

	return nil
}

func (c *Config) FindProfile(name string) (int, *Profile) {
	for i, p := range c.Profiles {
		if p.Name == name {
			return i, &c.Profiles[i]
		}
	}
	return -1, nil
}

func (c *Config) AddProfile(profile Profile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	idx, _ := c.FindProfile(profile.Name)
	if idx != -1 {
		return ErrDuplicateProfile
	}

	c.Profiles = append(c.Profiles, profile)
	return nil
}

func (c *Config) RemoveProfile(name string) error {
	idx, _ := c.FindProfile(name)
	if idx == -1 {
		return ErrProfileNotFound
	}

	c.Profiles = append(c.Profiles[:idx], c.Profiles[idx+1:]...)
	return nil
}

func (c *Config) UpdateProfile(profile Profile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	idx, _ := c.FindProfile(profile.Name)
	if idx == -1 {
		return ErrProfileNotFound
	}

	c.Profiles[idx] = profile
	return nil
}

func (c *Config) SetDefaults() {
	if c.Defaults.PageSize == 0 {
		c.Defaults.PageSize = 100
	}
	if c.Defaults.TimeoutMs == 0 {
		c.Defaults.TimeoutMs = 5000
	}
	if c.Clients.Web.DefaultPort == 0 {
		c.Clients.Web.DefaultPort = 8080
	}
	if c.Clients.TUI.Theme == "" {
		c.Clients.TUI.Theme = "default"
	}
}

package config

import "fmt"

type Override struct {
	ProfileName string
	Hosts       []string
	Port        int
	Keyspace    string
	Username    string
	Password    string
	SSLEnabled  *bool
	PageSize    int
	TimeoutMs   int
	WebPort     int
	Theme       string
	VimMode     *bool
}

func ApplyOverrides(config *Config, override *Override) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	if override == nil {
		return nil
	}

	if override.ProfileName != "" {
		profile, err := config.GetProfile(override.ProfileName)
		if err != nil {
			return fmt.Errorf("override profile not found: %w", err)
		}

		profileOverride := buildProfileOverride(override)
		if err := profile.MergeWith(profileOverride); err != nil {
			return fmt.Errorf("failed to merge profile override: %w", err)
		}
	}

	if override.PageSize > 0 {
		config.Defaults.PageSize = override.PageSize
	}

	if override.TimeoutMs > 0 {
		config.Defaults.TimeoutMs = override.TimeoutMs
	}

	if override.WebPort > 0 {
		config.Clients.Web.DefaultPort = override.WebPort
	}

	if override.Theme != "" {
		config.Clients.TUI.Theme = override.Theme
	}

	if override.VimMode != nil {
		config.Clients.TUI.VimMode = *override.VimMode
	}

	return config.Validate()
}

func buildProfileOverride(override *Override) *Profile {
	profileOverride := &Profile{
		Name: override.ProfileName,
	}

	if len(override.Hosts) > 0 {
		profileOverride.Hosts = override.Hosts
	}

	if override.Port > 0 {
		profileOverride.Port = override.Port
	}

	if override.Keyspace != "" {
		profileOverride.Keyspace = override.Keyspace
	}

	if override.Username != "" || override.Password != "" {
		profileOverride.Auth = &AuthConfig{}
		if override.Username != "" {
			profileOverride.Auth.Username = override.Username
		}
		if override.Password != "" {
			profileOverride.Auth.Password = override.Password
		}
	}

	if override.SSLEnabled != nil {
		profileOverride.SSL = &SSLConfig{
			Enabled: *override.SSLEnabled,
		}
	}

	return profileOverride
}

func MergeConfigs(base *Config, override *Config) (*Config, error) {
	if base == nil {
		return nil, fmt.Errorf("base config is nil")
	}
	if override == nil {
		return base.clone(), nil
	}

	merged := base.clone()

	for _, overrideProfile := range override.Profiles {
		idx, existingProfile := merged.FindProfile(overrideProfile.Name)
		if idx != -1 {
			if err := existingProfile.MergeWith(&overrideProfile); err != nil {
				return nil, fmt.Errorf("failed to merge profile %s: %w", overrideProfile.Name, err)
			}
		} else {
			if err := merged.AddProfile(overrideProfile); err != nil {
				return nil, fmt.Errorf("failed to add profile %s: %w", overrideProfile.Name, err)
			}
		}
	}

	if override.Defaults.PageSize > 0 {
		merged.Defaults.PageSize = override.Defaults.PageSize
	}

	if override.Defaults.TimeoutMs > 0 {
		merged.Defaults.TimeoutMs = override.Defaults.TimeoutMs
	}

	if override.Defaults.DefaultProfile != "" {
		merged.Defaults.DefaultProfile = override.Defaults.DefaultProfile
	}

	if override.Clients.Web.DefaultPort > 0 {
		merged.Clients.Web.DefaultPort = override.Clients.Web.DefaultPort
	}

	merged.Clients.Web.AutoOpenBrowser = override.Clients.Web.AutoOpenBrowser

	if override.Clients.TUI.Theme != "" {
		merged.Clients.TUI.Theme = override.Clients.TUI.Theme
	}

	merged.Clients.TUI.VimMode = override.Clients.TUI.VimMode

	if err := merged.Validate(); err != nil {
		return nil, fmt.Errorf("merged config validation failed: %w", err)
	}

	return merged, nil
}

func (c *Config) clone() *Config {
	clone := &Config{
		Version:  c.Version,
		Profiles: make([]Profile, len(c.Profiles)),
		Defaults: c.Defaults,
		Clients:  c.Clients,
	}

	for i := range c.Profiles {
		clone.Profiles[i] = *c.Profiles[i].Clone()
	}

	return clone
}

func ApplyProfileOverride(profile *Profile, override *Override) error {
	if profile == nil {
		return fmt.Errorf("profile is nil")
	}
	if override == nil {
		return nil
	}

	profileOverride := buildProfileOverride(override)
	if err := profile.MergeWith(profileOverride); err != nil {
		return err
	}

	return profile.Validate()
}

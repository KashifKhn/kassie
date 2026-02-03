package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	ErrInvalidVarSyntax = errors.New("invalid environment variable syntax")
	ErrVarNotFound      = errors.New("environment variable not found")
	ErrCircularRef      = errors.New("circular reference in environment variables")
)

var envVarRegex = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)

func InterpolateEnvVars(value string) (string, error) {
	return interpolateEnvVarsRecursive(value, make(map[string]bool), 0)
}

func interpolateEnvVarsRecursive(value string, visited map[string]bool, depth int) (string, error) {
	if depth > 10 {
		return "", ErrCircularRef
	}

	if !strings.Contains(value, "${") {
		return value, nil
	}

	result := value
	matches := envVarRegex.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		fullMatch := match[0]
		varName := match[1]

		if visited[varName] {
			return "", ErrCircularRef
		}

		envValue, exists := os.LookupEnv(varName)
		if !exists {
			return "", fmt.Errorf("%w: %s", ErrVarNotFound, varName)
		}

		visited[varName] = true

		interpolated, err := interpolateEnvVarsRecursive(envValue, visited, depth+1)
		if err != nil {
			return "", err
		}

		result = strings.ReplaceAll(result, fullMatch, interpolated)

		delete(visited, varName)
	}

	return result, nil
}

func InterpolateProfile(profile *Profile) error {
	if profile.Auth != nil {
		if strings.Contains(profile.Auth.Password, "${") {
			interpolated, err := InterpolateEnvVars(profile.Auth.Password)
			if err != nil {
				return fmt.Errorf("failed to interpolate password: %w", err)
			}
			profile.Auth.Password = interpolated
		}

		if strings.Contains(profile.Auth.Username, "${") {
			interpolated, err := InterpolateEnvVars(profile.Auth.Username)
			if err != nil {
				return fmt.Errorf("failed to interpolate username: %w", err)
			}
			profile.Auth.Username = interpolated
		}
	}

	if profile.SSL != nil {
		if strings.Contains(profile.SSL.CertPath, "${") {
			interpolated, err := InterpolateEnvVars(profile.SSL.CertPath)
			if err != nil {
				return fmt.Errorf("failed to interpolate cert path: %w", err)
			}
			profile.SSL.CertPath = interpolated
		}

		if strings.Contains(profile.SSL.KeyPath, "${") {
			interpolated, err := InterpolateEnvVars(profile.SSL.KeyPath)
			if err != nil {
				return fmt.Errorf("failed to interpolate key path: %w", err)
			}
			profile.SSL.KeyPath = interpolated
		}

		if strings.Contains(profile.SSL.CAPath, "${") {
			interpolated, err := InterpolateEnvVars(profile.SSL.CAPath)
			if err != nil {
				return fmt.Errorf("failed to interpolate ca path: %w", err)
			}
			profile.SSL.CAPath = interpolated
		}
	}

	return nil
}

func InterpolateConfig(config *Config) error {
	for i := range config.Profiles {
		if err := InterpolateProfile(&config.Profiles[i]); err != nil {
			return fmt.Errorf("failed to interpolate profile %s: %w", config.Profiles[i].Name, err)
		}
	}
	return nil
}

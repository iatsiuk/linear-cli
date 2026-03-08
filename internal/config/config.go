package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configFileName = "config.yaml"
	configDirName  = "linear-cli"
	envAPIKey      = "LINEAR_API_KEY"
	// envConfigDir allows overriding config directory (used in tests)
	envConfigDir = "LINEAR_CONFIG_DIR"
)

// Config holds CLI configuration values.
type Config struct {
	APIKey      string `yaml:"api_key"`
	DefaultTeam string `yaml:"default_team,omitempty"`
}

// configPath returns the path to the config file.
func configPath() (string, error) {
	if dir := os.Getenv(envConfigDir); dir != "" {
		return filepath.Join(dir, configFileName), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(base, configDirName, configFileName), nil
}

// Load reads config from disk and applies env overrides.
// Returns empty Config if file does not exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err == nil && len(data) > 0 {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}

	if key := os.Getenv(envAPIKey); key != "" {
		cfg.APIKey = key
	}

	return cfg, nil
}

// Save writes cfg to the config file, creating the directory if needed.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// Package config loads and validates the ec2s multi-account configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Account describes one AWS account/profile ec2s should aggregate instances from.
type Account struct {
	Name    string   `yaml:"name"`
	Profile string   `yaml:"profile"`
	Region  string   `yaml:"region,omitempty"`
	Regions []string `yaml:"regions,omitempty"`
}

// Config is the top-level ec2s configuration file schema.
type Config struct {
	Accounts []Account `yaml:"accounts"`
}

// Load resolves the config file path and loads it. If no config file is found
// at any candidate location, it falls back to auto-discovery (see discover.go)
// instead of returning an error, so ec2s is usable with zero configuration.
func Load(explicitPath string) (*Config, error) {
	path, err := resolvePath(explicitPath)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return Fallback()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %q: %w", path, err)
	}

	if err := cfg.normalizeAndValidate(); err != nil {
		return nil, fmt.Errorf("config %q: %w", path, err)
	}

	return &cfg, nil
}

// resolvePath returns the first existing config path in priority order, or ""
// if none exist (signaling the caller should fall back to auto-discovery).
// Priority: explicit --config flag > $EC2S_CONFIG > ./ec2s.yaml > user config dir.
func resolvePath(explicitPath string) (string, error) {
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); err != nil {
			return "", fmt.Errorf("config path %q: %w", explicitPath, err)
		}
		return explicitPath, nil
	}

	candidates := []string{}
	if env := os.Getenv("EC2S_CONFIG"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates, "ec2s.yaml")

	if userCfgDir, err := os.UserConfigDir(); err == nil {
		candidates = append(candidates, filepath.Join(userCfgDir, "ec2s", "config.yaml"))
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	return "", nil
}

// normalizeAndValidate normalizes the region/regions shorthand and validates
// account entries, failing fast since this is static user configuration.
func (c *Config) normalizeAndValidate() error {
	if len(c.Accounts) == 0 {
		return fmt.Errorf("no accounts defined")
	}

	seen := make(map[string]bool, len(c.Accounts))
	for i := range c.Accounts {
		a := &c.Accounts[i]

		if a.Name == "" {
			return fmt.Errorf("account %d: name is required", i)
		}
		if seen[a.Name] {
			return fmt.Errorf("account %q: duplicate name", a.Name)
		}
		seen[a.Name] = true

		if a.Profile == "" {
			return fmt.Errorf("account %q: profile is required", a.Name)
		}

		switch {
		case a.Region != "" && len(a.Regions) > 0:
			return fmt.Errorf("account %q: specify either region or regions, not both", a.Name)
		case a.Region != "":
			a.Regions = []string{a.Region}
			a.Region = ""
		case len(a.Regions) == 0:
			return fmt.Errorf("account %q: at least one region is required", a.Name)
		}
	}

	return nil
}

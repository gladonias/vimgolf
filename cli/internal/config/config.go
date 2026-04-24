package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

const (
	DefaultHost = "https://www.vimgolf.com"
	dirName     = ".vimgolf"
	configFile  = "config.yaml"
)

type Config struct {
	Key string `yaml:"key"`
}

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, dirName)
}

func PutDir() string {
	return filepath.Join(Dir(), "put")
}

func Host() string {
	if h := os.Getenv("GOLFHOST"); h != "" {
		return h
	}
	return DefaultHost
}

func Load() (*Config, error) {
	data, err := os.ReadFile(filepath.Join(Dir(), configFile))
	if err != nil {
		return nil, fmt.Errorf("config not found — run 'vimgolf setup' first")
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Key == "" {
		return nil, fmt.Errorf("API key not set — run 'vimgolf setup' first")
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(Dir(), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := os.MkdirAll(PutDir(), 0o755); err != nil {
		return fmt.Errorf("failed to create put directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(filepath.Join(Dir(), configFile), data, 0o644)
}

var keyPattern = regexp.MustCompile(`^[a-zA-Z0-9]{32}$`)

func ValidateKey(key string) bool {
	return keyPattern.MatchString(key)
}

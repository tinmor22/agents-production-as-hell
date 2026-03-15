package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrNotConfigured is returned when no config file exists.
var ErrNotConfigured = errors.New("statcard no esta configurado: ejecuta 'statcard init --api-key <clave>'")

const (
	dirName      = ".statcard"
	fileName     = "config.json"
	defaultPlan  = "free"
	defaultLimit = 5
	dirPerm      = 0700
	filePerm     = 0600
)

// Config holds persistent StatCard configuration.
type Config struct {
	APIKey     string `json:"api_key"`
	Watermark  string `json:"watermark"`
	Plan       string `json:"plan"`
	DailyLimit int    `json:"daily_limit"`
}

// DefaultDir returns the path to the ~/.statcard directory.
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, dirName), nil
}

// DefaultPath returns the full path to config.json.
func DefaultPath() (string, error) {
	dir, err := DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// Load reads config from the given path.
// If path is empty, it uses DefaultPath.
func Load(path string) (*Config, error) {
	if path == "" {
		p, err := DefaultPath()
		if err != nil {
			return nil, fmt.Errorf("config path: %w", err)
		}
		path = p
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotConfigured
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyDefaults()
	return &cfg, nil
}

// Save writes config to the given path.
// If path is empty, it uses DefaultPath.
func Save(cfg *Config, path string) error {
	if path == "" {
		p, err := DefaultPath()
		if err != nil {
			return fmt.Errorf("config path: %w", err)
		}
		path = p
	}
	cfg.applyDefaults()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, filePerm); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c *Config) applyDefaults() {
	if c.Plan == "" {
		c.Plan = defaultPlan
	}
	if c.DailyLimit <= 0 {
		c.DailyLimit = defaultLimit
	}
}

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Custom error types
var (
	ErrRepoNotFound  = errors.New("repository not found")
	ErrRepoExists    = errors.New("repository already exists")
	ErrInvalidPath   = errors.New("invalid path")
	ErrNotAGitRepo   = errors.New("not a git repository")
	ErrRepoNotExists = errors.New("repository does not exist")
)

// Repository represents a tracked git repository
type Repository struct {
	Path     string     `yaml:"path" json:"path"`
	Name     string     `yaml:"name" json:"name"`
	Group    string     `yaml:"group,omitempty" json:"group,omitempty"`
	LastSync *time.Time `yaml:"last_sync,omitempty" json:"last_sync,omitempty"`
}

// Config represents the go-gitter configuration
type Config struct {
	Repositories []Repository `yaml:"repositories" json:"repositories"`
	Settings     Settings     `yaml:"settings,omitempty" json:"settings,omitempty"`
}

// Settings holds application settings
type Settings struct {
	AutoFetch   bool `yaml:"auto_fetch" json:"auto_fetch"`
	SyncTimeout int  `yaml:"sync_timeout" json:"sync_timeout"`
}

// Cached config for performance
var (
	configCache     *Config
	configCacheOnce sync.Once
	configCacheErr  error
)

// GetConfig returns a cached config or loads a new one
func GetConfig() (*Config, error) {
	configCacheOnce.Do(func() {
		configCache, configCacheErr = Load()
	})
	if configCacheErr != nil {
		return nil, configCacheErr
	}
	return configCache, nil
}

// RefreshConfig reloads config from disk and updates cache
func RefreshConfig() (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	configCache = cfg
	configCacheErr = nil
	return cfg, nil
}

// InvalidateCache clears the config cache
func InvalidateCache() {
	configCacheOnce = sync.Once{}
	configCache = nil
	configCacheErr = nil
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		Repositories: []Repository{},
		Settings: Settings{
			AutoFetch:   false,
			SyncTimeout: 300,
		},
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configHome, "go-gitter", "config.yaml")
}

// EnsureDir ensures the config directory exists
func EnsureDir() error {
	configPath := ConfigPath()
	dir := filepath.Dir(configPath)
	return os.MkdirAll(dir, 0755)
}

// Load reads the config from disk (supports both YAML and JSON)
func Load() (*Config, error) {
	configPath := ConfigPath()

	// If file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config

	// Detect format from extension
	if strings.HasSuffix(configPath, ".json") {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	} else {
		// Default to YAML
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	}

	if cfg.Repositories == nil {
		cfg.Repositories = []Repository{}
	}
	if cfg.Settings.SyncTimeout == 0 {
		cfg.Settings.SyncTimeout = 300
	}

	return &cfg, nil
}

// Save writes the config to disk (supports both YAML and JSON)
func Save(cfg *Config) error {
	if err := EnsureDir(); err != nil {
		return err
	}

	configPath := ConfigPath()
	var data []byte
	var err error

	if strings.HasSuffix(configPath, ".json") {
		data, err = json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
	} else {
		data, err = yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
	}

	// Use 0600 for security - config might contain sensitive paths
	return os.WriteFile(configPath, data, 0600)
}

// ValidateRepo checks if a repository path is valid
func ValidateRepo(path string) error {
	if path == "" {
		return ErrInvalidPath
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrRepoNotExists
		}
		return err
	}

	if !info.IsDir() {
		return ErrInvalidPath
	}

	// Check if it's a git repository
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return ErrNotAGitRepo
	}

	return nil
}

// ValidateConfig checks all repositories in config and returns issues
func ValidateConfig(cfg *Config) []error {
	var errors []error

	for _, repo := range cfg.Repositories {
		if err := ValidateRepo(repo.Path); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// GetRepoByName finds a repository by name
func GetRepoByName(cfg *Config, name string) *Repository {
	for i := range cfg.Repositories {
		if cfg.Repositories[i].Name == name {
			return &cfg.Repositories[i]
		}
	}
	return nil
}

// GetRepoByPath finds a repository by path
func GetRepoByPath(cfg *Config, path string) *Repository {
	for i := range cfg.Repositories {
		if cfg.Repositories[i].Path == path {
			return &cfg.Repositories[i]
		}
	}
	return nil
}

// AddRepo adds a repository to the config
func AddRepo(cfg *Config, path, name, group string) error {
	// Validate the path first
	if err := ValidateRepo(path); err != nil {
		return err
	}

	// Check if repo already exists by path
	for _, r := range cfg.Repositories {
		if r.Path == path {
			return ErrRepoExists
		}
	}

	// Check for duplicate name
	for _, r := range cfg.Repositories {
		if r.Name == name {
			return fmt.Errorf("%w: name '%s' already used", ErrRepoExists, name)
		}
	}

	repo := Repository{
		Path:  path,
		Name:  name,
		Group: group,
	}

	cfg.Repositories = append(cfg.Repositories, repo)
	return Save(cfg)
}

// RemoveRepo removes a repository by path or name
func RemoveRepo(cfg *Config, identifier string) error {
	repos := []Repository{}
	found := false

	for _, r := range cfg.Repositories {
		if r.Path != identifier && r.Name != identifier {
			repos = append(repos, r)
		} else {
			found = true
		}
	}

	if !found {
		return nil // Not found, no error
	}

	cfg.Repositories = repos
	return Save(cfg)
}

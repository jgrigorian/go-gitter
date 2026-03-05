package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Repository represents a tracked git repository
type Repository struct {
	Path     string     `yaml:"path"`
	Name     string     `yaml:"name"`
	Group    string     `yaml:"group,omitempty"`
	LastSync *time.Time `yaml:"last_sync,omitempty"`
}

// Config represents the go-gitter configuration
type Config struct {
	Repositories []Repository `yaml:"repositories"`
	Settings     Settings     `yaml:"settings,omitempty"`
}

// Settings holds application settings
type Settings struct {
	AutoFetch   bool `yaml:"auto_fetch"`
	SyncTimeout int  `yaml:"sync_timeout"`
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

// Load reads the config from disk
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
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Repositories == nil {
		cfg.Repositories = []Repository{}
	}
	if cfg.Settings.SyncTimeout == 0 {
		cfg.Settings.SyncTimeout = 300
	}

	return &cfg, nil
}

// Save writes the config to disk
func Save(cfg *Config) error {
	if err := EnsureDir(); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), data, 0644)
}

// AddRepo adds a repository to the config
func AddRepo(cfg *Config, path, name, group string) error {
	// Check if repo already exists
	for _, r := range cfg.Repositories {
		if r.Path == path {
			return nil // Already exists
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

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.Repositories) != 0 {
		t.Errorf("expected empty repositories, got %d", len(cfg.Repositories))
	}

	if cfg.Settings.SyncTimeout != 300 {
		t.Errorf("expected default timeout 300, got %d", cfg.Settings.SyncTimeout)
	}

	if cfg.Settings.AutoFetch != false {
		t.Errorf("expected AutoFetch false, got %v", cfg.Settings.AutoFetch)
	}
}

func TestConfigPath(t *testing.T) {
	// Test default path
	os.Unsetenv("XDG_CONFIG_HOME")
	path := ConfigPath()

	expected := filepath.Join(os.Getenv("HOME"), ".config", "go-gitter", "config.yaml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}

	// Test with XDG_CONFIG_HOME
	os.Setenv("XDG_CONFIG_HOME", "/custom/config")
	defer os.Unsetenv("XDG_CONFIG_HOME")

	path = ConfigPath()
	expected = filepath.Join("/custom/config", "go-gitter", "config.yaml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestRepositoryOperations(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Save original config path and restore after test
	origPath := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if origPath != "" {
			os.Setenv("XDG_CONFIG_HOME", origPath)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := DefaultConfig()

	// Test adding a repo (without validation since we can't create a real git repo in test)
	repo := Repository{
		Path:  "/test/path",
		Name:  "test-repo",
		Group: "test-group",
	}
	cfg.Repositories = append(cfg.Repositories, repo)

	if len(cfg.Repositories) != 1 {
		t.Errorf("expected 1 repository, got %d", len(cfg.Repositories))
	}

	if cfg.Repositories[0].Name != "test-repo" {
		t.Errorf("expected name 'test-repo', got '%s'", cfg.Repositories[0].Name)
	}

	// Test GetRepoByName
	found := GetRepoByName(cfg, "test-repo")
	if found == nil {
		t.Error("expected to find repo by name")
	}

	found = GetRepoByName(cfg, "nonexistent")
	if found != nil {
		t.Error("expected nil for nonexistent repo")
	}

	// Test GetRepoByPath
	found = GetRepoByPath(cfg, "/test/path")
	if found == nil {
		t.Error("expected to find repo by path")
	}

	found = GetRepoByPath(cfg, "/nonexistent")
	if found != nil {
		t.Error("expected nil for nonexistent path")
	}
}

func TestRemoveRepo(t *testing.T) {
	tmpDir := t.TempDir()

	origPath := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if origPath != "" {
			os.Setenv("XDG_CONFIG_HOME", origPath)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := DefaultConfig()

	// Add test repos
	now := time.Now()
	cfg.Repositories = []Repository{
		{Path: "/path/1", Name: "repo1", Group: "group1", LastSync: &now},
		{Path: "/path/2", Name: "repo2", Group: "group2"},
		{Path: "/path/3", Name: "repo3"},
	}

	// Remove by name
	err := RemoveRepo(cfg, "repo1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(cfg.Repositories) != 2 {
		t.Errorf("expected 2 repos, got %d", len(cfg.Repositories))
	}

	// Remove by path
	err = RemoveRepo(cfg, "/path/2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(cfg.Repositories) != 1 {
		t.Errorf("expected 1 repo, got %d", len(cfg.Repositories))
	}

	// Remove nonexistent (should not error)
	err = RemoveRepo(cfg, "nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	origPath := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origPath != "" {
			os.Setenv("XDG_CONFIG_HOME", origPath)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	now := time.Now()
	cfg := &Config{
		Repositories: []Repository{
			{Path: "/test/path", Name: "test", Group: "test", LastSync: &now},
		},
		Settings: Settings{
			AutoFetch:   true,
			SyncTimeout: 600,
		},
	}

	err := Save(cfg)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if len(loaded.Repositories) != 1 {
		t.Errorf("expected 1 repo, got %d", len(loaded.Repositories))
	}

	if loaded.Repositories[0].Name != "test" {
		t.Errorf("expected name 'test', got '%s'", loaded.Repositories[0].Name)
	}

	if !loaded.Settings.AutoFetch {
		t.Error("expected AutoFetch to be true")
	}

	if loaded.Settings.SyncTimeout != 600 {
		t.Errorf("expected timeout 600, got %d", loaded.Settings.SyncTimeout)
	}
}

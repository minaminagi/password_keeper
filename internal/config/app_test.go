package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppConfigCreatesDefaultConfig(t *testing.T) {
	t.Setenv(configEnvName, "")
	t.Setenv(dbPathEnvName, "")

	configPath := filepath.Join(t.TempDir(), "config.toml")

	loaded, err := LoadAppConfig(configPath)
	if err != nil {
		t.Fatalf("LoadAppConfig() error = %v", err)
	}

	wantDBPath := filepath.Join(filepath.Dir(configPath), "password.db")
	if loaded.ConfigPath != configPath {
		t.Fatalf("ConfigPath = %q, want %q", loaded.ConfigPath, configPath)
	}
	if loaded.DBPath != wantDBPath {
		t.Fatalf("DBPath = %q, want %q", loaded.DBPath, wantDBPath)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("default config was not created: %v", err)
	}
}

func TestLoadAppConfigResolvesRelativeDBPathFromConfigDirectory(t *testing.T) {
	t.Setenv(configEnvName, "")
	t.Setenv(dbPathEnvName, "")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "nested", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(configPath, []byte("DBPath = \"data/vault.db\"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	loaded, err := LoadAppConfig(configPath)
	if err != nil {
		t.Fatalf("LoadAppConfig() error = %v", err)
	}

	wantDBPath := filepath.Join(filepath.Dir(configPath), "data", "vault.db")
	if loaded.DBPath != wantDBPath {
		t.Fatalf("DBPath = %q, want %q", loaded.DBPath, wantDBPath)
	}
}

func TestLoadAppConfigAllowsDBPathEnvOverride(t *testing.T) {
	t.Setenv(configEnvName, "")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	envDBPath := filepath.Join(dir, "override", "password.db")
	t.Setenv(dbPathEnvName, envDBPath)

	loaded, err := LoadAppConfig(configPath)
	if err != nil {
		t.Fatalf("LoadAppConfig() error = %v", err)
	}

	if loaded.DBPath != envDBPath {
		t.Fatalf("DBPath = %q, want %q", loaded.DBPath, envDBPath)
	}
}

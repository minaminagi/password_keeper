package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	configEnvName = "PASSWORD_KEEPER_CONFIG"
	dbPathEnvName = "PASSWORD_KEEPER_DB_PATH"
	appDirName    = "password_keeper"
)

type AppConfig struct {
	DBPath string `toml:"DBPath"`
}

type LoadedAppConfig struct {
	AppConfig
	ConfigPath string
}

func LoadAppConfig(explicitPath string) (LoadedAppConfig, error) {
	configPath, err := resolveConfigPath(explicitPath)
	if err != nil {
		return LoadedAppConfig{}, err
	}

	if err := ensureConfigFile(configPath); err != nil {
		return LoadedAppConfig{}, err
	}

	var appConfig AppConfig
	if _, err := toml.DecodeFile(configPath, &appConfig); err != nil {
		return LoadedAppConfig{}, fmt.Errorf("read config %q: %w", configPath, err)
	}

	if envDBPath := os.Getenv(dbPathEnvName); envDBPath != "" {
		appConfig.DBPath = envDBPath
	}
	if appConfig.DBPath == "" {
		appConfig.DBPath = defaultDBPath(configPath)
	}

	dbPath, err := resolveDBPath(configPath, appConfig.DBPath)
	if err != nil {
		return LoadedAppConfig{}, err
	}
	appConfig.DBPath = dbPath

	if err := os.MkdirAll(filepath.Dir(appConfig.DBPath), 0o755); err != nil {
		return LoadedAppConfig{}, fmt.Errorf("create db directory: %w", err)
	}

	return LoadedAppConfig{
		AppConfig:  appConfig,
		ConfigPath: configPath,
	}, nil
}

func resolveConfigPath(explicitPath string) (string, error) {
	if explicitPath != "" {
		return filepath.Abs(explicitPath)
	}
	if envPath := os.Getenv(configEnvName); envPath != "" {
		return filepath.Abs(envPath)
	}
	if _, err := os.Stat("config.toml"); err == nil {
		return filepath.Abs("config.toml")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat local config: %w", err)
	}

	configDir, err := userConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.toml"), nil
}

func ensureConfigFile(configPath string) error {
	if _, err := os.Stat(configPath); err == nil {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat config %q: %w", configPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	content := fmt.Sprintf("DBPath = %q\n", defaultDBPath(configPath))
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("create config %q: %w", configPath, err)
	}
	return nil
}

func resolveDBPath(configPath string, dbPath string) (string, error) {
	if filepath.IsAbs(dbPath) {
		return filepath.Clean(dbPath), nil
	}
	return filepath.Abs(filepath.Join(filepath.Dir(configPath), dbPath))
}

func defaultDBPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "password.db")
}

func userConfigDir() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(baseDir, appDirName), nil
}

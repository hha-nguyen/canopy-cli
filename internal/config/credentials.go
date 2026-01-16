package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetCredentialsPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials"), nil
}

func SaveCredentials(apiKey string) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	credPath := filepath.Join(configDir, "credentials")

	if err := os.WriteFile(credPath, []byte(apiKey), 0600); err != nil {
		return fmt.Errorf("write credentials file: %w", err)
	}

	return nil
}

func LoadCredentials() (string, error) {
	credPath, err := GetCredentialsPath()
	if err != nil {
		return "", err
	}

	info, err := os.Stat(credPath)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("stat credentials file: %w", err)
	}

	mode := info.Mode()
	if mode&0077 != 0 {
		fmt.Fprintln(os.Stderr, "Warning: credentials file has insecure permissions. Run: chmod 600", credPath)
	}

	data, err := os.ReadFile(credPath)
	if err != nil {
		return "", fmt.Errorf("read credentials file: %w", err)
	}

	return string(data), nil
}

func ClearCredentials() error {
	credPath, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	if err := os.Remove(credPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove credentials file: %w", err)
	}

	return nil
}

func GetAPIKey() string {
	if key := os.Getenv("CANOPY_API_KEY"); key != "" {
		return key
	}

	cfg, err := Load()
	if err == nil && cfg.APIKey != "" {
		return cfg.APIKey
	}

	key, err := LoadCredentials()
	if err == nil && key != "" {
		return key
	}

	return ""
}

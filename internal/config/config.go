package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	APIURL   string         `yaml:"api_url" mapstructure:"api_url"`
	APIKey   string         `yaml:"api_key" mapstructure:"api_key"`
	Defaults DefaultsConfig `yaml:"defaults" mapstructure:"defaults"`
	Output   OutputConfig   `yaml:"output" mapstructure:"output"`
}

type DefaultsConfig struct {
	Platform  string `yaml:"platform" mapstructure:"platform"`
	Format    string `yaml:"format" mapstructure:"format"`
	Threshold string `yaml:"threshold" mapstructure:"threshold"`
	Timeout   string `yaml:"timeout" mapstructure:"timeout"`
}

type OutputConfig struct {
	Color    bool `yaml:"color" mapstructure:"color"`
	Progress bool `yaml:"progress" mapstructure:"progress"`
	Quiet    bool `yaml:"quiet" mapstructure:"quiet"`
}

func DefaultConfig() *Config {
	return &Config{
		APIURL: "https://api.canopy.app",
		Defaults: DefaultsConfig{
			Platform:  "both",
			Format:    "text",
			Threshold: "blocker",
			Timeout:   "5m",
		},
		Output: OutputConfig{
			Color:    true,
			Progress: true,
			Quiet:    false,
		},
	}
}

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, ".canopy"), nil
}

func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	configPath, err := GetConfigPath()
	if err != nil {
		return cfg, nil
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func Get(key string) string {
	return viper.GetString(key)
}

func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "api_url":
		cfg.APIURL = value
	case "api_key":
		cfg.APIKey = value
	case "default_platform", "defaults.platform":
		cfg.Defaults.Platform = value
	case "default_format", "defaults.format":
		cfg.Defaults.Format = value
	case "default_threshold", "defaults.threshold":
		cfg.Defaults.Threshold = value
	case "default_timeout", "defaults.timeout":
		cfg.Defaults.Timeout = value
	case "output.color":
		cfg.Output.Color = value == "true"
	case "output.progress":
		cfg.Output.Progress = value == "true"
	case "output.quiet":
		cfg.Output.Quiet = value == "true"
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return Save(cfg)
}

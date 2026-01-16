package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hha-nguyen/canopy-cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `View and modify CLI configuration settings.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create default config file",
	Long:  `Create a default configuration file at ~/.canopy/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(configPath); err == nil {
			fmt.Println("Config file already exists at:", configPath)
			return nil
		}

		cfg := config.DefaultConfig()
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("create config file: %w", err)
		}

		fmt.Println("Created config file at:", configPath)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value",
	Long:  `Get a configuration value by key.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var value string
		switch key {
		case "api_url":
			value = cfg.APIURL
		case "api_key":
			if cfg.APIKey != "" {
				value = maskAPIKey(cfg.APIKey)
			}
		case "default_platform", "defaults.platform":
			value = cfg.Defaults.Platform
		case "default_format", "defaults.format":
			value = cfg.Defaults.Format
		case "default_threshold", "defaults.threshold":
			value = cfg.Defaults.Threshold
		case "default_timeout", "defaults.timeout":
			value = cfg.Defaults.Timeout
		case "output.color":
			value = fmt.Sprintf("%t", cfg.Output.Color)
		case "output.progress":
			value = fmt.Sprintf("%t", cfg.Output.Progress)
		case "output.quiet":
			value = fmt.Sprintf("%t", cfg.Output.Quiet)
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}

		if value != "" {
			fmt.Println(value)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Long:  `Set a configuration value by key.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if err := config.Set(key, value); err != nil {
			return err
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all config values",
	Long:  `List all configuration values.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.APIKey != "" {
			cfg.APIKey = maskAPIKey(cfg.APIKey)
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}

		fmt.Print(string(data))
		return nil
	},
}

func maskAPIKey(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:12] + strings.Repeat("*", len(key)-12)
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

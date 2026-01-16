package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	apiKey    string
	apiURL    string
	quiet     bool
	noColor   bool
	debug     bool
	version   string
	commit    string
	buildDate string
)

var rootCmd = &cobra.Command{
	Use:   "canopy",
	Short: "Canopy CLI - Mobile app policy compliance scanner",
	Long: `Canopy CLI helps you scan mobile app projects for policy compliance issues
before submitting to Apple App Store or Google Play Store.

Use this tool in your CI/CD pipeline to catch guideline violations early.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	buildDate = d
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.canopy/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication (or CANOPY_API_KEY env)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "https://api.canopy.app", "API base URL")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("api_url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: could not find home directory:", err)
			return
		}

		viper.AddConfigPath(home + "/.canopy")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("CANOPY")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if debug {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}

func GetAPIKey() string {
	if apiKey != "" {
		return apiKey
	}
	return viper.GetString("api_key")
}

func GetAPIURL() string {
	if apiURL != "" {
		return apiURL
	}
	url := viper.GetString("api_url")
	if url == "" {
		return "https://api.canopy.app"
	}
	return url
}

func IsQuiet() bool {
	return quiet || viper.GetBool("quiet")
}

func IsNoColor() bool {
	return noColor || viper.GetBool("no_color")
}

func IsDebug() bool {
	return debug || viper.GetBool("debug")
}

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/hha-nguyen/canopy-cli/internal/api"
	"github.com/hha-nguyen/canopy-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  `Authenticate with the Canopy API and manage credentials.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and store API key",
	Long: `Store your API key for future use.

The API key will be stored securely in ~/.canopy/credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("api-key")

		if key == "" {
			fmt.Print("Enter your API key: ")
			if term.IsTerminal(int(syscall.Stdin)) {
				byteKey, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("read API key: %w", err)
				}
				key = string(byteKey)
				fmt.Println()
			} else {
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("read API key: %w", err)
				}
				key = strings.TrimSpace(input)
			}
		}

		if key == "" {
			return fmt.Errorf("API key is required")
		}

		if !strings.HasPrefix(key, "cpk_") {
			return fmt.Errorf("invalid API key format (should start with cpk_)")
		}

		client := api.NewClient(GetAPIURL(), key)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		status, err := client.GetAuthStatus(ctx)
		if err != nil {
			return fmt.Errorf("validate API key: %w", err)
		}

		if !status.Authenticated {
			return fmt.Errorf("invalid API key")
		}

		if err := config.SaveCredentials(key); err != nil {
			return fmt.Errorf("save credentials: %w", err)
		}

		color.Green("✓ Logged in successfully")
		if status.Email != "" {
			fmt.Printf("  Email: %s\n", status.Email)
		}

		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Long:  `Remove the stored API key from ~/.canopy/credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.ClearCredentials(); err != nil {
			return fmt.Errorf("clear credentials: %w", err)
		}

		color.Green("✓ Logged out successfully")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  `Check if you are currently authenticated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := config.GetAPIKey()
		if key == "" {
			fmt.Println("Not logged in")
			fmt.Println("\nTo login, run: canopy auth login")
			return nil
		}

		client := api.NewClient(GetAPIURL(), key)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		status, err := client.GetAuthStatus(ctx)
		if err != nil {
			fmt.Println("Authentication error:", err)
			return nil
		}

		if !status.Authenticated {
			fmt.Println("Stored API key is invalid")
			return nil
		}

		color.Green("✓ Authenticated")
		if status.Email != "" {
			fmt.Printf("  Email: %s\n", status.Email)
		}
		if status.UserID != "" {
			fmt.Printf("  User ID: %s\n", status.UserID)
		}

		return nil
	},
}

var authTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage API tokens",
	Long:  `Create, list, and revoke API tokens.`,
}

var authTokenCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API token",
	Long:  `Create a new API token for CLI or CI/CD use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		scopes, _ := cmd.Flags().GetString("scopes")

		if name == "" {
			return fmt.Errorf("--name is required")
		}

		key := config.GetAPIKey()
		if key == "" {
			return fmt.Errorf("not logged in. Run: canopy auth login")
		}

		client := api.NewClient(GetAPIURL(), key)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.CreateAPIKey(ctx, name, scopes)
		if err != nil {
			return fmt.Errorf("create API key: %w", err)
		}

		color.Green("✓ Created API key: %s", resp.Name)
		fmt.Println()
		color.Yellow("⚠  Save this key now - it won't be shown again!")
		fmt.Println()
		fmt.Printf("Key: %s\n", resp.Key)
		fmt.Printf("ID: %s\n", resp.ID)
		fmt.Printf("Scopes: %s\n", resp.Scopes)

		return nil
	},
}

var authTokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API tokens",
	Long:  `List all API tokens associated with your account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := config.GetAPIKey()
		if key == "" {
			return fmt.Errorf("not logged in. Run: canopy auth login")
		}

		client := api.NewClient(GetAPIURL(), key)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.ListAPIKeys(ctx)
		if err != nil {
			return fmt.Errorf("list API keys: %w", err)
		}

		if len(resp.APIKeys) == 0 {
			fmt.Println("No API keys found")
			return nil
		}

		fmt.Printf("%-36s  %-20s  %-12s  %-20s  %s\n", "ID", "Name", "Prefix", "Last Used", "Created")
		fmt.Println(strings.Repeat("-", 120))

		for _, k := range resp.APIKeys {
			lastUsed := "Never"
			if k.LastUsedAt != nil {
				lastUsed = k.LastUsedAt.Format("2006-01-02 15:04")
			}

			status := ""
			if !k.IsActive {
				status = " (inactive)"
			}

			fmt.Printf("%-36s  %-20s  %-12s  %-20s  %s%s\n",
				k.ID,
				truncate(k.Name, 20),
				k.KeyPrefix,
				lastUsed,
				k.CreatedAt.Format("2006-01-02 15:04"),
				status,
			)
		}

		return nil
	},
}

var authTokenRevokeCmd = &cobra.Command{
	Use:   "revoke <id>",
	Short: "Revoke an API token",
	Long:  `Revoke an API token by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		key := config.GetAPIKey()
		if key == "" {
			return fmt.Errorf("not logged in. Run: canopy auth login")
		}

		client := api.NewClient(GetAPIURL(), key)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := client.RevokeAPIKey(ctx, id); err != nil {
			return fmt.Errorf("revoke API key: %w", err)
		}

		color.Green("✓ Revoked API key: %s", id)
		return nil
	},
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func init() {
	rootCmd.AddCommand(authCmd)

	authLoginCmd.Flags().String("api-key", "", "API key to store")
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)

	authCmd.AddCommand(authTokenCmd)
	authTokenCreateCmd.Flags().StringP("name", "n", "", "Name for the API key")
	authTokenCreateCmd.Flags().StringP("scopes", "s", "", "Comma-separated list of scopes")
	authTokenCmd.AddCommand(authTokenCreateCmd)
	authTokenCmd.AddCommand(authTokenListCmd)
	authTokenCmd.AddCommand(authTokenRevokeCmd)
}

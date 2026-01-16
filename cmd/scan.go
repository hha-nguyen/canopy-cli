package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hha-nguyen/canopy-cli/internal/api"
	"github.com/hha-nguyen/canopy-cli/internal/archive"
	"github.com/hha-nguyen/canopy-cli/internal/config"
	"github.com/hha-nguyen/canopy-cli/internal/exit"
	"github.com/hha-nguyen/canopy-cli/internal/output"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a project for policy violations",
	Long: `Scan a project directory or archive file for mobile app policy violations.

If a directory is provided, it will be compressed before upload.
Supported archive formats: .zip, .tar.gz, .tgz`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

var (
	scanPlatform   string
	scanFormat     string
	scanOutput     string
	scanThreshold  string
	scanProjectID  string
	scanTimeout    time.Duration
	scanNoProgress bool
	scanFailOnErr  bool
)

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringVarP(&scanPlatform, "platform", "p", "both", "Target platform: apple, google, both")
	scanCmd.Flags().StringVarP(&scanFormat, "format", "f", "text", "Output format: text, json, sarif")
	scanCmd.Flags().StringVarP(&scanOutput, "output", "o", "", "Write output to file instead of stdout")
	scanCmd.Flags().StringVarP(&scanThreshold, "threshold", "t", "blocker", "Minimum severity to fail: blocker, high, medium, low")
	scanCmd.Flags().StringVar(&scanProjectID, "project", "", "Associate scan with existing project ID")
	scanCmd.Flags().DurationVar(&scanTimeout, "timeout", 5*time.Minute, "Scan timeout")
	scanCmd.Flags().BoolVar(&scanNoProgress, "no-progress", false, "Disable progress updates")
	scanCmd.Flags().BoolVar(&scanFailOnErr, "fail-on-error", true, "Exit with error if scan fails")
}

func runScan(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	apiKey := config.GetAPIKey()
	if apiKey == "" {
		if key := GetAPIKey(); key != "" {
			apiKey = key
		}
	}

	if apiKey == "" {
		return fmt.Errorf("authentication required. Run: canopy auth login")
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("access path: %w", err)
	}

	var archivePath string
	var cleanup func()

	if info.IsDir() {
		if !IsQuiet() {
			fmt.Printf("Compressing %s...\n", absPath)
		}

		tmpFile, err := os.CreateTemp("", "canopy-*.tar.gz")
		if err != nil {
			return fmt.Errorf("create temp file: %w", err)
		}
		tmpFile.Close()

		archivePath = tmpFile.Name()
		cleanup = func() { os.Remove(archivePath) }

		if err := archive.CompressDirectory(absPath, archivePath, nil); err != nil {
			cleanup()
			return fmt.Errorf("compress directory: %w", err)
		}
	} else {
		if !archive.IsArchive(absPath) {
			return fmt.Errorf("unsupported file type. Use .zip or .tar.gz, or provide a directory path")
		}

		if _, err := archive.ValidateArchive(absPath); err != nil {
			return err
		}

		archivePath = absPath
		cleanup = func() {}
	}
	defer cleanup()

	client := api.NewClient(GetAPIURL(), apiKey)

	platform := parsePlatform(scanPlatform)

	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	if !IsQuiet() {
		fmt.Printf("Uploading to Canopy...\n")
	}

	scanResp, err := client.CreateScan(ctx, archivePath, platform)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 401 || apiErr.StatusCode == 403 {
				return fmt.Errorf("authentication failed: %s", apiErr.Message)
			}
		}
		return fmt.Errorf("create scan: %w", err)
	}

	if !IsQuiet() {
		fmt.Printf("Scan started: %s\n", scanResp.ID)
	}

	var bar *progressbar.ProgressBar
	if !scanNoProgress && !IsQuiet() {
		bar = progressbar.NewOptions(100,
			progressbar.OptionSetDescription("Scanning"),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowCount(),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerHead:    ">",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
	}

	var result *api.ScanResult
	pollTicker := time.NewTicker(2 * time.Second)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("scan timeout")
		case <-pollTicker.C:
			result, err = client.GetScan(ctx, scanResp.ID)
			if err != nil {
				return fmt.Errorf("get scan status: %w", err)
			}

			switch result.Status {
			case "COMPLETED":
				if bar != nil {
					bar.Set(100)
					fmt.Println()
				}
				return outputResults(result)

			case "FAILED":
				if bar != nil {
					fmt.Println()
				}
				if scanFailOnErr {
					return fmt.Errorf("scan failed")
				}
				return outputResults(result)

			case "PROCESSING":
				if bar != nil && result.Summary != nil {
					progress := 50
					bar.Set(progress)
				}

			case "EVALUATING":
				if bar != nil {
					bar.Set(75)
				}
			}
		}
	}
}

func outputResults(result *api.ScanResult) error {
	formatter := output.NewFormatter(scanFormat, IsNoColor())

	formatted, err := formatter.Format(result)
	if err != nil {
		return fmt.Errorf("format output: %w", err)
	}

	if scanOutput != "" {
		if err := os.WriteFile(scanOutput, formatted, 0644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
		if !IsQuiet() {
			color.Green("✓ Results written to %s", scanOutput)
		}
	} else {
		fmt.Print(string(formatted))
	}

	summary := exit.ScanSummary{}
	if result.Summary != nil {
		summary.Blocker = result.Summary.Blocker
		summary.High = result.Summary.High
		summary.Medium = result.Summary.Medium
		summary.Low = result.Summary.Low
	}

	exitCode := exit.DetermineExitCode(summary, exit.ParseThreshold(scanThreshold))
	if exitCode != exit.Success {
		os.Exit(exitCode)
	}

	return nil
}

func parsePlatform(s string) api.Platform {
	switch strings.ToLower(s) {
	case "apple", "ios":
		return api.PlatformApple
	case "google", "android":
		return api.PlatformGoogle
	default:
		return api.PlatformBoth
	}
}

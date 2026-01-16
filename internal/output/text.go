package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/hha-nguyen/canopy-cli/internal/api"
)

type TextFormatter struct {
	noColor bool
}

func NewTextFormatter(noColor bool) *TextFormatter {
	if noColor {
		color.NoColor = true
	}
	return &TextFormatter{noColor: noColor}
}

func (f *TextFormatter) Format(result *api.ScanResult) ([]byte, error) {
	var sb strings.Builder

	sb.WriteString("Canopy Scan Results\n")
	sb.WriteString("==================\n\n")

	sb.WriteString(fmt.Sprintf("Scan ID: %s\n", result.ID))
	sb.WriteString(fmt.Sprintf("Platform: %s\n", formatPlatform(result.Platform)))
	sb.WriteString(fmt.Sprintf("Policy Version: %s\n", result.PolicyVersion))
	sb.WriteString(fmt.Sprintf("Duration: %dms\n", result.DurationMs))
	sb.WriteString("\n")

	if result.Summary != nil {
		sb.WriteString("Summary\n")
		sb.WriteString("-------\n")

		if result.RiskAssessment != nil {
			riskColor := getRiskColor(result.RiskAssessment.Score)
			sb.WriteString(fmt.Sprintf("Risk Score: %s%d/100 (%s)%s\n",
				riskColor,
				result.RiskAssessment.Score,
				result.RiskAssessment.Interpretation,
				colorReset()))
		}

		sb.WriteString(fmt.Sprintf("Total Issues: %d\n", result.Summary.Total))
		if result.Summary.Blocker > 0 {
			sb.WriteString(fmt.Sprintf("  %sBLOCKER: %d%s\n", colorRed(), result.Summary.Blocker, colorReset()))
		}
		if result.Summary.High > 0 {
			sb.WriteString(fmt.Sprintf("  %sHIGH: %d%s\n", colorYellow(), result.Summary.High, colorReset()))
		}
		if result.Summary.Medium > 0 {
			sb.WriteString(fmt.Sprintf("  MEDIUM: %d\n", result.Summary.Medium))
		}
		if result.Summary.Low > 0 {
			sb.WriteString(fmt.Sprintf("  LOW: %d\n", result.Summary.Low))
		}
		sb.WriteString("\n")
	}

	if len(result.Findings) > 0 {
		sb.WriteString("Issues\n")
		sb.WriteString("------\n\n")

		for _, finding := range result.Findings {
			icon := getSeverityIcon(finding.Severity)
			severityColor := getSeverityColor(finding.Severity)

			sb.WriteString(fmt.Sprintf("%s %s%s%s: %s - %s\n",
				icon,
				severityColor,
				finding.Severity,
				colorReset(),
				finding.RuleCode,
				finding.RuleName,
			))
			sb.WriteString(fmt.Sprintf("   %s\n", finding.Message))

			if finding.FilePath != "" {
				sb.WriteString(fmt.Sprintf("   File: %s\n", finding.FilePath))
			}

			if finding.Remediation != nil && finding.Remediation.Template != "" {
				sb.WriteString(fmt.Sprintf("   Fix: %s\n", finding.Remediation.Template))
			}

			if finding.DocsURL != "" {
				sb.WriteString(fmt.Sprintf("   Docs: %s\n", finding.DocsURL))
			}

			sb.WriteString("\n")
		}
	}

	blockerCount := 0
	if result.Summary != nil {
		blockerCount = result.Summary.Blocker
	}

	if blockerCount > 0 {
		sb.WriteString(fmt.Sprintf("%sScan completed with %d BLOCKER issues. Fix these before submission.%s\n",
			colorRed(), blockerCount, colorReset()))
	} else {
		sb.WriteString(fmt.Sprintf("%sScan completed with no BLOCKER issues.%s\n",
			colorGreen(), colorReset()))
	}

	return []byte(sb.String()), nil
}

func formatPlatform(platform string) string {
	switch strings.ToUpper(platform) {
	case "APPLE":
		return "Apple App Store"
	case "GOOGLE":
		return "Google Play"
	case "BOTH":
		return "Apple App Store, Google Play"
	default:
		return platform
	}
}

func getSeverityIcon(severity string) string {
	switch strings.ToUpper(severity) {
	case "BLOCKER":
		return "❌"
	case "HIGH":
		return "⚠️"
	case "MEDIUM":
		return "⚡"
	case "LOW":
		return "ℹ️"
	default:
		return "•"
	}
}

func getSeverityColor(severity string) string {
	switch strings.ToUpper(severity) {
	case "BLOCKER":
		return colorRed()
	case "HIGH":
		return colorYellow()
	case "MEDIUM":
		return colorCyan()
	default:
		return ""
	}
}

func getRiskColor(score int) string {
	if score >= 80 {
		return colorRed()
	} else if score >= 50 {
		return colorYellow()
	}
	return colorGreen()
}

func colorRed() string {
	if color.NoColor {
		return ""
	}
	return "\033[31m"
}

func colorYellow() string {
	if color.NoColor {
		return ""
	}
	return "\033[33m"
}

func colorGreen() string {
	if color.NoColor {
		return ""
	}
	return "\033[32m"
}

func colorCyan() string {
	if color.NoColor {
		return ""
	}
	return "\033[36m"
}

func colorReset() string {
	if color.NoColor {
		return ""
	}
	return "\033[0m"
}

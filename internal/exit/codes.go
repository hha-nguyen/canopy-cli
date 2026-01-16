package exit

import "strings"

const (
	Success           = 0
	IssuesFound       = 1
	ScanFailed        = 2
	AuthenticationErr = 3
	NetworkErr        = 4
	InvalidArgs       = 5
	Timeout           = 6
)

type Threshold string

const (
	ThresholdBlocker Threshold = "blocker"
	ThresholdHigh    Threshold = "high"
	ThresholdMedium  Threshold = "medium"
	ThresholdLow     Threshold = "low"
)

func ParseThreshold(s string) Threshold {
	switch strings.ToLower(s) {
	case "blocker":
		return ThresholdBlocker
	case "high":
		return ThresholdHigh
	case "medium":
		return ThresholdMedium
	case "low":
		return ThresholdLow
	default:
		return ThresholdBlocker
	}
}

type ScanSummary struct {
	Blocker int
	High    int
	Medium  int
	Low     int
}

func DetermineExitCode(summary ScanSummary, threshold Threshold) int {
	switch threshold {
	case ThresholdBlocker:
		if summary.Blocker > 0 {
			return IssuesFound
		}
	case ThresholdHigh:
		if summary.Blocker > 0 || summary.High > 0 {
			return IssuesFound
		}
	case ThresholdMedium:
		if summary.Blocker > 0 || summary.High > 0 || summary.Medium > 0 {
			return IssuesFound
		}
	case ThresholdLow:
		if summary.Blocker > 0 || summary.High > 0 || summary.Medium > 0 || summary.Low > 0 {
			return IssuesFound
		}
	}

	return Success
}

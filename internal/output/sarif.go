package output

import (
	"encoding/json"
	"strings"

	"github.com/hha-nguyen/canopy-cli/internal/api"
)

const sarifSchema = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"

type SARIFFormatter struct{}

func NewSARIFFormatter() *SARIFFormatter {
	return &SARIFFormatter{}
}

type SARIFReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

type SARIFDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationUri string      `json:"informationUri"`
	Rules          []SARIFRule `json:"rules"`
}

type SARIFRule struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	ShortDescription SARIFMessage    `json:"shortDescription"`
	FullDescription  SARIFMessage    `json:"fullDescription,omitempty"`
	HelpUri          string          `json:"helpUri,omitempty"`
	DefaultConfig    SARIFRuleConfig `json:"defaultConfiguration"`
}

type SARIFRuleConfig struct {
	Level string `json:"level"`
}

type SARIFMessage struct {
	Text string `json:"text"`
}

type SARIFResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   SARIFMessage    `json:"message"`
	Locations []SARIFLocation `json:"locations,omitempty"`
}

type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
}

type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

func (f *SARIFFormatter) Format(result *api.ScanResult) ([]byte, error) {
	rulesMap := make(map[string]SARIFRule)
	var results []SARIFResult

	for _, finding := range result.Findings {
		if _, exists := rulesMap[finding.RuleCode]; !exists {
			rulesMap[finding.RuleCode] = SARIFRule{
				ID:   finding.RuleCode,
				Name: toCamelCase(finding.RuleName),
				ShortDescription: SARIFMessage{
					Text: finding.RuleName,
				},
				FullDescription: SARIFMessage{
					Text: finding.Message,
				},
				HelpUri: finding.DocsURL,
				DefaultConfig: SARIFRuleConfig{
					Level: mapSeverityToSARIF(finding.Severity),
				},
			}
		}

		sarifResult := SARIFResult{
			RuleID: finding.RuleCode,
			Level:  mapSeverityToSARIF(finding.Severity),
			Message: SARIFMessage{
				Text: finding.Message,
			},
		}

		if finding.FilePath != "" {
			sarifResult.Locations = []SARIFLocation{
				{
					PhysicalLocation: SARIFPhysicalLocation{
						ArtifactLocation: SARIFArtifactLocation{
							URI: finding.FilePath,
						},
					},
				},
			}
		}

		results = append(results, sarifResult)
	}

	var rules []SARIFRule
	for _, rule := range rulesMap {
		rules = append(rules, rule)
	}

	report := SARIFReport{
		Schema:  sarifSchema,
		Version: "2.1.0",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:           "Canopy",
						Version:        result.PolicyVersion,
						InformationUri: "https://canopy.app",
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}

	return json.MarshalIndent(report, "", "  ")
}

func mapSeverityToSARIF(severity string) string {
	switch strings.ToUpper(severity) {
	case "BLOCKER", "HIGH":
		return "error"
	case "MEDIUM":
		return "warning"
	case "LOW":
		return "note"
	default:
		return "none"
	}
}

func toCamelCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

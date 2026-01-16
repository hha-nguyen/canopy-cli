package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Platform string

const (
	PlatformApple  Platform = "APPLE"
	PlatformGoogle Platform = "GOOGLE"
	PlatformBoth   Platform = "BOTH"
)

type CreateScanResponse struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id,omitempty"`
	Status       string    `json:"status"`
	Platform     string    `json:"platform"`
	CreatedAt    time.Time `json:"created_at"`
	WebSocketURL string    `json:"ws_url,omitempty"`
}

type ScanResult struct {
	ID             string          `json:"id"`
	Status         string          `json:"status"`
	Platform       string          `json:"platform"`
	PolicyVersion  string          `json:"policy_version"`
	DurationMs     int64           `json:"duration_ms"`
	RiskAssessment *RiskAssessment `json:"risk_assessment,omitempty"`
	Summary        *ScanSummary    `json:"summary,omitempty"`
	Findings       []Finding       `json:"findings"`
	Errors         []string        `json:"errors,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
}

type RiskAssessment struct {
	Score          int    `json:"score"`
	Confidence     string `json:"confidence"`
	Interpretation string `json:"interpretation"`
}

type ScanSummary struct {
	Total   int `json:"total"`
	Blocker int `json:"blocker"`
	High    int `json:"high"`
	Medium  int `json:"medium"`
	Low     int `json:"low"`
	Info    int `json:"info"`
	Passed  int `json:"passed"`
}

type Finding struct {
	ID          string                 `json:"id"`
	RuleCode    string                 `json:"rule_code"`
	RuleName    string                 `json:"rule_name"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	FilePath    string                 `json:"file_path,omitempty"`
	Evidence    map[string]interface{} `json:"evidence,omitempty"`
	Remediation *Remediation           `json:"remediation,omitempty"`
	DocsURL     string                 `json:"docs_url,omitempty"`
}

type Remediation struct {
	Action   string `json:"action"`
	Template string `json:"template,omitempty"`
}

func (c *Client) CreateScan(ctx context.Context, filePath string, platform Platform) (*CreateScanResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	body, contentType, err := createMultipartForm(file, filepath.Base(filePath), string(platform))
	if err != nil {
		return nil, fmt.Errorf("create multipart form: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/scans", body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	var result CreateScanResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetScan(ctx context.Context, id string) (*ScanResult, error) {
	var result ScanResult
	if err := c.Get(ctx, "/api/v1/scans/"+id, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CancelScan(ctx context.Context, id string) error {
	return c.Post(ctx, "/api/v1/scans/"+id+"/cancel", nil, nil)
}

func createMultipartForm(file *os.File, filename, platform string) (io.Reader, string, error) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		if err := writer.WriteField("platform", platform); err != nil {
			pw.CloseWithError(err)
			return
		}

		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		if _, err := io.Copy(part, file); err != nil {
			pw.CloseWithError(err)
			return
		}
	}()

	return pr, writer.FormDataContentType(), nil
}

package jasper

import (
	"adminbe/internal/app/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client handles JasperServer REST API operations
type Client struct {
	config *models.JasperServerConfig
	client *http.Client
}

// NewClient creates a new JasperServer client
func NewClient(config *models.JasperServerConfig) *Client {
	return &Client{
		config: config,
		client: &http.Client{},
	}
}

// createRequest creates HTTP request with basic auth
func (c *Client) createRequest(method, url string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// Add basic auth
	req.SetBasicAuth(c.config.Username, c.config.Password)

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add organization header if specified
	if c.config.Organization != "" {
		req.Header.Set("organization", c.config.Organization)
	}

	return req, nil
}

// RunReport runs a JasperServer report
func (c *Client) RunReport(req *models.JasperReportRequest) (*models.JasperReportResponse, []byte, error) {
	// Build URL for report execution
	runURL := fmt.Sprintf("%s/rest_v2/reports%s.%s", c.config.BaseURL, req.ReportPath, req.OutputFormat)

	// Add query parameters
	if len(req.Parameters) > 0 {
		params := []string{}
		for key, value := range req.Parameters {
			params = append(params, fmt.Sprintf("%s=%v", key, value))
		}
		runURL += "?" + strings.Join(params, "&")
	}

	httpReq, err := c.createRequest("GET", runURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// For reports with POST parameters (if any), use POST method
	if len(req.Parameters) > 0 {
		runURL = fmt.Sprintf("%s/rest_v2/reports%s.%s", c.config.BaseURL, req.ReportPath, req.OutputFormat)
		paramBody := map[string]interface{}{
			"reportParameter": req.Parameters,
		}
		if req.Interactive {
			paramBody["interactive"] = true
		}
		if req.Page > 0 {
			paramBody["page"] = req.Page
		}
		if req.Pages != "" {
			paramBody["pages"] = req.Pages
		}

		httpReq, err = c.createRequest("POST", runURL, paramBody)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create POST request: %w", err)
		}
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("JasperServer returned status %d: %s", resp.StatusCode, string(body))
	}

	// For binary content (PDF, etc.), return directly
	if req.OutputFormat == "pdf" || req.OutputFormat == "excel" || req.OutputFormat == "pptx" ||
		req.OutputFormat == "rtf" || req.OutputFormat == "docx" || req.OutputFormat == "xlsx" ||
		req.OutputFormat == "xls" || req.OutputFormat == "png" {
		response := &models.JasperReportResponse{
			ID:     "success",
			Status: "ready",
		}
		return response, body, nil
	}

	// For JSON/HTML content, parse JSON response
	var jasResp models.JasperReportResponse
	if err := json.Unmarshal(body, &jasResp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &jasResp, body, nil
}

// GetServerInfo retrieves JasperServer information
func (c *Client) GetServerInfo() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/rest_v2/serverInfo", c.config.BaseURL)

	req, err := c.createRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get server info: %s", body)
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return info, nil
}

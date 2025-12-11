package models

// JasperServerConfig holds JasperServer configuration
type JasperServerConfig struct {
	BaseURL      string `yaml:"base_url" json:"base_url"`
	Username     string `yaml:"username" json:"username"`
	Password     string `yaml:"password" json:"password"`
	Organization string `yaml:"organization" json:"organization"`
}

// JasperReportRequest represents a request to run a report
type JasperReportRequest struct {
	ReportPath   string                 `json:"report_path" binding:"required"`
	OutputFormat string                 `json:"output_format" binding:"oneof=pdf html excel pptx rtf docx xlsx xls png"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Interactive  bool                   `json:"interactive,omitempty"`
	Page         uint                   `json:"page,omitempty"`
	Pages        string                 `json:"pages,omitempty"`
}

// JasperReportResponse represents the response from running a report
type JasperReportResponse struct {
	ID           string `json:"id"`
	Output       string `json:"output,omitempty"`
	Status       string `json:"status"`
	ReportURI    string `json:"reportURI,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Permissions  string `json:"permissions,omitempty"`
}

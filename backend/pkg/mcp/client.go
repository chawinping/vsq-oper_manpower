package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"vsq-oper-manpower/backend/internal/config"
)

type Client struct {
	serverURL string
	apiKey    string
	enabled   bool
	client    *http.Client
}

type SuggestionRequest struct {
	BranchID        string  `json:"branch_id"`
	Date            string  `json:"date"`
	ExpectedRevenue float64 `json:"expected_revenue"`
	CurrentStaff    []string `json:"current_staff"`
	AvailableRotationStaff []string `json:"available_rotation_staff"`
}

type SuggestionResponse struct {
	Suggestions []AssignmentSuggestion `json:"suggestions"`
}

type AssignmentSuggestion struct {
	RotationStaffID string `json:"rotation_staff_id"`
	BranchID        string `json:"branch_id"`
	Date            string `json:"date"`
	AssignmentLevel int    `json:"assignment_level"`
	Confidence      float64 `json:"confidence"`
	Reason          string  `json:"reason"`
}

func NewClient(cfg config.MCPConfig) *Client {
	return &Client{
		serverURL: cfg.ServerURL,
		apiKey:    cfg.APIKey,
		enabled:   cfg.Enabled,
		client:    &http.Client{},
	}
}

func (c *Client) GetSuggestions(req SuggestionRequest) (*SuggestionResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("MCP client is not enabled")
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.serverURL+"/suggestions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MCP server returned error: %s", string(body))
	}

	var suggestionResp SuggestionResponse
	if err := json.NewDecoder(resp.Body).Decode(&suggestionResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &suggestionResp, nil
}

func (c *Client) RegenerateSuggestions(req SuggestionRequest) (*SuggestionResponse, error) {
	return c.GetSuggestions(req)
}



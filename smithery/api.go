package smithery

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Server represents a Smithery MCP server entry from the list endpoint.
// Based on https://smithery.ai/docs/registry/llms.txt
type Server struct {
	QualifiedName string `json:"qualifiedName"`
	DisplayName   string `json:"displayName"`
	Description   string `json:"description"`
	Homepage      string `json:"homepage"`
	UseCount      int    `json:"useCount"` // Changed type to int based on runtime error
	IsDeployed    bool   `json:"isDeployed"`
	CreatedAt     string `json:"createdAt"`
}

// --- Get Server Detail Types ---
// Based on https://smithery.ai/docs/registry/llms.txt

type Connection struct {
	Type         string          `json:"type"`
	URL          *string         `json:"url,omitempty"` // Use pointer for optional field
	ConfigSchema json.RawMessage `json:"configSchema"`  // Keep as raw JSON for now
}

type Security struct {
	ScanPassed bool `json:"scanPassed"`
}

type Tool struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"` // Use pointer for optional field
}

// ServerDetail represents the detailed information for a single server.
type ServerDetail struct {
	QualifiedName string          `json:"qualifiedName"`
	DisplayName   string          `json:"displayName"`
	Remote        bool            `json:"remote"`
	IconURL       *string         `json:"iconUrl,omitempty"`       // Use pointer for optional field
	DeploymentURL *string         `json:"deploymentUrl,omitempty"` // Use pointer for optional field
	ConfigSchema  json.RawMessage `json:"configSchema"`            // Keep as raw JSON for now
	Connections   []Connection    `json:"connections"`
	Security      *Security       `json:"security,omitempty"` // Use pointer for optional field
	Tools         []Tool          `json:"tools,omitempty"`    // Use pointer for optional array field
}

// Pagination represents the pagination info from the Smithery API.
type Pagination struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalPages  int `json:"totalPages"`
	TotalCount  int `json:"totalCount"`
}

// ListResponse represents the structure of the server list API response.
type ListResponse struct {
	Servers    []Server   `json:"servers"`
	Pagination Pagination `json:"pagination"`
}

const registryURL = "https://registry.smithery.ai/servers"

// FetchServers retrieves a list of servers from the Smithery Registry API.
func FetchServers(apiKey string, query string, page int) (*ListResponse, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Smithery API key is required")
	}

	req, err := http.NewRequest("GET", registryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Smithery API request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	// Add query parameters
	qParams := req.URL.Query()
	if query != "" {
		qParams.Add("q", query)
	}
	if page > 1 {
		qParams.Add("page", fmt.Sprintf("%d", page))
	}
	// qParams.Add("pageSize", "20") // Example: request more items per page
	req.URL.RawQuery = qParams.Encode()

	log.Println("Fetching Smithery servers from:", req.URL.String())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Smithery API request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Smithery API response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Smithery API Error: Status %d, Body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("Smithery API request failed with status %d", resp.StatusCode)
	}

	var listResponse ListResponse
	if err := json.Unmarshal(body, &listResponse); err != nil {
		log.Printf("ERROR: Failed to unmarshal Smithery API response. Status: %d, Body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("failed to unmarshal Smithery API response: %w", err)
	}

	return &listResponse, nil
}

// FetchServerDetail retrieves detailed information for a specific server.
func FetchServerDetail(apiKey string, qualifiedName string) (*ServerDetail, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Smithery API key is required")
	}
	if qualifiedName == "" {
		return nil, fmt.Errorf("qualifiedName is required")
	}

	// Construct the URL: https://registry.smithery.ai/servers/{qualifiedName}
	// Use url.JoinPath for safer path construction
	baseURL, _ := url.Parse(registryURL) // Assuming registryURL is "https://registry.smithery.ai/servers"
	detailURL, err := url.JoinPath(baseURL.String(), qualifiedName)
	if err != nil {
		return nil, fmt.Errorf("failed to construct detail URL: %w", err)
	}

	req, err := http.NewRequest("GET", detailURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Smithery detail API request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	log.Println("Fetching Smithery server detail from:", req.URL.String())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Smithery detail API request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Smithery detail API response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Smithery Detail API Error: Status %d, Body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("Smithery detail API request failed with status %d", resp.StatusCode)
	}

	var serverDetail ServerDetail
	if err := json.Unmarshal(body, &serverDetail); err != nil {
		log.Printf("ERROR: Failed to unmarshal Smithery detail API response. Status: %d, Body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("failed to unmarshal Smithery detail API response: %w", err)
	}

	return &serverDetail, nil
}

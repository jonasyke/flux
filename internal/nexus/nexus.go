package nexus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	baseURL        = "https://api.nexusmods.com/v1"
	GameDomainName = "readyornot"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type ModDetails struct {
	ModID          int    `json:"mod_id"`
	Name           string `json:"name"`
	Summary        string `json:"summary"`
	Version        string `json:"version"`
	UpdatedAt      int64  `json:"updated_timestamp"`
	GameDomainName string `json:"domain_name"`
}

func (c *Client) ValidateKey() error {
	req, err := http.NewRequest("GET", baseURL+"/users/validate.json", nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid API key or unauthorized (status: %d)", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetModDetails(gameDomain string, modID int) (*ModDetails, error) {
	url := fmt.Sprintf("%s/games/%s/mods/%d.json", baseURL, gameDomain, modID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mod: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("mod ID %d not found on %s", modID, gameDomain)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nexus API error (status: %d)", resp.StatusCode)
	}

	var mod ModDetails
	if err := json.NewDecoder(resp.Body).Decode(&mod); err != nil {
		return nil, fmt.Errorf("failed to parse mod response: %w", err)
	}

	return &mod, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Application-Name", "FluxModManager")
	req.Header.Set("Application-Version", "1.0.0")
}

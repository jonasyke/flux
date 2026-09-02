package nexus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL        = "https://api.nexusmods.com/v1"
	GameDomainName = "readyornot"
)

type NXMLink struct {
	GameDomain string
	ModID      int
	FileID     int
	Key        string
	Expires    int64
	UserID     int
}

type DownloadLink struct {
	Name string `json:"name"`
	URI  string `json:"URI"`
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type ModDetails struct {
	ModID          int    `json:"mod_id"`
	Name           string `json:"name"`
	Summary        string `json:"summary"`
	Version        string `json:"version"`
	UpdatedAt      int64  `json:"updated_timestamp"`
	GameDomainName string `json:"domain_name"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
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

func ParseNXM(raw string) (*NXMLink, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if !strings.EqualFold(u.Scheme, "nxm") {
		return nil, fmt.Errorf("not an nxm:// link")
	}

	game := strings.ToLower(u.Host)
	if game == "" {
		return nil, fmt.Errorf("missing game domain")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 4 || parts[0] != "mods" || parts[2] != "files" {
		return nil, fmt.Errorf("unexpected nxm path: %s", u.Path)
	}

	modID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid mod id: %w", err)
	}
	fileID, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid file id: %w", err)
	}

	q := u.Query()
	expires, _ := strconv.ParseInt(q.Get("expires"), 10, 64)
	userID, _ := strconv.Atoi(q.Get("user_id"))

	return &NXMLink{
		GameDomain: game,
		ModID:      modID,
		FileID:     fileID,
		Key:        q.Get("key"),
		Expires:    expires,
		UserID:     userID,
	}, nil
}

func (c *Client) GetDownloadURLs(gameDomain string, modID, fileID int, key string, expires int64) ([]DownloadLink, error) {
	apiURL := fmt.Sprintf("%s/games/%s/mods/%d/files/%d/download_link.json",
		baseURL, gameDomain, modID, fileID)

	if key != "" && expires > 0 {
		apiURL += fmt.Sprintf("?key=%s&expires=%d", url.QueryEscape(key), expires)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	// Watch timeout
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download link request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("file %d not found for mod %d", fileID, modID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nexus API error (status %d)", resp.StatusCode)
	}

	var links []DownloadLink
	if err := json.NewDecoder(resp.Body).Decode(&links); err != nil {
		return nil, fmt.Errorf("failed to parse download links: %w", err)
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("no download links returned")
	}
	return links, nil
}

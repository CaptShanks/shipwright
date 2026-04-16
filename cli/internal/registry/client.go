package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultOwner = "CaptShanks"
	defaultRepo  = "shipwright"
	apiBase      = "https://api.github.com"
	cacheTTL     = 1 * time.Hour
)

type Client struct {
	httpClient *http.Client
	owner      string
	repo       string
	token      string
	cacheDir   string
}

func NewClient() *Client {
	home, _ := os.UserHomeDir()
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		owner:      defaultOwner,
		repo:       defaultRepo,
		token:      os.Getenv("GITHUB_TOKEN"),
		cacheDir:   filepath.Join(home, ".shipwright", "cache"),
	}
}

// FetchMarketplace returns the marketplace manifest, using cache when fresh.
func (c *Client) FetchMarketplace() (*Marketplace, error) {
	if cached, err := c.loadCachedMarketplace(); err == nil {
		return cached, nil
	}

	url := fmt.Sprintf("%s/repos/%s/%s/contents/.claude-plugin/marketplace.json", apiBase, c.owner, c.repo)
	body, err := c.fetchFile(url)
	if err != nil {
		return nil, fmt.Errorf("fetching marketplace.json: %w", err)
	}

	var gh GitHubContent
	if err := json.Unmarshal(body, &gh); err != nil {
		return nil, fmt.Errorf("parsing github response: %w", err)
	}

	raw, err := c.download(gh.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("downloading marketplace.json: %w", err)
	}

	var m Marketplace
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parsing marketplace.json: %w", err)
	}

	_ = c.cacheMarketplace(&m)
	return &m, nil
}

// FetchPluginManifest returns a plugin's plugin.json from the repo.
func (c *Client) FetchPluginManifest(pluginSource string) (*PluginManifest, error) {
	path := fmt.Sprintf("%s/.claude-plugin/plugin.json", normalizeSource(pluginSource))
	raw, err := c.fetchRawFile(path)
	if err != nil {
		return nil, fmt.Errorf("fetching plugin manifest for %s: %w", pluginSource, err)
	}

	var pm PluginManifest
	if err := json.Unmarshal(raw, &pm); err != nil {
		return nil, fmt.Errorf("parsing plugin.json for %s: %w", pluginSource, err)
	}
	return &pm, nil
}

// FetchMcpManifest returns an MCP server config from a plugin's .mcp.json.
// The source should be the marketplace source path (e.g. "./plugins/context7").
func (c *Client) FetchMcpManifest(pluginSource string) (*McpManifest, error) {
	path := fmt.Sprintf("%s/.mcp.json", normalizeSource(pluginSource))
	raw, err := c.fetchRawFile(path)
	if err != nil {
		return nil, fmt.Errorf("fetching MCP manifest for %s: %w", pluginSource, err)
	}

	var servers map[string]struct {
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
	}
	if err := json.Unmarshal(raw, &servers); err != nil {
		return nil, fmt.Errorf("parsing .mcp.json for %s: %w", pluginSource, err)
	}

	for name, srv := range servers {
		return &McpManifest{
			Name:    name,
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
		}, nil
	}
	return nil, fmt.Errorf("no server entries in .mcp.json for %s", pluginSource)
}

// FetchFileContent downloads a single file's raw content by repo path.
func (c *Client) FetchFileContent(repoPath string) ([]byte, error) {
	return c.fetchRawFile(repoPath)
}

// FetchDirectoryListing returns the contents of a directory in the repo.
func (c *Client) FetchDirectoryListing(repoPath string) ([]GitHubContent, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", apiBase, c.owner, c.repo, repoPath)
	body, err := c.fetchFile(url)
	if err != nil {
		return nil, err
	}

	var entries []GitHubContent
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("parsing directory listing for %s: %w", repoPath, err)
	}
	return entries, nil
}

// FetchSkillTree recursively downloads an entire skill directory from the repo.
func (c *Client) FetchSkillTree(skillRepoPath string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	return files, c.fetchTreeRecursive(skillRepoPath, files)
}

func (c *Client) fetchTreeRecursive(dirPath string, files map[string][]byte) error {
	entries, err := c.FetchDirectoryListing(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Type == "file" {
			content, err := c.download(entry.DownloadURL)
			if err != nil {
				return fmt.Errorf("downloading %s: %w", entry.Path, err)
			}
			// Store with path relative to the skill root
			relPath := strings.TrimPrefix(entry.Path, dirPath+"/")
			if relPath == entry.Path {
				relPath = filepath.Base(entry.Path)
			}
			files[relPath] = content
		} else if entry.Type == "dir" {
			if err := c.fetchTreeRecursive(entry.Path, files); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) fetchFile(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) fetchRawFile(repoPath string) ([]byte, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", apiBase, c.owner, c.repo, repoPath)
	body, err := c.fetchFile(url)
	if err != nil {
		return nil, err
	}

	var gh GitHubContent
	if err := json.Unmarshal(body, &gh); err != nil {
		return nil, err
	}
	if gh.DownloadURL == "" {
		return nil, fmt.Errorf("no download URL for %s", repoPath)
	}
	return c.download(gh.DownloadURL)
}

func (c *Client) download(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d for %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// NormalizeSource strips a leading "./" from a marketplace source path so it
// can be used directly as a repo-relative path.
func NormalizeSource(source string) string {
	return normalizeSource(source)
}

func normalizeSource(source string) string {
	return strings.TrimPrefix(source, "./")
}

func (c *Client) loadCachedMarketplace() (*Marketplace, error) {
	path := filepath.Join(c.cacheDir, "marketplace.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cached CachedMarketplace
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}

	if time.Since(cached.FetchedAt) > cacheTTL {
		return nil, fmt.Errorf("cache expired")
	}
	return &cached.Marketplace, nil
}

func (c *Client) cacheMarketplace(m *Marketplace) error {
	if err := os.MkdirAll(c.cacheDir, 0o755); err != nil {
		return err
	}

	cached := CachedMarketplace{
		FetchedAt:   time.Now(),
		Marketplace: *m,
	}
	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.cacheDir, "marketplace.json"), data, 0o644)
}

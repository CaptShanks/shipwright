package registry

import "time"

type Marketplace struct {
	Name     string            `json:"name"`
	Owner    MarketplaceOwner  `json:"owner"`
	Metadata MarketplaceMeta   `json:"metadata"`
	Plugins  []MarketplaceItem `json:"plugins"`
}

type MarketplaceOwner struct {
	Name string `json:"name"`
}

type MarketplaceMeta struct {
	Description string `json:"description"`
	Version     string `json:"version"`
	PluginRoot  string `json:"pluginRoot"`
}

type MarketplaceItem struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Keywords    []string `json:"keywords"`
}

type PluginManifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Agents      []string `json:"agents"`
	Skills      []string `json:"skills"`
}

type CachedMarketplace struct {
	FetchedAt   time.Time   `json:"fetched_at"`
	Marketplace Marketplace `json:"marketplace"`
}

// McpManifest represents an MCP server template from the _mcps/ directory.
type McpManifest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	Tags        []string          `json:"tags"`
}

// GitHubContent represents a file or directory entry from the GitHub Contents API.
type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
	SHA         string `json:"sha"`
}

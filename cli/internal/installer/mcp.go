package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// McpConfig is the JSON shape written into each tool's config entry.
type McpConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// McpInstaller defines how an AI tool receives MCP server configurations.
type McpInstaller interface {
	Name() string
	ConfigPath(scope Scope) string
	Install(name string, cfg McpConfig, scope Scope) error
	Remove(name string, scope Scope) error
	ListInstalled(scope Scope) ([]string, error)
}

// AllMcp returns MCP installer instances for every supported tool.
func AllMcp() []McpInstaller {
	return []McpInstaller{
		&McpCursorInstaller{},
		&McpClaudeInstaller{},
		&McpVSCodeInstaller{},
	}
}

// McpForTarget returns MCP installers matching the target flag.
func McpForTarget(target string) ([]McpInstaller, error) {
	if target == "all" {
		return AllMcp(), nil
	}
	for _, inst := range AllMcp() {
		if inst.Name() == target {
			return []McpInstaller{inst}, nil
		}
	}
	return nil, fmt.Errorf("unknown target %q (valid: cursor, claude, vscode, all)", target)
}

// --- Cursor MCP Installer ---

type McpCursorInstaller struct{}

func (m *McpCursorInstaller) Name() string { return "cursor" }

func (m *McpCursorInstaller) ConfigPath(scope Scope) string {
	if scope == ScopeGlobal {
		return filepath.Join(homeDir(), ".cursor", "mcp.json")
	}
	return filepath.Join(ProjectRoot(), ".cursor", "mcp.json")
}

func (m *McpCursorInstaller) Install(name string, cfg McpConfig, scope Scope) error {
	return upsertMcpEntry(m.ConfigPath(scope), "mcpServers", name, cfg)
}

func (m *McpCursorInstaller) Remove(name string, scope Scope) error {
	return removeMcpEntry(m.ConfigPath(scope), "mcpServers", name)
}

func (m *McpCursorInstaller) ListInstalled(scope Scope) ([]string, error) {
	return listMcpEntries(m.ConfigPath(scope), "mcpServers")
}

// --- Claude MCP Installer ---

type McpClaudeInstaller struct{}

func (m *McpClaudeInstaller) Name() string { return "claude" }

func (m *McpClaudeInstaller) ConfigPath(scope Scope) string {
	if scope == ScopeGlobal {
		return filepath.Join(homeDir(), ".claude.json")
	}
	return filepath.Join(ProjectRoot(), ".claude.json")
}

func (m *McpClaudeInstaller) Install(name string, cfg McpConfig, scope Scope) error {
	return upsertMcpEntry(m.ConfigPath(scope), "mcpServers", name, cfg)
}

func (m *McpClaudeInstaller) Remove(name string, scope Scope) error {
	return removeMcpEntry(m.ConfigPath(scope), "mcpServers", name)
}

func (m *McpClaudeInstaller) ListInstalled(scope Scope) ([]string, error) {
	return listMcpEntries(m.ConfigPath(scope), "mcpServers")
}

// --- VS Code MCP Installer ---

type McpVSCodeInstaller struct{}

func (m *McpVSCodeInstaller) Name() string { return "vscode" }

func (m *McpVSCodeInstaller) ConfigPath(scope Scope) string {
	if scope == ScopeGlobal {
		return vscodeGlobalMcpPath()
	}
	return filepath.Join(ProjectRoot(), ".vscode", "mcp.json")
}

func (m *McpVSCodeInstaller) Install(name string, cfg McpConfig, scope Scope) error {
	return upsertMcpEntry(m.ConfigPath(scope), "servers", name, cfg)
}

func (m *McpVSCodeInstaller) Remove(name string, scope Scope) error {
	return removeMcpEntry(m.ConfigPath(scope), "servers", name)
}

func (m *McpVSCodeInstaller) ListInstalled(scope Scope) ([]string, error) {
	return listMcpEntries(m.ConfigPath(scope), "servers")
}

func vscodeGlobalMcpPath() string {
	home := homeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Code", "User", "mcp.json")
	case "linux":
		return filepath.Join(home, ".config", "Code", "User", "mcp.json")
	default: // windows
		return filepath.Join(os.Getenv("APPDATA"), "Code", "User", "mcp.json")
	}
}

// --- Shared JSON manipulation helpers ---

func loadJSONFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return make(map[string]any), nil
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return obj, nil
}

func saveJSONFile(path string, obj map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func upsertMcpEntry(path, rootKey, name string, cfg McpConfig) error {
	obj, err := loadJSONFile(path)
	if err != nil {
		return err
	}

	servers, ok := obj[rootKey].(map[string]any)
	if !ok {
		servers = make(map[string]any)
	}

	entry := map[string]any{
		"command": cfg.Command,
		"args":    cfg.Args,
	}
	if len(cfg.Env) > 0 {
		entry["env"] = cfg.Env
	}

	servers[name] = entry
	obj[rootKey] = servers
	return saveJSONFile(path, obj)
}

func removeMcpEntry(path, rootKey, name string) error {
	obj, err := loadJSONFile(path)
	if err != nil {
		return err
	}

	servers, ok := obj[rootKey].(map[string]any)
	if !ok {
		return nil
	}

	delete(servers, name)
	obj[rootKey] = servers
	return saveJSONFile(path, obj)
}

func listMcpEntries(path, rootKey string) ([]string, error) {
	obj, err := loadJSONFile(path)
	if err != nil {
		return nil, err
	}

	servers, ok := obj[rootKey].(map[string]any)
	if !ok {
		return nil, nil
	}

	names := make([]string, 0, len(servers))
	for k := range servers {
		names = append(names, k)
	}
	return names, nil
}

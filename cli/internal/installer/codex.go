package installer

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

type CodexInstaller struct{}

func NewCodexInstaller() *CodexInstaller { return &CodexInstaller{} }
func (c *CodexInstaller) Name() string   { return "codex" }

func (c *CodexInstaller) AgentDir(scope Scope) string {
	if scope == ScopeGlobal {
		return filepath.Join(homeDir(), ".codex", "agents")
	}
	return filepath.Join(ProjectRoot(), ".codex", "agents")
}

func (c *CodexInstaller) SkillDir(scope Scope, skillName string) string {
	if scope == ScopeGlobal {
		return filepath.Join(homeDir(), ".agents", "skills", skillName)
	}
	return filepath.Join(ProjectRoot(), ".agents", "skills", skillName)
}

func (c *CodexInstaller) InstallAgent(name string, content []byte, scope Scope) (string, error) {
	toml, err := mdAgentToTOML(content)
	if err != nil {
		return "", fmt.Errorf("converting agent %s to TOML: %w", name, err)
	}
	path := filepath.Join(c.AgentDir(scope), name+".toml")
	return path, writeFile(path, toml)
}

func (c *CodexInstaller) InstallSkill(name string, files map[string][]byte, scope Scope) ([]string, error) {
	base := c.SkillDir(scope, name)
	var installed []string
	for relPath, content := range files {
		fullPath := filepath.Join(base, relPath)
		if err := writeFile(fullPath, content); err != nil {
			return installed, err
		}
		installed = append(installed, fullPath)
	}
	return installed, nil
}

func (c *CodexInstaller) UninstallFiles(files []string) error {
	return removeFiles(files)
}

// mdAgentToTOML converts a shipwright agent markdown file (with YAML
// frontmatter) into a Codex-compatible TOML agent definition.
//
// Codex agents are .toml files with three required fields:
//   - name (string)
//   - description (string)
//   - developer_instructions (string)
func mdAgentToTOML(md []byte) ([]byte, error) {
	frontmatter, body, err := splitFrontmatter(md)
	if err != nil {
		return nil, err
	}

	name := extractYAMLField(frontmatter, "name")
	desc := extractYAMLField(frontmatter, "description")
	if name == "" {
		return nil, fmt.Errorf("agent markdown missing 'name' in frontmatter")
	}
	if desc == "" {
		desc = name
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "name = %q\n", name)
	fmt.Fprintf(&buf, "description = %q\n", desc)
	fmt.Fprintf(&buf, "developer_instructions = %q\n", strings.TrimSpace(body))
	return buf.Bytes(), nil
}

// splitFrontmatter splits a markdown file into YAML frontmatter and body.
// Frontmatter must be delimited by --- on the first line and a closing ---.
func splitFrontmatter(md []byte) (frontmatter string, body string, err error) {
	s := string(md)
	if !strings.HasPrefix(s, "---") {
		return "", s, nil
	}

	rest := s[3:]
	if idx := strings.Index(rest, "\n"); idx >= 0 {
		rest = rest[idx+1:]
	}

	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", s, fmt.Errorf("unterminated YAML frontmatter")
	}

	fm := rest[:end]
	remaining := rest[end+4:]
	if idx := strings.Index(remaining, "\n"); idx >= 0 {
		remaining = remaining[idx+1:]
	}

	return fm, remaining, nil
}

// extractYAMLField does a simple line-based extraction of a top-level scalar
// YAML field. Handles both single-line values and multi-line values that
// continue on the next line with indentation.
func extractYAMLField(yaml, field string) string {
	prefix := field + ":"
	for _, line := range strings.Split(yaml, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			return val
		}
	}
	return ""
}

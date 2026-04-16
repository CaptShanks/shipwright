package installer

import "path/filepath"

type ClaudeInstaller struct{}

func NewClaudeInstaller() *ClaudeInstaller { return &ClaudeInstaller{} }
func (c *ClaudeInstaller) Name() string    { return "claude" }

func (c *ClaudeInstaller) AgentDir(scope Scope) string {
	if scope == ScopeGlobal {
		return filepath.Join(homeDir(), ".claude", "agents")
	}
	return filepath.Join(ProjectRoot(), ".claude", "agents")
}

func (c *ClaudeInstaller) SkillDir(scope Scope, skillName string) string {
	if scope == ScopeGlobal {
		return filepath.Join(homeDir(), ".claude", "skills", skillName)
	}
	return filepath.Join(ProjectRoot(), ".claude", "skills", skillName)
}

func (c *ClaudeInstaller) InstallAgent(name string, content []byte, scope Scope) (string, error) {
	path := filepath.Join(c.AgentDir(scope), name+".md")
	return path, writeFile(path, content)
}

func (c *ClaudeInstaller) InstallSkill(name string, files map[string][]byte, scope Scope) ([]string, error) {
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

func (c *ClaudeInstaller) UninstallFiles(files []string) error {
	return removeFiles(files)
}

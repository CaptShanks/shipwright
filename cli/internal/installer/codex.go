package installer

import "path/filepath"

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
		return filepath.Join(homeDir(), ".codex", "skills", skillName)
	}
	return filepath.Join(ProjectRoot(), ".codex", "skills", skillName)
}

func (c *CodexInstaller) InstallAgent(name string, content []byte, scope Scope) (string, error) {
	path := filepath.Join(c.AgentDir(scope), name+".md")
	return path, writeFile(path, content)
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

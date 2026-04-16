package installer

import (
	"os"
	"path/filepath"
)

type CursorInstaller struct{}

func NewCursorInstaller() *CursorInstaller { return &CursorInstaller{} }
func (c *CursorInstaller) Name() string    { return "cursor" }

func (c *CursorInstaller) AgentDir(scope Scope) string {
	if scope == ScopeGlobal {
		return filepath.Join(homeDir(), ".cursor", "agents")
	}
	return filepath.Join(ProjectRoot(), ".cursor", "agents")
}

func (c *CursorInstaller) SkillDir(scope Scope, skillName string) string {
	if scope == ScopeGlobal {
		return filepath.Join(homeDir(), ".cursor", "skills-cursor", skillName)
	}
	return filepath.Join(ProjectRoot(), ".cursor", "skills", skillName)
}

func (c *CursorInstaller) InstallAgent(name string, content []byte, scope Scope) (string, error) {
	path := filepath.Join(c.AgentDir(scope), name+".md")
	return path, writeFile(path, content)
}

func (c *CursorInstaller) InstallSkill(name string, files map[string][]byte, scope Scope) ([]string, error) {
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

func (c *CursorInstaller) UninstallFiles(files []string) error {
	return removeFiles(files)
}

func removeFiles(files []string) error {
	dirs := make(map[string]bool)
	for _, f := range files {
		_ = os.Remove(f)
		dirs[filepath.Dir(f)] = true
	}
	for dir := range dirs {
		removeEmptyDirs(dir)
	}
	return nil
}

// removeEmptyDirs removes the directory and its parents if they are empty,
// stopping at the first non-empty or non-existent directory.
func removeEmptyDirs(dir string) {
	for {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

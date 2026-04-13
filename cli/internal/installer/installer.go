package installer

import (
	"fmt"
	"os"
	"path/filepath"
)

type Scope string

const (
	ScopeLocal  Scope = "local"
	ScopeGlobal Scope = "global"
)

// Installer defines how a specific AI tool receives agents and skills.
type Installer interface {
	Name() string
	AgentDir(scope Scope) string
	SkillDir(scope Scope, skillName string) string
	InstallAgent(name string, content []byte, scope Scope) (string, error)
	InstallSkill(name string, files map[string][]byte, scope Scope) ([]string, error)
	UninstallFiles(files []string) error
}

// All returns installer instances for every supported tool.
func All() []Installer {
	return []Installer{
		NewCursorInstaller(),
		NewClaudeInstaller(),
		NewCodexInstaller(),
	}
}

// ForTarget returns the installers matching the target flag.
func ForTarget(target string) ([]Installer, error) {
	if target == "all" {
		return All(), nil
	}

	for _, inst := range All() {
		if inst.Name() == target {
			return []Installer{inst}, nil
		}
	}
	return nil, fmt.Errorf("unknown target %q (valid: cursor, claude, codex, all)", target)
}

// ProjectRoot walks up from cwd to find a .git directory, returning the project root.
// Falls back to cwd if no .git is found.
func ProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return cwd
		}
		dir = parent
	}
}

func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

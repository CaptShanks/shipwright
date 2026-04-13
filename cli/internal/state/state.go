package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const stateFile = "state.json"

type Store struct {
	path string
}

type State struct {
	Installations []Installation `json:"installations"`
}

type Installation struct {
	Plugin      string    `json:"plugin"`
	Version     string    `json:"version"`
	Target      string    `json:"target"`
	Scope       string    `json:"scope"`
	ProjectPath string    `json:"project_path,omitempty"`
	InstalledAt time.Time `json:"installed_at"`
	Files       []string  `json:"files"`
}

func NewStore() *Store {
	home, _ := os.UserHomeDir()
	return &Store{
		path: filepath.Join(home, ".shipwright", stateFile),
	}
}

func (s *Store) Load() (*State, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, err
	}

	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *Store) Save(st *State) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// Record adds or updates an installation entry.
func (s *Store) Record(inst Installation) error {
	st, err := s.Load()
	if err != nil {
		return err
	}

	updated := false
	for i, existing := range st.Installations {
		if existing.Plugin == inst.Plugin && existing.Target == inst.Target &&
			existing.Scope == inst.Scope && existing.ProjectPath == inst.ProjectPath {
			st.Installations[i] = inst
			updated = true
			break
		}
	}
	if !updated {
		st.Installations = append(st.Installations, inst)
	}
	return s.Save(st)
}

// Remove deletes an installation entry and returns the files that were tracked.
func (s *Store) Remove(plugin, target, scope, projectPath string) ([]string, error) {
	st, err := s.Load()
	if err != nil {
		return nil, err
	}

	var files []string
	filtered := make([]Installation, 0, len(st.Installations))
	for _, inst := range st.Installations {
		if inst.Plugin == plugin && inst.Target == target &&
			inst.Scope == scope && inst.ProjectPath == projectPath {
			files = inst.Files
			continue
		}
		filtered = append(filtered, inst)
	}
	st.Installations = filtered
	return files, s.Save(st)
}

// List returns installations matching the given filters.
// Empty strings match everything.
func (s *Store) List(target, scope, projectPath string) ([]Installation, error) {
	st, err := s.Load()
	if err != nil {
		return nil, err
	}

	var results []Installation
	for _, inst := range st.Installations {
		if target != "" && target != "all" && inst.Target != target {
			continue
		}
		if scope != "" && inst.Scope != scope {
			continue
		}
		if projectPath != "" && inst.Scope == "local" && inst.ProjectPath != projectPath {
			continue
		}
		results = append(results, inst)
	}
	return results, nil
}

// FindInstallation returns a specific installation entry.
func (s *Store) FindInstallation(plugin, target, scope, projectPath string) (*Installation, error) {
	st, err := s.Load()
	if err != nil {
		return nil, err
	}

	for _, inst := range st.Installations {
		if inst.Plugin == plugin && inst.Target == target &&
			inst.Scope == scope && inst.ProjectPath == projectPath {
			return &inst, nil
		}
	}
	return nil, nil
}

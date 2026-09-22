package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/djranoia/discforge/internal/jobs"
)

type Snapshot struct {
	Jobs []jobs.Job `json:"jobs"`
}
type Store struct{ path string }

func New(path string) Store { return Store{path: path} }

func Default() Store {
	if override := os.Getenv("DISCFORGE_STATE_FILE"); override != "" {
		return New(override)
	}
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return New(filepath.Join(".", "discforge-state.json"))
		}
		base = filepath.Join(home, ".local", "state")
	}
	return New(filepath.Join(base, "discforge", "state.json"))
}

func (s Store) Load() (Snapshot, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return Snapshot{}, nil
	}
	if err != nil {
		return Snapshot{}, err
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, err
	}
	for i := range snapshot.Jobs {
		if snapshot.Jobs[i].Status == jobs.Preparing || snapshot.Jobs[i].Status == jobs.Running {
			snapshot.Jobs[i].Status = jobs.Interrupted
		}
	}
	return snapshot, nil
}

func (s Store) Save(snapshot Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".state-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.path)
}

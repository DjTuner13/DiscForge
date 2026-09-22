package logs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Store struct{ Root string }

func (s Store) Open(jobID string) (io.WriteCloser, error) {
	if jobID == "" {
		return nil, fmt.Errorf("job id is required")
	}
	if err := os.MkdirAll(s.Root, 0o700); err != nil {
		return nil, err
	}
	return os.OpenFile(filepath.Join(s.Root, jobID+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
}

func (s Store) Read(jobID string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.Root, jobID+".log"))
}

package benchmark

import (
	"path/filepath"
	"testing"
	"time"
)

func TestResultSave(t *testing.T) { path := filepath.Join(t.TempDir(), "result.json"); if err := (Result{CPU: "fixture", Profile: "dvd", Frames: 100, Duration: time.Second, FPS: 100}).Save(path); err != nil { t.Fatal(err) } }


package hardware

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageTemperature(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "hwmon0"); if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(dir, "temp1_label"), []byte("Package id 0\n"), 0o644); err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(dir, "temp1_input"), []byte("68000\n"), 0o644); err != nil { t.Fatal(err) }
	temp, ok := PackageTemperature(filepath.Dir(dir)); if !ok || temp.Celsius != 68 { t.Fatalf("temperature = %#v, %v", temp, ok) }
}


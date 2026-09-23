package hardware

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Temperature struct {
	Celsius float64
	Label   string
	Source  string
}

func PackageTemperature(root string) (Temperature, bool) {
	entries, err := filepath.Glob(filepath.Join(root, "hwmon*"))
	if err != nil {
		return Temperature{}, false
	}
	for _, dir := range entries {
		for _, path := range valueFiles(dir) {
			labelBytes, _ := os.ReadFile(strings.TrimSuffix(path, "_input") + "_label")
			label := strings.TrimSpace(string(labelBytes))
			if label == "" {
				label = filepath.Base(path)
			}
			if !strings.Contains(strings.ToLower(label), "package") {
				continue
			}
			value, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			millidegrees, err := strconv.ParseFloat(strings.TrimSpace(string(value)), 64)
			if err != nil {
				continue
			}
			return Temperature{Celsius: millidegrees / 1000, Label: label, Source: path}, true
		}
	}
	return Temperature{}, false
}
func valueFiles(dir string) []string {
	paths, _ := filepath.Glob(filepath.Join(dir, "temp*_input"))
	return paths
}

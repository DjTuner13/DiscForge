package benchmark

import (
	"encoding/json"
	"os"
	"time"
)

type Result struct { CPU string `json:"cpu"`; Profile string `json:"profile"`; Frames int64 `json:"frames"`; Duration time.Duration `json:"duration"`; FPS float64 `json:"fps"`; Realtime float64 `json:"realtime"`; AvgTemperature float64 `json:"avg_temperature,omitempty"`; MaxTemperature float64 `json:"max_temperature,omitempty"` }

func (r Result) Save(path string) error { data, err := json.MarshalIndent(r, "", "  "); if err != nil { return err }; return os.WriteFile(path, append(data, '\n'), 0o600) }


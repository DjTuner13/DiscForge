package jobs

import (
	"testing"
	"time"
)

func TestEstimateETA(t *testing.T) {
	if got := EstimateETA(1000, 500, 10); got != 50*time.Second {
		t.Fatalf("eta = %s", got)
	}
}

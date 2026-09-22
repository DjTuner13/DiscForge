package jobs

import (
	"bufio"
	"strings"
	"testing"
	"time"
)

func TestParseProgress(t *testing.T) {
	input := "frame=48291\nfps=13.1\nout_time_us=102030000\nspeed=0.219x\nprogress=end\n"
	p := Progress{}
	ParseProgress(bufio.NewScanner(strings.NewReader(input)), &p)
	if p.Frame != 48291 || p.FPS != 13.1 || p.OutTime != 102030*time.Millisecond || p.Speed != 0.219 || !p.Done {
		t.Fatalf("parsed progress = %#v", p)
	}
}

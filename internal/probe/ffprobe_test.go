package probe

import "testing"

func TestMediaStreamCountsAndDuration(t *testing.T) {
	m := Media{Format: Format{Duration: "12.5"}, Streams: []Stream{{CodecType: "video"}, {CodecType: "audio"}, {CodecType: "subtitle"}}}
	audio, subs := m.StreamCounts(); if audio != 1 || subs != 1 || m.DurationSeconds() != 12.5 { t.Fatalf("media summary = %d, %d, %f", audio, subs, m.DurationSeconds()) }
}


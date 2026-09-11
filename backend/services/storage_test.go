package services

import "testing"

func TestParseByteRange(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		size      int64
		wantStart int64
		wantEnd   int64
		wantOK    bool
	}{
		{name: "bounded", value: "bytes=0-99", size: 1000, wantStart: 0, wantEnd: 99, wantOK: true},
		{name: "open ended", value: "bytes=900-", size: 1000, wantStart: 900, wantEnd: 999, wantOK: true},
		{name: "clamped end", value: "bytes=900-1200", size: 1000, wantStart: 900, wantEnd: 999, wantOK: true},
		{name: "suffix", value: "bytes=-100", size: 1000, wantStart: 900, wantEnd: 999, wantOK: true},
		{name: "suffix larger than object", value: "bytes=-2000", size: 1000, wantStart: 0, wantEnd: 999, wantOK: true},
		{name: "invalid start", value: "bytes=-", size: 1000, wantOK: false},
		{name: "start past end", value: "bytes=1000-", size: 1000, wantOK: false},
		{name: "reversed", value: "bytes=20-10", size: 1000, wantOK: false},
		{name: "wrong unit", value: "items=0-10", size: 1000, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, ok := parseByteRange(tt.value, tt.size)
			if ok != tt.wantOK || (ok && (start != tt.wantStart || end != tt.wantEnd)) {
				t.Fatalf("parseByteRange(%q, %d) = (%d, %d, %v), want (%d, %d, %v)", tt.value, tt.size, start, end, ok, tt.wantStart, tt.wantEnd, tt.wantOK)
			}
		})
	}
}

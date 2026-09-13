package main

import "testing"

func TestValidateSegment(t *testing.T) {
	cases := []struct { name string; start,end,duration,max float64; wantErr bool }{
		{"valid", 5, 35, 60, 60, false},
		{"negative start", -1, 10, 60, 60, true},
		{"reversed", 20, 10, 60, 60, true},
		{"past duration", 0, 61, 60, 60, true},
		{"too long", 0, 61, 120, 60, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateSegment(tc.start, tc.end, tc.duration, tc.max); (err != nil) != tc.wantErr {
				t.Fatalf("validateSegment() error=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestProviderRegistryDefaultsToInternal(t *testing.T) {
	t.Setenv("MUSIC_PROVIDER", "internal")
	p := newProviderRegistry().current()
	if p.Name() != "internal" || !p.Ready() { t.Fatalf("unexpected default provider: %s", p.Name()) }
}

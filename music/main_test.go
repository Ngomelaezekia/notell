package main

import (
	"encoding/json"
	"testing"
	"time"
)

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

func TestMemoryTrendStoreRanksAndCountsEvents(t *testing.T) {
	now := time.Now().UTC()
	s := &memoryTrendStore{}
	for _, e := range []MusicEvent{
		{TrackID: "track-a", Type: "play", Country: "TZ", City: "Dar es Salaam", At: now.Add(-30 * time.Minute)},
		{TrackID: "track-a", Type: "use", Country: "TZ", City: "Dar es Salaam", At: now.Add(-20 * time.Minute)},
		{TrackID: "track-a", Type: "save", Country: "TZ", City: "Dar es Salaam", At: now.Add(-10 * time.Minute)},
		{TrackID: "track-b", Type: "play", Country: "TZ", City: "Arusha", At: now.Add(-15 * time.Minute)},
	} {
		if err := s.Add(e); err != nil { t.Fatal(err) }
	}
	items, err := s.Rank("TZ", "Dar es Salaam", time.Hour)
	if err != nil { t.Fatal(err) }
	if len(items) != 1 || items[0].TrackID != "track-a" { t.Fatalf("unexpected ranked tracks: %#v", items) }
	if items[0].Plays != 1 || items[0].Uses != 1 || items[0].Saves != 1 { t.Fatalf("unexpected counters: %#v", items[0]) }
	if items[0].Score != 14 { t.Fatalf("unexpected score: %v", items[0].Score) }
}

func TestPositiveEnvInt(t *testing.T) {
	t.Setenv("MUSIC_TEST_INT", "12")
	if got := positiveEnvInt("MUSIC_TEST_INT", 7); got != 12 { t.Fatalf("got %d, want 12", got) }
	t.Setenv("MUSIC_TEST_INT", "0")
	if got := positiveEnvInt("MUSIC_TEST_INT", 7); got != 7 { t.Fatalf("got %d, want fallback 7", got) }
	t.Setenv("MUSIC_TEST_INT", "invalid")
	if got := positiveEnvInt("MUSIC_TEST_INT", 7); got != 7 { t.Fatalf("got %d, want fallback 7", got) }
}

func TestStartTrendMaintenanceNoopForMemoryStore(t *testing.T) {
	startTrendMaintenance(&memoryTrendStore{})
}

func TestRightsSyncPayloadRejectsUnknownFields(t *testing.T) {
	payload := `{"records":[{"provider":"massivemusic","providerTrackId":"1","territory":"tz","licensed":true,"ugcUse":true,"streaming":true,"canUseInPost":true,"status":"active","unexpected":true}]}`
	var decoded rightsSyncPayload
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err == nil { t.Fatal("expected unknown field rejection") }
}

func TestRightsSyncFullSnapshotRequiresSingleProvider(t *testing.T) {
	payload := rightsSyncPayload{Full: true, Records: []rightsSyncRecord{
		{Provider: "MassiveMusic", ProviderTrackID: "1", Territory: "tz"},
		{Provider: "7digital", ProviderTrackID: "2", Territory: "tz"},
	}}
	provider := ""
	for _, record := range payload.Records {
		p := strings.ToLower(strings.TrimSpace(record.Provider))
		if provider == "" { provider = p; continue }
		if p != provider { return }
	}
	t.Fatal("expected provider mismatch to be rejected")
}

func TestRightsSyncNormalization(t *testing.T) {
	record := rightsSyncRecord{Provider: " MassiveMusic ", ProviderTrackID: " track-1 ", Territory: " tz "}
	record.Provider = strings.ToLower(strings.TrimSpace(record.Provider))
	record.ProviderTrackID = strings.TrimSpace(record.ProviderTrackID)
	record.Territory = strings.ToUpper(strings.TrimSpace(record.Territory))
	if record.Provider != "massivemusic" || record.ProviderTrackID != "track-1" || record.Territory != "TZ" { t.Fatalf("unexpected normalization: %#v", record) }
}

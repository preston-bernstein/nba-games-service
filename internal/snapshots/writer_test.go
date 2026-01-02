package snapshots

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"nba-data-service/internal/domain"
)

func TestWriterWritesSnapshotAndManifest(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir, 10)

	today := time.Now().Format("2006-01-02")
	snap := domain.TodayResponse{
		Date:  today,
		Games: []domain.Game{{ID: "g1"}},
	}

	writeSnapshot(t, w, today, snap)

	// Verify snapshot file exists.
	data, err := os.ReadFile(filepath.Join(dir, "games", today+".json"))
	if err != nil {
		t.Fatalf("expected snapshot file, got err %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected snapshot content")
	}

	// Verify manifest was written.
	mBytes, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatalf("expected manifest, got err %v", err)
	}
	if len(mBytes) == 0 {
		t.Fatalf("expected manifest content")
	}
}

func TestWriterPrunesOldSnapshots(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir, 1) // 1-day retention

	oldDate := time.Now().AddDate(0, 0, -5).Format("2006-01-02")
	newDate := time.Now().Format("2006-01-02")

	// Write an old snapshot and a new one.
	for _, d := range []string{oldDate, newDate} {
		snap := domain.TodayResponse{
			Date:  d,
			Games: []domain.Game{{ID: d}},
		}
		writeSnapshot(t, w, d, snap)
	}

	// Old snapshot should be pruned.
	if _, err := os.Stat(filepath.Join(dir, "games", oldDate+".json")); err == nil {
		t.Fatalf("expected old snapshot to be pruned")
	}
	if _, err := os.Stat(filepath.Join(dir, "games", newDate+".json")); err != nil {
		t.Fatalf("expected new snapshot to exist")
	}
}

func TestWriterHandlesNilAndEmptyDate(t *testing.T) {
	var w *Writer
	if err := w.WriteGamesSnapshot("2024-01-01", domain.TodayResponse{}); err == nil {
		t.Fatalf("expected error for nil writer")
	}

	w = NewWriter(t.TempDir(), 1)
	if err := w.WriteGamesSnapshot("", domain.TodayResponse{}); err == nil {
		t.Fatalf("expected error for empty date")
	}
}

func TestNewWriterDefaultsRetention(t *testing.T) {
	w := NewWriter(t.TempDir(), 0)
	if w.retentionDays <= 0 {
		t.Fatalf("expected retention to default when non-positive provided")
	}
}

func TestListGameDatesIgnoresNonJSONAndDirs(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "games", "nested"), 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "games", "2024-01-01.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to write snapshot: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "games", "ignore.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to write extra file: %v", err)
	}

	w := NewWriter(dir, 1)
	dates, err := w.listGameDates()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(dates) != 1 || dates[0] != "2024-01-01" {
		t.Fatalf("expected only json snapshots, got %v", dates)
	}
}

func TestBasePathExposesRoot(t *testing.T) {
	base := t.TempDir()
	w := NewWriter(base, 1)
	if w.BasePath() != base {
		t.Fatalf("expected base path %s, got %s", base, w.BasePath())
	}
}

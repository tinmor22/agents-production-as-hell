package metrics

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAppend_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	e := Entry{
		EntityA: "Messi",
		EntityB: "Cristiano",
		Formats: 2,
		Elapsed: 1.5,
		Success: true,
	}
	if err := Append(dir, e); err != nil {
		t.Fatalf("Append: %v", err)
	}

	path := filepath.Join(dir, fileName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("metrics file is empty")
	}
}

func TestCountSince(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	for i := 0; i < 3; i++ {
		e := Entry{
			Timestamp: now.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			EntityA:   "A",
			EntityB:   "B",
			Success:   true,
		}
		if err := Append(dir, e); err != nil {
			t.Fatalf("Append %d: %v", i, err)
		}
	}

	count, err := CountSince(dir, now.Add(90*time.Minute))
	if err != nil {
		t.Fatalf("CountSince: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestCountSince_MissingFile(t *testing.T) {
	count, err := CountSince(t.TempDir(), time.Now())
	if err != nil {
		t.Fatalf("CountSince: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

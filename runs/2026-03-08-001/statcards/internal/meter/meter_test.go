package meter

import (
	"errors"
	"testing"
	"time"
)

func TestCheck_Allow(t *testing.T) {
	dc := &DailyCounter{Date: today(), Count: 4}
	if err := Check(dc, 5); err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
}

func TestCheck_Block(t *testing.T) {
	dc := &DailyCounter{Date: today(), Count: 5}
	err := Check(dc, 5)
	if err == nil {
		t.Fatal("expected error when at limit")
	}
	if !errors.Is(err, ErrDailyLimitReached) {
		t.Errorf("expected ErrDailyLimitReached, got: %v", err)
	}
}

func TestCheck_DateReset(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	dc := &DailyCounter{Date: yesterday, Count: 99}

	if err := Check(dc, 5); err != nil {
		t.Errorf("expected nil after date reset, got: %v", err)
	}
	if dc.Count != 0 {
		t.Errorf("expected count=0 after reset, got: %d", dc.Count)
	}
	if dc.Date != today() {
		t.Errorf("expected date=%s after reset, got: %s", today(), dc.Date)
	}
}

func TestIncrement_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	dc, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if dc.Count != 0 {
		t.Fatalf("expected fresh counter count=0, got: %d", dc.Count)
	}

	Increment(dc)
	if dc.Count != 1 {
		t.Errorf("expected count=1 after increment, got: %d", dc.Count)
	}

	if err := Save(dc, dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if loaded.Count != 1 {
		t.Errorf("expected count=1 after reload, got: %d", loaded.Count)
	}
}

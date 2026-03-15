package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Missing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "config.json")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	if !errors.Is(err, ErrNotConfigured) {
		t.Errorf("expected ErrNotConfigured, got: %v", err)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	original := &Config{
		APIKey:    "testkey123",
		Watermark: "@futbol_stats",
	}
	if err := Save(original, path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.APIKey != original.APIKey {
		t.Errorf("APIKey: got %q, want %q", loaded.APIKey, original.APIKey)
	}
	if loaded.Watermark != original.Watermark {
		t.Errorf("Watermark: got %q, want %q", loaded.Watermark, original.Watermark)
	}
}

func TestDailyLimit_Default(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := &Config{APIKey: "key"}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.DailyLimit != defaultLimit {
		t.Errorf("DailyLimit: got %d, want %d", loaded.DailyLimit, defaultLimit)
	}
}

func TestPlan_Default(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := &Config{APIKey: "key"}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Plan != defaultPlan {
		t.Errorf("Plan: got %q, want %q", loaded.Plan, defaultPlan)
	}
}

func TestSave_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "deep")
	path := filepath.Join(dir, "config.json")

	cfg := &Config{APIKey: "key"}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory to be created")
	}
}

package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/statcard/statcard/internal/config"
	"github.com/statcard/statcard/internal/meter"
)

func TestVersion_Subcommand(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "statcard") {
		t.Errorf("expected 'statcard' in output, got: %q", out)
	}
	if !strings.Contains(out, "dev") {
		t.Errorf("expected 'dev' in output, got: %q", out)
	}
}

func TestRootHelp(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("help failed: %v", err)
	}
	out := buf.String()
	for _, sub := range []string{"init", "status", "version"} {
		if !strings.Contains(out, sub) {
			t.Errorf("help output missing subcommand %q", sub)
		}
	}
}

func TestInit_SavesConfig(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()
	_ = tmpDir

	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"init", "--api-key", "testkey123"})

	if err := root.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if !strings.Contains(buf.String(), "exitosamente") {
		t.Errorf("expected success message, got: %q", buf.String())
	}
}

func TestStatus_AfterInit(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Init config
	cfg := &config.Config{APIKey: "testkey123"}
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("save config: %v", err)
	}

	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"status"})
	if err := root.Execute(); err != nil {
		t.Fatalf("status failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "free") {
		t.Errorf("expected 'free' plan, got: %q", out)
	}
	if !strings.Contains(out, "0/5") {
		t.Errorf("expected '0/5' usage, got: %q", out)
	}
}

func TestStatus_WithUsage(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := &config.Config{APIKey: "testkey123"}
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("save config: %v", err)
	}

	dc := &meter.DailyCounter{
		Date:  time.Now().Format("2006-01-02"),
		Count: 2,
	}
	if err := meter.Save(dc, meterDir); err != nil {
		t.Fatalf("save meter: %v", err)
	}

	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"status"})
	if err := root.Execute(); err != nil {
		t.Fatalf("status failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "2/5") {
		t.Errorf("expected '2/5' usage, got: %q", out)
	}
	if !strings.Contains(out, "Restantes: 3") {
		t.Errorf("expected 'Restantes: 3', got: %q", out)
	}
}

func TestStatus_BeforeInit(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"status"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for status before init")
	}
}

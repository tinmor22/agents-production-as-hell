package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/statcard/statcard/internal/apifootball"
	"github.com/statcard/statcard/internal/config"
	"github.com/statcard/statcard/internal/meter"
)

// mockClient implements apifootball.ClientInterface for testing.
type mockClient struct {
	playersA []apifootball.PlayerResult
	playersB []apifootball.PlayerResult
	h2h      *apifootball.H2HStats
	err      error
	calls    int
}

func (m *mockClient) SearchPlayer(name string) ([]apifootball.PlayerResult, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	if m.calls == 1 {
		return m.playersA, nil
	}
	return m.playersB, nil
}

func (m *mockClient) SearchTeam(name string) ([]apifootball.TeamResult, error) {
	return nil, nil
}

func (m *mockClient) GetH2H(teamA, teamB int) (*apifootball.H2HStats, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.h2h, nil
}

func setupTestEnv(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()
	tmpDir = t.TempDir()

	oldCfgPath := configPath
	oldMeterDir := meterDir
	oldMetricsDir := metricsDir
	oldClient := apiClient

	configPath = filepath.Join(tmpDir, "config.json")
	meterDir = filepath.Join(tmpDir, "meter")
	metricsDir = filepath.Join(tmpDir, "metrics")

	return tmpDir, func() {
		configPath = oldCfgPath
		meterDir = oldMeterDir
		metricsDir = oldMetricsDir
		apiClient = oldClient
	}
}

func TestGenerate_Happy(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Setup config
	cfg := &config.Config{APIKey: "testkey123"}
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("save config: %v", err)
	}

	// Setup mock client
	mc := &mockClient{
		playersA: []apifootball.PlayerResult{
			{ID: 1, Name: "Lionel Messi", TeamName: "Inter Miami", Goals: 20, Assists: 16, Games: 19},
		},
		playersB: []apifootball.PlayerResult{
			{ID: 2, Name: "Cristiano Ronaldo", TeamName: "Al Nassr", Goals: 35, Assists: 11, Games: 31},
		},
	}
	apiClient = mc

	outDir := filepath.Join(tmpDir, "output")
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"--output-dir", outDir, "Messi vs Cristiano"})

	if err := root.Execute(); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, ".png") {
		t.Errorf("expected .png paths in output, got: %q", out)
	}
	if strings.Count(out, ".png") < 2 {
		t.Errorf("expected at least 2 .png paths, got: %q", out)
	}
}

func TestGenerate_NoConfig(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"Messi vs Cristiano"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when config missing")
	}
}

func TestGenerate_LimitReached(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := &config.Config{APIKey: "testkey123"}
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("save config: %v", err)
	}

	// Pre-fill the meter to exactly the limit (5)
	dc := &meter.DailyCounter{
		Date:  time.Now().Format("2006-01-02"),
		Count: 5,
	}
	if err := meter.Save(dc, meterDir); err != nil {
		t.Fatalf("save meter: %v", err)
	}

	apiClient = &mockClient{
		playersA: []apifootball.PlayerResult{{ID: 1, Name: "Messi"}},
		playersB: []apifootball.PlayerResult{{ID: 2, Name: "Cristiano"}},
	}

	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"Messi vs Cristiano"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when limit reached")
	}
	if !strings.Contains(err.Error(), "limite diario") {
		t.Errorf("expected Spanish limit error, got: %v", err)
	}
}

func TestGenerate_ParseError(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := &config.Config{APIKey: "testkey123"}
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("save config: %v", err)
	}

	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"solo messi"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for bad prompt")
	}
	if !strings.Contains(err.Error(), "separador") {
		t.Errorf("expected Spanish error about separator, got: %v", err)
	}
}

func TestGenerate_NoArgs_ShowsHelp(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "statcard") {
		t.Error("expected help output")
	}
}

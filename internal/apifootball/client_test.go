package apifootball

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
)

// mockHTTP returns a fixture file as the HTTP response.
type mockHTTP struct {
	fixturePath string
	callCount   atomic.Int32
}

func (m *mockHTTP) Do(req *http.Request) (*http.Response, error) {
	m.callCount.Add(1)
	data, err := os.ReadFile(m.fixturePath)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(data)),
		Header:     make(http.Header),
	}, nil
}

func TestSearchPlayer_Fixture(t *testing.T) {
	mock := &mockHTTP{fixturePath: "testdata/player-search-messi.json"}
	client := NewClient("testkey", t.TempDir(), mock)

	results, err := client.SearchPlayer("Messi")
	if err != nil {
		t.Fatalf("SearchPlayer: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Name != "Lionel Messi" {
		t.Errorf("Name: got %q, want 'Lionel Messi'", results[0].Name)
	}
	if results[0].ID != 154 {
		t.Errorf("ID: got %d, want 154", results[0].ID)
	}
	if results[0].Goals != 20 {
		t.Errorf("Goals: got %d, want 20", results[0].Goals)
	}
	if results[0].Assists != 16 {
		t.Errorf("Assists: got %d, want 16", results[0].Assists)
	}
	if results[0].Games != 19 {
		t.Errorf("Games: got %d, want 19", results[0].Games)
	}
}

func TestSearchPlayer_CacheHit(t *testing.T) {
	mock := &mockHTTP{fixturePath: "testdata/player-search-messi.json"}
	cacheDir := t.TempDir()
	client := NewClient("testkey", cacheDir, mock)

	// First call: hits HTTP
	_, err := client.SearchPlayer("Messi")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Second call: should hit cache, not HTTP
	_, err = client.SearchPlayer("Messi")
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if got := mock.callCount.Load(); got != 1 {
		t.Errorf("expected Do called once, got %d", got)
	}
}

func TestSearchPlayer_EmptyResults(t *testing.T) {
	mock := &mockHTTP{fixturePath: "testdata/player-search-empty.json"}
	client := NewClient("testkey", t.TempDir(), mock)

	results, err := client.SearchPlayer("xyznonexistent")
	if err != nil {
		t.Fatalf("SearchPlayer: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty slice, got %d results", len(results))
	}
}

func TestGetH2H_Fixture(t *testing.T) {
	mock := &mockHTTP{fixturePath: "testdata/h2h-fixture.json"}
	client := NewClient("testkey", t.TempDir(), mock)

	stats, err := client.GetH2H(529, 541)
	if err != nil {
		t.Fatalf("GetH2H: %v", err)
	}

	if stats.Matches != 3 {
		t.Errorf("Matches: got %d, want 3", stats.Matches)
	}
	// Fixture: Barca 3-1 RM (WinA), RM 2-2 Barca (Draw), Barca 0-1 RM (WinB)
	// teamA=529 (Barcelona)
	if stats.WinsA != 1 {
		t.Errorf("WinsA: got %d, want 1", stats.WinsA)
	}
	if stats.WinsB != 1 {
		t.Errorf("WinsB: got %d, want 1", stats.WinsB)
	}
	if stats.Draws != 1 {
		t.Errorf("Draws: got %d, want 1", stats.Draws)
	}
	// GoalsA: 3 + 2 + 0 = 5 (Barcelona's goals)
	if stats.GoalsA != 5 {
		t.Errorf("GoalsA: got %d, want 5", stats.GoalsA)
	}
	// GoalsB: 1 + 2 + 1 = 4 (Real Madrid's goals)
	if stats.GoalsB != 4 {
		t.Errorf("GoalsB: got %d, want 4", stats.GoalsB)
	}
}

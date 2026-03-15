package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	fileName = "metrics.jsonl"
	filePerm = 0600
)

// Entry records one card generation event.
type Entry struct {
	Timestamp string  `json:"timestamp"`
	EntityA   string  `json:"entity_a"`
	EntityB   string  `json:"entity_b"`
	Formats   int     `json:"formats"`
	Elapsed   float64 `json:"elapsed_seconds"`
	Success   bool    `json:"success"`
	Error     string  `json:"error,omitempty"`
}

// Append writes a metrics entry to the JSONL file in the given directory.
func Append(dir string, entry Entry) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create metrics dir: %w", err)
	}
	path := filepath.Join(dir, fileName)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePerm)
	if err != nil {
		return fmt.Errorf("open metrics file: %w", err)
	}
	defer f.Close()

	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal metrics entry: %w", err)
	}
	data = append(data, '\n')
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write metrics entry: %w", err)
	}
	return nil
}

// CountSince returns the number of entries since the given time.
func CountSince(dir string, since time.Time) (int, error) {
	path := filepath.Join(dir, fileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read metrics: %w", err)
	}

	count := 0
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		t, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			continue
		}
		if !t.Before(since) {
			count++
		}
	}
	return count, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

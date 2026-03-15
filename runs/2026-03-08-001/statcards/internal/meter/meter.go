package meter

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ErrDailyLimitReached is returned when the daily card limit is exceeded.
var ErrDailyLimitReached = errors.New("limite diario alcanzado: intenta de nuevo manana")

const (
	fileName = "counter.json"
	filePerm = 0600
)

// DailyCounter tracks daily card generation count.
type DailyCounter struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// Load reads the daily counter from the given directory.
// Returns a zero counter for today if the file does not exist.
func Load(dir string) (*DailyCounter, error) {
	path := filepath.Join(dir, fileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &DailyCounter{Date: today()}, nil
		}
		return nil, fmt.Errorf("read counter: %w", err)
	}
	var dc DailyCounter
	if err := json.Unmarshal(data, &dc); err != nil {
		return nil, fmt.Errorf("parse counter: %w", err)
	}
	return &dc, nil
}

// Save writes the daily counter to the given directory.
func Save(dc *DailyCounter, dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create counter dir: %w", err)
	}
	path := filepath.Join(dir, fileName)
	data, err := json.Marshal(dc)
	if err != nil {
		return fmt.Errorf("marshal counter: %w", err)
	}
	if err := os.WriteFile(path, data, filePerm); err != nil {
		return fmt.Errorf("write counter: %w", err)
	}
	return nil
}

// Check validates that the daily limit has not been reached.
// Resets the counter if the date has changed.
func Check(dc *DailyCounter, limit int) error {
	if dc.Date != today() {
		dc.Date = today()
		dc.Count = 0
	}
	if dc.Count >= limit {
		return ErrDailyLimitReached
	}
	return nil
}

// Increment adds one to the daily count, setting today's date.
func Increment(dc *DailyCounter) {
	dc.Date = today()
	dc.Count++
}

func today() string {
	return time.Now().Format("2006-01-02")
}

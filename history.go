package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type HistoryEntry struct {
	Pillar    string `yaml:"pillar"`
	Timestamp string `yaml:"timestamp"`
	Date      string `yaml:"date"`
	Block     int    `yaml:"block"`
}

type History struct {
	Entries []HistoryEntry `yaml:"entries"`
}

const historyPath = "data/history.yaml"

func loadHistory() (History, error) {
	var history History

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return History{Entries: []HistoryEntry{}}, nil
		}
		return history, fmt.Errorf("could not read history: %w", err)
	}

	err = yaml.Unmarshal(data, &history)
	if err != nil {
		return history, fmt.Errorf("could not parse history: %w", err)
	}

	if history.Entries == nil {
		history.Entries = []HistoryEntry{}
	}

	return history, nil
}

func saveHistory(history History) error {
	data, err := yaml.Marshal(&history)
	if err != nil {
		return fmt.Errorf("could not serialize history: %w", err)
	}

	err = os.WriteFile(historyPath, data, 0644)
	if err != nil {
		return fmt.Errorf("could not write history: %w", err)
	}

	return nil
}

func isBlockDue(entries []HistoryEntry, blockDays int) bool {
	return isBlockDueAt(entries, blockDays, time.Now())
}

func isBlockDueAt(entries []HistoryEntry, blockDays int, now time.Time) bool {
	if len(entries) == 0 {
		return true
	}

	last := entries[len(entries)-1]
	if last.Date == "" {
		return true
	}

	lastDate, err := time.ParseInLocation("2006-01-02", last.Date, now.Location())
	if err != nil {
		return true
	}

	return calendarDaysBetween(lastDate, now) >= blockDays
}

func calendarDaysBetween(from, to time.Time) int {
	y1, m1, d1 := from.Date()
	y2, m2, d2 := to.Date()
	utcFrom := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	utcTo := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
	return int(utcTo.Sub(utcFrom).Hours() / 24)
}


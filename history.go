package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type HistoryEntry struct {
	Pillar    string `yaml:"pillar"`
	Timestamp string `yaml:"timestamp"`
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

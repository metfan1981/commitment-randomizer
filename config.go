package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Pillar struct {
	Name    string `yaml:"name"`
	Weight  int    `yaml:"weight"`
	Journal string `yaml:"journal,omitempty"`
}

type Config struct {
	BlockDays        int      `yaml:"block_days"`
	MaxConsecutive   int      `yaml:"max_consecutive"`
	CorrectionFactor float64  `yaml:"correction_factor"`
	Pillars          []Pillar `yaml:"pillars"`
}

const configPath = "data/config.yaml"

func loadConfig() (Config, error) {
	var config Config

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("could not read config: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("could not parse config: %w", err)
	}

	return config, nil
}

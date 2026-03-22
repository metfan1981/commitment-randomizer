package main

import (
	"fmt"
	"testing"
)

func TestSimulate(t *testing.T) {
	config, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}

	yearBlocks := 365 / config.BlockDays
	quarterBlocks := 90 / config.BlockDays

	totalWeight := 0
	for _, p := range config.Pillars {
		totalWeight += p.Weight
	}

	history := History{Entries: []HistoryEntry{}}

	fmt.Printf("\n=== Simulating 1 year: %d blocks (%d-day blocks) ===\n",
		yearBlocks, config.BlockDays)
	fmt.Printf("Config: max_consecutive=%d, correction_factor=%.0f\n\n",
		config.MaxConsecutive, config.CorrectionFactor)

	quarterCounts := make(map[string]int)

	for i := 0; i < yearBlocks; i++ {
		streaks := recentStreaks(history.Entries, config.MaxConsecutive)
		candidates := eligiblePillars(config.Pillars, streaks, config.MaxConsecutive)
		candidates = adjustedWeights(candidates, history.Entries, config.CorrectionFactor)
		chosen := weightedPick(candidates)

		entry := HistoryEntry{
			Pillar: chosen.Name,
			Block:  nextBlockIndex(history.Entries),
		}
		history.Entries = append(history.Entries, entry)

		fmt.Printf("  Block %3d: %s\n", entry.Block, chosen.Name)

		if i < quarterBlocks {
			quarterCounts[chosen.Name]++
		}
	}

	yearCounts := make(map[string]int)
	for _, e := range history.Entries {
		yearCounts[e.Pillar]++
	}

	printDistribution := func(label string, days int, numBlocks int, counts map[string]int) {
		fmt.Printf("\n=== %s: %d blocks (%d days) ===\n", label, numBlocks, days)
		fmt.Printf("  %-45s %6s %6s %6s %6s %7s\n", "Pillar", "Target", "Actual", "Tgt#", "Act#", "Delta")
		fmt.Println("  " + "---------------------------------------------------------------------------------")

		for _, p := range config.Pillars {
			targetPct := float64(p.Weight) / float64(totalWeight) * 100
			actualPct := float64(counts[p.Name]) / float64(numBlocks) * 100
			targetN := float64(p.Weight) / float64(totalWeight) * float64(numBlocks)
			actualN := counts[p.Name]
			delta := actualPct - targetPct
			sign := "+"
			if delta <= 0 {
				sign = ""
			}
			fmt.Printf("  %-45s %5.1f%% %5.1f%% %5.1f  %4d   %s%.1f%%\n",
				p.Name, targetPct, actualPct, targetN, actualN, sign, delta)
		}

		fmt.Println()
		fmt.Printf("  %-45s %6s\n", "Pillar", "Days")
		fmt.Println("  " + "---------------------------------------------------------------------------------")
		for _, p := range config.Pillars {
			targetDays := float64(p.Weight) / float64(totalWeight) * float64(days)
			actualDays := counts[p.Name] * config.BlockDays
			daysDelta := actualDays - int(targetDays+0.5)
			sign := "+"
			if daysDelta <= 0 {
				sign = ""
			}
			fmt.Printf("  %-45s target %3.0f  actual %3d  (%s%d days)\n",
				p.Name, targetDays, actualDays, sign, daysDelta)
		}
	}

	printDistribution("First 90 days", 90, quarterBlocks, quarterCounts)
	printDistribution("Full year", 365, yearBlocks, yearCounts)

	fmt.Println("\n=== Streaks observed (full year) ===")
	fmt.Printf("  %-45s %10s\n", "Pillar", "Max streak")
	fmt.Println("  " + "---------------------------------------------------------------------------------")
	for _, p := range config.Pillars {
		maxS := 0
		cur := 0
		for _, e := range history.Entries {
			if e.Pillar == p.Name {
				cur++
				if cur > maxS {
					maxS = cur
				}
			} else {
				cur = 0
			}
		}
		fmt.Printf("  %-45s %6d blocks (%d days)\n",
			p.Name, maxS, maxS*config.BlockDays)
	}
	fmt.Println()
}

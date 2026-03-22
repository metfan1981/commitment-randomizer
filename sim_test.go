package main

import (
	"fmt"
	"testing"
)

const simDays = 90
const monteCarloRuns = 1000

func TestSimulate(t *testing.T) {
	config, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}

	numBlocks := simDays / config.BlockDays

	totalWeight := 0
	for _, p := range config.Pillars {
		totalWeight += p.Weight
	}

	history := History{Entries: []HistoryEntry{}}

	fmt.Printf("\n=== Simulating %d block assignments (%d days, %d-day blocks) ===\n",
		numBlocks, simDays, config.BlockDays)
	fmt.Printf("Config: max_consecutive=%d, correction_factor=%.0f\n\n",
		config.MaxConsecutive, config.CorrectionFactor)

	for i := 0; i < numBlocks; i++ {
		streaks := recentStreaks(history.Entries, config.MaxConsecutive)
		candidates := eligiblePillars(config.Pillars, streaks, config.MaxConsecutive)
		candidates = adjustedWeights(candidates, history.Entries, config.CorrectionFactor)
		chosen := weightedPick(candidates)

		entry := HistoryEntry{
			Pillar: chosen.Name,
			Block:  nextBlockIndex(history.Entries),
		}
		history.Entries = append(history.Entries, entry)

		fmt.Printf("  Block %2d: %s\n", entry.Block, chosen.Name)
	}

	counts := make(map[string]int)
	for _, e := range history.Entries {
		counts[e.Pillar]++
	}

	fmt.Println("\n=== Distribution vs Target ===")
	fmt.Printf("  %-45s %6s %6s %6s %6s\n", "Pillar", "Target", "Actual", "Tgt#", "Act#")
	fmt.Println("  " + "-----------------------------------------------------------------------")

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

	fmt.Printf("=== Monte Carlo: %d runs of %d blocks (average distribution) ===\n", monteCarloRuns, numBlocks)
	totalCounts := make(map[string]int)

	for r := 0; r < monteCarloRuns; r++ {
		h := History{Entries: []HistoryEntry{}}
		for i := 0; i < numBlocks; i++ {
			streaks := recentStreaks(h.Entries, config.MaxConsecutive)
			candidates := eligiblePillars(config.Pillars, streaks, config.MaxConsecutive)
			candidates = adjustedWeights(candidates, h.Entries, config.CorrectionFactor)
			chosen := weightedPick(candidates)
			h.Entries = append(h.Entries, HistoryEntry{Pillar: chosen.Name, Block: i + 1})
		}
		for _, e := range h.Entries {
			totalCounts[e.Pillar]++
		}
	}

	fmt.Printf("  %-45s %6s %6s %6s\n", "Pillar", "Target", "Avg", "Delta")
	fmt.Println("  " + "-----------------------------------------------------------------------")
	for _, p := range config.Pillars {
		targetPct := float64(p.Weight) / float64(totalWeight) * 100
		avgPct := float64(totalCounts[p.Name]) / float64(monteCarloRuns*numBlocks) * 100
		delta := avgPct - targetPct
		sign := "+"
		if delta <= 0 {
			sign = ""
		}
		fmt.Printf("  %-45s %5.1f%% %5.1f%%  %s%.1f%%\n",
			p.Name, targetPct, avgPct, sign, delta)
	}
	fmt.Println()

	fmt.Printf("=== Streak stats (from %d runs) ===\n", monteCarloRuns)
	type streakStat struct {
		maxStreak int
		total     int
		count     int
	}
	streakStats := make(map[string]*streakStat)
	for _, p := range config.Pillars {
		streakStats[p.Name] = &streakStat{}
	}

	for r := 0; r < monteCarloRuns; r++ {
		h := History{Entries: []HistoryEntry{}}
		for i := 0; i < numBlocks; i++ {
			s := recentStreaks(h.Entries, config.MaxConsecutive)
			c := eligiblePillars(config.Pillars, s, config.MaxConsecutive)
			c = adjustedWeights(c, h.Entries, config.CorrectionFactor)
			chosen := weightedPick(c)
			h.Entries = append(h.Entries, HistoryEntry{Pillar: chosen.Name, Block: i + 1})
		}

		for _, p := range config.Pillars {
			maxS := 0
			cur := 0
			for _, e := range h.Entries {
				if e.Pillar == p.Name {
					cur++
					if cur > maxS {
						maxS = cur
					}
				} else {
					cur = 0
				}
			}
			ss := streakStats[p.Name]
			if maxS > ss.maxStreak {
				ss.maxStreak = maxS
			}
			ss.total += maxS
			ss.count++
		}
	}

	fmt.Printf("  %-45s %10s %10s\n", "Pillar", "Avg Max", "Overall Max")
	fmt.Println("  " + "-----------------------------------------------------------------------")
	for _, p := range config.Pillars {
		ss := streakStats[p.Name]
		fmt.Printf("  %-45s %8.1f %10d\n",
			p.Name, float64(ss.total)/float64(ss.count), ss.maxStreak)
	}
	fmt.Println()
}

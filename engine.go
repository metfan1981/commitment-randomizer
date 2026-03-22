package main

import (
	"fmt"
	"math/rand"
)

// Looks at the tail of history and counts how many times
// the same pillar appears consecutively.
func recentStreaks(entries []HistoryEntry, maxLookback int) map[string]int {
	streaks := make(map[string]int)

	if len(entries) == 0 {
		return streaks
	}

	lastPillar := ""
	count := 0

	for i := len(entries) - 1; i >= 0 && count < maxLookback; i-- {
		current := entries[i].Pillar

		if current == lastPillar || lastPillar == "" {
			lastPillar = current
			count++
		} else {
			break
		}
	}

	if lastPillar != "" {
		streaks[lastPillar] = count
	}

	return streaks
}

func eligiblePillars(pillars []Pillar, streaks map[string]int, maxConsecutive int) []Pillar {
	var eligible []Pillar

	for _, p := range pillars {
		consecutiveCount, exists := streaks[p.Name]
		if !exists || consecutiveCount < maxConsecutive {
			eligible = append(eligible, p)
		}
	}

	if len(eligible) == 0 {
		fmt.Println("(All pillars hit max streak — resetting eligibility)")
		return pillars
	}

	return eligible
}

// Compares each pillar's target share vs actual share in history,
// then boosts underrepresented pillars and dampens overrepresented ones.
// At correction_factor=0 this is a no-op (pure raw weights).
// At 3-5 it self-corrects monthly.
func adjustedWeights(pillars []Pillar, entries []HistoryEntry, correctionFactor float64) []Pillar {
	totalBlocks := len(entries)

	if totalBlocks < 3 || correctionFactor == 0 {
		return pillars
	}

	counts := make(map[string]int)
	for _, e := range entries {
		counts[e.Pillar]++
	}

	totalWeight := 0
	for _, p := range pillars {
		totalWeight += p.Weight
	}

	adjusted := make([]Pillar, len(pillars))
	for i, p := range pillars {
		targetShare := float64(p.Weight) / float64(totalWeight)
		actualShare := float64(counts[p.Name]) / float64(totalBlocks)
		deficit := targetShare - actualShare

		multiplier := 1.0 + deficit*correctionFactor
		if multiplier < 0.1 {
			multiplier = 0.1
		}

		effectiveWeight := max(int(float64(p.Weight)*multiplier+0.5), 1)
		adjusted[i] = Pillar{Name: p.Name, Weight: effectiveWeight, Journal: p.Journal}
	}

	return adjusted
}

// Weighted random selection: cumulative weight array,
// roll a random number, pick where it lands.
func weightedPick(pillars []Pillar) Pillar {
	totalWeight := 0
	for _, p := range pillars {
		totalWeight += p.Weight
	}

	roll := rand.Intn(totalWeight)

	cumulative := 0
	for _, p := range pillars {
		cumulative += p.Weight
		if roll < cumulative {
			return p
		}
	}

	return pillars[len(pillars)-1]
}

func nextBlockIndex(entries []HistoryEntry) int {
	if len(entries) == 0 {
		return 1
	}
	return entries[len(entries)-1].Block + 1
}

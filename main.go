package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

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

type HistoryEntry struct {
	Pillar    string `yaml:"pillar"`
	Timestamp string `yaml:"timestamp"`
	Block     int    `yaml:"block"`
}

type History struct {
	Entries []HistoryEntry `yaml:"entries"`
}

const configPath = "config.yaml"
const historyPath = "history.yaml"


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


func loadHistory() (History, error) {
	var history History

	data, err := os.ReadFile(historyPath)
	if err != nil {
		// First run — no history yet, that's fine
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

// --- Save history back to YAML ---

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

// --- Count recent consecutive assignments for each pillar ---
// Looks at the tail of history and counts how many times
// the same pillar appears in a row, per pillar.

func recentStreaks(entries []HistoryEntry, maxLookback int) map[string]int {
	streaks := make(map[string]int)

	if len(entries) == 0 {
		return streaks
	}

	// Walk backwards from most recent entry
	lastPillar := ""
	count := 0

	for i := len(entries) - 1; i >= 0 && count < maxLookback; i-- {
		current := entries[i].Pillar

		if current == lastPillar || lastPillar == "" {
			lastPillar = current
			count++
		} else {
			// Streak broken — we only care about the most recent streak
			break
		}
	}

	if lastPillar != "" {
		streaks[lastPillar] = count
	}

	return streaks
}

// --- Filter out pillars that hit max_consecutive ---

func eligiblePillars(pillars []Pillar, streaks map[string]int, maxConsecutive int) []Pillar {
	var eligible []Pillar

	for _, p := range pillars {
		consecutiveCount, exists := streaks[p.Name]
		if !exists || consecutiveCount < maxConsecutive {
			eligible = append(eligible, p)
		}
	}

	// Fallback safety: if everything got filtered, reset and allow all
	if len(eligible) == 0 {
		fmt.Println("(All pillars hit max streak — resetting eligibility)")
		return pillars
	}

	return eligible
}

// --- Deficit-aware weight adjustment ---
// Compares each pillar's target share vs actual share in history,
// then boosts underrepresented pillars and dampens overrepresented ones.
// correction_factor controls how aggressively it corrects.
// At 0 this is a no-op (pure raw weights). At 3-5 it self-corrects monthly.

func adjustedWeights(pillars []Pillar, entries []HistoryEntry, correctionFactor float64) []Pillar {
	totalBlocks := len(entries)

	// Not enough history to correct — use raw weights
	if totalBlocks < 3 || correctionFactor == 0 {
		return pillars
	}

	// Count how many times each pillar was picked
	counts := make(map[string]int)
	for _, e := range entries {
		counts[e.Pillar]++
	}

	// Sum of all raw weights (for computing target share)
	totalWeight := 0
	for _, p := range pillars {
		totalWeight += p.Weight
	}

	adjusted := make([]Pillar, len(pillars))
	for i, p := range pillars {
		targetShare := float64(p.Weight) / float64(totalWeight)
		actualShare := float64(counts[p.Name]) / float64(totalBlocks)
		deficit := targetShare - actualShare

		// Multiply raw weight by (1 + deficit * correction_factor)
		// Deficit is positive when underrepresented → boost
		// Deficit is negative when overrepresented → dampen
		multiplier := 1.0 + deficit*correctionFactor
		if multiplier < 0.1 {
			multiplier = 0.1 // floor — never fully zero out a pillar
		}

		effectiveWeight := max(int(float64(p.Weight)*multiplier+0.5), 1)

		adjusted[i] = Pillar{Name: p.Name, Weight: effectiveWeight, Journal: p.Journal}
	}

	return adjusted
}

// --- Weighted random selection ---
// Classic approach: build a cumulative weight array,
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

	// Should never reach here, but just in case
	return pillars[len(pillars)-1]
}

// --- Next block index ---

func nextBlockIndex(entries []HistoryEntry) int {
	if len(entries) == 0 {
		return 1
	}
	return entries[len(entries)-1].Block + 1
}

// --- Main ---

func main() {
	force := flag.Bool("f", false, "Force run (bypass Sunday rest day guard)")

	flag.Usage = func() {
		fmt.Print(`
  commitment-randomizer — pick your next focus block

  Usage:
    randomizer        Generate the next focus block assignment
    randomizer -f     Force run (bypass Sunday rest day guard)
    randomizer -h     Show this help

  How it works:
    Each run picks one focus pillar using weighted randomness,
    records it to history.yaml, and prints the result.
    Run it once every 2 days (or per your config's block_days).

  Files:
    config.yaml       Pillars, weights, and rules (edit freely)
    history.yaml      Past assignments (auto-created, append-only)

`)
	}

	flag.Parse()

	// Sunday guard (skip with -f)
	if !*force && time.Now().Weekday() == time.Sunday {
		fmt.Print("\n  Sunday — rest day. No block assigned. Use -f to override.\n\n")
		return
	}

	// 1. Load config
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// 2. Load history
	history, err := loadHistory()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// 3. Check streaks and filter
	streaks := recentStreaks(history.Entries, config.MaxConsecutive)
	candidates := eligiblePillars(config.Pillars, streaks, config.MaxConsecutive)

	// 4. Adjust weights based on historical deficit
	candidates = adjustedWeights(candidates, history.Entries, config.CorrectionFactor)

	// 5. Pick one
	chosen := weightedPick(candidates)

	// 6. Record it
	entry := HistoryEntry{
		Pillar:    chosen.Name,
		Timestamp: time.Now().Format("Mon Jan 2 15:04 MST"),
		Block:     nextBlockIndex(history.Entries),
	}
	history.Entries = append(history.Entries, entry)

	err = saveHistory(history)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// 7. Tell the human what to do
	fmt.Println()
	fmt.Printf("  Next %d-day focus block: %s\n", config.BlockDays, chosen.Name)
	fmt.Println()
	fmt.Printf("  Block #%d — assigned %s\n", entry.Block, time.Now().Format("Mon Jan 2"))
	fmt.Println()

	// 8. Show journal if one exists for this pillar
	if chosen.Journal != "" {
		data, err := os.ReadFile(chosen.Journal)
		if err != nil {
			fmt.Printf("  (journal not found: %s)\n\n", chosen.Journal)
		} else {
			fmt.Println("  ---")
			fmt.Println()
			fmt.Printf("%s\n", data)
		}
	}
}

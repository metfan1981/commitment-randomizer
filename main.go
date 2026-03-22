package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

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
    records it to data/history.yaml, and prints the result.
    Run it once every 2 days (or per your config's block_days).

  Files:
    data/config.yaml    Pillars, weights, and rules (edit freely)
    data/history.yaml   Past assignments (auto-created, append-only)

`)
	}

	flag.Parse()

	if !*force && time.Now().Weekday() == time.Sunday {
		fmt.Print("\n  Sunday — rest day. No block assigned. Use -f to override.\n\n")
		return
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	history, err := loadHistory()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	streaks := recentStreaks(history.Entries, config.MaxConsecutive)
	candidates := eligiblePillars(config.Pillars, streaks, config.MaxConsecutive)
	candidates = adjustedWeights(candidates, history.Entries, config.CorrectionFactor)
	chosen := weightedPick(candidates)

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

	fmt.Println()
	fmt.Printf("  Next %d-day focus block: %s\n", config.BlockDays, chosen.Name)
	fmt.Println()
	fmt.Printf("  Block #%d — assigned %s\n", entry.Block, time.Now().Format("Mon Jan 2"))
	fmt.Println()

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

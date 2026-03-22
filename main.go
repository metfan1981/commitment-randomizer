package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	force := flag.Bool("f", false, "Force run (bypass block boundary check)")

	flag.Usage = func() {
		fmt.Print(`
  commitment-randomizer — pick your next focus block

  Usage:
    randomizer            Pick next block (skips if not due yet)
    randomizer -f         Force pick (bypass block boundary check)
    randomizer -h         Show this help

  How it works:
    Each run checks if the current block has ended (based on block_days),
    picks one focus pillar using weighted randomness, records it to
    data/history.yaml, and prints the result.

    If TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID are set in data/.env,
    the result is also sent as a Telegram message.

  Files:
    data/config.yaml    Pillars, weights, and rules (edit freely)
    data/history.yaml   Past assignments (auto-created, append-only)
    data/.env           Telegram credentials (optional)

`)
	}

	flag.Parse()
	loadEnv()

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

	if !*force && !isBlockDue(history.Entries, config.BlockDays) {
		last := history.Entries[len(history.Entries)-1]
		fmt.Printf("\nCurrent block: %s (assigned %s, %d-day block)\n\n",
			last.Pillar, last.Date, config.BlockDays)
		return
	}

	streaks := recentStreaks(history.Entries, config.MaxConsecutive)
	candidates := eligiblePillars(config.Pillars, streaks, config.MaxConsecutive)
	candidates = adjustedWeights(candidates, history.Entries, config.CorrectionFactor)
	chosen := weightedPick(candidates)

	now := time.Now()
	entry := HistoryEntry{
		Pillar:    chosen.Name,
		Timestamp: now.Format("Mon Jan 2 2006 15:04 MST"),
		Date:      now.Format("2006-01-02"),
		Block:     nextBlockIndex(history.Entries),
	}
	history.Entries = append(history.Entries, entry)

	err = saveHistory(history)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	var msg strings.Builder
	fmt.Fprintf(&msg, "Next %d-day focus block: %s\n", config.BlockDays, chosen.Name)
	fmt.Fprintf(&msg, "Block #%d — assigned %s", entry.Block, now.Format("Mon Jan 2"))

	if chosen.Journal != "" {
		data, err := os.ReadFile(chosen.Journal)
		if err != nil {
			fmt.Fprintf(&msg, "\n\n(journal not found: %s)", chosen.Journal)
		} else {
			fmt.Fprintf(&msg, "\n\n---\n\n%s", strings.TrimSpace(string(data)))
		}
	}

	fmt.Printf("\n%s\n\n", msg.String())

	if telegramConfigured() {
		if err := sendTelegram(msg.String()); err != nil {
			fmt.Printf("(telegram: %v)\n\n", err)
		}
	}
}

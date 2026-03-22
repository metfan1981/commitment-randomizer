package main

import (
	"testing"
	"time"
)

func TestCalendarDaysBetween_DST(t *testing.T) {
	ams, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		t.Fatal(err)
	}

	// 2026 DST spring-forward: March 29 at 02:00 (CET→CEST, 23-hour day)
	tests := []struct {
		name     string
		from     time.Time
		to       time.Time
		wantDays int
	}{
		{
			name:     "normal 2 days, no DST",
			from:     time.Date(2026, 3, 1, 0, 0, 0, 0, ams),
			to:       time.Date(2026, 3, 3, 0, 0, 0, 0, ams),
			wantDays: 2,
		},
		{
			name:     "across spring-forward (23h day)",
			from:     time.Date(2026, 3, 28, 0, 0, 0, 0, ams),
			to:       time.Date(2026, 3, 30, 0, 0, 0, 0, ams),
			wantDays: 2,
		},
		{
			name:     "spring-forward day itself",
			from:     time.Date(2026, 3, 28, 17, 0, 0, 0, ams),
			to:       time.Date(2026, 3, 29, 17, 0, 0, 0, ams),
			wantDays: 1,
		},
		{
			name:     "same day",
			from:     time.Date(2026, 6, 15, 8, 0, 0, 0, ams),
			to:       time.Date(2026, 6, 15, 22, 0, 0, 0, ams),
			wantDays: 0,
		},
		// 2026 DST fall-back: October 25 at 03:00 (CEST→CET, 25-hour day)
		{
			name:     "across fall-back (25h day)",
			from:     time.Date(2026, 10, 24, 0, 0, 0, 0, ams),
			to:       time.Date(2026, 10, 26, 0, 0, 0, 0, ams),
			wantDays: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calendarDaysBetween(tt.from, tt.to)
			if got != tt.wantDays {
				t.Errorf("calendarDaysBetween(%s, %s) = %d, want %d",
					tt.from.Format("2006-01-02 15:04 MST"),
					tt.to.Format("2006-01-02 15:04 MST"),
					got, tt.wantDays)
			}
		})
	}
}

func TestIsBlockDueAt_DST(t *testing.T) {
	ams, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		t.Fatal(err)
	}

	entries := []HistoryEntry{{
		Pillar: "Test",
		Date:   "2026-03-28",
		Block:  1,
	}}

	// March 30 in Amsterdam = 2 calendar days after March 28
	// With block_days=2, this should be due
	now := time.Date(2026, 3, 30, 17, 0, 0, 0, ams)

	if !isBlockDueAt(entries, 2, now) {
		t.Error("block should be due: 2 calendar days have passed across DST spring-forward")
	}

	// March 29 = 1 calendar day, should NOT be due for block_days=2
	now = time.Date(2026, 3, 29, 17, 0, 0, 0, ams)

	if isBlockDueAt(entries, 2, now) {
		t.Error("block should NOT be due: only 1 calendar day has passed")
	}
}

func TestIsBlockDueAt_EmptyHistory(t *testing.T) {
	if !isBlockDueAt(nil, 2, time.Now()) {
		t.Error("block should be due when history is empty")
	}
}

func TestIsBlockDueAt_MissingDate(t *testing.T) {
	entries := []HistoryEntry{{Pillar: "Test", Block: 1}}

	if !isBlockDueAt(entries, 2, time.Now()) {
		t.Error("block should be due when last entry has no date (legacy format)")
	}
}

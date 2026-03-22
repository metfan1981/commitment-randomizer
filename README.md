# Commitment Randomizer

Stop deciding. Start doing.

If you've ever burned half a morning wondering "what should I focus on today?" — or spent more energy planning your week than actually living it — this tool is for you.

**Commitment Randomizer** is a tiny CLI that picks your next focus area so you don't have to. One command, one answer, zero negotiation with yourself.

### Why it works

- **One focus per block.** No juggling five things. The app picks one, and that's your priority for the next N days (configurable).
- **No decisions required.** You run it, it tells you what to do. The "what should I work on?" loop is gone.
- **Weighted balance over time.** You control how often each area shows up long-term. The system handles the rest.
- **Anti-streak protection.** Won't give you the same thing too many times in a row.
- **Deficit correction.** If something hasn't shown up in a while, it gets a boost — no pillar gets forgotten for months.
- **Block boundary check.** Run it daily (cron, manually, whatever) — it only assigns a new block when the current one has ended.
- **Telegram notifications.** Optional — sends the assignment to your phone when configured.

Built for people with ADHD, decision fatigue, or anyone who knows that the hardest part isn't doing the work — it's picking which work to do.

This is not a calendar or planner. It is a **commitment generator**.

### Operating model

- Fixed commitments (e.g. work, gym) live in your calendar outside this system.
- The app assigns **one additional focus pillar** per block. Focus means priority, not exclusivity.
- The program can be called as often as you want — it silently exits if the current block hasn't ended yet.

## Step 1: Define your pillars

Open `data/config.yaml` and make it yours. This is where you decide what matters — your focus areas, how much weight each one gets, and the rules of the game.

```yaml
# EXAMPLE — replace with your own pillars and weights
block_days: 2
max_consecutive: 2
correction_factor: 4

pillars:
  - name: Drawing
    weight: 40
  - name: Language
    weight: 25
  - name: Home Tasks
    weight: 15
    journal: journal/home-tasks
  - name: Free time
    weight: 20
```

Ask yourself:
- What are the areas of my life I want to grow, but keep postponing?
- Which ones deserve more time than others? Give those higher weights.
- What's already a fixed routine (work, gym, etc.)? Leave those out — this tool handles the *rest*.

### Fields

| Field | What it controls |
|---|---|
| `block_days` | How many days each focus block lasts |
| `max_consecutive` | Max times the same pillar can be picked in a row before it's temporarily excluded |
| `correction_factor` | How aggressively the system corrects distribution drift (0 = pure random, 3-5 = tight balance) |
| `pillars` | List of focus areas with relative weights |

### Weights

Weights don't need to sum to 100 — they're relative. A pillar with weight 40 is roughly 2x more likely than one with weight 20.

To add a new pillar, just append it:

```yaml
  - name: Reading
    weight: 15
```

To remove one, delete or comment out the entry.

### Journal

Some pillars need context. "Home Tasks" — okay, but *which* tasks? Add a `journal` field pointing to a text file, and the app prints its contents when that pillar is selected.

```yaml
  - name: Home Tasks
    weight: 15
    journal: journal/home-tasks
```

The journal file is just plain text — a task list, notes, links, whatever helps you hit the ground running:

```
- fix leaking kitchen faucet
- organize garage shelves
- schedule annual furnace inspection
```

No journal field? No problem — the pillar just prints without extra context.

### Deficit correction

The system tracks each pillar's actual share vs its target share over all history. Underrepresented pillars get a weight boost, overrepresented ones get dampened. This prevents "anomaly months" where a pillar vanishes from the schedule. Set `correction_factor: 0` to disable and use pure weighted randomness.

## Step 2: Build and run

### Requirements

- [Go 1.24+](https://go.dev/dl/) (for building from source)
- `make` (optional, for convenience targets)

```bash
# one-time build
make build

# run (picks a block if one is due, otherwise silently exits)
./randomizer

# force a pick regardless of block boundary
./randomizer -f

# simulate a full year to check weight distribution
make sim
```

Output:

```
Next 2-day focus block: Drawing
Block #1 — assigned Mon Feb 16
```

### Flags

| Flag | What it does |
|---|---|
| `-h` | Show usage help |
| `-f` | Force pick — bypass the block boundary check |

## Telegram notifications (optional)

The app can send each assignment to your phone via Telegram. To set it up:

1. Open Telegram and message **@BotFather**
2. Send `/newbot`, pick a name and username
3. Copy the bot token BotFather gives you
4. Send any message to your new bot, then visit `https://api.telegram.org/bot<TOKEN>/getUpdates` to find your `chat_id`
5. Copy `data/.env.example` to `data/.env` and fill in both values:

```
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
TELEGRAM_CHAT_ID=123456789
```

When configured, every block assignment is sent to your Telegram chat automatically. If `data/.env` is missing or empty, the app works normally without notifications.

### Automation (cron)

Set up a daily cron job — the program handles the rest. It only assigns a new block when the previous one has ended (based on `block_days`).

```bash
crontab -e
```

```
0 17 * * * /path/to/commitment-randomizer/run.sh
```

## History

Past assignments are stored in `data/history.yaml` (auto-created on first run). The file is append-only and looks like this:

```yaml
entries:
    - pillar: Drawing
      timestamp: Mon Feb 16 2026 08:30 CET
      date: "2026-02-16"
      block: 1
    - pillar: Language
      timestamp: Wed Feb 18 2026 09:12 CET
      date: "2026-02-18"
      block: 2
```

History is used for:
- Block boundary check (is it time for a new block?)
- Streak detection (anti-repetition)
- Long-term distribution correction

To start fresh, delete `data/history.yaml`.

## How selection works

1. Check if the current block has ended (compare last entry's date against `block_days`)
2. If not due, exit silently — unless `-f` is used
3. Load config and history
4. Check if the most recent pillar has been picked `max_consecutive` times in a row
5. If yes, exclude it from this round's candidates
6. Adjust weights based on historical deficit (boost underrepresented, dampen overrepresented)
7. Perform weighted random selection among adjusted candidates
8. Append result to history
9. Print the assignment (and journal, if configured)
10. Send to Telegram if configured

If all pillars somehow get excluded (shouldn't happen with normal config), the system resets and allows everything — no crash, no stuck state.

## Running from anywhere

A `run.sh` script is included. It `cd`s into the project directory before running the binary, so it works from any location:

```bash
# add to ~/.zshrc or ~/.bashrc
alias focus="$HOME/path/to/commitment-randomizer/run.sh"
```

Then: `focus`, `focus -f`, `focus -h`.

## Simulation

Run `make sim` to simulate a full year of block assignments using your current config. It shows distribution vs target at 90 days and 365 days, plus max streak stats — useful for tuning weights and `correction_factor` before committing to a setup.

## Project structure

```
.
├── main.go            # CLI entry point — flags, output, glue
├── config.go          # Pillar/Config types and loader
├── history.go         # History types, load/save, block boundary check
├── engine.go          # Core algorithm: streaks, eligibility, weights, pick
├── telegram.go        # Env loader + Telegram send
├── sim_test.go        # Yearly simulation for config tuning
├── data/
│   ├── config.yaml    # Editable settings (pillars, weights, rules)
│   ├── history.yaml   # Auto-generated assignment log (gitignored)
│   ├── .env           # Telegram credentials (gitignored)
│   └── .env.example   # Template for .env
├── journal/           # Optional per-pillar context files
├── run.sh             # Run from anywhere wrapper
├── Makefile           # build / run / sim targets
├── go.mod
└── go.sum
```

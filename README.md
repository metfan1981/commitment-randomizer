# Commitment Randomizer

Stop deciding. Start doing.

If you've ever burned half a morning wondering "what should I focus on today?" — or spent more energy planning your week than actually living it — this tool is for you.

**Commitment Randomizer** is a tiny CLI that picks your next focus area so you don't have to. One command, one answer, zero negotiation with yourself.

### Why it works

- **One focus per block.** No juggling five things. The app picks one, and that's your priority for the next 2 days (configurable).
- **No decisions required.** You run it, it tells you what to do. The "what should I work on?" loop is gone.
- **Weighted balance over time.** You control how often each area shows up long-term. The system handles the rest.
- **Anti-streak protection.** Won't give you the same thing five times in a row.
- **Deficit correction.** If something hasn't shown up in a while, it gets a gentle boost — no pillar gets forgotten for months.

Built for people with ADHD, decision fatigue, or anyone who knows that the hardest part isn't doing the work — it's picking which work to do.

This is not a calendar or planner. It is a **commitment generator**.

### Operating model

- **Sunday is excluded** — rest day. The app won't roll on Sundays (use `-f` to override).
- **Monday–Saturday** are the active days.
- Fixed commitments (e.g. work, gym) live in your calendar outside this system.
- The app assigns **one additional focus pillar** per block. Focus means priority, not exclusivity.

## Step 1: Define your pillars

Before running anything, open `config.yaml` and make it yours. This is where you decide what matters — your focus areas, how much weight each one gets, and the rules of the game.

```yaml
# EXAMPLE — replace with your own pillars and weights
block_days: 2
max_consecutive: 3
correction_factor: 3

pillars:
  - name: Drawing
    weight: 40
  - name: Language
    weight: 25
  - name: Exercise Research
    weight: 15
  - name: Home Tasks
    weight: 20
    journal: journal/home-tasks
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
| `correction_factor` | How aggressively the system corrects distribution drift (0 = pure random, 3-5 = tight monthly balance) |
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
    weight: 20
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

## Requirements

- [Go 1.21+](https://go.dev/dl/) (for building from source)
- `make` (optional, for convenience targets)

## Step 2: Build and run

```bash
# one-time build
make build

# run
./randomizer
# usage help
./randomizer -h
```

Output:

```
  Next 2-day focus block: Drawing

  Block #1 — assigned Mon Feb 16
```

That's it. Run it once every 2 days (or whenever your current block ends).

### Flags

| Flag | What it does |
|---|---|
| `-h` | Show usage help |
| `-f` | Force run — bypass the Sunday rest day guard |

## History

Past assignments are stored in `history.yaml` (auto-created on first run). The file is append-only and looks like this:

```yaml
entries:
    - pillar: Drawing
      timestamp: Mon Feb 16 08:30 CET
      block: 1
    - pillar: Language
      timestamp: Wed Feb 18 09:12 CET
      block: 2
```

History is used for:
- Streak detection (anti-repetition)
- Long-term distribution integrity

To start fresh, delete `history.yaml`.

## How selection works

1. Load config and history
2. Check if the most recent pillar has been picked `max_consecutive` times in a row
3. If yes, exclude it from this round's candidates
4. Adjust weights based on historical deficit (boost underrepresented, dampen overrepresented)
5. Perform weighted random selection among adjusted candidates
6. Append result to history
7. Print the assignment (and journal, if configured)

If all pillars somehow get excluded (shouldn't happen with normal config), the system resets and allows everything — no crash, no stuck state.

## Running from anywhere

A `run.sh` script is included. It `cd`s into the project directory before running the binary, so it works from any location. Add an alias to your shell config:

**macOS / Linux (zsh)** — add to `~/.zshrc`:

```bash
alias focus="$HOME/path/to/commitment-randomizer/run.sh"
```

**Linux (bash)** — add to `~/.bashrc`:

```bash
alias focus="$HOME/path/to/commitment-randomizer/run.sh"
```

**Windows (Git Bash / WSL)** — add to `~/.bashrc`:

```bash
alias focus="/c/Users/yourname/path/to/commitment-randomizer/run.sh"
```

Replace the path with wherever you cloned the repo. Then reload your shell (`source ~/.zshrc` etc.) and use from anywhere: `focus`, `focus -f`, `focus -h`.

## Project structure

```
.
├── main.go          # All application logic
├── config.yaml      # Editable settings (pillars, weights, rules)
├── history.yaml     # Auto-generated assignment log (gitignored)
├── journal/         # Optional per-pillar context files
├── run.sh           # Run from anywhere wrapper
├── Makefile         # build / run targets
├── go.mod
└── go.sum
```

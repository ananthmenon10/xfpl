---
name: pp-xfpl
description: "Printing Press CLI for Livefpl. Combined CLI for multiple API services"
author: "Ananth Menon"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - xfpl
---

# Livefpl — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `xfpl` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer:
   ```bash
   npx -y @mvanhorn/printing-press install xfpl --cli-only
   ```
2. Verify: `xfpl --version`
3. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

If the `npx` install fails before this CLI has a public-library category, install Node or use the category-specific Go fallback after publish.

If `--version` reports "command not found" after install, the install step did not put the binary on `$PATH`. Do not proceed with skill commands until verification succeeds.

Combined CLI for multiple API services

## When Not to Use This CLI

Do not activate this CLI for requests that require creating, updating, deleting, publishing, commenting, upvoting, inviting, ordering, sending messages, booking, purchasing, or changing remote state. This printed CLI exposes read-only commands for inspection, export, sync, and analysis.

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

## Command Reference

**bootstrap-static** — Manage bootstrap static

- `xfpl bootstrap-static` — ~1.5MB. The single most important read in the FPL API. Loaded once and cached locally for all entity lookups.

**dream-team** — Manage dream team

- `xfpl dream-team <eventId>` — Highest-scoring XI for a gameweek

**element-summary** — Manage element summary

- `xfpl element-summary <elementId>` — Per-player detail (history + upcoming fixtures)

**elite-json** — Manage elite json

- `xfpl elite-json` — Aggregated picks, captaincy %, ownership, and chip usage among the top 10,000 ranked managers. ~10KB, refreshed each...

**entry** — Manage entry

- `xfpl entry <managerId>` — Manager profile (name, country, leagues, summary score)

**event** — Manage event


**event-status** — Manage event status

- `xfpl event-status` — Bonus and league update status across active GW

**fixtures** — Manage fixtures

- `xfpl fixtures` — All season fixtures (optionally filtered by event)

**games-json** — Manage games json

- `xfpl games-json` — Array of arrays — one per Premier League fixture in the active gameweek. Each entry: [home, away, hscore, ascore,...

**leagues-classic** — Manage leagues classic


**leagues-h2h-matches** — Manage leagues h2h matches

- `xfpl leagues-h2h-matches <leagueId>` — H2H league matches

**lh-api2** — Manage lh api2

- `xfpl lh-api2` — Returns FT, bank, chips_available, base_gw, eleven, picks, benches, bought_values per element_id, manager_name,...

**livefplapi** — Manage livefplapi

- `xfpl livefplapi <teamId>` — THE primary LiveFPL transcendence endpoint. Returns: - GWrank: current gameweek rank - GWrank2: projected/secondary...

**locals-league-id-json** — Manage locals league id json

- `xfpl locals-league-id-json <leagueId>` — Available for indexed (popular) leagues; small leagues 404 here and must use the dynamic livefplapi path on...

**me** — Manage me

- `xfpl me` — Authenticated user's own profile

**my-team** — Manage my team

- `xfpl my-team <managerId>` — Requires a session cookie from a logged-in fantasy.premierleague.com session. Opt-in only; we surface this behind a...

**prices-json** — Manage prices json

- `xfpl prices-json` — Dict keyed by FPL element_id. Each value: {name, team, type (GK/DEF/MID/FWD), type_code, team_code, cost, progress,...

**team** — Manage team

- `xfpl team` — Set-piece taker notes per club

**top-transfers-json** — Manage top transfers json

- `xfpl top-transfers-json` — Array of [out_player_id, in_player_id, count_million, pct_share]. Sorted descending by volume. Updated every few...

**version-json** — Manage version json

- `xfpl version-json` — Returns {gen, version, meter, store, league}. `gen` increments every time LiveFPL's backend re-aggregates; clients...


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
xfpl which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Auth Setup
Run `xfpl auth setup` to print the URL and steps for getting a key (add `--launch` to open the URL). Then set:

```bash
export FANTASY_PREMIER_LEAGUE_SESSION_COOKIE="<your-key>"
```

Or persist it in `~/.config/xfpl/config.toml`.

Run `xfpl doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  xfpl bootstrap-static --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Read-only** — do not use this CLI for create, update, delete, publish, comment, upvote, invite, order, send, or other mutating requests

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal AND no machine-format flag (`--json`, `--csv`, `--compact`, `--quiet`, `--plain`, `--select`) is set — piped/agent consumers and explicit-format runs get pure JSON on stdout.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
xfpl feedback "the --since flag is inclusive but docs say exclusive"
xfpl feedback --stdin < notes.txt
xfpl feedback list --json --limit 10
```

Entries are stored locally at `~/.xfpl/feedback.jsonl`. They are never POSTed unless `LIVEFPL_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `LIVEFPL_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
xfpl profile save briefing --json
xfpl --profile briefing bootstrap-static
xfpl profile list --json
xfpl profile show briefing
xfpl profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `xfpl --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

Install the MCP binary from this CLI's published public-library entry or pre-built release, then register it:

```bash
claude mcp add xfpl-mcp -- xfpl-mcp
```

Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which xfpl`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   xfpl <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `xfpl <command> --help`.

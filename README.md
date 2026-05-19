# xfpl — the only Fantasy Premier League CLI that knows what's actually about to happen

[![CI](https://github.com/ananthmenon10/xfpl/actions/workflows/ci.yml/badge.svg)](https://github.com/ananthmenon10/xfpl/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/ananthmenon10/xfpl.svg)](https://pkg.go.dev/github.com/ananthmenon10/xfpl)

A read-only CLI that combines the official Fantasy Premier League API with LiveFPL.net's live rank, EO, threats, and price-change data. Agent-friendly (`--json --agent`), offline-capable (local SQLite cache), and shipped as an MCP server for Claude Desktop / Cursor / Claude Code.

Printed by [@ananthmenon10](https://github.com/ananthmenon10) (Ananth Menon) on the [Printing Press](https://github.com/mvanhorn/cli-printing-press).

## Highlights

- **`xfpl captain --top 5 --explain`** — captain ranker scored by `ep_next × FDR × home/away × xGI/90 × minutes-risk × DGW multiplier`. No other FPL CLI exposes a captain ranker.
- **`xfpl explain rank <teamId>`** — top 5 player contributions to the current gameweek rank for any manager.
- **`xfpl chip-plan <teamId>`** — cross-references your remaining chips against detected Blank GWs / Double GWs in the remaining fixture list.
- **`xfpl player <name>`** / **`xfpl compare <name> <name> ...`** — name-based player lookup and side-by-side comparison (no other CLI accepts names — they all require numeric element IDs).
- **`xfpl captains-history <teamId>`** / **`xfpl points <teamId>`** / **`xfpl cup <teamId>`** — past captains, live GW score breakdown, and FPL Cup status.

## Install

### Recommended: `go install`

```bash
go install github.com/ananthmenon10/xfpl/cmd/xfpl@latest
go install github.com/ananthmenon10/xfpl/cmd/xfpl-mcp@latest
```

Requires Go 1.26+. Ensure `$GOPATH/bin` (or `~/go/bin`) is on your `PATH`. Verify with `xfpl --version`.

### Homebrew (after first release)

```bash
brew install ananthmenon10/xfpl/xfpl
```

### Printing Press (CLI + agent skill in one shot)

The recommended path installs both the `xfpl` binary and the `pp-xfpl` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press install xfpl
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press install xfpl --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press install xfpl --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press install xfpl --agent claude-code
npx -y @mvanhorn/printing-press install xfpl --agent claude-code --agent codex
```

### Without Node

The generated install path is category-agnostic until this CLI is published. If `npx` is not available before publish, install Node or use the category-specific Go fallback from the public-library entry after publish.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/xfpl-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-xfpl --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-xfpl --force
```

## Install for OpenClaw

Tell your OpenClaw agent (copy this):

```
Install the pp-xfpl skill from https://github.com/mvanhorn/printing-press-library/tree/main/cli-skills/pp-xfpl. The skill defines how its required CLI can be installed.
```

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/xfpl-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.
3. Fill in `FANTASY_PREMIER_LEAGUE_SESSION_COOKIE` when Claude Desktop prompts you.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


Install the MCP binary from this CLI's published public-library entry or pre-built release.

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "xfpl": {
      "command": "xfpl-mcp",
      "env": {
        "FANTASY_PREMIER_LEAGUE_SESSION_COOKIE": "<your-key>"
      }
    }
  }
}
```

</details>

## Quick Start

### 1. Install

See [Install](#install) above.

### 2. Set Up Credentials

Get your API key from your API provider's developer portal. The key typically looks like a long alphanumeric string.

```bash
export FANTASY_PREMIER_LEAGUE_SESSION_COOKIE="<paste-your-key>"
```

You can also persist this in your config file at `~/.config/xfpl/config.toml`.

### 3. Verify Setup

```bash
xfpl doctor
```

This checks your configuration and credentials.

### 4. Try Your First Command

```bash
xfpl bootstrap-static
```

## Usage

Run `xfpl --help` for the full command reference and flag list.

## Commands

### bootstrap-static

Manage bootstrap static

- **`xfpl bootstrap-static`** - ~1.5MB. The single most important read in the FPL API. Loaded once
and cached locally for all entity lookups.

### dream-team

Manage dream team

- **`xfpl dream-team <eventId>`** - Highest-scoring XI for a gameweek

### element-summary

Manage element summary

- **`xfpl element-summary <elementId>`** - Per-player detail (history + upcoming fixtures)

### elite-json

Manage elite json

- **`xfpl elite-json`** - Aggregated picks, captaincy %, ownership, and chip usage among the
top 10,000 ranked managers. ~10KB, refreshed each gameweek.

### entry

Manage entry

- **`xfpl entry <managerId>`** - Manager profile (name, country, leagues, summary score)

### event

Manage event


### event-status

Manage event status

- **`xfpl event-status`** - Bonus and league update status across active GW

### fixtures

Manage fixtures

- **`xfpl fixtures`** - All season fixtures (optionally filtered by event)

### games-json

Manage games json

- **`xfpl games-json`** - Array of arrays — one per Premier League fixture in the active gameweek.
Each entry: [home, away, hscore, ascore, status, source, h_scorers,
h_assists, a_scorers, a_assists, bps_top, ..., per-player live entries
with BPS, minutes, clean sheets, bonus]. Update frequency: ~1-2 min
during live matches.

### leagues-classic

Manage leagues classic


### leagues-h2h-matches

Manage leagues h2h matches

- **`xfpl leagues-h2h-matches <leagueId>`** - H2H league matches

### lh-api2

Manage lh api2

- **`xfpl lh-api2`** - Returns FT, bank, chips_available, base_gw, eleven, picks, benches,
bought_values per element_id, manager_name, original picks (start of
GW), epicks (effective picks after auto-subs). Used by LiveFPL's
transfer planner UI.

### livefplapi

Manage livefplapi

- **`xfpl livefplapi <teamId>`** - THE primary LiveFPL transcendence endpoint. Returns:
- GWrank: current gameweek rank
- GWrank2: projected/secondary rank
- a_e / a_o / a_t: average scores (entry/overall/top tiers)
- arrow_*: visual rank-direction indicators
- avg_similarity: how template the team is
- bench, benchzz: bench points + bench element_ids
- buds: similar-manager comparisons
- cache: cache key
- calcplayerptsd: live calculated points per element_id
- and ~40 more fields
Single endpoint per team; ~80KB. Public, no auth.

### locals-league-id-json

Manage locals league id json

- **`xfpl locals-league-id-json <leagueId>`** - Available for indexed (popular) leagues; small leagues 404 here and
must use the dynamic livefplapi path on www.livefpl.net.

### me

Manage me

- **`xfpl me`** - Authenticated user's own profile

### my-team

Manage my team

- **`xfpl my-team <managerId>`** - Requires a session cookie from a logged-in fantasy.premierleague.com
session. Opt-in only; we surface this behind a `--with-auth` flag and
an env var (PP_FPL_SESSION_COOKIE) so accidental agent loops cannot
trigger it.

### prices-json

Manage prices json

- **`xfpl prices-json`** - Dict keyed by FPL element_id. Each value: {name, team, type (GK/DEF/MID/FWD),
type_code, team_code, cost, progress, progress_tonight, per_hour}.
Used by LiveFPL's price-change predictor.

### team

Manage team

- **`xfpl team`** - Set-piece taker notes per club

### top-transfers-json

Manage top transfers json

- **`xfpl top-transfers-json`** - Array of [out_player_id, in_player_id, count_million, pct_share].
Sorted descending by volume. Updated every few minutes during
transfer windows.

### version-json

Manage version json

- **`xfpl version-json`** - Returns {gen, version, meter, store, league}. `gen` increments every
time LiveFPL's backend re-aggregates; clients use it as a cache key.


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
xfpl bootstrap-static

# JSON for scripting and agents
xfpl bootstrap-static --json

# Filter to specific fields
xfpl bootstrap-static --json --select id,name,status

# Dry run — show the request without sending
xfpl bootstrap-static --dry-run

# Agent mode — JSON + compact + no prompts in one flag
xfpl bootstrap-static --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Read-only by default** - this CLI does not create, update, delete, publish, send, or mutate remote resources
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
xfpl doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Config file: `~/.config/xfpl/config.toml`

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `FANTASY_PREMIER_LEAGUE_SESSION_COOKIE` | harvested | Yes | Populated automatically by auth login. |

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `xfpl doctor` to check credentials
- Verify the environment variable is set: `echo $FANTASY_PREMIER_LEAGUE_SESSION_COOKIE`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)

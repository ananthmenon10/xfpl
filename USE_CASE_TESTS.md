# xfpl — End-to-End Use Case Test Report

**Date:** 2026-05-19 (FPL 2025-26 season, last GW upcoming is GW38 on 2026-05-24)
**Build:** `xfpl 1.0.0` (commit `892308e`)
**Test approach:** real API calls, three team IDs covering the user's own team and two top-ranked managers globally.

## Team IDs under test

| ID | Manager | Notes |
|----|---------|-------|
| 5505524 | Ananth Menon | Test user (the repo author) |
| 3027768 | Erik Ibsen | Overall #1 in the global "Overall" league (2492 pts) |
| 2569513 | steve connor | Overall #2 (2473 pts) |

## Use Case 1 — "Should I captain Haaland or Watkins this week?"

**Commands run**

```
xfpl compare Haaland Watkins --agent
xfpl captain --top 15 --agent
```

**Findings**

| Player | Cost | Form | EP next | Ownership | Status | Captain rank |
|--------|------|------|---------|-----------|--------|--------------|
| Haaland | £14.7m | 5.2 | 5.7 | 64.5% | a | #2 (score 10.25, MCI home vs AVL FDR 3) |
| Watkins | £8.7m | 8.6 | 8.1 | 13.4% | a | not in top 15 (AVL away at MCI = FDR 5) |

**Verdict:** Captain ranker picks Haaland despite Watkins's higher `ep_next` because home/away + FDR dominate. Form alone (Watkins 8.6 vs Haaland 5.2) would mislead a naive picker. **The v0.2 scorer makes the right call.**

## Use Case 2 — "What's my chip strategy?"

**Commands run**

```
xfpl chip-plan 5505524 --agent
xfpl chip-plan 3027768 --agent
```

**Findings**

Both managers (test user + overall #1) have **0 chips remaining and 0 upcoming BGW/DGW**. The recommendation engine correctly returns `"no obvious chip windows in remaining gameweeks"`. This is correct for end-of-season — the chip-plan command is most valuable mid-season around DGWs in GW25/GW34/etc., which is when its detection logic shines.

## Use Case 3 — "Which players carried/disappointed last GW?"

**Commands run**

```
xfpl explain rank 5505524 --agent
xfpl captains-history 5505524 --limit 3 --agent
```

**Findings — Ananth GW37 top contributors (after explain-rank bug fix)**

```
gw=37 pts=50 cap=20
  Saka          (ARS) 20pts (CAPTAIN, ×2 mult)
  B.Fernandes   (MUN) 9pts
  Saliba        (ARS) 8pts
  Gabriel       (ARS) 6pts
  Groß          (BHA) 3pts
```

**Captain history (last 3 GWs)**

| GW | Captain | Mult | Captain pts | GW pts | GW rank |
|----|---------|------|-------------|--------|---------|
| 35 | Haaland | ×2 | 14 | 49 | 7.0M |
| 36 | Haaland | ×3 (TC) | 33 | 87 | 1.7M |
| 37 | Saka    | ×2 | 20 | 50 | 2.2M |

**Bug found and fixed mid-test:** the original `explain rank` cross-referenced LiveFPL's `calcplayerptsd` (global live points dict) but did not filter by the team's own picks, so it surfaced the highest-scoring players in the gameweek globally rather than this manager's actual contributors. Fixed in commit `7d20333` by additionally fetching `/entry/{id}/event/{gw}/picks/` and applying each pick's multiplier (0 for bench, 1 for starter, 2 for captain, 3 for triple captain) before sorting. Verified working across all three test IDs — each returns its own different captain.

## Use Case 4 — "Help me decide my next transfer"

**Commands run**

```
xfpl compare Salah Saka Palmer Watkins --agent
xfpl top-transfers-json --agent
```

**Compare output (4 attackers)**

| Player | Cost | Form | EP next | Ownership |
|--------|------|------|---------|-----------|
| Salah | £14.0m | 2.4 | **2.9** | 13.7% |
| Saka | £10.0m | 4.8 | 5.8 | 13.0% |
| Palmer | £10.3m | 0.5 | **0.5** | 12.5% |
| Watkins | £8.7m | 8.6 | **8.1** | 13.4% |

**Top 7 transfers globally (out → in)**

| Out | In | Volume | Share |
|-----|----|--------|-------|
| Thiago | Calvert-Lewin | 412M | 0.90% |
| Thiago | Osula | 389M | 0.85% |
| Wilson | Gibbs-White | 360M | 0.79% |
| Wilson | Anderson | 337M | 0.74% |
| Thiago | Bowen | 331M | 0.73% |
| Gyökeres | Calvert-Lewin | 302M | 0.66% |
| Gyökeres | Osula | 285M | 0.63% |

**Verdict:** Watkins is the math-optimal transfer-in for GW38 by ep_next, but the FPL crowd is pivoting hard to Forest mid + striker rotation. The CLI's `compare` + `top-transfers-json` together give exactly the data needed to triangulate.

## Use Case 5 — Fresh install + top-10k meta analysis (sub-agent)

**Dispatched to a fresh sub-agent** simulating a brand-new user with no FPL session cookie.

**Sub-agent results (verbatim summary):**

1. **`go install` succeeded** in under 5 seconds. One friction point: binary lands in `~/go/bin/` which isn't on the default `$PATH`. A README note would help. *(Already addressed in README "Install" section.)*
2. **All commands ran without auth.** No command required a session cookie. Confirms the `auth_optional: true` setting in `.printing-press.json` is accurate.
3. **Meta findings (GW38 prep):**
   - **Most captained in top 10k:** B.Fernandes (EO 1.125), Haaland (EO 1.043) — the only two players above 1.0 EO (signal of heavy captaincy).
   - **Highest EO:** Gabriel (1.000), Saka (0.982), O'Reilly (0.904) — the "skeleton" Arsenal-heavy template.
   - **Transfers-in trio:** Calvert-Lewin, Osula, Gibbs-White — the Forest rotation move dominates.
   - **CLI ↔ crowd divergence:** the CLI's captain ranker tops with Gibbs-White (math-optimal home vs BOU); top10k managers are massing on B.Fernandes/Haaland instead. The CLI is recommending the differential play.
4. **UX rough edges identified:**
   1. Spurious `warning: N items skipped (no extractable ID field found)` on stderr for endpoints that return maps/arrays. Pollutes `2>&1 | jq` pipelines.
   2. `elite-json` has no schema labels — values above 1.0 are unmarked (caller must know it's EO).
   3. `--human-friendly` flag is silently a no-op for the `captain` command.

## Overall coverage

| Command surface | Tested |
|---|---|
| Endpoint mirrors (bootstrap-static, fixtures, prices-json, top-transfers-json, etc.) | ✔ |
| Novel/transcendence commands (captain, explain rank, chip-plan, player, compare) | ✔ |
| Hermes-inspired extras (captains-history, points, cup) | ✔ |
| `--agent` JSON mode | ✔ (used on every command) |
| Anonymous (no session cookie) operation | ✔ (sub-agent confirmed) |
| Multi-team-ID isolation (per-team results differ correctly) | ✔ (3 IDs, distinct outputs) |
| Fresh-install path via `go install` | ✔ (sub-agent) |

## Issues fixed during testing

1. `explain rank` returned global GW top scorers, not this team's contributors → **fixed** (commit `7d20333`).
2. Player names missing from rank/contributor output → **fixed** (enriched with bootstrap-static).
3. `cup` errored on managers not in the FPL Cup → **fixed** to return clean `{"status": "not_in_cup"}` (commit `9315ee7`).

## Issues identified for v0.1.x follow-ups

1. Spurious `warning: items skipped` on stderr for non-list endpoint responses (low-severity stderr noise).
2. `elite-json` JSON schema labels for ownership/EO/captaincy% values.
3. `--human-friendly` table renderer for `captain`.
4. `xfpl which "captain ranker"` returns 0 matches — the capability index in `which.go` does not include captain/explain/chip-plan terms.

## Post-rename verification (2026-05-19, after `livefpl` → `xfpl`)

After the full rename (module path, cmd dirs, binary names, docs, GitHub repo), every use case was re-run against the new `xfpl` binary built from `github.com/ananthmenon10/xfpl`. Results below confirm parity with the pre-rename run on every probe.

| Probe | Pre-rename | Post-rename | Match |
|---|---|---|---|
| `captain --top 5` GW38 #1 pick | Gibbs-White (14.09) | Gibbs-White (14.09) | ✔ |
| `compare Haaland Watkins` ep_next | 5.7 / 8.1 | 5.7 / 8.1 | ✔ |
| `chip-plan 5505524` recommendation | "no obvious chip windows…" | "no obvious chip windows…" | ✔ |
| `explain rank 5505524` GW37 captain | Saka (×2, 20pts) | Saka (×2, 20pts) | ✔ |
| `explain rank 3027768` GW37 top | Gabriel | Gabriel | ✔ |
| `explain rank 2569513` GW37 top | B.Fernandes | B.Fernandes | ✔ |
| `top-transfers-json` top pair | Thiago→Calvert-Lewin | Thiago→Calvert-Lewin | ✔ |

**Fresh install path verified:**

```
$ go install github.com/ananthmenon10/xfpl/cmd/xfpl@latest
$ xfpl --version
xfpl 1.0.0
```

Clean-room run on an empty `$GOBIN` produced a 65MB binary in ~7 seconds. No environment errors, no auth prompts, no manual steps.

**CI verified:** GitHub Actions on the renamed repo passed in 2m44s on commit `9f8006b` — gofmt + vet + tests + cross-platform build all green.

**Repo URL:** `https://github.com/ananthmenon10/xfpl` (the old `…/livefpl` URL auto-redirects via GitHub).

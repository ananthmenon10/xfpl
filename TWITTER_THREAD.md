# xfpl — Twitter / X thread draft

Five-tweet thread. Each is under 280 chars. Copy each block one at a time into a chained X post.

---

## Tweet 1 — hook

shipped xfpl: a CLI that gives any AI agent live FPL data, league standings, and gameweek breakdowns.

no auth, no login, one install. ask Claude Code / Cursor / Codex "how did my team do last gameweek and which players hurt me" and it just answers.

---

## Tweet 2 — what it does

what it does:
- live points + ranks while matches are playing (faster than the official site)
- which players carried your league rank vs which cost you points
- captain ranker (FDR + xGI + home/away)
- chip planner for blank/double GWs
- compare any 2-4 players head to head

---

## Tweet 3 — install

install:
go install github.com/ananthmenon10/xfpl/cmd/xfpl@latest

every command supports --agent for clean JSON. paste the repo link into your agent of choice and it figures the rest out. tested clean-room on a fresh install in under 5 seconds.

---

## Tweet 4 — credits

credit where it's due:
- @LiveFPLnet (Ragabolly) for the live-rank backend this leans on. livefpl.net is the gold standard
- @mvanhorn for the Printing Press, the agent-native CLI factory I generated this with. printingpress.dev

both worth a follow.

---

## Tweet 5 — link

https://github.com/ananthmenon10/xfpl

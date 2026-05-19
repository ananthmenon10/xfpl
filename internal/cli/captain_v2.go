// Copyright 2026 ananth-menon. Licensed under Apache-2.0. See LICENSE.
//
// Captain ranker v0.2.
//
// Formula:
//
//	score = ep_next
//	      * (2.2 - 0.3 * fdr)   // 1.9x (FDR 1) → 0.7x (FDR 5)
//	      * homeBoost            // 1.10 home, 0.95 away
//	      * (1 + 0.3 * xgi_90)   // attackers with xGI/90 > 0 get a lift
//	      * minRisk              // 0.5 → 1.0 based on minutes played
//	      * dgwMultiplier        // 1 + count(extra fixtures) * 0.85
//
// Inputs come from bootstrap-static (elements, teams, events with is_next)
// and the fixtures endpoint. Element-summary is intentionally NOT fetched —
// the per-90 xGI columns already in bootstrap-static give us recent attacking
// volume without 30 extra API calls.
//
// v0.1 (deprecated): form*0.4 + ep_next*0.5 + selected/100*0.1

package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
)

type captainElement struct {
	ID            int     `json:"id"`
	WebName       string  `json:"web_name"`
	Team          int     `json:"team"`
	ElementType   int     `json:"element_type"`
	NowCost       int     `json:"now_cost"`
	Form          string  `json:"form"`
	EPNext        string  `json:"ep_next"`
	SelectedByPct string  `json:"selected_by_percent"`
	Status        string  `json:"status"`
	Minutes       int     `json:"minutes"`
	TotalPoints   int     `json:"total_points"`
	ExpectedGI90  float64 `json:"expected_goal_involvements_per_90"`
	News          string  `json:"news"`
	StartsPer90   float64 `json:"starts_per_90"`
}

type captainFixture struct {
	Event    int  `json:"event"`
	TeamH    int  `json:"team_h"`
	TeamA    int  `json:"team_a"`
	HDiff    int  `json:"team_h_difficulty"`
	ADiff    int  `json:"team_a_difficulty"`
	Finished bool `json:"finished"`
}

type captainEvent struct {
	ID        int  `json:"id"`
	IsNext    bool `json:"is_next"`
	IsCurrent bool `json:"is_current"`
	Finished  bool `json:"finished"`
}

type captainScore struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Team      string   `json:"team"`
	Position  string   `json:"position"`
	GW        int      `json:"gw"`
	EPNext    float64  `json:"ep_next"`
	FDR       float64  `json:"fdr_avg"`
	XGI90     float64  `json:"xgi_per_90"`
	HomeBoost float64  `json:"home_boost"`
	MinRisk   float64  `json:"minutes_risk"`
	DGWMult   float64  `json:"dgw_multiplier"`
	Score     float64  `json:"captain_score"`
	Fixtures  []string `json:"fixtures"`
}

// scoreCaptain is the pure scoring function — no IO. Exposed so unit tests
// can verify the formula across fixed inputs without hitting the network.
func scoreCaptain(epNext, fdrAvg, xgi90 float64, minutesSeason, dgwExtra int, allHome bool) float64 {
	if epNext <= 0 {
		return 0
	}
	fdrMult := 2.2 - 0.3*fdrAvg
	if fdrMult < 0.3 {
		fdrMult = 0.3
	}
	homeBoost := 1.10
	if !allHome {
		// Mixed (DGW with one home/one away) or pure away.
		homeBoost = 0.95
	}
	xgiBoost := 1.0 + 0.3*xgi90
	// Minutes risk: 0.5 floor for players under 5 starts equivalent (~450 min);
	// linear to 1.0 above 1350 min (15 starts).
	minRisk := 0.5
	switch {
	case minutesSeason >= 1350:
		minRisk = 1.0
	case minutesSeason >= 450:
		minRisk = 0.5 + 0.5*(float64(minutesSeason-450)/900.0)
	}
	dgwMult := 1.0 + 0.85*float64(dgwExtra)
	return epNext * fdrMult * homeBoost * xgiBoost * minRisk * dgwMult
}

func resolveNextGW(events []captainEvent, override int) (int, error) {
	if override > 0 {
		return override, nil
	}
	for _, e := range events {
		if e.IsNext {
			return e.ID, nil
		}
	}
	for _, e := range events {
		if e.IsCurrent && !e.Finished {
			return e.ID, nil
		}
	}
	return 0, fmt.Errorf("could not determine next gameweek")
}

func newCaptainCmd(flags *rootFlags) *cobra.Command {
	var top int
	var gw int
	var explain bool
	var minMinutes int
	cmd := &cobra.Command{
		Use:   "captain",
		Short: "Ranked captain picks for the next gameweek (FDR + xGI + home/away)",
		Long: `Captain ranker v0.2.

Scores each available outfield player and goalkeeper by:
  ep_next × FDR multiplier × home/away boost × (1 + 0.3·xGI/90) × minutes risk × DGW multiplier

Add --explain to see each component, --gw N to score a future gameweek, and
--min-minutes to filter out players with low season minutes.`,
		Example: "  livefpl captain --top 5\n  livefpl captain --top 10 --explain\n  livefpl captain --gw 38",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			boot, err := c.Get("https://fantasy.premierleague.com/api/bootstrap-static/", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var b struct {
				Elements []captainElement `json:"elements"`
				Events   []captainEvent   `json:"events"`
				Teams    []bootTeam       `json:"teams"`
			}
			if err := json.Unmarshal(boot, &b); err != nil {
				return apiErr(fmt.Errorf("decoding bootstrap-static: %w", err))
			}

			targetGW, err := resolveNextGW(b.Events, gw)
			if err != nil {
				return apiErr(err)
			}

			fix, err := c.Get("https://fantasy.premierleague.com/api/fixtures/", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var fixtures []captainFixture
			if err := json.Unmarshal(fix, &fixtures); err != nil {
				return apiErr(fmt.Errorf("decoding fixtures: %w", err))
			}

			type teamFixture struct {
				Opponent int
				FDR      int
				IsHome   bool
			}
			byTeam := map[int][]teamFixture{}
			for _, f := range fixtures {
				if f.Event != targetGW || f.Finished {
					continue
				}
				byTeam[f.TeamH] = append(byTeam[f.TeamH], teamFixture{f.TeamA, f.HDiff, true})
				byTeam[f.TeamA] = append(byTeam[f.TeamA], teamFixture{f.TeamH, f.ADiff, false})
			}

			teamShortByID := map[int]string{}
			for _, t := range b.Teams {
				teamShortByID[t.ID] = t.ShortName
			}

			picks := make([]captainScore, 0, 32)
			for _, e := range b.Elements {
				if e.Status != "a" {
					continue
				}
				if e.Minutes < minMinutes {
					continue
				}
				ep, _ := strconv.ParseFloat(e.EPNext, 64)
				if ep <= 0 {
					continue
				}
				xgi := e.ExpectedGI90
				fxs := byTeam[e.Team]
				if len(fxs) == 0 {
					continue
				}
				totalFDR := 0
				allHome := true
				fixStrs := make([]string, 0, len(fxs))
				for _, f := range fxs {
					totalFDR += f.FDR
					if !f.IsHome {
						allHome = false
					}
					opp := teamShortByID[f.Opponent]
					if f.IsHome {
						fixStrs = append(fixStrs, fmt.Sprintf("%s (H, FDR %d)", opp, f.FDR))
					} else {
						fixStrs = append(fixStrs, fmt.Sprintf("%s (A, FDR %d)", opp, f.FDR))
					}
				}
				fdrAvg := float64(totalFDR) / float64(len(fxs))
				dgwExtra := len(fxs) - 1
				score := scoreCaptain(ep, fdrAvg, xgi, e.Minutes, dgwExtra, allHome)
				homeBoost := 1.10
				if !allHome {
					homeBoost = 0.95
				}
				p := captainScore{
					ID:        e.ID,
					Name:      e.WebName,
					Team:      teamShortByID[e.Team],
					Position:  positionByType[e.ElementType],
					GW:        targetGW,
					EPNext:    ep,
					FDR:       fdrAvg,
					XGI90:     xgi,
					HomeBoost: homeBoost,
					MinRisk:   captainMinRisk(e.Minutes),
					DGWMult:   1.0 + 0.85*float64(dgwExtra),
					Score:     score,
					Fixtures:  fixStrs,
				}
				picks = append(picks, p)
			}

			sort.Slice(picks, func(i, j int) bool { return picks[i].Score > picks[j].Score })
			if top > 0 && len(picks) > top {
				picks = picks[:top]
			}

			payload := map[string]any{
				"gw":     targetGW,
				"method": "ep_next * (2.2-0.3*fdr) * home_boost * (1+0.3*xgi_per_90) * minutes_risk * dgw_multiplier",
				"picks":  picks,
			}
			if !explain {
				slim := make([]map[string]any, 0, len(picks))
				for _, p := range picks {
					slim = append(slim, map[string]any{
						"id":            p.ID,
						"name":          p.Name,
						"team":          p.Team,
						"captain_score": p.Score,
						"fixtures":      p.Fixtures,
					})
				}
				payload["picks"] = slim
			}
			out, _ := json.MarshalIndent(payload, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
	cmd.Flags().IntVar(&top, "top", 5, "Number of captain picks to return")
	cmd.Flags().IntVar(&gw, "gw", 0, "Gameweek to score (default: next)")
	cmd.Flags().BoolVar(&explain, "explain", false, "Show per-component breakdown")
	cmd.Flags().IntVar(&minMinutes, "min-minutes", 270, "Filter out players with less than N season minutes (default: 3 GW equivalents)")
	return cmd
}

func captainMinRisk(minutesSeason int) float64 {
	if minutesSeason >= 1350 {
		return 1.0
	}
	if minutesSeason >= 450 {
		return 0.5 + 0.5*(float64(minutesSeason-450)/900.0)
	}
	return 0.5
}

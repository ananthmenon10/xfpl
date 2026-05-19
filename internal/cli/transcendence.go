// Copyright 2026 ananth-menon. Licensed under Apache-2.0. See LICENSE.
//
// Hand-authored novel commands (manifest C2/C4/C5):
//
//   livefpl explain rank <teamId>  — top contributors to current GW rank
//   livefpl captain                — ranked captain picks for the next GW
//   livefpl chip-plan <teamId>     — blank/double GWs vs remaining chips
//
// These synthesize multiple endpoints; the underlying single-endpoint
// commands remain available (livefplapi, elite-json, lh-api2, etc.).

package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
)

// ---------- explain ----------

func newExplainCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain",
		Short: "Explain rank / live-points changes",
	}
	cmd.AddCommand(newExplainRankCmd(flags))
	return cmd
}

func newExplainRankCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "rank <teamId>",
		Short:   "Top contributors to current gameweek rank for a team",
		Example: "  livefpl explain rank 5505524",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			teamID := args[0]

			data, err := c.Get("https://www.livefpl.net/livefplapi/"+teamID, nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var snap map[string]json.RawMessage
			if err := json.Unmarshal(data, &snap); err != nil {
				return apiErr(fmt.Errorf("decoding livefpl snapshot: %w", err))
			}

			var gwRank int64
			_ = json.Unmarshal(snap["GWrank"], &gwRank)
			var benchPts int64
			_ = json.Unmarshal(snap["bench"], &benchPts)
			var capPts int64
			_ = json.Unmarshal(snap["cap_pts"], &capPts)
			var gwPoints int
			var scores []int
			if err := json.Unmarshal(snap["scores"], &scores); err == nil && len(scores) > 0 {
				gwPoints = scores[0]
			}
			var pts map[string]int64
			_ = json.Unmarshal(snap["calcplayerptsd"], &pts)

			// Determine current GW so we can fetch FPL-side picks.
			bootData, err := c.Get("https://fantasy.premierleague.com/api/bootstrap-static/", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var bs struct {
				Elements []bootElement `json:"elements"`
				Teams    []bootTeam    `json:"teams"`
				Events   []struct {
					ID        int  `json:"id"`
					IsCurrent bool `json:"is_current"`
					IsNext    bool `json:"is_next"`
					Finished  bool `json:"finished"`
				} `json:"events"`
			}
			_ = json.Unmarshal(bootData, &bs)
			nameByID := map[int]string{}
			teamByID := map[int]string{}
			b := &bootStatic{Elements: bs.Elements, Teams: bs.Teams}
			for _, e := range bs.Elements {
				nameByID[e.ID] = e.WebName
				teamByID[e.ID] = teamShort(b, e.Team)
			}
			curGW := 0
			for _, ev := range bs.Events {
				if ev.IsCurrent {
					curGW = ev.ID
					break
				}
			}
			if curGW == 0 {
				// Fallback to last finished GW.
				for _, ev := range bs.Events {
					if ev.Finished && ev.ID > curGW {
						curGW = ev.ID
					}
				}
			}

			type contrib struct {
				ElementID  int    `json:"element_id"`
				Name       string `json:"name,omitempty"`
				Team       string `json:"team,omitempty"`
				Raw        int    `json:"raw_points"`
				Multiplier int    `json:"multiplier"`
				Effective  int    `json:"effective_points"`
				IsCaptain  bool   `json:"is_captain,omitempty"`
			}
			top := make([]contrib, 0, 11)

			if curGW > 0 {
				picksData, perr := c.Get(
					fmt.Sprintf("https://fantasy.premierleague.com/api/entry/%s/event/%d/picks/", teamID, curGW),
					nil,
				)
				if perr == nil {
					var picks struct {
						Picks []struct {
							Element    int  `json:"element"`
							Multiplier int  `json:"multiplier"`
							IsCaptain  bool `json:"is_captain"`
						} `json:"picks"`
					}
					if json.Unmarshal(picksData, &picks) == nil {
						for _, p := range picks.Picks {
							raw := int(pts[strconv.Itoa(p.Element)])
							if p.Multiplier == 0 {
								continue
							}
							top = append(top, contrib{
								ElementID:  p.Element,
								Name:       nameByID[p.Element],
								Team:       teamByID[p.Element],
								Raw:        raw,
								Multiplier: p.Multiplier,
								Effective:  raw * p.Multiplier,
								IsCaptain:  p.IsCaptain,
							})
						}
						sort.Slice(top, func(i, j int) bool { return top[i].Effective > top[j].Effective })
						if len(top) > 5 {
							top = top[:5]
						}
					}
				}
			}

			payload := map[string]any{
				"team_id":          teamID,
				"gw":               curGW,
				"gw_points":        gwPoints,
				"gw_rank":          gwRank,
				"bench_points":     benchPts,
				"captain_points":   capPts,
				"top_contributors": top,
			}
			out, _ := json.MarshalIndent(payload, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
}

// ---------- chip-plan ----------

func newChipPlanCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "chip-plan <teamId>",
		Short:   "Remaining chips × upcoming blank / double gameweeks",
		Example: "  livefpl chip-plan 5505524",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			plan, err := c.Get(
				"https://livefpl-api-489391001748.europe-west4.run.app/LH_api2/planner/snapshot",
				map[string]string{"id": args[0]},
			)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var p struct {
				Data struct {
					ChipsAvailable map[string]int `json:"chips_available"`
					BaseGW         int            `json:"base_gw"`
					ManagerName    string         `json:"manager_name"`
				} `json:"data"`
			}
			_ = json.Unmarshal(plan, &p)

			boot, err := c.Get("https://fantasy.premierleague.com/api/bootstrap-static/", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var b struct {
				Events []struct {
					ID        int    `json:"id"`
					Name      string `json:"name"`
					Finished  bool   `json:"finished"`
					IsCurrent bool   `json:"is_current"`
					IsNext    bool   `json:"is_next"`
				} `json:"events"`
			}
			_ = json.Unmarshal(boot, &b)

			fix, err := c.Get("https://fantasy.premierleague.com/api/fixtures/", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var fixtures []struct {
				Event    int  `json:"event"`
				TeamH    int  `json:"team_h"`
				TeamA    int  `json:"team_a"`
				Finished bool `json:"finished"`
			}
			_ = json.Unmarshal(fix, &fixtures)

			// Count fixtures per team per event. >1 = DGW, 0 = BGW.
			teamPerEvent := map[[2]int]int{}
			for _, f := range fixtures {
				if f.Event == 0 {
					continue
				}
				teamPerEvent[[2]int{f.Event, f.TeamH}]++
				teamPerEvent[[2]int{f.Event, f.TeamA}]++
			}
			summaries := make([]eventSummary, 0)
			for _, ev := range b.Events {
				if ev.Finished {
					continue
				}
				blanks, doubles := 0, 0
				for t := 1; t <= 20; t++ {
					n := teamPerEvent[[2]int{ev.ID, t}]
					if n == 0 {
						blanks++
					} else if n > 1 {
						doubles++
					}
				}
				if blanks > 0 || doubles > 0 {
					summaries = append(summaries, eventSummary{ev.ID, blanks, doubles})
				}
			}

			out, _ := json.MarshalIndent(map[string]any{
				"manager":          p.Data.ManagerName,
				"chips_available":  p.Data.ChipsAvailable,
				"base_gw":          p.Data.BaseGW,
				"upcoming_bgw_dgw": summaries,
				"recommendation":   chipRecommendation(p.Data.ChipsAvailable, summaries),
			}, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
}

type eventSummary struct {
	GW      int `json:"gw"`
	Blanks  int `json:"blanks"`
	Doubles int `json:"doubles"`
}

func chipRecommendation(avail map[string]int, evs []eventSummary) []string {
	var out []string
	for _, e := range evs {
		if e.Doubles >= 4 && avail["bboost"] > 0 {
			out = append(out, fmt.Sprintf("GW%d: Bench Boost (%d doubles)", e.GW, e.Doubles))
		}
		if e.Doubles >= 3 && avail["3xc"] > 0 {
			out = append(out, fmt.Sprintf("GW%d: Triple Captain on a DGW player", e.GW))
		}
		if e.Blanks >= 8 && avail["freehit"] > 0 {
			out = append(out, fmt.Sprintf("GW%d: Free Hit (%d teams blank)", e.GW, e.Blanks))
		}
		if e.Blanks >= 5 && avail["wildcard"] > 0 {
			out = append(out, fmt.Sprintf("GW%d: Wildcard before this blank-heavy week", e.GW))
		}
	}
	if len(out) == 0 {
		out = append(out, "no obvious chip windows in remaining gameweeks")
	}
	return out
}

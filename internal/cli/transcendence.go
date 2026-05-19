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
			data, err := c.Get("https://www.livefpl.net/livefplapi/"+args[0], nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var snap map[string]json.RawMessage
			if err := json.Unmarshal(data, &snap); err != nil {
				return apiErr(fmt.Errorf("decoding livefpl snapshot: %w", err))
			}

			var gwRank int64
			_ = json.Unmarshal(snap["GWrank"], &gwRank)

			var bench int64
			_ = json.Unmarshal(snap["bench"], &bench)

			var pts map[string]int64
			_ = json.Unmarshal(snap["calcplayerptsd"], &pts)

			type contrib struct {
				ElementID int   `json:"element_id"`
				Points    int64 `json:"points"`
			}
			out := make([]contrib, 0, len(pts))
			for k, v := range pts {
				id, _ := strconv.Atoi(k)
				out = append(out, contrib{ElementID: id, Points: v})
			}
			sort.Slice(out, func(i, j int) bool { return out[i].Points > out[j].Points })
			if len(out) > 5 {
				out = out[:5]
			}

			payload := map[string]any{
				"team_id":           args[0],
				"gw_rank":           gwRank,
				"bench_points":      bench,
				"top_contributors":  out,
			}
			b, _ := json.MarshalIndent(payload, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
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
					ID                  int    `json:"id"`
					Name                string `json:"name"`
					Finished            bool   `json:"finished"`
					IsCurrent           bool   `json:"is_current"`
					IsNext              bool   `json:"is_next"`
				} `json:"events"`
			}
			_ = json.Unmarshal(boot, &b)

			fix, err := c.Get("https://fantasy.premierleague.com/api/fixtures/", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var fixtures []struct {
				Event  int  `json:"event"`
				TeamH  int  `json:"team_h"`
				TeamA  int  `json:"team_a"`
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
				"manager":         p.Data.ManagerName,
				"chips_available": p.Data.ChipsAvailable,
				"base_gw":         p.Data.BaseGW,
				"upcoming_bgw_dgw": summaries,
				"recommendation": chipRecommendation(p.Data.ChipsAvailable, summaries),
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

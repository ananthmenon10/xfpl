// Copyright 2026 ananth-menon. Licensed under Apache-2.0. See LICENSE.
//
// Hand-authored hermes-inspired commands:
//
//   livefpl captains-history <teamId>  — captain picks per gameweek with points
//   livefpl points <teamId>            — current GW points breakdown by player
//   livefpl cup <teamId>               — FPL Cup status and match history
//
// Each wraps multiple FPL endpoints and joins with bootstrap-static for names.

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// ---------- captains-history ----------

func newCaptainsHistoryCmd(flags *rootFlags) *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:     "captains-history <teamId>",
		Short:   "Captain picks per gameweek with points scored",
		Example: "  livefpl captains-history 5505524 --limit 10",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamID := args[0]
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			boot, err := fetchBootstrap(flags)
			if err != nil {
				return err
			}
			nameByID := map[int]string{}
			for _, e := range boot.Elements {
				nameByID[e.ID] = e.WebName
			}

			histData, err := c.Get("https://fantasy.premierleague.com/api/entry/"+teamID+"/history/", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var hist struct {
				Current []struct {
					Event  int `json:"event"`
					Points int `json:"points"`
					Rank   int `json:"rank"`
				} `json:"current"`
			}
			if err := json.Unmarshal(histData, &hist); err != nil {
				return apiErr(fmt.Errorf("decoding history: %w", err))
			}

			events := hist.Current
			if limit > 0 && len(events) > limit {
				events = events[len(events)-limit:]
			}

			type captainEntry struct {
				GW          int    `json:"gw"`
				CaptainID   int    `json:"captain_id"`
				CaptainName string `json:"captain_name"`
				Multiplier  int    `json:"multiplier"`
				Points      int    `json:"captain_points"`
				GWPoints    int    `json:"gw_points"`
				GWRank      int    `json:"gw_rank"`
			}
			out := make([]captainEntry, 0, len(events))
			for _, ev := range events {
				picksData, err := c.Get(
					fmt.Sprintf("https://fantasy.premierleague.com/api/entry/%s/event/%d/picks/", teamID, ev.Event),
					nil,
				)
				if err != nil {
					// Skip unfinished/inaccessible GWs but don't kill the whole run.
					continue
				}
				var picks struct {
					Picks []struct {
						Element    int  `json:"element"`
						IsCaptain  bool `json:"is_captain"`
						Multiplier int  `json:"multiplier"`
					} `json:"picks"`
				}
				if err := json.Unmarshal(picksData, &picks); err != nil {
					continue
				}

				liveData, _ := c.Get(
					fmt.Sprintf("https://fantasy.premierleague.com/api/event/%d/live/", ev.Event),
					nil,
				)
				var live struct {
					Elements []struct {
						ID    int `json:"id"`
						Stats struct {
							TotalPoints int `json:"total_points"`
						} `json:"stats"`
					} `json:"elements"`
				}
				_ = json.Unmarshal(liveData, &live)
				ptsByID := map[int]int{}
				for _, e := range live.Elements {
					ptsByID[e.ID] = e.Stats.TotalPoints
				}

				for _, p := range picks.Picks {
					if !p.IsCaptain {
						continue
					}
					rawPts := ptsByID[p.Element]
					out = append(out, captainEntry{
						GW:          ev.Event,
						CaptainID:   p.Element,
						CaptainName: nameByID[p.Element],
						Multiplier:  p.Multiplier,
						Points:      rawPts * p.Multiplier,
						GWPoints:    ev.Points,
						GWRank:      ev.Rank,
					})
					break
				}
			}

			payload := map[string]any{
				"team_id":  teamID,
				"captains": out,
			}
			b, _ := json.MarshalIndent(payload, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 10, "Number of most-recent gameweeks to include")
	return cmd
}

// ---------- points ----------

func newPointsCmd(flags *rootFlags) *cobra.Command {
	var gw int
	cmd := &cobra.Command{
		Use:     "points <teamId>",
		Short:   "Per-player points breakdown for a gameweek (live during play)",
		Example: "  livefpl points 5505524\n  livefpl points 5505524 --gw 30",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamID := args[0]
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			boot, err := fetchBootstrap(flags)
			if err != nil {
				return err
			}
			nameByID := map[int]string{}
			for _, e := range boot.Elements {
				nameByID[e.ID] = e.WebName
			}

			if gw == 0 {
				bootData, err := c.Get("https://fantasy.premierleague.com/api/bootstrap-static/", nil)
				if err != nil {
					return classifyAPIError(err, flags)
				}
				var bs struct {
					Events []struct {
						ID        int  `json:"id"`
						IsCurrent bool `json:"is_current"`
						IsNext    bool `json:"is_next"`
					} `json:"events"`
				}
				_ = json.Unmarshal(bootData, &bs)
				for _, ev := range bs.Events {
					if ev.IsCurrent {
						gw = ev.ID
						break
					}
				}
				if gw == 0 {
					for _, ev := range bs.Events {
						if ev.IsNext {
							gw = ev.ID
							break
						}
					}
				}
				if gw == 0 {
					return apiErr(fmt.Errorf("could not determine current gameweek"))
				}
			}

			picksData, err := c.Get(
				fmt.Sprintf("https://fantasy.premierleague.com/api/entry/%s/event/%d/picks/", teamID, gw),
				nil,
			)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var picks struct {
				EntryHistory struct {
					Points    int `json:"points"`
					TotalPoints int `json:"total_points"`
					Rank      int `json:"rank"`
					EventTransfersCost int `json:"event_transfers_cost"`
					Bank      int `json:"bank"`
					Value     int `json:"value"`
				} `json:"entry_history"`
				Picks []struct {
					Element    int  `json:"element"`
					Position   int  `json:"position"`
					Multiplier int  `json:"multiplier"`
					IsCaptain  bool `json:"is_captain"`
					IsViceCaptain bool `json:"is_vice_captain"`
				} `json:"picks"`
			}
			if err := json.Unmarshal(picksData, &picks); err != nil {
				return apiErr(fmt.Errorf("decoding picks: %w", err))
			}

			liveData, err := c.Get(
				fmt.Sprintf("https://fantasy.premierleague.com/api/event/%d/live/", gw),
				nil,
			)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var live struct {
				Elements []struct {
					ID    int `json:"id"`
					Stats struct {
						TotalPoints int `json:"total_points"`
						Minutes     int `json:"minutes"`
						Goals       int `json:"goals_scored"`
						Assists     int `json:"assists"`
						Bonus       int `json:"bonus"`
					} `json:"stats"`
				} `json:"elements"`
			}
			if err := json.Unmarshal(liveData, &live); err != nil {
				return apiErr(fmt.Errorf("decoding live: %w", err))
			}
			byID := map[int]struct {
				Pts, Min, G, A, B int
			}{}
			for _, e := range live.Elements {
				byID[e.ID] = struct{ Pts, Min, G, A, B int }{
					Pts: e.Stats.TotalPoints, Min: e.Stats.Minutes,
					G: e.Stats.Goals, A: e.Stats.Assists, B: e.Stats.Bonus,
				}
			}

			type pickBreakdown struct {
				Position   int    `json:"position"`
				Element    int    `json:"element"`
				Name       string `json:"name"`
				Raw        int    `json:"raw_points"`
				Multiplier int    `json:"multiplier"`
				Effective  int    `json:"effective_points"`
				Minutes    int    `json:"minutes"`
				Goals      int    `json:"goals"`
				Assists    int    `json:"assists"`
				Bonus      int    `json:"bonus"`
				Captain    bool   `json:"is_captain,omitempty"`
				Vice       bool   `json:"is_vice_captain,omitempty"`
			}
			breakdown := make([]pickBreakdown, 0, 15)
			for _, p := range picks.Picks {
				s := byID[p.Element]
				breakdown = append(breakdown, pickBreakdown{
					Position:   p.Position,
					Element:    p.Element,
					Name:       nameByID[p.Element],
					Raw:        s.Pts,
					Multiplier: p.Multiplier,
					Effective:  s.Pts * p.Multiplier,
					Minutes:    s.Min,
					Goals:      s.G,
					Assists:    s.A,
					Bonus:      s.B,
					Captain:    p.IsCaptain,
					Vice:       p.IsViceCaptain,
				})
			}

			payload := map[string]any{
				"team_id":           teamID,
				"gw":                gw,
				"gw_points":         picks.EntryHistory.Points,
				"gw_rank":           picks.EntryHistory.Rank,
				"transfer_cost":     picks.EntryHistory.EventTransfersCost,
				"team_value":        float64(picks.EntryHistory.Value) / 10.0,
				"bank":              float64(picks.EntryHistory.Bank) / 10.0,
				"breakdown":         breakdown,
			}
			b, _ := json.MarshalIndent(payload, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		},
	}
	cmd.Flags().IntVar(&gw, "gw", 0, "Gameweek to inspect (default: current)")
	return cmd
}

// ---------- cup ----------

func newCupCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "cup <teamId>",
		Short:   "FPL Cup status and match history",
		Example: "  livefpl cup 5505524",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamID := args[0]
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			payload := map[string]any{
				"team_id": teamID,
			}

			data, err := c.Get(
				"https://fantasy.premierleague.com/api/entry/"+teamID+"/cup-status/",
				nil,
			)
			if err != nil {
				// 404 is the legitimate "not in cup" signal; surface it cleanly.
				payload["status"] = "not_in_cup"
				payload["note"] = "manager is not in the FPL Cup this season (or cup not yet started)"
				b, _ := json.MarshalIndent(payload, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(b))
				return nil
			}
			var resp map[string]json.RawMessage
			if err := json.Unmarshal(data, &resp); err != nil {
				return apiErr(fmt.Errorf("decoding cup-status: %w", err))
			}
			payload["cup"] = resp

			// Also try /entry/{id}/cup/ for match history.
			matchData, err := c.Get(
				"https://fantasy.premierleague.com/api/entry/"+teamID+"/cup/",
				nil,
			)
			if err == nil {
				var matches map[string]json.RawMessage
				if json.Unmarshal(matchData, &matches) == nil {
					payload["matches"] = matches
				}
			}

			b, _ := json.MarshalIndent(payload, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		},
	}
}


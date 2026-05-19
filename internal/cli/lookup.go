// Copyright 2026 ananth-menon. Licensed under Apache-2.0. See LICENSE.
//
// Hand-authored manifest commands:
//
//   livefpl compare <name> <name> [<name> ...]  — side-by-side player view (A4)
//   livefpl player  <name>                      — name-based player lookup (A2)
//
// Both resolve names against bootstrap-static.elements[].web_name with a
// simple case-insensitive substring match.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type bootElement struct {
	ID            int    `json:"id"`
	WebName       string `json:"web_name"`
	FirstName     string `json:"first_name"`
	SecondName    string `json:"second_name"`
	Team          int    `json:"team"`
	ElementType   int    `json:"element_type"`
	NowCost       int    `json:"now_cost"`
	Form          string `json:"form"`
	TotalPoints   int    `json:"total_points"`
	SelectedByPct string `json:"selected_by_percent"`
	Status        string `json:"status"`
	News          string `json:"news"`
	EPNext        string `json:"ep_next"`
	Minutes       int    `json:"minutes"`
}

type bootTeam struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

type bootStatic struct {
	Elements []bootElement `json:"elements"`
	Teams    []bootTeam    `json:"teams"`
}

func fetchBootstrap(flags *rootFlags) (*bootStatic, error) {
	c, err := flags.newClient()
	if err != nil {
		return nil, err
	}
	data, err := c.Get("https://fantasy.premierleague.com/api/bootstrap-static/", nil)
	if err != nil {
		return nil, classifyAPIError(err, flags)
	}
	var b bootStatic
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, apiErr(fmt.Errorf("decoding bootstrap-static: %w", err))
	}
	return &b, nil
}

func resolvePlayerByName(b *bootStatic, query string) (*bootElement, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil, usageErr(fmt.Errorf("player name required"))
	}
	var matches []bootElement
	for _, e := range b.Elements {
		web := strings.ToLower(e.WebName)
		first := strings.ToLower(e.FirstName)
		second := strings.ToLower(e.SecondName)
		full := strings.ToLower(e.FirstName + " " + e.SecondName)
		if web == q || full == q {
			return &e, nil
		}
		if strings.Contains(web, q) || strings.Contains(second, q) || strings.Contains(first, q) {
			matches = append(matches, e)
		}
	}
	switch len(matches) {
	case 0:
		return nil, notFoundErr(fmt.Errorf("no player matched %q", query))
	case 1:
		m := matches[0]
		return &m, nil
	default:
		names := make([]string, 0, len(matches))
		for _, m := range matches {
			names = append(names, fmt.Sprintf("%s (id=%d)", m.WebName, m.ID))
		}
		return nil, usageErr(fmt.Errorf("ambiguous player %q matched %d players: %s — use --id or refine the name",
			query, len(matches), strings.Join(names, ", ")))
	}
}

func teamShort(b *bootStatic, id int) string {
	for _, t := range b.Teams {
		if t.ID == id {
			return t.ShortName
		}
	}
	return fmt.Sprintf("T%d", id)
}

type playerView struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Team          string  `json:"team"`
	Position      string  `json:"position"`
	NowCost       float64 `json:"now_cost"`
	Form          string  `json:"form"`
	TotalPoints   int     `json:"total_points"`
	SelectedByPct string  `json:"selected_by_percent"`
	Status        string  `json:"status"`
	News          string  `json:"news,omitempty"`
	EPNext        string  `json:"ep_next"`
}

var positionByType = map[int]string{1: "GK", 2: "DEF", 3: "MID", 4: "FWD"}

func viewOf(b *bootStatic, e *bootElement) playerView {
	return playerView{
		ID:            e.ID,
		Name:          e.WebName,
		Team:          teamShort(b, e.Team),
		Position:      positionByType[e.ElementType],
		NowCost:       float64(e.NowCost) / 10.0,
		Form:          e.Form,
		TotalPoints:   e.TotalPoints,
		SelectedByPct: e.SelectedByPct,
		Status:        e.Status,
		News:          e.News,
		EPNext:        e.EPNext,
	}
}

func newPlayerLookupCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "player <name>",
		Aliases: []string{"who"},
		Short:   "Find a player by name (case-insensitive substring match)",
		Example: "  livefpl player Haaland",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := fetchBootstrap(flags)
			if err != nil {
				return err
			}
			e, err := resolvePlayerByName(b, args[0])
			if err != nil {
				return err
			}
			v := viewOf(b, e)
			out, _ := json.MarshalIndent(v, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
}

func newCompareCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "compare <name> <name> [<name>...]",
		Short:   "Side-by-side comparison of 2+ players (price, form, points, ownership)",
		Example: "  livefpl compare Haaland Salah Watkins",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := fetchBootstrap(flags)
			if err != nil {
				return err
			}
			views := make([]playerView, 0, len(args))
			for _, name := range args {
				e, err := resolvePlayerByName(b, name)
				if err != nil {
					return err
				}
				views = append(views, viewOf(b, e))
			}
			out, _ := json.MarshalIndent(map[string]any{"players": views}, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
}

// Copyright 2026 ananth-menon. Licensed under Apache-2.0. See LICENSE.
//
// Hand-authored: map manifest-friendly names (rank, prices, live, planner,
// top10k, trending, player, league, picks, gw) onto the auto-generated
// endpoint commands via Cobra aliases. Kept in one file so a future
// regen-merge can preserve it without touching DO-NOT-EDIT promoted files.

package cli

import "github.com/spf13/cobra"

// manifestAliases is the source-of-truth mapping from auto-derived endpoint
// command names to the human-friendly aliases the manifest commits to.
//
// Update both this table and the manifest in lockstep if either changes.
var manifestAliases = map[string][]string{
	"livefplapi":          {"rank"},
	"prices-json":         {"prices"},
	"games-json":          {"live"},
	"top-transfers-json":  {"trending"},
	"lh-api2":             {"planner"},
	"elite-json":          {"top10k", "eo"},
	"version-json":        {"livefpl-version"},
	"element-summary":     {"player-id"},
	"leagues-classic":     {"league"},
	"leagues-h2h-matches": {"h2h"},
	"entry":               {"manager"},
	"dream-team":          {"dt"},
	"bootstrap-static":    {"static"},
	"team":                {"set-pieces"},
}

// applyManifestAliases walks the root command tree and attaches the
// manifest-friendly aliases to each generated subcommand.  Safe to call
// after every newRootCmd().
func applyManifestAliases(root *cobra.Command) {
	for _, sub := range root.Commands() {
		if extra, ok := manifestAliases[sub.Name()]; ok {
			// Preserve any aliases the generator may have set.
			sub.Aliases = append(sub.Aliases, extra...)
		}
	}
}

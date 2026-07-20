package main

import (
	"github.com/spf13/cobra"
)

var devxCmd = &cobra.Command{
	Use:   "devx",
	Short: "OmniDevX developer-experience telemetry",
	Long: `Read collected OmniDevX events (Claude Code, Codex CLI, git, GitHub)
from the local store and build period reports and dashboards from them.

Requires events already collected into the local store
(~/.plexusone/omnidevx/data/ by default) via the omnidevx-core providers;
this command only reads and reports, it does not collect.`,
}

func init() {
	rootCmd.AddCommand(devxCmd)
}

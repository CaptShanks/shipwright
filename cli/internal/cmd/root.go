package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cliVersion  string
	targetFlag  string
	globalFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "ship",
	Short: "Shipwright plugin manager for AI development tools",
	Long: `Ship is a universal CLI that installs shipwright agents and skills
into any supported AI tool's native directory structure.

Supported tools: cursor, claude, codex
Install modes: local (per-project, default) or global (user-wide)`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&targetFlag, "target", "t", "all", "target AI tool (cursor, claude, codex, all)")
	rootCmd.PersistentFlags().BoolVarP(&globalFlag, "global", "g", false, "install globally instead of per-project")
}

func SetVersion(v string) {
	cliVersion = v
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}

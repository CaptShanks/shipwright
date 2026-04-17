package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/CaptShanks/shipwright/cli/internal/tui"
)

var (
	cliVersion string
	targetFlag string
	globalFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "ship",
	Short: "Shipwright plugin manager for AI development tools",
	Long: `Ship is a universal CLI that installs shipwright agents and skills
into any supported AI tool's native directory structure.

Supported tools: cursor, claude, codex
Install modes: local (per-project, default) or global (user-wide)

Run with no arguments to launch the interactive TUI dashboard.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		app := tui.NewApp(cliVersion)
		p := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
		return nil
	},
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

package cmd

import (
	"fmt"

	"github.com/CaptShanks/shipwright/cli/internal/installer"
	"github.com/CaptShanks/shipwright/cli/internal/state"
	"github.com/CaptShanks/shipwright/cli/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed shipwright plugins",
	Long: `Show all plugins currently installed in the project or globally.

Examples:
  ship list
  ship list --target cursor
  ship list --global`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	scope := "local"
	if globalFlag {
		scope = "global"
	}

	projectPath := ""
	if scope == "local" {
		projectPath = installer.ProjectRoot()
	}

	store := state.NewStore()
	installs, err := store.List(targetFlag, scope, projectPath)
	if err != nil {
		return fmt.Errorf("reading state: %w", err)
	}

	if len(installs) == 0 {
		ui.Warn("no plugins installed")
		return nil
	}

	ui.Header("Installed plugins")

	var rows [][]string
	for _, inst := range installs {
		rows = append(rows, []string{
			inst.Plugin,
			inst.Target,
			inst.Scope,
			inst.Version,
			fmt.Sprintf("%d files", len(inst.Files)),
		})
	}

	ui.Table(
		[]string{"PLUGIN", "TARGET", "SCOPE", "VERSION", "FILES"},
		rows,
	)
	fmt.Println()
	return nil
}

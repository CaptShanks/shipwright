package cmd

import (
	"fmt"

	"github.com/CaptShanks/shipwright/cli/internal/installer"
	"github.com/CaptShanks/shipwright/cli/internal/state"
	"github.com/CaptShanks/shipwright/cli/internal/ui"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <plugin>",
	Short: "Remove an installed shipwright plugin",
	Long: `Remove a previously installed plugin and clean up all its files.

Examples:
  ship uninstall architect-agent
  ship uninstall architect-agent --target cursor
  ship uninstall shipwright-full --global`,
	Args: cobra.ExactArgs(1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	scope := installer.ScopeLocal
	if globalFlag {
		scope = installer.ScopeGlobal
	}

	store := state.NewStore()
	installers, err := installer.ForTarget(targetFlag)
	if err != nil {
		return err
	}

	ui.Header("Uninstalling " + ui.Bold(pluginName))

	projectPath := ""
	if scope == installer.ScopeLocal {
		projectPath = installer.ProjectRoot()
	}

	removed := 0
	for _, inst := range installers {
		files, err := store.Remove(pluginName, inst.Name(), string(scope), projectPath)
		if err != nil {
			ui.Error("%s: failed to read state: %v", inst.Name(), err)
			continue
		}
		if len(files) == 0 {
			continue
		}

		if err := inst.UninstallFiles(files); err != nil {
			ui.Error("%s: failed to remove files: %v", inst.Name(), err)
			continue
		}

		ui.Success("%s: removed %d files", inst.Name(), len(files))
		removed++
	}

	if removed == 0 {
		ui.Warn("no installations found for %s", pluginName)
	}

	fmt.Println()
	return nil
}

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/CaptShanks/shipwright/cli/internal/installer"
	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/state"
	"github.com/CaptShanks/shipwright/cli/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [plugin]",
	Short: "Update installed plugins to latest versions",
	Long: `Re-download and reinstall plugins from the marketplace.
Without arguments, updates all installed plugins matching the current scope.

Examples:
  ship update
  ship update architect-agent
  ship update --target cursor --global`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	scope := installer.ScopeLocal
	if globalFlag {
		scope = installer.ScopeGlobal
	}

	store := state.NewStore()
	projectPath := ""
	if scope == installer.ScopeLocal {
		projectPath = installer.ProjectRoot()
	}

	installs, err := store.List(targetFlag, string(scope), projectPath)
	if err != nil {
		return fmt.Errorf("reading state: %w", err)
	}

	if len(args) > 0 {
		pluginName := args[0]
		filtered := make([]state.Installation, 0)
		for _, inst := range installs {
			if inst.Plugin == pluginName {
				filtered = append(filtered, inst)
			}
		}
		installs = filtered
	}

	if len(installs) == 0 {
		ui.Warn("no plugins to update")
		return nil
	}

	client := registry.NewClient()
	marketplace, err := client.FetchMarketplace()
	if err != nil {
		return fmt.Errorf("fetching marketplace: %w", err)
	}

	ui.Header("Updating plugins")

	// Deduplicate plugins (same plugin may be installed for multiple targets)
	type pluginKey struct{ name, target string }
	seen := make(map[pluginKey]bool)

	for _, inst := range installs {
		key := pluginKey{inst.Plugin, inst.Target}
		if seen[key] {
			continue
		}
		seen[key] = true

		var item *registry.MarketplaceItem
		for i := range marketplace.Plugins {
			if marketplace.Plugins[i].Name == inst.Plugin {
				item = &marketplace.Plugins[i]
				break
			}
		}
		if item == nil {
			ui.Warn("%s: no longer in marketplace, skipping", inst.Plugin)
			continue
		}

		manifest, err := client.FetchPluginManifest(item.Source)
		if err != nil {
			ui.Error("%s: %v", inst.Plugin, err)
			continue
		}

		toolInstallers, err := installer.ForTarget(inst.Target)
		if err != nil {
			ui.Error("%s: %v", inst.Plugin, err)
			continue
		}

		for _, toolInst := range toolInstallers {
			// Remove old files
			_ = toolInst.UninstallFiles(inst.Files)

			var allFiles []string

			for _, agentPath := range manifest.Agents {
				repoPath := fmt.Sprintf("%s/%s", registry.NormalizeSource(item.Source), agentPath)
				content, err := client.FetchFileContent(repoPath)
				if err != nil {
					ui.Error("%s/%s: agent download failed: %v", inst.Plugin, toolInst.Name(), err)
					continue
				}
				name := strings.TrimSuffix(agentPath, ".md")
				name = strings.TrimPrefix(name, "agents/")
				path, err := toolInst.InstallAgent(name, content, scope)
				if err != nil {
					ui.Error("%s/%s: agent install failed: %v", inst.Plugin, toolInst.Name(), err)
					continue
				}
				allFiles = append(allFiles, path)
			}

			for _, skillPath := range manifest.Skills {
				skillName := skillPath
				if idx := strings.LastIndex(skillPath, "/"); idx >= 0 {
					skillName = skillPath[idx+1:]
				}
				repoPath := "_skills/" + skillName
				files, err := client.FetchSkillTree(repoPath)
				if err != nil {
					ui.Error("%s/%s: skill download failed: %v", inst.Plugin, toolInst.Name(), err)
					continue
				}
				paths, err := toolInst.InstallSkill(skillName, files, scope)
				if err != nil {
					ui.Error("%s/%s: skill install failed: %v", inst.Plugin, toolInst.Name(), err)
					continue
				}
				allFiles = append(allFiles, paths...)
			}

			_ = store.Record(state.Installation{
				Plugin:      inst.Plugin,
				Version:     manifest.Version,
				Target:      toolInst.Name(),
				Scope:       string(scope),
				ProjectPath: projectPath,
				InstalledAt: time.Now(),
				Files:       allFiles,
			})

			ui.Success("%s/%s: updated to v%s (%d files)", inst.Plugin, toolInst.Name(), manifest.Version, len(allFiles))
		}
	}

	fmt.Println()
	return nil
}

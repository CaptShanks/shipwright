package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/CaptShanks/shipwright/cli/internal/installer"
	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/state"
	"github.com/CaptShanks/shipwright/cli/internal/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <plugin>",
	Short: "Install a shipwright plugin (agent, skill, or bundle)",
	Long: `Install a plugin into the target AI tool's native directory structure.

By default, installs locally into the current project for all supported tools.
Use --target to limit to a specific tool, or --global for user-wide installation.

Examples:
  ship install architect-agent
  ship install architect-agent --target cursor
  ship install shipwright-full --global`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	scope := installer.ScopeLocal
	if globalFlag {
		scope = installer.ScopeGlobal
	}

	client := registry.NewClient()
	store := state.NewStore()

	ui.Header("Installing " + ui.Bold(pluginName))

	marketplace, err := client.FetchMarketplace()
	if err != nil {
		return fmt.Errorf("fetching marketplace: %w", err)
	}

	var item *registry.MarketplaceItem
	for i := range marketplace.Plugins {
		if marketplace.Plugins[i].Name == pluginName {
			item = &marketplace.Plugins[i]
			break
		}
	}
	if item == nil {
		return fmt.Errorf("plugin %q not found in marketplace", pluginName)
	}

	manifest, err := client.FetchPluginManifest(item.Source)
	if err != nil {
		return fmt.Errorf("fetching plugin manifest: %w", err)
	}

	ui.Info("Found %s v%s (%s)", manifest.Name, manifest.Version, item.Category)
	if len(manifest.Agents) > 0 {
		ui.Dim("Agents: %s", strings.Join(manifest.Agents, ", "))
	}
	if len(manifest.Skills) > 0 {
		skillNames := make([]string, len(manifest.Skills))
		for i, s := range manifest.Skills {
			skillNames[i] = filepath.Base(s)
		}
		ui.Dim("Skills: %s", strings.Join(skillNames, ", "))
	}

	// Download all agent and skill content from the repo
	agentContents := make(map[string][]byte)
	for _, agentPath := range manifest.Agents {
		repoPath := fmt.Sprintf("plugins/%s/%s", item.Source, agentPath)
		content, err := client.FetchFileContent(repoPath)
		if err != nil {
			return fmt.Errorf("downloading agent %s: %w", agentPath, err)
		}
		name := strings.TrimSuffix(filepath.Base(agentPath), ".md")
		agentContents[name] = content
	}

	skillContents := make(map[string]map[string][]byte)
	for _, skillPath := range manifest.Skills {
		skillName := filepath.Base(skillPath)
		// Skills are symlinked to _skills/ in the repo, fetch from canonical source
		repoPath := "_skills/" + skillName
		files, err := client.FetchSkillTree(repoPath)
		if err != nil {
			return fmt.Errorf("downloading skill %s: %w", skillName, err)
		}
		skillContents[skillName] = files
	}

	installers, err := installer.ForTarget(targetFlag)
	if err != nil {
		return err
	}

	type result struct {
		target string
		files  []string
		err    error
	}

	var wg sync.WaitGroup
	results := make(chan result, len(installers))

	for _, inst := range installers {
		wg.Add(1)
		go func(inst installer.Installer) {
			defer wg.Done()
			var allFiles []string

			for name, content := range agentContents {
				path, err := inst.InstallAgent(name, content, scope)
				if err != nil {
					results <- result{target: inst.Name(), err: fmt.Errorf("agent %s: %w", name, err)}
					return
				}
				allFiles = append(allFiles, path)
			}

			for name, files := range skillContents {
				paths, err := inst.InstallSkill(name, files, scope)
				if err != nil {
					results <- result{target: inst.Name(), err: fmt.Errorf("skill %s: %w", name, err)}
					return
				}
				allFiles = append(allFiles, paths...)
			}

			results <- result{target: inst.Name(), files: allFiles}
		}(inst)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	projectPath := ""
	if scope == installer.ScopeLocal {
		projectPath = installer.ProjectRoot()
	}

	hadError := false
	for res := range results {
		if res.err != nil {
			ui.Error("%s: %v", res.target, res.err)
			hadError = true
			continue
		}

		ui.Success("%s: %d files installed", res.target, len(res.files))

		err := store.Record(state.Installation{
			Plugin:      pluginName,
			Version:     manifest.Version,
			Target:      res.target,
			Scope:       string(scope),
			ProjectPath: projectPath,
			InstalledAt: time.Now(),
			Files:       res.files,
		})
		if err != nil {
			ui.Warn("failed to record state for %s: %v", res.target, err)
		}
	}

	if hadError {
		return fmt.Errorf("some installations failed")
	}

	fmt.Println()
	return nil
}

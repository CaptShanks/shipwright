package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/ui"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <plugin>",
	Short: "Show detailed information about a plugin",
	Long: `Display a plugin's description, agents, skills, tags, and keywords.

Examples:
  ship info architect-agent
  ship info shipwright-full`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	client := registry.NewClient()
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

	ui.Header(ui.Bold(manifest.Name) + " v" + manifest.Version)
	fmt.Printf("  %s\n\n", manifest.Description)

	fmt.Printf("  %-12s %s\n", "Category:", item.Category)

	if len(manifest.Agents) > 0 {
		names := make([]string, len(manifest.Agents))
		for i, a := range manifest.Agents {
			names[i] = strings.TrimSuffix(filepath.Base(a), ".md")
		}
		fmt.Printf("  %-12s %s\n", "Agents:", strings.Join(names, ", "))
	}

	if len(manifest.Skills) > 0 {
		names := make([]string, len(manifest.Skills))
		for i, s := range manifest.Skills {
			names[i] = filepath.Base(s)
		}
		fmt.Printf("  %-12s %s\n", "Skills:", strings.Join(names, ", "))
	}

	if len(item.Tags) > 0 {
		fmt.Printf("  %-12s %s\n", "Tags:", strings.Join(item.Tags, ", "))
	}
	if len(item.Keywords) > 0 {
		fmt.Printf("  %-12s %s\n", "Keywords:", strings.Join(item.Keywords, ", "))
	}

	fmt.Println()
	return nil
}

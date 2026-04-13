package cmd

import (
	"fmt"
	"strings"

	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/ui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search available shipwright plugins",
	Long: `Browse the shipwright marketplace. Without a query, lists all plugins.
With a query, filters by name, description, tags, and keywords.

Examples:
  ship search
  ship search security
  ship search golang`,
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 0 {
		query = strings.ToLower(strings.Join(args, " "))
	}

	client := registry.NewClient()
	marketplace, err := client.FetchMarketplace()
	if err != nil {
		return fmt.Errorf("fetching marketplace: %w", err)
	}

	var rows [][]string
	for _, p := range marketplace.Plugins {
		if query != "" && !matchesQuery(p, query) {
			continue
		}
		rows = append(rows, []string{
			p.Name,
			p.Category,
			p.Version,
			p.Description,
		})
	}

	if len(rows) == 0 {
		ui.Warn("no plugins match %q", query)
		return nil
	}

	ui.Header("Available plugins")
	ui.Table(
		[]string{"NAME", "CATEGORY", "VERSION", "DESCRIPTION"},
		rows,
	)
	fmt.Println()
	return nil
}

func matchesQuery(p registry.MarketplaceItem, q string) bool {
	haystack := strings.ToLower(
		p.Name + " " + p.Description + " " +
			strings.Join(p.Tags, " ") + " " +
			strings.Join(p.Keywords, " ") + " " +
			p.Category,
	)
	for _, word := range strings.Fields(q) {
		if !strings.Contains(haystack, word) {
			return false
		}
	}
	return true
}

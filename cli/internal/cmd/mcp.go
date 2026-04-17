package cmd

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/CaptShanks/shipwright/cli/internal/installer"
	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/state"
	"github.com/CaptShanks/shipwright/cli/internal/ui"
	"github.com/spf13/cobra"
)

var forceFlag bool

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP (Model Context Protocol) servers",
	Long: `Install, remove, and list MCP servers across AI tools.

MCPs are configured in each tool's native config file:
  Cursor:  .cursor/mcp.json
  Claude:  .claude.json
  VS Code: .vscode/mcp.json (local) or User settings (global)`,
}

var mcpInstallCmd = &cobra.Command{
	Use:   "install <mcp-name>",
	Short: "Install an MCP server into AI tool configs",
	Long: `Add an MCP server configuration to the target tool's config file.

By default, installs locally (project-level) for all supported tools.
Use --target to limit to a specific tool, or --global for user-wide config.

Examples:
  ship mcp install context7
  ship mcp install mcp-atlassian --target cursor
  ship mcp install serena --global`,
	Args: cobra.ExactArgs(1),
	RunE: runMcpInstall,
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove <mcp-name>",
	Short: "Remove an MCP server from AI tool configs",
	Long: `Remove an MCP server entry from the target tool's config file.

Examples:
  ship mcp remove context7
  ship mcp remove mcp-atlassian --target cursor
  ship mcp remove serena --global`,
	Args: cobra.ExactArgs(1),
	RunE: runMcpRemove,
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and installed MCP servers",
	Long: `Show all MCP servers available in the shipwright marketplace.

Examples:
  ship mcp list`,
	RunE: runMcpList,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpInstallCmd)
	mcpCmd.AddCommand(mcpRemoveCmd)
	mcpCmd.AddCommand(mcpListCmd)

	mcpInstallCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "overwrite existing MCP entry")
}

func runMcpInstall(cmd *cobra.Command, args []string) error {
	mcpName := args[0]
	scope := installer.ScopeLocal
	if globalFlag {
		scope = installer.ScopeGlobal
	}

	client := registry.NewClient()
	store := state.NewStore()

	ui.Header("Installing MCP: " + ui.Bold(mcpName))

	marketplace, err := client.FetchMarketplace()
	if err != nil {
		return fmt.Errorf("fetching marketplace: %w", err)
	}

	var item *registry.MarketplaceItem
	for i := range marketplace.Plugins {
		if marketplace.Plugins[i].Name == mcpName && marketplace.Plugins[i].Category == "mcps" {
			item = &marketplace.Plugins[i]
			break
		}
	}
	if item == nil {
		return fmt.Errorf("MCP %q not found in marketplace", mcpName)
	}

	manifest, err := client.FetchMcpManifest(item.Source)
	if err != nil {
		return fmt.Errorf("fetching MCP manifest: %w", err)
	}

	ui.Info("Found %s: %s", manifest.Name, manifest.Description)

	cfg := installer.McpConfig{
		Command: manifest.Command,
		Args:    manifest.Args,
		Env:     manifest.Env,
	}

	installers, err := installer.McpForTarget(targetFlag)
	if err != nil {
		return err
	}

	type result struct {
		target     string
		configPath string
		err        error
	}

	var wg sync.WaitGroup
	results := make(chan result, len(installers))

	for _, inst := range installers {
		wg.Add(1)
		go func(inst installer.McpInstaller) {
			defer wg.Done()
			err := inst.Install(manifest.Name, cfg, scope, forceFlag)
			results <- result{
				target:     inst.Name(),
				configPath: inst.ConfigPath(scope),
				err:        err,
			}
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
			if errors.Is(res.err, installer.ErrAlreadyExists) {
				ui.Warn("%s: %s already configured in %s (use --force to overwrite)", res.target, mcpName, res.configPath)
				continue
			}
			ui.Error("%s: %v", res.target, res.err)
			hadError = true
			continue
		}

		ui.Success("%s: added %s to %s", res.target, mcpName, res.configPath)

		_ = store.Record(state.Installation{
			Plugin:      "mcp:" + mcpName,
			Version:     item.Version,
			Target:      res.target,
			Scope:       string(scope),
			ProjectPath: projectPath,
			InstalledAt: time.Now(),
			Files:       []string{res.configPath},
		})
	}

	envVars := requiredEnvVars(manifest.Env)
	if len(envVars) > 0 {
		fmt.Println()
		ui.Warn("%d env vars need values -- edit the config file(s) to set:", len(envVars))
		ui.Dim("%s", strings.Join(envVars, ", "))
	}

	if hadError {
		return fmt.Errorf("some installations failed")
	}

	fmt.Println()
	return nil
}

func runMcpRemove(cmd *cobra.Command, args []string) error {
	mcpName := args[0]
	scope := installer.ScopeLocal
	if globalFlag {
		scope = installer.ScopeGlobal
	}

	store := state.NewStore()
	installers, err := installer.McpForTarget(targetFlag)
	if err != nil {
		return err
	}

	ui.Header("Removing MCP: " + ui.Bold(mcpName))

	projectPath := ""
	if scope == installer.ScopeLocal {
		projectPath = installer.ProjectRoot()
	}

	removed := 0
	for _, inst := range installers {
		if err := inst.Remove(mcpName, scope); err != nil {
			ui.Error("%s: %v", inst.Name(), err)
			continue
		}

		_, _ = store.Remove("mcp:"+mcpName, inst.Name(), string(scope), projectPath)
		ui.Success("%s: removed %s from %s", inst.Name(), mcpName, inst.ConfigPath(scope))
		removed++
	}

	if removed == 0 {
		ui.Warn("no configurations found for %s", mcpName)
	}

	fmt.Println()
	return nil
}

func runMcpList(cmd *cobra.Command, args []string) error {
	client := registry.NewClient()
	marketplace, err := client.FetchMarketplace()
	if err != nil {
		return fmt.Errorf("fetching marketplace: %w", err)
	}

	var rows [][]string
	for _, p := range marketplace.Plugins {
		if p.Category != "mcps" {
			continue
		}

		manifest, err := client.FetchMcpManifest(p.Source)
		envCount := "none"
		if err == nil && len(manifest.Env) > 0 {
			envCount = fmt.Sprintf("%d required", len(manifest.Env))
		}

		rows = append(rows, []string{
			p.Name,
			p.Description,
			envCount,
		})
	}

	if len(rows) == 0 {
		ui.Warn("no MCPs available")
		return nil
	}

	ui.Header("Available MCPs")
	ui.Table(
		[]string{"NAME", "DESCRIPTION", "ENV VARS"},
		rows,
	)
	fmt.Println()
	return nil
}

func requiredEnvVars(env map[string]string) []string {
	var vars []string
	for k, v := range env {
		if v == "" {
			vars = append(vars, k)
		}
	}
	return vars
}

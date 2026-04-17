package common

import (
	"time"

	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/state"
)

// MarketplaceFetchedMsg is sent when marketplace data has been loaded.
type MarketplaceFetchedMsg struct {
	Marketplace *registry.Marketplace
	Err         error
}

// ManifestFetchedMsg is sent when a plugin manifest has been loaded.
type ManifestFetchedMsg struct {
	Manifest *registry.PluginManifest
	Err      error
}

// McpManifestFetchedMsg is sent when an MCP manifest has been loaded.
type McpManifestFetchedMsg struct {
	Manifest *registry.McpManifest
	Err      error
}

// StateFetchedMsg is sent when installation state has been loaded.
type StateFetchedMsg struct {
	Installations []state.Installation
	Err           error
}

// InstallCompleteMsg is sent when a plugin/MCP install finishes.
type InstallCompleteMsg struct {
	Plugin string
	Target string
	Err    error
}

// UninstallCompleteMsg is sent when an uninstall finishes.
type UninstallCompleteMsg struct {
	Plugin string
	Target string
	Err    error
}

// UpdateCompleteMsg is sent when an update finishes.
type UpdateCompleteMsg struct {
	Plugin string
	Target string
	Err    error
}

// ToastMsg triggers a transient notification.
type ToastMsg struct {
	Message string
	Level   ToastLevel
}

type ToastLevel int

const (
	ToastSuccess ToastLevel = iota
	ToastError
	ToastWarn
	ToastInfo
)

// ToastExpiredMsg dismisses the current toast.
type ToastExpiredMsg struct {
	ID int
}

// NavigateMsg requests navigation to a view, optionally carrying context.
type NavigateMsg struct {
	View    int
	Context any
}

// McpConfigLoadedMsg is sent when MCP config has been read from disk.
type McpConfigLoadedMsg struct {
	Name    string
	Target  string
	Command string
	Args    []string
	Env     map[string]string
	Err     error
}

// McpConfigSavedMsg is sent after saving MCP config changes.
type McpConfigSavedMsg struct {
	Name string
	Err  error
}

// TickMsg is a periodic timer tick for animations and auto-dismiss.
type TickMsg time.Time

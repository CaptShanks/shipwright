package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
)

type Toast struct {
	message string
	level   common.ToastLevel
	id      int
	visible bool
	counter int
}

func NewToast() Toast {
	return Toast{}
}

func (t *Toast) Show(msg string, level common.ToastLevel) tea.Cmd {
	t.counter++
	t.id = t.counter
	t.message = msg
	t.level = level
	t.visible = true

	id := t.id
	return tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
		return common.ToastExpiredMsg{ID: id}
	})
}

func (t *Toast) HandleExpiry(msg common.ToastExpiredMsg) {
	if msg.ID == t.id {
		t.visible = false
	}
}

func (t Toast) View(width int) string {
	if !t.visible {
		return ""
	}

	var style lipgloss.Style
	switch t.level {
	case common.ToastSuccess:
		style = lipgloss.NewStyle().Foreground(common.ColorSuccess)
	case common.ToastError:
		style = lipgloss.NewStyle().Foreground(common.ColorError)
	case common.ToastWarn:
		style = lipgloss.NewStyle().Foreground(common.ColorWarn)
	default:
		style = lipgloss.NewStyle().Foreground(common.ColorPrimary)
	}

	icons := map[common.ToastLevel]string{
		common.ToastSuccess: "✓ ",
		common.ToastError:   "✗ ",
		common.ToastWarn:    "! ",
		common.ToastInfo:    "→ ",
	}

	rendered := style.Render(icons[t.level] + t.message)
	return lipgloss.PlaceHorizontal(width, lipgloss.Right, rendered)
}

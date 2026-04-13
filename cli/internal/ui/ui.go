package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	boldStyle    = lipgloss.NewStyle().Bold(true)
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
)

func Success(msg string, args ...any) {
	fmt.Println(successStyle.Render("  ✓ " + fmt.Sprintf(msg, args...)))
}

func Error(msg string, args ...any) {
	fmt.Println(errorStyle.Render("  ✗ " + fmt.Sprintf(msg, args...)))
}

func Warn(msg string, args ...any) {
	fmt.Println(warnStyle.Render("  ! " + fmt.Sprintf(msg, args...)))
}

func Info(msg string, args ...any) {
	fmt.Println(infoStyle.Render("  → " + fmt.Sprintf(msg, args...)))
}

func Dim(msg string, args ...any) {
	fmt.Println(dimStyle.Render("    " + fmt.Sprintf(msg, args...)))
}

func Header(msg string) {
	fmt.Println()
	fmt.Println(headerStyle.Render(msg))
}

func Bold(msg string) string {
	return boldStyle.Render(msg)
}

// Table renders a simple table with headers and rows.
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, col := range row {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	headerLine := "  "
	separatorLine := "  "
	for i, h := range headers {
		headerLine += fmt.Sprintf("%-*s  ", widths[i], h)
		separatorLine += strings.Repeat("─", widths[i]) + "  "
	}
	fmt.Println(headerStyle.Render(headerLine))
	fmt.Println(dimStyle.Render(separatorLine))

	for _, row := range rows {
		line := "  "
		for i, col := range row {
			if i < len(widths) {
				line += fmt.Sprintf("%-*s  ", widths[i], col)
			}
		}
		fmt.Println(line)
	}
}

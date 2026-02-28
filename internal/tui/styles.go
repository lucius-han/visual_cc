package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorBlue   = lipgloss.Color("#5C9CF5")
	colorGreen  = lipgloss.Color("#5CB85C")
	colorRed    = lipgloss.Color("#D9534F")
	colorPurple = lipgloss.Color("#9B59B6")
	colorYellow = lipgloss.Color("#F0AD4E")
	colorGray   = lipgloss.Color("#666666")
	colorDim    = lipgloss.Color("#444444")
	colorBorder = lipgloss.Color("#333333")
	colorHeader = lipgloss.Color("#CCCCCC")
	colorTeal   = lipgloss.Color("#00B4D8")
	colorChild  = lipgloss.Color("#555555")

	stylePreTool    = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	stylePostOK     = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	stylePostErr    = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	styleStop       = lipgloss.NewStyle().Foreground(colorPurple).Bold(true)
	styleNotif      = lipgloss.NewStyle().Foreground(colorYellow)
	styleAgent      = lipgloss.NewStyle().Foreground(colorTeal).Bold(true)
	styleChildPrefix = lipgloss.NewStyle().Foreground(colorChild)
	styleChildEvent  = lipgloss.NewStyle().Foreground(colorChild)
	styleBadgeOK    = lipgloss.NewStyle().Foreground(colorGreen)
	styleBadgePend  = lipgloss.NewStyle().Foreground(colorGray)
	styleBadgeUnread = lipgloss.NewStyle().Foreground(colorYellow)

	styleTime   = lipgloss.NewStyle().Foreground(colorGray)
	styleDim    = lipgloss.NewStyle().Foreground(colorDim)
	styleIndent = lipgloss.NewStyle().Foreground(colorDim).PaddingLeft(12)

	styleDuration = lipgloss.NewStyle().Foreground(colorGray).Align(lipgloss.Right)

	stylePanelBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder).
				Padding(0, 1)

	styleHeader = lipgloss.NewStyle().
			Foreground(colorHeader).
			Bold(true)

	styleDivider = lipgloss.NewStyle().Foreground(colorBorder)

	styleHelp = lipgloss.NewStyle().Foreground(colorDim)
)

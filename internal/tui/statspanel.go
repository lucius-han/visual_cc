package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucius-han/visual_cc/internal/store"
)

const panelWidth = 22

// RenderStatsPanel renders the right-side statistics panel.
func RenderStatsPanel(stats store.Stats, height int) string {
	elapsed := time.Since(stats.StartTime)
	lines := []string{
		styleHeader.Render("통계"),
		styleDivider.Render(strings.Repeat("─", panelWidth-2)),
		fmt.Sprintf("⏱  %s", formatDuration(elapsed)),
		fmt.Sprintf("🔤 %s tokens", formatNumber(stats.TotalTokens)),
		fmt.Sprintf("💰 $%.4f", stats.TotalCostUSD),
		"",
		styleDim.Render("Tool 호출"),
		styleDivider.Render(strings.Repeat("─", panelWidth-2)),
	}

	type kv struct {
		k string
		v int
	}
	var sorted []kv
	for k, v := range stats.ToolCounts {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })

	maxCount := 1
	if len(sorted) > 0 {
		maxCount = sorted[0].v
	}

	for _, item := range sorted {
		bar := renderBar(item.v, maxCount, 6)
		lines = append(lines, fmt.Sprintf("%-6s %s %d", item.k, bar, item.v))
	}

	for len(lines) < height-4 {
		lines = append(lines, "")
	}

	lines = append(lines,
		styleDivider.Render(strings.Repeat("─", panelWidth-2)),
		styleHelp.Render("q quit  c clear"),
	)

	content := strings.Join(lines, "\n")
	return stylePanelBorder.Width(panelWidth).Render(content)
}

func renderBar(value, max, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}
	filled := (value * width) / max
	bar := stylePreTool.Render(strings.Repeat("█", filled)) +
		styleDim.Render(strings.Repeat("░", width-filled))
	return bar
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func formatNumber(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// StatsPanelWidth returns the total width of the stats panel including border.
func StatsPanelWidth() int {
	return panelWidth + 2
}

// suppress unused import warning - lipgloss is used via styles.go in same package
var _ = lipgloss.NewStyle()

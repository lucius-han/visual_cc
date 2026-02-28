package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucius-han/visual_cc/internal/event"
	"github.com/lucius-han/visual_cc/internal/store"
)

// NewEventMsg carries a new event to the TUI model.
type NewEventMsg event.Event

// Model is the bubbletea application model.
type Model struct {
	store      *store.Store
	viewport   viewport.Model
	width      int
	height     int
	autoScroll bool
	logBuf     strings.Builder
}

// NewModel creates a new Model backed by the given store.
func NewModel(s *store.Store) Model {
	return Model{
		store:      s,
		autoScroll: true,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "c":
			m.store.Reset()
			m.logBuf.Reset()
			m.viewport.SetContent("")
			m.autoScroll = true
			return m, nil
		case "n":
			m.store.MarkNotifsRead()
			return m, nil
		case "G":
			m.autoScroll = true
			m.viewport.GotoBottom()
		case "up", "k":
			m.autoScroll = false
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		panelW := StatsPanelWidth() + 2
		logW := m.width - panelW - 1
		if logW < 1 {
			logW = 1
		}
		headerH := 3
		footerH := 1
		vpHeight := m.height - headerH - footerH
		if vpHeight < 1 {
			vpHeight = 1
		}
		m.viewport = viewport.New(logW, vpHeight)
		m.viewport.SetContent(m.logBuf.String())
		if m.autoScroll {
			m.viewport.GotoBottom()
		}

	case NewEventMsg:
		e := event.Event(msg)
		m.store.Add(e)
		mainID := m.store.MainSessionID()
		isChild := mainID != "" && e.SessionID != "" && e.SessionID != mainID
		rendered := RenderEvent(e, isChild)
		if rendered != "" {
			m.logBuf.WriteString(rendered)
			m.viewport.SetContent(m.logBuf.String())
			if m.autoScroll {
				m.viewport.GotoBottom()
			}
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "초기화 중..."
	}

	header := renderHeader(m.width)
	divider := styleDivider.Render(strings.Repeat("━", m.width))

	statsH := m.height - 3 - 1
	if statsH < 1 {
		statsH = 1
	}
	stats := RenderStatsPanel(m.store.Stats(), statsH)

	logView := m.viewport.View()
	body := lipgloss.JoinHorizontal(lipgloss.Top, logView, " ", stats)

	return lipgloss.JoinVertical(lipgloss.Left, header, divider, body)
}

func renderHeader(width int) string {
	left := styleHeader.Render(" visual_cc")
	right := styleDim.Render("q quit  ↑↓ scroll  G bottom  c clear  n read ")
	space := strings.Repeat(" ", max(0, width-lipgloss.Width(left)-lipgloss.Width(right)))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, space, right)
}


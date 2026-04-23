package tui

import (
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/cemc-oper/hpc-model-go"
)

// viewState represents the current view state of the TUI.
type viewState int

const (
	listView   viewState = iota // job list view
	detailView                  // job detail view
)

// model is the main Bubble Tea model holding all TUI state.
type model struct {
	state        viewState  // current view (list or detail)
	list         list.Model // job list component
	detailItem   *jobItem   // currently selected job for detail view
	detailOffset int        // scroll offset in detail view
	lastUpdate   time.Time  // last data refresh timestamp
	err          error      // most recent fetch error
	width        int        // terminal width
	height       int        // terminal height
	columns      []ColDef   // list column definitions
	isDark       bool       // whether terminal has a dark background
	loading      bool       // whether data is being fetched
}

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#A0A0A0"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5050"))
	labelStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	loadingStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FACC15")).Background(lipgloss.Color("#713F12")).Padding(0, 1)
)

// initialModel creates the initial TUI model, configures the list component, and enters loading state.
func initialModel() model {
	const defaultWidth = 80
	const listHeight = 20

	columns := BuildColumns()
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	delegate := jobDelegate{columns: columns, isDark: isDark}
	l := list.New([]list.Item{}, delegate, defaultWidth, listHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	return model{
		state:   listView,
		list:    l,
		columns: columns,
		isDark:  isDark,
		loading: true,
	}
}

// Init is called when the TUI starts; triggers the first data fetch.
func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchJobsCmd(),
	)
}

// Update handles all input events and async messages to drive state transitions.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reserve 4 lines for title, summary/status, footer hint, and spacing.
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil

	case jobsFetchedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			m.lastUpdate = time.Now()
			var listItems []list.Item
			for _, ji := range msg.items {
				listItems = append(listItems, ji)
			}
			m.list.SetItems(listItems)

			hpcItems := make([]hpcmodel.Item, len(msg.items))
			for i, ji := range msg.items {
				hpcItems[i] = ji.item
			}
			widths := computeColWidths(m.columns, hpcItems)
			m.list.SetDelegate(jobDelegate{columns: m.columns, colWidths: widths, isDark: m.isDark})
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case listView:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "r":
				m.loading = true
				return m, fetchJobsCmd()
			case "enter", "right":
				if i, ok := m.list.SelectedItem().(jobItem); ok {
					m.detailItem = &i
					m.state = detailView
					m.detailOffset = 0
				}
				return m, nil
			case "pgup", "pgdown":
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(msg)
				return m, cmd
			}
		case detailView:
			switch msg.String() {
			case "q", "esc", "left", "ctrl+c":
				m.state = listView
				m.detailItem = nil
				m.detailOffset = 0
				return m, nil
			case "up":
				if m.detailOffset > 0 {
					m.detailOffset--
				}
				return m, nil
			case "down":
				m.detailOffset++
				return m, nil
			case "pgup":
				pageSize := m.height - 4
				if pageSize < 1 {
					pageSize = 1
				}
				m.detailOffset -= pageSize
				if m.detailOffset < 0 {
					m.detailOffset = 0
				}
				return m, nil
			case "pgdown":
				pageSize := m.height - 4
				if pageSize < 1 {
					pageSize = 1
				}
				m.detailOffset += pageSize
				return m, nil
			}
		}
	case tea.MouseMsg:
		if m.state == detailView {
			mouse := msg.Mouse()
			switch mouse.Button {
			case tea.MouseWheelUp:
				if m.detailOffset > 0 {
					m.detailOffset -= 3
					if m.detailOffset < 0 {
						m.detailOffset = 0
					}
				}
				return m, nil
			case tea.MouseWheelDown:
				m.detailOffset += 3
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the list or detail interface based on the current state and returns the final view configuration.
func (m model) View() tea.View {
	var v tea.View
	if m.state == detailView && m.detailItem != nil {
		v = m.detailView()
	} else {
		v = m.listView()
	}
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// TUICommand starts the Bubble Tea event loop and blocks until the user exits.
func TUICommand() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
	}
}

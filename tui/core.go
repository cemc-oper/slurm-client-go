package tui

import (
	"fmt"
	"os"
	"strings"
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
	state         viewState       // current view (list or detail)
	list          list.Model      // flat job list component
	treeList      list.Model      // tree view list component
	detailItem    *jobItem        // currently selected job for detail view
	detailOffset  int             // scroll offset in detail view
	lastUpdate    time.Time       // last data refresh timestamp
	err           error           // most recent fetch error
	width         int             // terminal width
	height        int             // terminal height
	columns       []ColDef        // list column definitions
	isDark        bool            // whether terminal has a dark background
	loading       bool            // whether data is being fetched
	displayMode   displayMode     // list or tree view
	groupBy       []string        // grouping levels for tree view
	treeCollapsed map[string]bool // group path key -> collapsed state (true = folded)
	allItems      []jobItem       // cached raw items for mode switching
	colWidths     []int           // cached column widths
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

	treeL := list.New([]list.Item{}, delegate, defaultWidth, listHeight)
	treeL.SetShowTitle(false)
	treeL.SetShowStatusBar(false)
	treeL.SetShowHelp(false)
	treeL.SetFilteringEnabled(false)

	return model{
		state:         listView,
		list:          l,
		treeList:      treeL,
		columns:       columns,
		isDark:        isDark,
		loading:       true,
		displayMode:   modeList,
		groupBy:       []string{"squeue.user"},
		treeCollapsed: make(map[string]bool),
	}
}

// rebuildList reconstructs both the flat list and tree list with current data.
func (m *model) rebuildList() {
	var listItems []list.Item
	for _, ji := range m.allItems {
		listItems = append(listItems, ji)
	}
	m.list.SetItems(listItems)
	m.list.SetDelegate(jobDelegate{columns: m.columns, colWidths: m.colWidths, isDark: m.isDark})

	treeItems := buildTreeItems(m.allItems, m.groupBy, m.treeCollapsed)
	var treeListItems []list.Item
	for _, ti := range treeItems {
		treeListItems = append(treeListItems, ti)
	}
	m.treeList.SetItems(treeListItems)
	m.treeList.SetDelegate(treeDelegate{
		jobDel: jobDelegate{columns: m.columns, colWidths: m.colWidths, isDark: m.isDark, indent: 0},
		isDark: m.isDark,
	})
}

// Init is called when the TUI starts; triggers the first data fetch.
func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchJobsCmd(),
	)
}

// handleListKey processes keyboard input in list view. Returns (cmd, true) if the key was handled.
func (m *model) handleListKey(key string) (tea.Cmd, bool) {
	switch key {
	case "q", "ctrl+c":
		return tea.Quit, true
	case "r":
		m.loading = true
		return fetchJobsCmd(), true
	case "t":
		if m.displayMode == modeList {
			m.displayMode = modeTree
		} else {
			m.displayMode = modeList
		}
		return nil, true
	case "c":
		currentIdx := 0
		for i, preset := range groupByPresets {
			if len(preset.levels) == len(m.groupBy) {
				match := true
				for j := range preset.levels {
					if preset.levels[j] != m.groupBy[j] {
						match = false
						break
					}
				}
				if match {
					currentIdx = i
					break
				}
			}
		}
		nextIdx := (currentIdx + 1) % len(groupByPresets)
		m.groupBy = append([]string(nil), groupByPresets[nextIdx].levels...)
		if m.displayMode == modeTree {
			m.treeCollapsed = make(map[string]bool)
		}
		m.rebuildList()
		return nil, true
	case "a":
		if m.displayMode == modeTree {
			treeItems := buildTreeItems(m.allItems, m.groupBy, m.treeCollapsed)
			groupPaths := make(map[string]struct{})
			for _, ti := range treeItems {
				if ti.itemType == treeItemGroup {
					groupPaths[ti.pathKey] = struct{}{}
				}
			}
			allCollapsed := true
			for path := range groupPaths {
				if !m.treeCollapsed[path] {
					allCollapsed = false
					break
				}
			}
			if allCollapsed {
				m.treeCollapsed = make(map[string]bool)
			} else {
				for path := range groupPaths {
					m.treeCollapsed[path] = true
				}
			}
			m.rebuildList()
			return nil, true
		}
	case "enter", "right":
		if m.displayMode == modeTree {
			if ti, ok := m.treeList.SelectedItem().(treeItem); ok {
				if ti.itemType == treeItemGroup {
					if ti.isExpanded {
						m.treeCollapsed[ti.pathKey] = true
					} else {
						delete(m.treeCollapsed, ti.pathKey)
					}
					m.rebuildList()
					return nil, true
				}
				m.detailItem = &ti.job
				m.state = detailView
				m.detailOffset = 0
				return nil, true
			}
		} else {
			if i, ok := m.list.SelectedItem().(jobItem); ok {
				m.detailItem = &i
				m.state = detailView
				m.detailOffset = 0
				return nil, true
			}
		}
	case "left":
		if m.displayMode == modeTree {
			if ti, ok := m.treeList.SelectedItem().(treeItem); ok {
				if ti.itemType == treeItemGroup && ti.isExpanded {
					m.treeCollapsed[ti.pathKey] = true
					m.rebuildList()
					return nil, true
				}
			}
		}
	}
	return nil, false
}

// handleDetailKey processes keyboard input in detail view. Returns (cmd, true) if the key was handled.
func (m *model) handleDetailKey(key string) (tea.Cmd, bool) {
	switch key {
	case "q", "esc", "left", "ctrl+c":
		m.state = listView
		m.detailItem = nil
		m.detailOffset = 0
		return nil, true
	case "up":
		if m.detailOffset > 0 {
			m.detailOffset--
		}
		return nil, true
	case "down":
		m.detailOffset++
		return nil, true
	case "pgup":
		pageSize := m.height - 4
		if pageSize < 1 {
			pageSize = 1
		}
		m.detailOffset -= pageSize
		if m.detailOffset < 0 {
			m.detailOffset = 0
		}
		return nil, true
	case "pgdown":
		pageSize := m.height - 4
		if pageSize < 1 {
			pageSize = 1
		}
		m.detailOffset += pageSize
		return nil, true
	}
	return nil, false
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
		m.treeList.SetWidth(msg.Width)
		m.treeList.SetHeight(msg.Height - 4)
		return m, nil

	case jobsFetchedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			m.lastUpdate = time.Now()
			m.allItems = msg.items
			hpcItems := make([]hpcmodel.Item, len(m.allItems))
			for i, ji := range m.allItems {
				hpcItems[i] = ji.item
			}
			m.colWidths = computeColWidths(m.columns, hpcItems)
			m.rebuildList()
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case listView:
			if cmd, handled := m.handleListKey(msg.String()); handled {
				return m, cmd
			}
		case detailView:
			if cmd, handled := m.handleDetailKey(msg.String()); handled {
				return m, cmd
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
	if m.displayMode == modeTree {
		m.treeList, cmd = m.treeList.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

// View renders the list or detail interface based on the current state and returns the final view configuration.
func (m model) View() tea.View {
	var v tea.View
	if m.state == detailView && m.detailItem != nil {
		v = m.detailView()
	} else if m.displayMode == modeTree {
		v = m.treeListView()
	} else {
		v = m.listView()
	}
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// renderJobList renders the common list frame for a given list component.
func (m model) renderJobList(l list.Model, title, hint string, summary string) tea.View {
	var b strings.Builder
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	if m.loading && len(l.Items()) == 0 {
		b.WriteString("\n")
		msg := loadingStyle.Render(" Fetching job data from Slurm... ")
		pad := (m.width - lipgloss.Width(msg)) / 2
		if pad < 0 {
			pad = 0
		}
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(msg)
		b.WriteString("\n")
	} else {
		if summary != "" {
			b.WriteString(summary)
			b.WriteString("\n")
		}
		b.WriteString(l.View())
		b.WriteString("\n")
		if m.loading {
			b.WriteString(loadingStyle.Render(" Loading jobs... "))
			b.WriteString(" ")
		}
	}
	if m.err != nil {
		b.WriteString(errStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString(" ")
	}
	if !m.lastUpdate.IsZero() {
		hint += fmt.Sprintf("  |  updated: %s", m.lastUpdate.Format("15:04:05"))
	}
	b.WriteString(infoStyle.Render(hint))
	return tea.NewView(b.String())
}

// TUICommand starts the Bubble Tea event loop and blocks until the user exits.
func TUICommand() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
	}
}

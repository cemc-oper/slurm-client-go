package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/slurm-client-go/common"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive TUI for Slurm jobs",
	Long:  "Launch an interactive terminal UI to browse and inspect Slurm jobs.",
	Run: func(cmd *cobra.Command, args []string) {
		TUICommand()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

// computeColWidths calculates the maximum rendered width for each column to align list items.
func computeColWidths(columns []colDef, items []hpcmodel.Item) []int {
	widths := make([]int, len(columns))
	for _, item := range items {
		for i, c := range columns {
			txt, _ := getProp(item, c.prop)
			w := lipgloss.Width(c.style.Render(txt))
			if w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

// ------------------------------------------------------------------
// Bubble Tea model
// ------------------------------------------------------------------

// viewState represents the current view state of the TUI.
type viewState int

const (
	listView   viewState = iota // job list view
	detailView                  // job detail view
)

// jobItem wraps hpcmodel.Item to implement the Bubble Tea list.Item interface for filtering.
type jobItem struct {
	item hpcmodel.Item
}

// FilterValue returns the text used for list filtering (job ID).
func (j jobItem) FilterValue() string {
	prop := j.item.GetProperty("squeue.job_id").(*hpcmodel.StringProperty)
	return prop.Text
}

// jobsFetchedMsg is sent when the async job fetch completes.
type jobsFetchedMsg struct {
	items []jobItem
	err   error
}

// fetchJobsCmd returns an async command that runs squeue and parses the result into jobItems.
func fetchJobsCmd() tea.Cmd {
	return func() tea.Msg {
		items, err := common.FetchSqueueItems()
		if err != nil {
			return jobsFetchedMsg{err: err}
		}
		var jobItems []jobItem
		for _, it := range items {
			jobItems = append(jobItems, jobItem{item: it})
		}
		return jobsFetchedMsg{items: jobItems}
	}
}

// jobDelegate implements the Bubble Tea list.ItemDelegate to render each job row.
type jobDelegate struct {
	columns   []colDef
	colWidths []int
	isDark    bool
}

func (d jobDelegate) Height() int                         { return 1 }
func (d jobDelegate) Spacing() int                        { return 0 }
func (d jobDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d jobDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	j, ok := listItem.(jobItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	selectedBG := lipgloss.Color("#1e293b")
	if !d.isDark {
		selectedBG = lipgloss.Color("#a1a1aa")
	}

	cells := make([]string, len(d.columns))
	for i, c := range d.columns {
		txt, _ := getProp(j.item, c.prop)
		style := c.style
		if isSelected {
			style = style.Background(selectedBG)
		}
		rendered := style.Render(txt)
		w := lipgloss.Width(rendered)
		pad := 0
		if i < len(d.columns)-1 && i < len(d.colWidths) {
			pad = d.colWidths[i] - w
			if pad < 0 {
				pad = 0
			}
		}
		if pad > 0 {
			cells[i] = style.Render(strings.Repeat(" ", pad) + txt)
		} else {
			cells[i] = rendered
		}
	}

	// Truncate the last column (command) if it would overflow.
	// The list component reserves 2 cells for the cursor/selection markers,
	// so the actual available width for content is m.Width()-2.
	lastIdx := len(cells) - 1
	usedWidth := 0
	for i := 0; i < lastIdx; i++ {
		usedWidth += lipgloss.Width(cells[i])
	}
	usedWidth += lastIdx               // spaces between columns
	avail := m.Width() - usedWidth - 2 // margin for cursor/selection markers
	if avail < 3 {
		avail = 3
	}
	if lipgloss.Width(cells[lastIdx]) > avail {
		cells[lastIdx] = ansi.Truncate(cells[lastIdx], avail, "...")
	}

	line := strings.Join(cells, " ")
	fmt.Fprint(w, line)
}

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
	columns      []colDef   // list column definitions
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

	columns := buildColumns()
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

// buildSummary counts total jobs, state distribution, and unique users for the status bar.
func buildSummary(items []list.Item) string {
	stateCounts := make(map[string]int)
	users := make(map[string]struct{})
	total := 0

	for _, it := range items {
		ji, ok := it.(jobItem)
		if !ok {
			continue
		}
		total++
		stateProp := ji.item.GetProperty("squeue.state").(*hpcmodel.StringProperty)
		stateCounts[stateProp.Text]++
		userProp := ji.item.GetProperty("squeue.user").(*hpcmodel.StringProperty)
		users[userProp.Text] = struct{}{}
	}

	if total == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Total: %d", total))

	// Show common states first if present; otherwise show whatever exists.
	preferredOrder := []string{"RUNNING", "PENDING", "COMPLETING", "COMPLETED", "CANCELLED", "FAILED", "TIMEOUT", "SUSPENDED"}
	shown := make(map[string]bool)
	for _, s := range preferredOrder {
		if c, ok := stateCounts[s]; ok {
			parts = append(parts, fmt.Sprintf("%s: %d", s, c))
			shown[s] = true
		}
	}
	for s, c := range stateCounts {
		if !shown[s] {
			parts = append(parts, fmt.Sprintf("%s: %d", s, c))
			shown[s] = true
		}
	}

	parts = append(parts, fmt.Sprintf("Users: %d", len(users)))
	return infoStyle.Render(strings.Join(parts, "  "))
}

// listView renders the job list view including title, loading indicator, summary, list body, and footer hints.
func (m model) listView() tea.View {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" Slurm Jobs "))
	b.WriteString("\n")

	// Show a friendly loading screen instead of the empty list's "No items".
	if m.loading && len(m.list.Items()) == 0 {
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
		if summary := buildSummary(m.list.Items()); summary != "" {
			b.WriteString(summary)
			b.WriteString("\n")
		}
		b.WriteString(m.list.View())
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

	hint := "[r] refresh  [enter/→] detail  [pgup/pgdown] page  [q] quit"
	if !m.lastUpdate.IsZero() {
		hint += fmt.Sprintf("  |  updated: %s", m.lastUpdate.Format("15:04:05"))
	}
	b.WriteString(infoStyle.Render(hint))
	return tea.NewView(b.String())
}

// detailView renders detailed information for a single job, supporting keyboard and mouse scrolling.
func (m model) detailView() tea.View {
	it := &m.detailItem.item

	getStr := func(id string) string {
		p := it.GetProperty(id)
		if p == nil {
			return "N/A"
		}
		if sp, ok := p.(*hpcmodel.StringProperty); ok {
			return sp.Text
		}
		if np, ok := p.(*hpcmodel.NumberProperty); ok {
			return np.Text
		}
		if dp, ok := p.(*hpcmodel.DateTimeProperty); ok {
			return dp.Text
		}
		if tp, ok := p.(*hpcmodel.TimeStringProperty); ok {
			return tp.Text
		}
		return "N/A"
	}

	section := func(name string) string {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(name)
	}

	lines := []string{
		titleStyle.Render(" Job Detail "),
		"",
		section("Basic Information"),
		fmt.Sprintf("  %s %s", labelStyle.Render("JOBID:"), getStr("squeue.job_id")),
		fmt.Sprintf("  %s %s", labelStyle.Render("NAME:"), getStr("squeue.name")),
		fmt.Sprintf("  %s %s", labelStyle.Render("USER:"), getStr("squeue.user")),
		fmt.Sprintf("  %s %s", labelStyle.Render("ACCOUNT:"), getStr("squeue.account")),
		fmt.Sprintf("  %s %s", labelStyle.Render("STATE:"), getStr("squeue.state")),
		fmt.Sprintf("  %s %s", labelStyle.Render("REASON:"), getStr("squeue.reason")),
		fmt.Sprintf("  %s %s", labelStyle.Render("PARTITION:"), getStr("squeue.partition")),
		fmt.Sprintf("  %s %s", labelStyle.Render("QOS:"), getStr("squeue.qos")),
		"",
		section("Resources"),
		fmt.Sprintf("  %s %s", labelStyle.Render("NODES:"), getStr("squeue.nodes")),
		fmt.Sprintf("  %s %s", labelStyle.Render("CPUS:"), getStr("squeue.cpus")),
		fmt.Sprintf("  %s %s", labelStyle.Render("MEMORY:"), getStr("squeue.min_memory")),
		fmt.Sprintf("  %s %s", labelStyle.Render("FEATURES:"), getStr("squeue.features")),
		fmt.Sprintf("  %s %s", labelStyle.Render("PRIORITY:"), getStr("squeue.priority")),
		"",
		section("Time"),
		fmt.Sprintf("  %s %s", labelStyle.Render("SUBMIT TIME:"), getStr("squeue.submit_time")),
		fmt.Sprintf("  %s %s", labelStyle.Render("START TIME:"), getStr("squeue.start_time")),
		fmt.Sprintf("  %s %s", labelStyle.Render("END TIME:"), getStr("squeue.end_time")),
		fmt.Sprintf("  %s %s", labelStyle.Render("TIME LIMIT:"), getStr("squeue.time_limit")),
		fmt.Sprintf("  %s %s", labelStyle.Render("RUN TIME:"), getStr("squeue.run_time")),
		fmt.Sprintf("  %s %s", labelStyle.Render("TIME LEFT:"), getStr("squeue.time_left")),
		"",
		section("Execution"),
		fmt.Sprintf("  %s %s", labelStyle.Render("WORK DIR:"), getStr("squeue.work_dir")),
		fmt.Sprintf("  %s %s", labelStyle.Render("NODELIST:"), getStr("squeue.nodelist")),
		fmt.Sprintf("  %s %s", labelStyle.Render("EXEC HOST:"), getStr("squeue.exec_host")),
		"",
		section("Command"),
		// Hard-wrap the command line so that scrolling knows the exact number
		// of lines. Relying on terminal soft-wrap would break offset math.
		wrapText(getStr("squeue.command"), m.width-8),
	}

	// Flatten hard-wrapped multi-line strings so every visual line has
	// exactly one entry in allLines. This is required for correct scroll
	// bounds: the terminal soft-wraps, but the TUI must see real line count.
	allLines := []string{}
	for _, line := range lines {
		allLines = append(allLines, strings.Split(line, "\n")...)
	}

	// Reserve 4 lines: top/bottom padding (2) + footer hint + spacing.
	contentHeight := m.height - 4
	if contentHeight < 1 {
		contentHeight = 1
	}
	maxOffset := len(allLines) - contentHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.detailOffset > maxOffset {
		m.detailOffset = maxOffset
	}
	end := m.detailOffset + contentHeight
	if end > len(allLines) {
		end = len(allLines)
	}
	visibleLines := allLines[m.detailOffset:end]

	hint := infoStyle.Render("[esc/←/q] back  [↑/↓/wheel] scroll  [pgup/pgdown] page  [ctrl+c] quit")
	visibleLines = append(visibleLines, "", hint)

	return tea.NewView(lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(visibleLines, "\n")))
}

// wrapText hard-wraps long text into multiple lines at a fixed width.
// This is necessary because the TUI scrolls by line count: if we relied
// on the terminal to soft-wrap, the program would not know how many
// visual lines the text actually occupies, breaking scroll bounds.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	var lines []string
	for len(text) > width {
		lines = append(lines, text[:width])
		text = text[width:]
	}
	lines = append(lines, text)
	return strings.Join(lines, "\n")
}

// TUICommand starts the Bubble Tea event loop and blocks until the user exits.
func TUICommand() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
	}
}

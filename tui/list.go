package tui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/slurm-client-go/common"
	"github.com/charmbracelet/x/ansi"
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
	columns   []ColDef
	colWidths []int
	isDark    bool
	indent    int // leading spaces for tree view job rows
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
		txt, _ := GetProp(j.item, c.Prop)
		style := c.Style
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
	usedWidth += lastIdx                          // spaces between columns
	avail := m.Width() - usedWidth - 2 - d.indent // margin for cursor/selection markers and indent
	if avail < 3 {
		avail = 3
	}
	if lipgloss.Width(cells[lastIdx]) > avail {
		cells[lastIdx] = ansi.Truncate(cells[lastIdx], avail, "...")
	}

	line := strings.Join(cells, " ")
	if d.indent > 0 {
		fmt.Fprint(w, strings.Repeat(" ", d.indent))
	}
	fmt.Fprint(w, line)
}

// computeColWidths calculates the maximum rendered width for each column to align list items.
func computeColWidths(columns []ColDef, items []hpcmodel.Item) []int {
	widths := make([]int, len(columns))
	for _, item := range items {
		for i, c := range columns {
			txt, _ := GetProp(item, c.Prop)
			w := lipgloss.Width(c.Style.Render(txt))
			if w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

// buildSummaryFromJobs counts total jobs, state distribution, and unique users from raw job items.
func buildSummaryFromJobs(jobs []jobItem) string {
	stateCounts := make(map[string]int)
	users := make(map[string]struct{})
	total := len(jobs)

	for _, ji := range jobs {
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

// buildSummary counts total jobs, state distribution, and unique users for the status bar.
func buildSummary(items []list.Item) string {
	var jobs []jobItem
	for _, it := range items {
		if ji, ok := it.(jobItem); ok {
			jobs = append(jobs, ji)
		}
	}
	return buildSummaryFromJobs(jobs)
}

// listView renders the job list view using the common list frame.
func (m model) listView() tea.View {
	summary := buildSummary(m.list.Items())
	hint := "[r] refresh  [t] tree  [enter/→] detail  [pgup/pgdown] page  [q] quit"
	return m.renderJobList(m.list, " Slurm Jobs ", hint, summary)
}

package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/cemc-oper/hpc-model-go"
)

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

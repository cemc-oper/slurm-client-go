package tui

import (
	"os"

	"charm.land/lipgloss/v2"
	"github.com/cemc-oper/hpc-model-go"
)

// ColDef defines the metadata for a table column: property ID and render style.
type ColDef struct {
	Prop  string
	Style lipgloss.Style
}

// BuildColumns builds the column definitions for TUI/query output with adaptive light/dark theme colors.
func BuildColumns() []ColDef {
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	adaptiveColor := lipgloss.LightDark(isDark)

	idStyle := lipgloss.NewStyle().Bold(true).
		Foreground(adaptiveColor(lipgloss.Color("#1B5E20"), lipgloss.Color("#81C784")))
	stateStyle := lipgloss.NewStyle().
		Foreground(adaptiveColor(lipgloss.Color("#F57F17"), lipgloss.Color("#FFD54F")))
	partitionStyle := lipgloss.NewStyle().
		Foreground(adaptiveColor(lipgloss.Color("#0D47A1"), lipgloss.Color("#64B5F6")))
	nodesStyle := lipgloss.NewStyle().
		Foreground(adaptiveColor(lipgloss.Color("#6A1B9A"), lipgloss.Color("#CE93D8")))
	userStyle := lipgloss.NewStyle().
		Foreground(adaptiveColor(lipgloss.Color("#006064"), lipgloss.Color("#80DEEA")))
	submitTimeStyle := lipgloss.NewStyle().
		Foreground(adaptiveColor(lipgloss.Color("#00695C"), lipgloss.Color("#4DB6AC")))
	runTimeStyle := lipgloss.NewStyle().
		Foreground(adaptiveColor(lipgloss.Color("#C62828"), lipgloss.Color("#EF9A9A")))

	return []ColDef{
		{"squeue.job_id", idStyle},
		{"squeue.state", stateStyle},
		{"squeue.partition", partitionStyle},
		{"squeue.nodes", nodesStyle},
		{"squeue.user", userStyle},
		{"squeue.submit_time", submitTimeStyle},
		{"squeue.run_time", runTimeStyle},
		{"squeue.command", lipgloss.NewStyle()},
	}
}

// GetProp extracts the text value and display name for a property from an Item, with type-specific assertions.
func GetProp(item hpcmodel.Item, propID string) (text, displayName string) {
	switch propID {
	case "squeue.job_id", "squeue.state", "squeue.partition", "squeue.user", "squeue.command":
		p := item.GetProperty(propID).(*hpcmodel.StringProperty)
		return p.Text, p.Category.DisplayName
	case "squeue.nodes":
		p := item.GetProperty(propID).(*hpcmodel.NumberProperty)
		return p.Text, p.Category.DisplayName
	case "squeue.submit_time":
		p := item.GetProperty(propID).(*hpcmodel.DateTimeProperty)
		return p.Text, p.Category.DisplayName
	case "squeue.run_time":
		p := item.GetProperty(propID).(*hpcmodel.TimeStringProperty)
		return p.Text, p.Category.DisplayName
	}
	return "", ""
}

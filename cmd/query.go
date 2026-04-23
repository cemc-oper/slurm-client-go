package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/slurm-client-go/common"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query jobs",
	Long:  "Query jobs in queue.",
	Run: func(cmd *cobra.Command, args []string) {
		QueryCommand(queryUsers, queryPartitions, querySortString, queryCommandPattern)
	},
}

var queryUsers []string
var queryPartitions []string
var querySortString string
var queryCommandPattern string

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().StringArrayVarP(
		&queryUsers, "user", "u", []string{}, "user")
	queryCmd.PersistentFlags().StringArrayVarP(
		&queryPartitions, "partition", "p", []string{}, "partition")
	queryCmd.PersistentFlags().StringVarP(
		&querySortString, "sort-keys", "s",
		"state:submit_time", "sort keys, split by :, such as status:query_date")
	queryCmd.PersistentFlags().StringVarP(
		&queryCommandPattern, "command-pattern", "c", "", "command pattern")
}

func QueryCommand(users []string, partitions []string, sortString string, commandPattern string) {
	params := []string{"-o", "%all"}

	filter := hpcmodel.Filter{}

	if len(users) > 0 {
		checker := hpcmodel.StringInValueChecker{
			ExpectedValues: users,
		}
		condition := hpcmodel.StringPropertyFilterCondition{
			ID:      "squeue.user",
			Checker: &checker,
		}
		filter.Conditions = append(filter.Conditions, &condition)
	}

	if len(partitions) > 0 {
		checker := hpcmodel.StringInValueChecker{
			ExpectedValues: partitions,
		}
		condition := hpcmodel.StringPropertyFilterCondition{
			ID:      "squeue.partition",
			Checker: &checker,
		}
		filter.Conditions = append(filter.Conditions, &condition)
	}

	if len(commandPattern) > 0 {
		checker := hpcmodel.StringContainChecker{
			ExpectedValue: commandPattern,
		}
		condition := hpcmodel.StringPropertyFilterCondition{
			ID:      "squeue.command",
			Checker: &checker,
		}
		filter.Conditions = append(filter.Conditions, &condition)
	}

	var sortKeys []string
	if len(sortString) > 0 {
		tokens := strings.Split(sortString, ":")
		for _, token := range tokens {
			sortKeys = append(sortKeys, "squeue."+token)
		}
	}

	lines, err := common.GetSqueueCommandResult(params)
	if err != nil {
		log.Fatalf("get query result error: %v", err)
	}

	model, err := common.GetSqueueQueryModel(lines)
	if err != nil {
		log.Fatalf("model build failed: %v", err)
	}

	targetItems := filter.Filter(model.Items)

	common.SortItems(targetItems, sortKeys)
	renderQueryTable(targetItems)
}

// colDef defines the metadata for a table column: property ID and render style.
type colDef struct {
	prop  string
	style lipgloss.Style
}

// buildColumns builds the column definitions for TUI/query output with adaptive light/dark theme colors.
func buildColumns() []colDef {
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

	return []colDef{
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

// getProp extracts the text value and display name for a property from an Item, with type-specific assertions.
func getProp(item hpcmodel.Item, propID string) (text, displayName string) {
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

func buildRow(columns []colDef, item hpcmodel.Item, header bool) []string {
	cells := make([]string, len(columns))
	for i, c := range columns {
		text, dn := getProp(item, c.prop)
		if header {
			cells[i] = c.style.Render(dn)
		} else {
			cells[i] = c.style.Render(text)
		}
	}
	return cells
}

func renderQueryTable(items []hpcmodel.Item) {
	columns := buildColumns()

	var rows [][]string
	if len(items) > 0 {
		rows = append(rows, buildRow(columns, items[0], true))
	}
	for _, item := range items {
		rows = append(rows, buildRow(columns, item, false))
	}

	widths := make([]int, len(columns))
	for _, row := range rows {
		for i, cell := range row {
			if w := lipgloss.Width(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}

	for _, row := range rows {
		for i, cell := range row {
			if i != len(columns)-1 {
				if pad := widths[i] - lipgloss.Width(cell); pad > 0 {
					cell = strings.Repeat(" ", pad) + cell
				}
			}
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(cell)
		}
		fmt.Println()
	}
}

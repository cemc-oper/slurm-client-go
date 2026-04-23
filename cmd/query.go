package cmd

import (
	"fmt"
	"log"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/slurm-client-go/common"
	"github.com/cemc-oper/slurm-client-go/tui"
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

func buildRow(columns []tui.ColDef, item hpcmodel.Item, header bool) []string {
	cells := make([]string, len(columns))
	for i, c := range columns {
		text, dn := tui.GetProp(item, c.Prop)
		if header {
			cells[i] = c.Style.Render(dn)
		} else {
			cells[i] = c.Style.Render(text)
		}
	}
	return cells
}

func renderQueryTable(items []hpcmodel.Item) {
	columns := tui.BuildColumns()

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

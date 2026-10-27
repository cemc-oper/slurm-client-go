package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	hpcmodel "github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/slurm-client-go/common"
	"github.com/cemc-oper/slurm-client-go/tui"
	"github.com/spf13/cobra"
)

type commandRunner interface {
	Squeue(params []string) ([]string, error)
	Sinfo(params []string) ([]string, error)
}

type slurmCommandRunner struct{}

func (slurmCommandRunner) Squeue(params []string) ([]string, error) {
	return common.GetSqueueCommandResult(params)
}

func (slurmCommandRunner) Sinfo(params []string) ([]string, error) {
	return common.GetSinfoCommandResult(params)
}

type squeueQueryOptions struct {
	users          []string
	partitions     []string
	jobs           []string
	sortString     string
	commandPattern string
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query jobs",
	Long:  "Query jobs in queue.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return queryCommand(cmd.OutOrStdout(), slurmCommandRunner{}, tui.BuildColumns(), queryUsers, queryPartitions, querySortString, queryCommandPattern)
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

func buildSqueueFilter(options squeueQueryOptions) hpcmodel.Filter {
	filter := hpcmodel.Filter{}
	if len(options.users) > 0 {
		checker := hpcmodel.StringInValueChecker{ExpectedValues: options.users}
		filter.Conditions = append(filter.Conditions, &hpcmodel.StringPropertyFilterCondition{ID: "squeue.user", Checker: &checker})
	}
	if len(options.partitions) > 0 {
		checker := hpcmodel.StringInValueChecker{ExpectedValues: options.partitions}
		filter.Conditions = append(filter.Conditions, &hpcmodel.StringPropertyFilterCondition{ID: "squeue.partition", Checker: &checker})
	}
	if len(options.jobs) > 0 {
		checker := hpcmodel.StringInValueChecker{ExpectedValues: options.jobs}
		filter.Conditions = append(filter.Conditions, &hpcmodel.StringPropertyFilterCondition{ID: "squeue.job_id", Checker: &checker})
	}
	if options.commandPattern != "" {
		checker := hpcmodel.StringContainChecker{ExpectedValue: options.commandPattern}
		filter.Conditions = append(filter.Conditions, &hpcmodel.StringPropertyFilterCondition{ID: "squeue.command", Checker: &checker})
	}
	return filter
}

func parseSortKeys(prefix, sortString string) []string {
	if sortString == "" {
		return nil
	}
	tokens := strings.Split(sortString, ":")
	sortKeys := make([]string, 0, len(tokens))
	for _, token := range tokens {
		sortKeys = append(sortKeys, prefix+"."+token)
	}
	return sortKeys
}

func sortItems(items []hpcmodel.Item, sortKeys []string) error {
	if len(sortKeys) == 0 || len(items) < 2 {
		return nil
	}
	for _, key := range sortKeys {
		property := items[0].GetProperty(key)
		if property == nil {
			return fmt.Errorf("sort key %q not found", key)
		}
		switch property.(type) {
		case *hpcmodel.StringProperty, *hpcmodel.NumberProperty, *hpcmodel.DateTimeProperty:
		default:
			return fmt.Errorf("sort key %q has unsupported property type %T", key, property)
		}
	}
	common.SortItems(items, sortKeys)
	return nil
}

func querySqueueLines(lines []string, options squeueQueryOptions) ([]hpcmodel.Item, error) {
	model, err := common.GetSqueueQueryModel(lines)
	if err != nil {
		return nil, fmt.Errorf("model build failed: %w", err)
	}
	filter := buildSqueueFilter(options)
	items := filter.Filter(model.Items)
	if err := sortItems(items, parseSortKeys("squeue", options.sortString)); err != nil {
		return nil, err
	}
	return items, nil
}

func fetchSqueueItems(runner commandRunner, options squeueQueryOptions) ([]hpcmodel.Item, error) {
	lines, err := runner.Squeue([]string{"-o", "%all"})
	if err != nil {
		return nil, fmt.Errorf("get query result error: %w", err)
	}
	return querySqueueLines(lines, options)
}

func QueryCommand(users []string, partitions []string, sortString string, commandPattern string) error {
	return queryCommand(os.Stdout, slurmCommandRunner{}, tui.BuildColumns(), users, partitions, sortString, commandPattern)
}

func queryCommand(output io.Writer, runner commandRunner, columns []tui.ColDef, users []string, partitions []string, sortString string, commandPattern string) error {
	items, err := fetchSqueueItems(runner, squeueQueryOptions{
		users:          users,
		partitions:     partitions,
		sortString:     sortString,
		commandPattern: commandPattern,
	})
	if err != nil {
		return err
	}
	if err := renderQueryTable(output, columns, items); err != nil {
		return fmt.Errorf("render query table: %w", err)
	}
	return nil
}

func buildRow(columns []tui.ColDef, item hpcmodel.Item, header bool) []string {
	cells := make([]string, len(columns))
	for i, c := range columns {
		text, displayName := tui.GetProp(item, c.Prop)
		if header {
			cells[i] = c.Style.Render(displayName)
		} else {
			cells[i] = c.Style.Render(text)
		}
	}
	return cells
}

func renderQueryTable(output io.Writer, columns []tui.ColDef, items []hpcmodel.Item) error {
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
			if width := lipgloss.Width(cell); width > widths[i] {
				widths[i] = width
			}
		}
	}

	for _, row := range rows {
		for i, cell := range row {
			if i != len(columns)-1 {
				if pad := widths[i] - lipgloss.Width(cell); pad > 0 {
					row[i] = strings.Repeat(" ", pad) + cell
				}
			}
		}
		if _, err := fmt.Fprintln(output, strings.Join(row, " ")); err != nil {
			return err
		}
	}
	return nil
}

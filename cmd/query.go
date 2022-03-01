package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/nwpc-oper/hpc-model-go"
	"github.com/nwpc-oper/slurm-client-go/common"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"text/tabwriter"
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
			ID:      "squeue.account",
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

	idColor := color.New(color.Bold).SprintFunc()
	partitionColor := color.New(color.FgBlue).SprintfFunc()
	accountColor := color.New(color.FgCyan).SprintfFunc()
	submitTimeColor := color.New(color.FgBlue).SprintfFunc()
	stateColor := color.New(color.FgYellow).SprintfFunc()

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 1, ' ', 0)

	targetItems := filter.Filter(model.Items)

	common.SortItems(targetItems, sortKeys)

	for _, item := range targetItems {
		jobID := item.GetProperty("squeue.job_id").(*hpcmodel.StringProperty)
		user := item.GetProperty("squeue.user").(*hpcmodel.StringProperty)
		partition := item.GetProperty("squeue.partition").(*hpcmodel.StringProperty)
		command := item.GetProperty("squeue.command").(*hpcmodel.StringProperty)
		state := item.GetProperty("squeue.state").(*hpcmodel.StringProperty)
		nodes := item.GetProperty("squeue.nodes").(*hpcmodel.NumberProperty)
		submitTime := item.GetProperty("squeue.submit_time").(*hpcmodel.DateTimeProperty)
		// workDir := item.GetProperty("squeue.work_dir").(*hpcmodel.StringProperty)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			idColor(jobID.Text),
			stateColor(state.Text),
			partitionColor(partition.Text),
			partitionColor(nodes.Text),
			accountColor(user.Text),
			submitTimeColor(submitTime.Text),
			command.Text)
	}
	w.Flush()

}

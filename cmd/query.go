package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/perillaroc/nwpc-hpc-model-go"
	"github.com/spf13/cobra"
	"log"
	"os"
	"slurm-client-go/common"
	"text/tabwriter"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query jobs",
	Long:  "Query jobs in queue.",
	Run: func(cmd *cobra.Command, args []string) {
		QueryCommand(users, partitions)
	},
}

var users []string
var partitions []string

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().StringArrayVarP(
		&users, "user", "u", []string{}, "user")
	queryCmd.PersistentFlags().StringArrayVarP(
		&partitions, "partition", "p", []string{}, "partition")
}

func QueryCommand(users []string, partitions []string) {
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

	for _, item := range targetItems {
		jobID := item.GetProperty("squeue.job_id").(*hpcmodel.StringProperty)
		account := item.GetProperty("squeue.account").(*hpcmodel.StringProperty)
		partition := item.GetProperty("squeue.partition").(*hpcmodel.StringProperty)
		command := item.GetProperty("squeue.command").(*hpcmodel.StringProperty)
		state := item.GetProperty("squeue.state").(*hpcmodel.StringProperty)
		submitTime := item.GetProperty("squeue.submit_time").(*hpcmodel.DateTimeProperty)
		// workDir := item.GetProperty("squeue.work_dir").(*hpcmodel.StringProperty)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			idColor(jobID.Text),
			stateColor(state.Text),
			partitionColor(partition.Text),
			accountColor(account.Text),
			submitTimeColor(submitTime.Text),
			command.Text)
	}
	w.Flush()

}

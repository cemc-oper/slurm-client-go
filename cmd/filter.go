package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/perillaroc/nwpc-hpc-model-go"
	"github.com/spf13/cobra"
	"log"
	"os"
	"slurm-client-go/common"
	"slurm-client-go/filters/long_time_job"
	"text/tabwriter"
)

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "job filter",
	Long:  "ItemFilter slurm jobs.",
	Run: func(cmd *cobra.Command, args []string) {
		FilterCommand()
	},
}

func init() {
	rootCmd.AddCommand(filterCmd)
}

func FilterCommand() {
	params := []string{"-o", "%all"}
	lines, err := common.GetSqueueCommandResult(params)
	if err != nil {
		log.Fatalf("get query result error: %v", err)
	}

	model, err := common.GetSqueueQueryModel(lines)
	if err != nil {
		log.Fatalf("model build failed: %v", err)
	}

	filter := long_time_job.CreateFilter()
	targetItems := filter.Apply(model.Items)

	boldColor := color.New(color.Bold).SprintFunc()
	partitionColor := color.New(color.FgBlue).SprintfFunc()
	accountColor := color.New(color.FgCyan).SprintfFunc()
	submitTimeColor := color.New(color.FgBlue).SprintfFunc()
	stateColor := color.New(color.FgYellow).SprintfFunc()

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 1, ' ', 0)

	common.SortItems(targetItems, []string{"squeue.state", "squeue.submit_time"})

	fmt.Fprintf(w, "%s\n", boldColor("long_time_job_filter:"))

	for _, item := range targetItems {
		jobID := item.GetProperty("squeue.job_id").(*hpcmodel.StringProperty)
		account := item.GetProperty("squeue.account").(*hpcmodel.StringProperty)
		partition := item.GetProperty("squeue.partition").(*hpcmodel.StringProperty)
		command := item.GetProperty("squeue.command").(*hpcmodel.StringProperty)
		state := item.GetProperty("squeue.state").(*hpcmodel.StringProperty)
		submitTime := item.GetProperty("squeue.submit_time").(*hpcmodel.DateTimeProperty)
		// workDir := item.GetProperty("squeue.work_dir").(*hpcmodel.StringProperty)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			boldColor(jobID.Text),
			stateColor(state.Text),
			partitionColor(partition.Text),
			accountColor(account.Text),
			submitTimeColor(submitTime.Text),
			command.Text)
	}
	fmt.Fprintf(w, "\n")
	w.Flush()
}

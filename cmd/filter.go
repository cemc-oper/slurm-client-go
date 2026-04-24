package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"charm.land/lipgloss/v2"
	"github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/slurm-client-go/common"
	"github.com/cemc-oper/slurm-client-go/filters/long_time_job"
	"github.com/spf13/cobra"
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

	idStyle := lipgloss.NewStyle().Bold(true)
	partitionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1976D2"))
	accountStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00838F"))
	submitTimeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1976D2"))
	stateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F9A825"))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 1, ' ', 0)

	common.SortItems(targetItems, []string{"squeue.state", "squeue.submit_time"})

	fmt.Fprintf(w, "%s\n", idStyle.Render("long_time_job_filter (experiment):"))

	for _, item := range targetItems {
		jobID := item.GetProperty("squeue.job_id").(*hpcmodel.StringProperty)
		account := item.GetProperty("squeue.account").(*hpcmodel.StringProperty)
		partition := item.GetProperty("squeue.partition").(*hpcmodel.StringProperty)
		command := item.GetProperty("squeue.command").(*hpcmodel.StringProperty)
		state := item.GetProperty("squeue.state").(*hpcmodel.StringProperty)
		submitTime := item.GetProperty("squeue.submit_time").(*hpcmodel.DateTimeProperty)
		// workDir := item.GetProperty("squeue.work_dir").(*hpcmodel.StringProperty)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			idStyle.Render(jobID.Text),
			stateStyle.Render(state.Text),
			partitionStyle.Render(partition.Text),
			accountStyle.Render(account.Text),
			submitTimeStyle.Render(submitTime.Text),
			command.Text)
	}
	fmt.Fprintf(w, "\n")
	w.Flush()
}

package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"charm.land/lipgloss/v2"
	hpcmodel "github.com/cemc-oper/hpc-model-go"
	"github.com/spf13/cobra"
)

var detailCmd = &cobra.Command{
	Use:   "detail",
	Short: "Query jobs in detail",
	Long:  "Query jobs in detail.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return detailCommand(cmd.OutOrStdout(), slurmCommandRunner{}, detailUsers, detailPartitions, detailSortString, detailCommandPattern)
	},
}

var detailUsers []string
var detailPartitions []string
var detailSortString string
var detailCommandPattern string

func init() {
	rootCmd.AddCommand(detailCmd)
	detailCmd.PersistentFlags().StringArrayVarP(
		&detailUsers, "user", "u", []string{}, "user")
	detailCmd.PersistentFlags().StringArrayVarP(
		&detailPartitions, "partition", "p", []string{}, "partition")
	detailCmd.PersistentFlags().StringVarP(
		&detailSortString, "sort-keys", "s",
		"state:submit_time", "sort keys, split by :, such as status:query_date")
	detailCmd.PersistentFlags().StringVarP(
		&detailCommandPattern, "command-pattern", "c", "", "command pattern")
}

func DetailCommand(users []string, partitions []string, sortString string, commandPattern string) error {
	return detailCommand(os.Stdout, slurmCommandRunner{}, users, partitions, sortString, commandPattern)
}

func detailCommand(output io.Writer, runner commandRunner, users []string, partitions []string, sortString string, commandPattern string) error {
	items, err := fetchSqueueItems(runner, squeueQueryOptions{
		users:          users,
		partitions:     partitions,
		sortString:     sortString,
		commandPattern: commandPattern,
	})
	if err != nil {
		return err
	}
	if err := renderDetailTable(output, items); err != nil {
		return fmt.Errorf("render job detail: %w", err)
	}
	return nil
}

func renderDetailTable(output io.Writer, items []hpcmodel.Item) error {
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AD1457"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00838F"))
	idStyle := lipgloss.NewStyle().Bold(true)
	stateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F9A825"))

	getStr := func(item hpcmodel.Item, id string) string {
		property := item.GetProperty(id)
		if property == nil {
			return "N/A"
		}
		switch value := property.(type) {
		case *hpcmodel.StringProperty:
			return value.Text
		case *hpcmodel.NumberProperty:
			return value.Text
		case *hpcmodel.DateTimeProperty:
			return value.Text
		case *hpcmodel.TimeStringProperty:
			return value.Text
		default:
			return "N/A"
		}
	}

	writer := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
	for i, item := range items {
		if i > 0 {
			fmt.Fprintln(writer, "────────────────────────────────────────────────────────────")
			fmt.Fprintln(writer)
		}

		fmt.Fprintf(writer, "%s\n", sectionStyle.Render(fmt.Sprintf("=== Job Detail (%s) ===", idStyle.Render(getStr(item, "squeue.job_id")))))
		fmt.Fprintln(writer)

		fmt.Fprintf(writer, "%s\n", sectionStyle.Render("Basic Information"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("JOBID:"), idStyle.Render(getStr(item, "squeue.job_id")))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("NAME:"), getStr(item, "squeue.name"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("USER:"), getStr(item, "squeue.user"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("ACCOUNT:"), getStr(item, "squeue.account"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("STATE:"), stateStyle.Render(getStr(item, "squeue.state")))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("REASON:"), getStr(item, "squeue.reason"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("PARTITION:"), getStr(item, "squeue.partition"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("QOS:"), getStr(item, "squeue.qos"))
		fmt.Fprintln(writer)

		fmt.Fprintf(writer, "%s\n", sectionStyle.Render("Resources"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("NODES:"), getStr(item, "squeue.nodes"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("CPUS:"), getStr(item, "squeue.cpus"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("MEMORY:"), getStr(item, "squeue.min_memory"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("FEATURES:"), getStr(item, "squeue.features"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("PRIORITY:"), getStr(item, "squeue.priority"))
		fmt.Fprintln(writer)

		fmt.Fprintf(writer, "%s\n", sectionStyle.Render("Time"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("SUBMIT TIME:"), getStr(item, "squeue.submit_time"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("START TIME:"), getStr(item, "squeue.start_time"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("END TIME:"), getStr(item, "squeue.end_time"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("TIME LIMIT:"), getStr(item, "squeue.time_limit"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("RUN TIME:"), getStr(item, "squeue.run_time"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("TIME LEFT:"), getStr(item, "squeue.time_left"))
		fmt.Fprintln(writer)

		fmt.Fprintf(writer, "%s\n", sectionStyle.Render("Execution"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("WORK DIR:"), getStr(item, "squeue.work_dir"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("NODELIST:"), getStr(item, "squeue.nodelist"))
		fmt.Fprintf(writer, "  %s\t%s\n", labelStyle.Render("EXEC HOST:"), getStr(item, "squeue.exec_host"))
		fmt.Fprintln(writer)

		fmt.Fprintf(writer, "%s\n", sectionStyle.Render("Command"))
		fmt.Fprintf(writer, "  %s\n", getStr(item, "squeue.command"))
	}
	return writer.Flush()
}

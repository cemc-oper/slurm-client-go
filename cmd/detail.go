package cmd

import (
	"charm.land/lipgloss/v2"
	"fmt"
	"github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/slurm-client-go/common"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"text/tabwriter"
)

var detailCmd = &cobra.Command{
	Use:   "detail",
	Short: "Query jobs in detail",
	Long:  "Query jobs in detail.",
	Run: func(cmd *cobra.Command, args []string) {
		DetailCommand(detailUsers, detailPartitions, detailSortString, detailCommandPattern)
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

func DetailCommand(users []string, partitions []string, sortString string, commandPattern string) {
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

	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AD1457"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00838F"))
	idStyle := lipgloss.NewStyle().Bold(true)
	stateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F9A825"))

	getStr := func(item hpcmodel.Item, id string) string {
		p := item.GetProperty(id)
		if p == nil {
			return "N/A"
		}
		switch v := p.(type) {
		case *hpcmodel.StringProperty:
			return v.Text
		case *hpcmodel.NumberProperty:
			return v.Text
		case *hpcmodel.DateTimeProperty:
			return v.Text
		case *hpcmodel.TimeStringProperty:
			return v.Text
		default:
			return "N/A"
		}
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 2, ' ', 0)

	targetItems := filter.Filter(model.Items)

	common.SortItems(targetItems, sortKeys)

	for i, item := range targetItems {
		if i > 0 {
			fmt.Fprintln(w, "────────────────────────────────────────────────────────────")
			fmt.Fprintln(w)
		}

		fmt.Fprintf(w, "%s\n", sectionStyle.Render(fmt.Sprintf("=== Job Detail (%s) ===", idStyle.Render(getStr(item, "squeue.job_id")))))
		fmt.Fprintln(w)

		fmt.Fprintf(w, "%s\n", sectionStyle.Render("Basic Information"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("JOBID:"), idStyle.Render(getStr(item, "squeue.job_id")))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("NAME:"), getStr(item, "squeue.name"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("USER:"), getStr(item, "squeue.user"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("ACCOUNT:"), getStr(item, "squeue.account"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("STATE:"), stateStyle.Render(getStr(item, "squeue.state")))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("REASON:"), getStr(item, "squeue.reason"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("PARTITION:"), getStr(item, "squeue.partition"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("QOS:"), getStr(item, "squeue.qos"))
		fmt.Fprintln(w)

		fmt.Fprintf(w, "%s\n", sectionStyle.Render("Resources"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("NODES:"), getStr(item, "squeue.nodes"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("CPUS:"), getStr(item, "squeue.cpus"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("MEMORY:"), getStr(item, "squeue.min_memory"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("FEATURES:"), getStr(item, "squeue.features"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("PRIORITY:"), getStr(item, "squeue.priority"))
		fmt.Fprintln(w)

		fmt.Fprintf(w, "%s\n", sectionStyle.Render("Time"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("SUBMIT TIME:"), getStr(item, "squeue.submit_time"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("START TIME:"), getStr(item, "squeue.start_time"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("END TIME:"), getStr(item, "squeue.end_time"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("TIME LIMIT:"), getStr(item, "squeue.time_limit"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("RUN TIME:"), getStr(item, "squeue.run_time"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("TIME LEFT:"), getStr(item, "squeue.time_left"))
		fmt.Fprintln(w)

		fmt.Fprintf(w, "%s\n", sectionStyle.Render("Execution"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("WORK DIR:"), getStr(item, "squeue.work_dir"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("NODELIST:"), getStr(item, "squeue.nodelist"))
		fmt.Fprintf(w, "  %s\t%s\n", labelStyle.Render("EXEC HOST:"), getStr(item, "squeue.exec_host"))
		fmt.Fprintln(w)

		fmt.Fprintf(w, "%s\n", sectionStyle.Render("Command"))
		fmt.Fprintf(w, "  %s\n", getStr(item, "squeue.command"))
	}
	w.Flush()

}

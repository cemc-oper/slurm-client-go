package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"charm.land/lipgloss/v2"
	hpcmodel "github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/slurm-client-go/common"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "query partition info",
	Long:  "Show partition info.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return infoCommand(cmd.OutOrStdout(), slurmCommandRunner{}, infoSortString)
	},
}

var infoSortString string

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.PersistentFlags().StringVarP(
		&infoSortString, "sort-keys", "s",
		"partition", "sort keys, split by :, such as partition")
}

func querySinfoLines(lines []string, sortString string) ([]hpcmodel.Item, error) {
	model, err := common.GetSinfoQueryModel(lines)
	if err != nil {
		return nil, fmt.Errorf("model build failed: %w", err)
	}
	items := model.Items
	if err := sortItems(items, parseSortKeys("sinfo", sortString)); err != nil {
		return nil, err
	}
	return items, nil
}

func InfoCommand(sortString string) error {
	return infoCommand(os.Stdout, slurmCommandRunner{}, sortString)
}

func infoCommand(output io.Writer, runner commandRunner, sortString string) error {
	lines, err := runner.Sinfo([]string{"-o", "%20P %.5a %.20F %.30C"})
	if err != nil {
		return fmt.Errorf("get query result error: %w", err)
	}
	items, err := querySinfoLines(lines, sortString)
	if err != nil {
		return err
	}
	if err := renderInfoTable(output, items); err != nil {
		return fmt.Errorf("render partition info: %w", err)
	}
	return nil
}

func renderInfoTable(output io.Writer, items []hpcmodel.Item) error {
	writer := tabwriter.NewWriter(output, 0, 0, 1, ' ', 0)
	partitionStyle := lipgloss.NewStyle().Bold(true)
	availStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1976D2"))
	nodesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00838F"))
	cpusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B1FA2"))

	for _, item := range items {
		partition := item.GetProperty("sinfo.partition").(*hpcmodel.StringProperty)
		avail := item.GetProperty("sinfo.avail").(*hpcmodel.StringProperty)
		nodes := item.GetProperty("sinfo.nodes").(*hpcmodel.StringProperty)
		cpus := item.GetProperty("sinfo.cpus").(*hpcmodel.StringProperty)
		if _, err := fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			partitionStyle.Render(partition.Text),
			availStyle.Render(avail.Text),
			nodesStyle.Render(nodes.Text),
			cpusStyle.Render(cpus.Text)); err != nil {
			return err
		}
	}
	return writer.Flush()
}

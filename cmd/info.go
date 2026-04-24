package cmd

import (
	"fmt"
	"github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/slurm-client-go/common"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"charm.land/lipgloss/v2"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "query partition info",
	Long:  "Show partition info.",
	Run: func(cmd *cobra.Command, args []string) {
		InfoCommand(infoSortString)
	},
}

var infoSortString string

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.PersistentFlags().StringVarP(
		&infoSortString, "sort-keys", "s",
		"partition", "sort keys, split by :, such as partition")
}

func InfoCommand(sortString string) {
	params := []string{"-o", "%20P %.5a %.20F %.30C"}

	var sortKeys []string
	if len(sortString) > 0 {
		tokens := strings.Split(sortString, ":")
		for _, token := range tokens {
			sortKeys = append(sortKeys, "sinfo."+token)
		}
	}

	lines, err := common.GetSinfoCommandResult(params)
	if err != nil {
		log.Fatalf("get query result error: %v", err)
	}

	model, err := common.GetSinfoQueryModel(lines)
	if err != nil {
		log.Fatalf("model build failed: %v", err)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 1, ' ', 0)

	partitionStyle := lipgloss.NewStyle().Bold(true)
	availStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1976D2"))
	nodesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00838F"))
	cpusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B1FA2"))

	targetItems := model.Items

	common.SortItems(targetItems, sortKeys)

	for _, item := range targetItems {
		partition := item.GetProperty("sinfo.partition").(*hpcmodel.StringProperty)
		avail := item.GetProperty("sinfo.avail").(*hpcmodel.StringProperty)
		nodes := item.GetProperty("sinfo.nodes").(*hpcmodel.StringProperty)
		cpus := item.GetProperty("sinfo.cpus").(*hpcmodel.StringProperty)
		// workDir := item.GetProperty("squeue.work_dir").(*hpcmodel.StringProperty)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			partitionStyle.Render(partition.Text),
			availStyle.Render(avail.Text),
			nodesStyle.Render(nodes.Text),
			cpusStyle.Render(cpus.Text))
	}
	w.Flush()
}

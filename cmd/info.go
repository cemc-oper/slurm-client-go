package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/perillaroc/nwpc-hpc-model-go"
	"github.com/spf13/cobra"
	"log"
	"os"
	"slurm-client-go/common"
	"strings"
	"text/tabwriter"
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

	partitionColor := color.New(color.Bold).SprintFunc()
	availColor := color.New(color.FgBlue).SprintfFunc()
	nodesColor := color.New(color.FgCyan).SprintfFunc()
	cpusColor := color.New(color.FgMagenta).SprintfFunc()

	targetItems := model.Items

	common.SortItems(targetItems, sortKeys)

	for _, item := range targetItems {
		partition := item.GetProperty("sinfo.partition").(*hpcmodel.StringProperty)
		avail := item.GetProperty("sinfo.avail").(*hpcmodel.StringProperty)
		nodes := item.GetProperty("sinfo.nodes").(*hpcmodel.StringProperty)
		cpus := item.GetProperty("sinfo.cpus").(*hpcmodel.StringProperty)
		// workDir := item.GetProperty("squeue.work_dir").(*hpcmodel.StringProperty)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			partitionColor(partition.Text),
			availColor(avail.Text),
			nodesColor(nodes.Text),
			cpusColor(cpus.Text))
	}
	w.Flush()
}

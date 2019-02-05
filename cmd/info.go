package cmd

import (
	"bytes"
	"fmt"
	"github.com/perillaroc/nwpc-hpc-model-go"
	"github.com/perillaroc/nwpc-hpc-model-go/slurm"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "query partition info",
	Long:  "Show partition info.",
	Run: func(cmd *cobra.Command, args []string) {
		InfoCommand()
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func InfoCommand() {
	params := []string{"-o", "%20P %.5a %.20F %.30C"}

	lines, err := getInfoResult(params)
	if err != nil {
		log.Fatalf("get query result error: %v", err)
	}

	model, err := GetQueryModel(lines)
	if err != nil {
		log.Fatalf("model build failed: %v", err)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 1, ' ', 0)

	for _, item := range model.Items {
		partition := item.GetProperty("sinfo.partition").(*hpcmodel.StringProperty)
		avail := item.GetProperty("sinfo.avail").(*hpcmodel.StringProperty)
		nodes := item.GetProperty("sinfo.nodes").(*hpcmodel.StringProperty)
		cpus := item.GetProperty("sinfo.cpus").(*hpcmodel.StringProperty)
		// workDir := item.GetProperty("squeue.work_dir").(*hpcmodel.StringProperty)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			partition.Text, avail.Text, nodes.Text, cpus.Text)
	}
	w.Flush()
}

func GetQueryModel(lines []string) (*slurm.Model, error) {
	categoryList := buildInfoCategoryList()
	model, err := slurm.BuildModel(lines, categoryList, " ")
	return model, err
}

func getInfoResult(params []string) ([]string, error) {
	cmd := exec.Command("sinfo", params...)
	//fmt.Println(cmd.Args)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("command ran error: %v", err)
	}
	s := out.String()
	lines := strings.Split(s, "\n")
	return lines, nil
}

func buildInfoCategoryList() slurm.QueryCategoryList {
	categoryList := slurm.QueryCategoryList{
		QueryCategoryList: hpcmodel.QueryCategoryList{
			CategoryList: []*hpcmodel.QueryCategory{
				{
					ID:                      "sinfo.partition",
					DisplayName:             "Partition",
					Label:                   "PARTITION",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "sinfo.avail",
					DisplayName:             "Avail",
					Label:                   "AVAIL",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "sinfo.nodes",
					DisplayName:             "Nodes(A/I/O/T)",
					Label:                   "NODES(A/I/O/T)",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "sinfo.cpus",
					DisplayName:             "CPUs(A/I/O/T)",
					Label:                   "CPUS(A/I/O/T)",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
			},
		},
	}
	return categoryList
}

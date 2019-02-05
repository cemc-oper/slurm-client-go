package cmd

import (
	"bytes"
	"fmt"
	"github.com/perillaroc/nwpc-hpc-model-go"
	"github.com/perillaroc/nwpc-hpc-model-go/slurm"
	"github.com/spf13/cobra"
	"log"
	"os/exec"
	"strings"
)

var users []string

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().StringArrayVarP(&users, "user", "u", []string{}, "user")

}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query jobs",
	Long:  "Query jobs in queue.",
	Run: func(cmd *cobra.Command, args []string) {
		QueryCommand(users)
	},
}

func QueryCommand(users []string) {
	params := []string{"-o %all"}
	for _, user := range users {
		params = append(params, "-u", user)
	}

	lines, err := getQueryResult(params)
	if err != nil {
		log.Fatalf("get query result error: %v", err)
	}

	categoryList := buildCategoryList()

	model, err := slurm.BuildModel(lines, categoryList, "|")
	if err != nil {
		log.Fatalf("model build failed: %v", err)
	}

	for _, item := range model.Items {
		jobID := item.GetProperty("squeue.job_id").(*hpcmodel.StringProperty)
		account := item.GetProperty("squeue.account").(*hpcmodel.StringProperty)
		partition := item.GetProperty("squeue.partition").(*hpcmodel.StringProperty)
		command := item.GetProperty("squeue.command").(*hpcmodel.StringProperty)
		state := item.GetProperty("squeue.state").(*hpcmodel.StringProperty)
		submitTime := item.GetProperty("squeue.submit_time").(*hpcmodel.DateTimeProperty)
		// workDir := item.GetProperty("squeue.work_dir").(*hpcmodel.StringProperty)
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n",
			jobID.Text, state.Text, partition.Text, account.Text, submitTime.Text, command.Text)
	}
}

func getQueryResult(params []string) ([]string, error) {
	cmd := exec.Command("squeue", params...)
	fmt.Println(cmd.Args)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("command ran error: %v", err)
	}
	s := out.String()
	fmt.Println(s)
	lines := strings.Split(s, "\n")
	return lines, nil
}

func buildCategoryList() slurm.QueryCategoryList {
	categoryList := slurm.QueryCategoryList{
		QueryCategoryList: hpcmodel.QueryCategoryList{
			CategoryList: []*hpcmodel.QueryCategory{
				{
					ID:                      "squeue.job_id",
					DisplayName:             "JOB ID",
					Label:                   "JOBID",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.account",
					DisplayName:             "account",
					Label:                   "ACCOUNT",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.partition",
					DisplayName:             "Partition",
					Label:                   "PARTITION",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.command",
					DisplayName:             "Command",
					Label:                   "COMMAND",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.state",
					DisplayName:             "State",
					Label:                   "STATE",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.submit_time",
					DisplayName:             "Submit Time",
					Label:                   "SUBMIT_TIME",
					PropertyClass:           "DateTimeProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.work_dir",
					DisplayName:             "Work Dir",
					Label:                   "WORK_DIR",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
			},
		},
	}
	return categoryList
}

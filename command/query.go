package command

import (
	"bytes"
	"fmt"
	"github.com/perillaroc/nwpc-hpc-model-go"
	"github.com/perillaroc/nwpc-hpc-model-go/slurm"
	"log"
	"os/exec"
	"strings"
)

func getQueryResult() ([]string, error) {
	cmd := exec.Command("squeue", "-o %all")
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

func buildCategoryList() slurm.QueryCategoryList {
	categoryList := slurm.QueryCategoryList{
		QueryCategoryList: hpcmodel.QueryCategoryList{
			CategoryList: []*hpcmodel.QueryCategory{
				{
					ID:                      "squeue.account",
					DisplayName:             "account",
					Label:                   "ACCOUNT",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.job_id",
					DisplayName:             "JOB ID",
					Label:                   "JOBID",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
			},
		},
	}
	return categoryList
}

func QueryCommand() {
	lines, err := getQueryResult()
	if err != nil {
		log.Fatalf("get query result error: %v", err)
	}

	categoryList := buildCategoryList()

	model, err := slurm.BuildModel(lines, categoryList, "|")
	if err != nil {
		log.Fatalf("model build failed: %v", err)
	}

	for _, item := range model.Items {
		jobIDProp := item.GetProperty("squeue.job_id").(*hpcmodel.StringProperty)
		accountProp := item.GetProperty("squeue.account").(*hpcmodel.StringProperty)
		fmt.Printf("%s\t%s\n", jobIDProp.Text, accountProp.Text)
	}

}

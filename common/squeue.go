package common

import (
	"bytes"
	"fmt"
	"github.com/nwpc-oper/hpc-model-go"
	"github.com/nwpc-oper/hpc-model-go/slurm"
	"os/exec"
	"strings"
)

func GetSqueueQueryModel(lines []string) (*slurm.Model, error) {
	categoryList := BuildSqueueCategoryList()
	model, err := slurm.BuildModel(lines, categoryList, "|")
	return model, err
}

func GetSqueueCommandResult(params []string) ([]string, error) {
	cmd := exec.Command("squeue", params...)
	//fmt.Println(cmd.Args)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("command ran error: %v", err)
	}
	s := out.String()
	//fmt.Println(s)
	lines := strings.Split(s, "\n")
	return lines, nil
}

func BuildSqueueCategoryList() slurm.QueryCategoryList {
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
				{
					ID:                      "squeue.cpus",
					DisplayName:             "CPUs",
					Label:                   "CPUS",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.nodes",
					DisplayName:             "NODEs",
					Label:                   "NODES",
					PropertyClass:           "NumberProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
			},
		},
	}
	return categoryList
}

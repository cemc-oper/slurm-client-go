package common

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/hpc-model-go/slurm"
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
		return nil, fmt.Errorf("command run error: %v", err)
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
					ID:                      "squeue.name",
					DisplayName:             "Name",
					Label:                   "NAME",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.account",
					DisplayName:             "Account",
					Label:                   "ACCOUNT",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.user",
					DisplayName:             "User",
					Label:                   "USER",
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
					ID:                      "squeue.qos",
					DisplayName:             "QoS",
					Label:                   "QOS",
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
					ID:                      "squeue.reason",
					DisplayName:             "Reason",
					Label:                   "REASON",
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
					ID:                      "squeue.submit_time",
					DisplayName:             "Submit Time",
					Label:                   "SUBMIT_TIME",
					PropertyClass:           "DateTimeProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.start_time",
					DisplayName:             "Start Time",
					Label:                   "START_TIME",
					PropertyClass:           "DateTimeProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.end_time",
					DisplayName:             "End Time",
					Label:                   "END_TIME",
					PropertyClass:           "DateTimeProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.run_time",
					DisplayName:             "Time",
					Label:                   "TIME",
					PropertyClass:           "TimeStringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.time_limit",
					DisplayName:             "Time Limit",
					Label:                   "TIME_LIMIT",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.time_left",
					DisplayName:             "Time Left",
					Label:                   "TIME_LEFT",
					PropertyClass:           "StringProperty",
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
				{
					ID:                      "squeue.min_memory",
					DisplayName:             "Memory",
					Label:                   "MIN_MEMORY",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.nodelist",
					DisplayName:             "Node List",
					Label:                   "NODELIST",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.exec_host",
					DisplayName:             "Exec Host",
					Label:                   "EXEC_HOST",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.features",
					DisplayName:             "Features",
					Label:                   "FEATURES",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
				{
					ID:                      "squeue.priority",
					DisplayName:             "Priority",
					Label:                   "PRIORITY",
					PropertyClass:           "StringProperty",
					PropertyCreateArguments: []string{},
					RecordParserClass:       "TokenRecordParser",
				},
			},
		},
	}
	return categoryList
}

package common

import (
	"bytes"
	"fmt"
	"github.com/perillaroc/nwpc-hpc-model-go"
	"github.com/perillaroc/nwpc-hpc-model-go/slurm"
	"os/exec"
	"strings"
)

func GetSinfoQueryModel(lines []string) (*slurm.Model, error) {
	categoryList := buildSinfoCategoryList()
	model, err := slurm.BuildModel(lines, categoryList, " ")
	return model, err
}

func GetSinfoCommandResult(params []string) ([]string, error) {
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

func buildSinfoCategoryList() slurm.QueryCategoryList {
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

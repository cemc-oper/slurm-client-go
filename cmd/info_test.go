package cmd_test

import (
	"fmt"
	"github.com/perillaroc/nwpc-hpc-model-go"
	"slurm-client-go/cmd"
	"strings"
	"testing"
)

func TestGetQueryModel(t *testing.T) {
	line := `PARTITION            AVAIL       NODES(A/I/O/T)                  CPUS(A/I/O/T)
serial                  up            24/0/0/24                  263/505/0/768
serial_op               up            24/0/0/24                  263/505/0/768
largemem                up          226/0/0/226                  7232/0/0/7232
normal                  up        1504/0/0/1504                48120/8/0/48128
operation               up        1504/0/0/1504                48120/8/0/48128`
	lines := strings.Split(line, "\n")
	fmt.Printf("%v\n", lines)
	model, err := cmd.GetQueryModel(lines)
	if err != nil {
		t.Errorf("get query model failed: %v", err)
	}

	tests := []struct {
		index int
		id    string
		data  string
	}{
		{
			0,
			"sinfo.partition",
			"serial",
		},
		{
			0,
			"sinfo.avail",
			"up",
		},
		{
			4,
			"sinfo.partition",
			"operation",
		},
	}

	for _, test := range tests {
		item := model.Items[test.index]
		prop := item.GetProperty(test.id).(*hpcmodel.StringProperty)
		if prop.Data != test.data {
			t.Errorf("Item %d property %s data is %s, required %s",
				test.index, test.id, prop.Data, test.data)
		}
	}
}

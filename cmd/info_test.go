package cmd

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

const testSinfoHeader = "PARTITION AVAIL NODES(A/I/O/T) CPUS(A/I/O/T)"

func TestQuerySinfoLinesSortsByMultipleKeys(t *testing.T) {
	lines := []string{
		testSinfoHeader,
		"gpu up 2/0/0/2 8/0/0/8",
		"cpu up 1/0/0/1 4/0/0/4",
		"beta down 3/0/0/3 12/0/0/12",
		"alpha down 2/0/0/2 8/0/0/8",
	}
	items, err := querySinfoLines(lines, "avail:partition")
	if err != nil {
		t.Fatalf("querySinfoLines() error = %v", err)
	}
	want := []string{"alpha", "beta", "cpu", "gpu"}
	if got := itemIDs(items, "sinfo.partition"); !reflect.DeepEqual(got, want) {
		t.Errorf("partitions = %v, want %v", got, want)
	}
}

func TestInfoCommandUsesRunnerAndWriter(t *testing.T) {
	lines := []string{testSinfoHeader, "gpu up 2/0/0/2 8/0/0/8", "cpu up 1/0/0/1 4/0/0/4"}
	runner := &fakeCommandRunner{sinfoResults: []commandResult{{lines: lines}}}
	var output bytes.Buffer
	if err := infoCommand(&output, runner, "partition"); err != nil {
		t.Fatalf("infoCommand() error = %v", err)
	}
	if !reflect.DeepEqual(runner.sinfoParams, [][]string{{"-o", "%20P %.5a %.20F %.30C"}}) {
		t.Errorf("sinfo params = %v", runner.sinfoParams)
	}
	text := output.String()
	cpuIndex := strings.Index(text, "cpu")
	gpuIndex := strings.Index(text, "gpu")
	if cpuIndex < 0 || gpuIndex < 0 || cpuIndex >= gpuIndex {
		t.Errorf("info output is not sorted by partition: %q", text)
	}
}

func TestInfoCommandEmptyInput(t *testing.T) {
	runner := &fakeCommandRunner{sinfoResults: []commandResult{{lines: []string{testSinfoHeader}}}}
	var output bytes.Buffer
	if err := infoCommand(&output, runner, "partition"); err != nil {
		t.Fatalf("infoCommand() error = %v", err)
	}
	if output.Len() != 0 {
		t.Errorf("empty info output = %q, want no output", output.String())
	}
}

func TestInfoCommandReturnsErrors(t *testing.T) {
	commandErr := errors.New("sinfo unavailable")
	tests := []struct {
		name    string
		result  commandResult
		wantErr string
	}{
		{name: "command execution", result: commandResult{err: commandErr}, wantErr: "get query result error"},
		{name: "model parsing", result: commandResult{lines: []string{testSinfoHeader, "cpu up"}}, wantErr: "model build failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{sinfoResults: []commandResult{tt.result}}
			var output bytes.Buffer
			err := infoCommand(&output, runner, "partition")
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("infoCommand() error = %v, want it to contain %q", err, tt.wantErr)
			}
			if tt.result.err != nil && !errors.Is(err, commandErr) {
				t.Errorf("infoCommand() error = %v, want wrapped command error", err)
			}
		})
	}
}

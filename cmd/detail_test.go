package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestDetailCommandFiltersSortsAndRenders(t *testing.T) {
	runner := &fakeCommandRunner{squeueResults: []commandResult{{lines: readSqueueFixture(t, "squeue_sc1_anonymized.txt")}}}
	var output bytes.Buffer
	err := detailCommand(&output, runner, []string{"user_a", "user_c", "user_d"}, []string{"partition_a", "partition_c"}, "state:submit_time", "")
	if err != nil {
		t.Fatalf("detailCommand() error = %v", err)
	}
	text := ansi.Strip(output.String())
	for _, section := range []string{"Basic Information", "Resources", "Time", "Execution", "Command"} {
		if !strings.Contains(text, section) {
			t.Errorf("detail output does not contain section %q", section)
		}
	}
	wantOrder := []string{"100001", "100003", "100004", "100006"}
	previous := -1
	for _, id := range wantOrder {
		index := strings.Index(text, id)
		if index < 0 || index <= previous {
			t.Errorf("detail output is missing or misorders job %s: %q", id, text)
		}
		previous = index
	}
	if strings.Contains(text, "100005") {
		t.Errorf("detail output contains a filtered job: %q", text)
	}
}

func TestDetailCommandNoMatches(t *testing.T) {
	runner := &fakeCommandRunner{squeueResults: []commandResult{{lines: readSqueueFixture(t, "squeue_sc1_anonymized.txt")}}}
	var output bytes.Buffer
	if err := detailCommand(&output, runner, []string{"nobody"}, nil, "state", ""); err != nil {
		t.Fatalf("detailCommand() error = %v", err)
	}
	if output.Len() != 0 {
		t.Errorf("detail output with no matches = %q, want no output", output.String())
	}
}

func TestDetailCommandReturnsErrors(t *testing.T) {
	commandErr := errors.New("squeue unavailable")
	tests := []struct {
		name    string
		result  commandResult
		wantErr string
	}{
		{name: "command execution", result: commandResult{err: commandErr}, wantErr: "get query result error"},
		{name: "model parsing", result: commandResult{lines: []string{"JOBID|USER"}}, wantErr: "model build failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{squeueResults: []commandResult{tt.result}}
			var output bytes.Buffer
			err := detailCommand(&output, runner, nil, nil, "state", "")
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("detailCommand() error = %v, want it to contain %q", err, tt.wantErr)
			}
			if tt.result.err != nil && !errors.Is(err, commandErr) {
				t.Errorf("detailCommand() error = %v, want wrapped command error", err)
			}
		})
	}
}

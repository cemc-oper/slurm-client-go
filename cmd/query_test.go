package cmd

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	hpcmodel "github.com/cemc-oper/hpc-model-go"
)

func itemIDs(items []hpcmodel.Item, propertyID string) []string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.GetProperty(propertyID).(*hpcmodel.StringProperty).Text
	}
	return ids
}

func TestQuerySqueueLinesCombinesFilters(t *testing.T) {
	items, err := querySqueueLines(readSqueueFixture(t, "squeue_sc1_anonymized.txt"), squeueQueryOptions{
		users:          []string{"user_a", "user_c", "user_d"},
		partitions:     []string{"partition_a", "partition_c"},
		commandPattern: "20240101",
	})
	if err != nil {
		t.Fatalf("querySqueueLines() error = %v", err)
	}
	if got, want := itemIDs(items, "squeue.job_id"), []string{"100006"}; !reflect.DeepEqual(got, want) {
		t.Errorf("job IDs = %v, want %v", got, want)
	}
}

func TestQuerySqueueLinesSortsByMultipleKeys(t *testing.T) {
	items, err := querySqueueLines(readSqueueFixture(t, "squeue_sc1_anonymized.txt"), squeueQueryOptions{
		users:      []string{"user_a", "user_c", "user_d"},
		partitions: []string{"partition_a", "partition_c"},
		sortString: "state:submit_time",
	})
	if err != nil {
		t.Fatalf("querySqueueLines() error = %v", err)
	}
	want := []string{"100001", "100003", "100004", "100006"}
	if got := itemIDs(items, "squeue.job_id"); !reflect.DeepEqual(got, want) {
		t.Errorf("job IDs = %v, want %v", got, want)
	}
}

func TestQuerySqueueLinesSortBoundaries(t *testing.T) {
	t.Run("empty sort preserves input order", func(t *testing.T) {
		items, err := querySqueueLines(readSqueueFixture(t, "squeue_sc1_anonymized.txt"), squeueQueryOptions{})
		if err != nil {
			t.Fatalf("querySqueueLines() error = %v", err)
		}
		if got, want := itemIDs(items, "squeue.job_id")[:2], []string{"100001", "100002"}; !reflect.DeepEqual(got, want) {
			t.Errorf("first job IDs = %v, want %v", got, want)
		}
	})

	t.Run("unknown sort key returns an error", func(t *testing.T) {
		_, err := querySqueueLines(readSqueueFixture(t, "squeue_sc1_anonymized.txt"), squeueQueryOptions{sortString: "missing"})
		if err == nil || !strings.Contains(err.Error(), "squeue.missing") {
			t.Errorf("querySqueueLines() error = %v, want an unknown sort key error", err)
		}
	})
}

func TestQueryCommandUsesRunnerAndWriter(t *testing.T) {
	runner := &fakeCommandRunner{squeueResults: []commandResult{{lines: readSqueueFixture(t, "squeue_sc1_anonymized.txt")}}}
	var output bytes.Buffer
	err := queryCommand(&output, runner, plainQueryColumns(), []string{"user_a"}, nil, "state:submit_time", "")
	if err != nil {
		t.Fatalf("queryCommand() error = %v", err)
	}
	if !reflect.DeepEqual(runner.squeueParams, [][]string{{"-o", "%all"}}) {
		t.Errorf("squeue params = %v, want [-o %%all]", runner.squeueParams)
	}
	text := output.String()
	for _, expected := range []string{"JOB ID", "100001", "100004"} {
		if !strings.Contains(text, expected) {
			t.Errorf("query output does not contain %q: %q", expected, text)
		}
	}
	if strings.Contains(text, "100003") {
		t.Errorf("query output contains a filtered job: %q", text)
	}
}

func TestQueryCommandEmptyQueue(t *testing.T) {
	runner := &fakeCommandRunner{squeueResults: []commandResult{{lines: readSqueueFixture(t, "squeue_sc1_anonymized.txt")[:1]}}}
	var output bytes.Buffer
	if err := queryCommand(&output, runner, plainQueryColumns(), nil, nil, "state:submit_time", ""); err != nil {
		t.Fatalf("queryCommand() error = %v", err)
	}
	if output.Len() != 0 {
		t.Errorf("empty queue output = %q, want no output", output.String())
	}
}

func TestQueryCommandReturnsErrors(t *testing.T) {
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
			err := queryCommand(&output, runner, plainQueryColumns(), nil, nil, "state", "")
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("queryCommand() error = %v, want it to contain %q", err, tt.wantErr)
			}
			if tt.result.err != nil && !errors.Is(err, commandErr) {
				t.Errorf("queryCommand() error = %v, want wrapped command error", err)
			}
		})
	}
}

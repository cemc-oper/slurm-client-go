package cmd

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestQuerySqueueLinesFiltersJobIDs(t *testing.T) {
	items, err := querySqueueLines(readSqueueFixture(t, "squeue_array_anonymized.txt"), squeueQueryOptions{
		users:      []string{"user_a", "user_b"},
		partitions: []string{"partition_a"},
		jobs:       []string{"100123_2"},
	})
	if err != nil {
		t.Fatalf("querySqueueLines() error = %v", err)
	}
	if got, want := itemIDs(items, "squeue.job_id"), []string{"100123_2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("job IDs = %v, want %v", got, want)
	}
}

func TestWatchCommandUsesInjectedClockAndStops(t *testing.T) {
	fixture := readSqueueFixture(t, "squeue_sc1_anonymized.txt")
	runner := &fakeCommandRunner{squeueResults: []commandResult{
		{lines: fixture},
		{lines: fixture[:1]},
	}}
	clock := &fakeWatchClock{times: []time.Time{
		time.Date(2026, 10, 26, 8, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 26, 8, 1, 0, 0, time.UTC),
	}}
	var output bytes.Buffer
	if err := watchCommand(&output, runner, clock, nil, nil, nil); err != nil {
		t.Fatalf("watchCommand() error = %v", err)
	}
	want := "[2026-10-26 08:00:00]checking jobs...COMPLETING: 1, PENDING: 2, RUNNING: 3\n" +
		"[2026-10-26 08:01:00]checking jobs...done\n"
	if got := output.String(); got != want {
		t.Errorf("watch output = %q, want %q", got, want)
	}
	if runner.squeueCalls != 2 {
		t.Errorf("squeue calls = %d, want 2", runner.squeueCalls)
	}
	if !reflect.DeepEqual(clock.sleeps, []time.Duration{time.Minute}) {
		t.Errorf("sleep durations = %v, want one minute", clock.sleeps)
	}
}

func TestWatchCommandStopsImmediatelyWithoutMatches(t *testing.T) {
	fixture := readSqueueFixture(t, "squeue_sc1_anonymized.txt")
	runner := &fakeCommandRunner{squeueResults: []commandResult{{lines: fixture}}}
	clock := &fakeWatchClock{}
	var output bytes.Buffer
	if err := watchCommand(&output, runner, clock, []string{"nobody"}, nil, nil); err != nil {
		t.Fatalf("watchCommand() error = %v", err)
	}
	if !strings.Contains(output.String(), "done") {
		t.Errorf("watch output = %q, want done", output.String())
	}
	if runner.squeueCalls != 1 || len(clock.sleeps) != 0 {
		t.Errorf("calls = %d, sleeps = %v; want one call and no sleep", runner.squeueCalls, clock.sleeps)
	}
}

func TestWatchCommandReturnsErrorsWithoutSleeping(t *testing.T) {
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
			clock := &fakeWatchClock{}
			var output bytes.Buffer
			err := watchCommand(&output, runner, clock, nil, nil, nil)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("watchCommand() error = %v, want it to contain %q", err, tt.wantErr)
			}
			if tt.result.err != nil && !errors.Is(err, commandErr) {
				t.Errorf("watchCommand() error = %v, want wrapped command error", err)
			}
			if len(clock.sleeps) != 0 {
				t.Errorf("sleep durations = %v, want none after an error", clock.sleeps)
			}
		})
	}
}

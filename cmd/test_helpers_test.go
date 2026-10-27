package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cemc-oper/slurm-client-go/tui"
)

type commandResult struct {
	lines []string
	err   error
}

type fakeCommandRunner struct {
	squeueResults []commandResult
	sinfoResults  []commandResult
	squeueCalls   int
	sinfoCalls    int
	squeueParams  [][]string
	sinfoParams   [][]string
}

func (f *fakeCommandRunner) Squeue(params []string) ([]string, error) {
	f.squeueParams = append(f.squeueParams, append([]string(nil), params...))
	index := f.squeueCalls
	f.squeueCalls++
	if index >= len(f.squeueResults) {
		return nil, fmt.Errorf("unexpected squeue call %d", index+1)
	}
	result := f.squeueResults[index]
	return result.lines, result.err
}

func (f *fakeCommandRunner) Sinfo(params []string) ([]string, error) {
	f.sinfoParams = append(f.sinfoParams, append([]string(nil), params...))
	index := f.sinfoCalls
	f.sinfoCalls++
	if index >= len(f.sinfoResults) {
		return nil, fmt.Errorf("unexpected sinfo call %d", index+1)
	}
	result := f.sinfoResults[index]
	return result.lines, result.err
}

type fakeWatchClock struct {
	times  []time.Time
	now    int
	sleeps []time.Duration
}

func (f *fakeWatchClock) Now() time.Time {
	if len(f.times) == 0 {
		return time.Date(2026, 10, 26, 8, 0, 0, 0, time.UTC)
	}
	index := f.now
	if index >= len(f.times) {
		index = len(f.times) - 1
	}
	f.now++
	return f.times[index]
}

func (f *fakeWatchClock) Sleep(duration time.Duration) {
	f.sleeps = append(f.sleeps, duration)
}

func readSqueueFixture(t *testing.T, name string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "common", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

func plainQueryColumns() []tui.ColDef {
	return []tui.ColDef{
		{Prop: "squeue.job_id"},
		{Prop: "squeue.state"},
		{Prop: "squeue.partition"},
		{Prop: "squeue.nodes"},
		{Prop: "squeue.user"},
		{Prop: "squeue.submit_time"},
		{Prop: "squeue.run_time"},
		{Prop: "squeue.command"},
	}
}

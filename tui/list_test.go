package tui

import (
	"reflect"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	hpcmodel "github.com/cemc-oper/hpc-model-go"
	"github.com/charmbracelet/x/ansi"
)

func TestComputeColWidths(t *testing.T) {
	longCommand := strings.Repeat("command ", 8)
	tests := []struct {
		name    string
		columns []ColDef
		items   []hpcmodel.Item
		want    []int
	}{
		{
			name:    "no items",
			columns: testColumns(),
			want:    make([]int, len(testColumns())),
		},
		{
			name:    "empty values",
			columns: []ColDef{{Prop: "squeue.command"}},
			items:   []hpcmodel.Item{newTestItem("1", "alice", "cpu", "PENDING", "")},
			want:    []int{0},
		},
		{
			name:    "long command",
			columns: []ColDef{{Prop: "squeue.command"}},
			items: []hpcmodel.Item{
				newTestItem("1", "alice", "cpu", "RUNNING", "short"),
				newTestItem("2", "alice", "cpu", "RUNNING", longCommand),
			},
			want: []int{len(longCommand)},
		},
		{
			name:    "Chinese and English display widths",
			columns: []ColDef{{Prop: "squeue.command"}, {Prop: "squeue.user"}},
			items: []hpcmodel.Item{
				newTestItem("1", "alice", "cpu", "RUNNING", "作业ab"),
				newTestItem("2", "李雷", "cpu", "RUNNING", "run"),
			},
			want: []int{6, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := computeColWidths(tt.columns, tt.items); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("computeColWidths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildSummaryFromJobs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := buildSummaryFromJobs(nil); got != "" {
			t.Errorf("buildSummaryFromJobs() = %q, want an empty summary", got)
		}
	})

	t.Run("known states and unique users", func(t *testing.T) {
		jobs := []jobItem{
			newTestJob("1", "bob", "cpu", "PENDING"),
			newTestJob("2", "alice", "cpu", "RUNNING"),
			newTestJob("3", "alice", "gpu", "RUNNING"),
			newTestJob("4", "carol", "gpu", "COMPLETING"),
		}
		got := ansi.Strip(buildSummaryFromJobs(jobs))
		want := "Total: 4  RUNNING: 2  PENDING: 1  COMPLETING: 1  Users: 3"
		if got != want {
			t.Errorf("buildSummaryFromJobs() = %q, want %q", got, want)
		}
	})

	t.Run("unknown states are sorted", func(t *testing.T) {
		jobs := []jobItem{
			newTestJob("1", "alice", "cpu", "REQUEUED"),
			newTestJob("2", "alice", "cpu", "PREEMPTED"),
			newTestJob("3", "bob", "cpu", "REQUEUED"),
			newTestJob("4", "bob", "cpu", "RUNNING"),
		}
		want := "Total: 4  RUNNING: 1  PREEMPTED: 1  REQUEUED: 2  Users: 2"
		for i := 0; i < 20; i++ {
			if got := ansi.Strip(buildSummaryFromJobs(jobs)); got != want {
				t.Fatalf("buildSummaryFromJobs() iteration %d = %q, want %q", i, got, want)
			}
		}
	})
}

func TestBuildSummaryIgnoresNonJobItems(t *testing.T) {
	items := []list.Item{newTestJob("1", "alice", "cpu", "RUNNING"), treeItem{itemType: treeItemGroup, groupKey: "alice"}}
	got := ansi.Strip(buildSummary(items))
	want := "Total: 1  RUNNING: 1  Users: 1"
	if got != want {
		t.Errorf("buildSummary() = %q, want %q", got, want)
	}
}

func TestJobDelegateTruncatesLongCommandToAvailableWidth(t *testing.T) {
	job := newTestItem("1001", "alice", "cpu", "RUNNING", strings.Repeat("x", 100))
	delegate := jobDelegate{
		columns:   []ColDef{{Prop: "squeue.job_id"}, {Prop: "squeue.command"}},
		colWidths: []int{4, 100},
	}
	listModel := list.New(nil, delegate, 24, 10)
	listModel.Select(1)
	var output strings.Builder
	delegate.Render(&output, listModel, 0, jobItem{item: job})

	line := ansi.Strip(output.String())
	if !strings.HasSuffix(line, "...") {
		t.Errorf("Render() = %q, want an ellipsis for the truncated command", line)
	}
	if got, want := lipgloss.Width(line), listModel.Width()-2; got != want {
		t.Errorf("rendered width = %d, want %d; line = %q", got, want, line)
	}
}

package common_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cemc-oper/hpc-model-go"
	"github.com/cemc-oper/slurm-client-go/common"
)

func readSqueueFixture(t *testing.T, name string) []string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	text := strings.TrimRight(string(content), "\r\n")
	return strings.Split(text, "\n")
}

func requireStringProperty(t *testing.T, item hpcmodel.Item, id string) *hpcmodel.StringProperty {
	t.Helper()
	property := item.GetProperty(id)
	if property == nil {
		t.Fatalf("property %q not found", id)
	}
	p, ok := property.(*hpcmodel.StringProperty)
	if !ok {
		t.Fatalf("property %q has type %T, want *StringProperty", id, property)
	}
	return p
}

func TestGetSqueueQueryModelAnonymizedFixtures(t *testing.T) {
	type expectedJob struct {
		id         string
		user       string
		partition  string
		state      string
		nodes      float64
		runTime    time.Duration
		submitTime time.Time
	}

	tests := []struct {
		name    string
		fixture string
		jobs    []expectedJob
	}{
		{
			name:    "sc1",
			fixture: "squeue_sc1_anonymized.txt",
			jobs: []expectedJob{
				{"100001", "user_a", "partition_a", "PENDING", 1, 0, time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)},
				{"100002", "user_b", "partition_b", "PENDING", 2, 0, time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)},
				{"100003", "user_c", "partition_c", "RUNNING", 4, 4 * time.Second, time.Date(2026, 1, 1, 3, 0, 0, 0, time.UTC)},
				{"100004", "user_a", "partition_a", "RUNNING", 1, time.Hour + 18*time.Minute + time.Second, time.Date(2026, 1, 1, 4, 0, 0, 0, time.UTC)},
				{"100005", "user_d", "partition_d", "COMPLETING", 2, 16 * time.Second, time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)},
				{"100006", "user_c", "partition_c", "RUNNING", 4, time.Minute + 34*time.Second, time.Date(2026, 1, 1, 6, 0, 0, 0, time.UTC)},
			},
		},
		{
			name:    "sc3",
			fixture: "squeue_sc3_anonymized.txt",
			jobs: []expectedJob{
				{"200001", "user_a", "partition_a", "RUNNING", 1, 4*time.Minute + 24*time.Second, time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)},
				{"200002", "user_b", "partition_b", "COMPLETING", 2, 3*24*time.Hour + 18*time.Second, time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)},
				{"200003", "user_c", "partition_c", "PENDING", 4, 0, time.Date(2026, 1, 1, 3, 0, 0, 0, time.UTC)},
				{"200004", "user_d", "partition_d", "PENDING", 1, 0, time.Date(2026, 1, 1, 4, 0, 0, 0, time.UTC)},
				{"200005", "user_a", "partition_a", "RUNNING", 2, 4*time.Hour + 3*time.Minute + 31*time.Second, time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)},
				{"200006", "user_c", "partition_c", "RUNNING", 4, 5*24*time.Hour + 5*time.Hour + 7*time.Minute + 30*time.Second, time.Date(2026, 1, 1, 6, 0, 0, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := common.GetSqueueQueryModel(readSqueueFixture(t, tt.fixture))
			if err != nil {
				t.Fatalf("GetSqueueQueryModel() error = %v", err)
			}
			if model == nil {
				t.Fatal("GetSqueueQueryModel() returned a nil model")
			}
			if got, want := len(model.Items), len(tt.jobs); got != want {
				t.Fatalf("len(Items) = %d, want %d", got, want)
			}

			categories := common.BuildSqueueCategoryList().CategoryList
			for i, want := range tt.jobs {
				item := model.Items[i]
				if got := len(item.Props); got != len(categories) {
					t.Errorf("job %s has %d properties, want %d", want.id, got, len(categories))
				}
				for _, category := range categories {
					if item.GetProperty(category.ID) == nil {
						t.Errorf("job %s property %q not found", want.id, category.ID)
					}
				}

				stringTests := []struct{ id, want string }{
					{"squeue.job_id", want.id},
					{"squeue.user", want.user},
					{"squeue.partition", want.partition},
					{"squeue.state", want.state},
				}
				for _, test := range stringTests {
					if got := requireStringProperty(t, item, test.id).Data; got != test.want {
						t.Errorf("job %s %s = %q, want %q", want.id, test.id, got, test.want)
					}
				}

				nodesProperty, ok := item.GetProperty("squeue.nodes").(*hpcmodel.NumberProperty)
				if !ok {
					t.Fatalf("job %s nodes has unexpected type %T", want.id, item.GetProperty("squeue.nodes"))
				}
				if nodesProperty.Data != want.nodes {
					t.Errorf("job %s nodes = %v, want %v", want.id, nodesProperty.Data, want.nodes)
				}

				runProperty, ok := item.GetProperty("squeue.run_time").(*hpcmodel.TimeStringProperty)
				if !ok {
					t.Fatalf("job %s run time has unexpected type %T", want.id, item.GetProperty("squeue.run_time"))
				}
				if runProperty.Data != want.runTime {
					t.Errorf("job %s run time = %v, want %v", want.id, runProperty.Data, want.runTime)
				}

				submitProperty, ok := item.GetProperty("squeue.submit_time").(*hpcmodel.DateTimeProperty)
				if !ok {
					t.Fatalf("job %s submit time has unexpected type %T", want.id, item.GetProperty("squeue.submit_time"))
				}
				if !submitProperty.Data.Equal(want.submitTime) {
					t.Errorf("job %s submit time = %v, want %v", want.id, submitProperty.Data, want.submitTime)
				}
			}
		})
	}
}

func TestGetSqueueQueryModelArrayJobs(t *testing.T) {
	model, err := common.GetSqueueQueryModel(readSqueueFixture(t, "squeue_array_anonymized.txt"))
	if err != nil {
		t.Fatalf("GetSqueueQueryModel() error = %v", err)
	}
	if model == nil {
		t.Fatal("GetSqueueQueryModel() returned a nil model")
	}
	if got, want := len(model.Items), 2; got != want {
		t.Fatalf("len(Items) = %d, want %d", got, want)
	}

	tests := []struct {
		row          int
		wantJobID    string
		wantState    string
		wantPriority string
	}{
		{row: 0, wantJobID: "100123_2", wantState: "RUNNING", wantPriority: "0.9"},
		{row: 1, wantJobID: "100123_3", wantState: "PENDING", wantPriority: "0.8"},
	}
	for _, tt := range tests {
		t.Run(tt.wantJobID, func(t *testing.T) {
			item := model.Items[tt.row]
			propertyTests := []struct{ id, want string }{
				{id: "squeue.job_id", want: tt.wantJobID},
				{id: "squeue.state", want: tt.wantState},
				{id: "squeue.priority", want: tt.wantPriority},
			}
			for _, test := range propertyTests {
				if got := requireStringProperty(t, item, test.id).Data; got != test.want {
					t.Errorf("%s = %q, want %q", test.id, got, test.want)
				}
			}
		})
	}
}

func TestSqueueFixtureSpecialFields(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		row     int
		id      string
		want    string
	}{
		{"pending reason", "squeue_sc1_anonymized.txt", 0, "squeue.reason", "Resources"},
		{"pending empty nodelist", "squeue_sc1_anonymized.txt", 0, "squeue.nodelist", ""},
		{"pending exec host", "squeue_sc1_anonymized.txt", 0, "squeue.exec_host", "n/a"},
		{"null features", "squeue_sc1_anonymized.txt", 0, "squeue.features", "(null)"},
		{"command with spaces", "squeue_sc1_anonymized.txt", 5, "squeue.command", "/work/user_c/100006/run.sh 20240101 2"},
		{"work directory", "squeue_sc1_anonymized.txt", 5, "squeue.work_dir", "/work/user_c/100006"},
		{"invalid time left", "squeue_sc3_anonymized.txt", 1, "squeue.time_left", "INVALID"},
		{"completing reason", "squeue_sc3_anonymized.txt", 1, "squeue.reason", "TimeLimit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := common.GetSqueueQueryModel(readSqueueFixture(t, tt.fixture))
			if err != nil {
				t.Fatalf("GetSqueueQueryModel() error = %v", err)
			}
			got := requireStringProperty(t, model.Items[tt.row], tt.id).Data
			if got != tt.want {
				t.Errorf("%s = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestSqueueFixtureColumnCounts(t *testing.T) {
	for _, fixture := range []string{"squeue_sc1_anonymized.txt", "squeue_sc3_anonymized.txt", "squeue_array_anonymized.txt"} {
		t.Run(fixture, func(t *testing.T) {
			lines := readSqueueFixture(t, fixture)
			wantColumns := strings.Count(lines[0], "|") + 1
			if wantColumns != 52 {
				t.Fatalf("header has %d columns, want 52", wantColumns)
			}
			for i, line := range lines {
				if got := strings.Count(line, "|") + 1; got != wantColumns {
					t.Errorf("line %d has %d columns, want %d", i+1, got, wantColumns)
				}
			}
		})
	}
}

func TestGetSqueueQueryModelHeaderOnlyAndTrailingBlankLine(t *testing.T) {
	for _, fixture := range []string{"squeue_sc1_anonymized.txt", "squeue_sc3_anonymized.txt", "squeue_array_anonymized.txt"} {
		t.Run(fixture, func(t *testing.T) {
			lines := readSqueueFixture(t, fixture)

			headerOnly, err := common.GetSqueueQueryModel(lines[:1])
			if err != nil {
				t.Fatalf("header-only input returned error: %v", err)
			}
			if headerOnly == nil || len(headerOnly.Items) != 0 {
				t.Errorf("header-only input returned %#v, want an empty model", headerOnly)
			}

			withBlankLine := append(append([]string(nil), lines...), "")
			model, err := common.GetSqueueQueryModel(withBlankLine)
			if err != nil {
				t.Fatalf("input with trailing blank line returned error: %v", err)
			}
			if got, want := len(model.Items), len(lines)-1; got != want {
				t.Errorf("len(Items) = %d, want %d", got, want)
			}
		})
	}
}

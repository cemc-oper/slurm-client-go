package common_test

import (
	"strings"
	"testing"

	"github.com/cemc-oper/slurm-client-go/common"
)

func TestGetSqueueQueryModelEmptyInput(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
	}{
		{name: "nil", lines: nil},
		{name: "empty slice", lines: []string{}},
		{name: "blank lines", lines: []string{"", "  ", "\t"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := common.GetSqueueQueryModel(tt.lines)
			if err != nil {
				t.Fatalf("GetSqueueQueryModel() error = %v", err)
			}
			if model == nil {
				t.Fatal("GetSqueueQueryModel() returned a nil model")
			}
			if len(model.Items) != 0 {
				t.Errorf("len(Items) = %d, want 0", len(model.Items))
			}
		})
	}
}

func TestGetSqueueQueryModelMissingColumns(t *testing.T) {
	tests := []struct {
		name    string
		missing []string
	}{
		{name: "one column", missing: []string{"USER"}},
		{name: "multiple columns", missing: []string{"USER", "WORK_DIR"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := readSqueueFixture(t, "squeue_sc1_anonymized.txt")
			excluded := make(map[string]bool)
			for _, label := range tt.missing {
				excluded[label] = true
			}
			header := strings.Split(lines[0], "|")
			filtered := make([]string, 0, len(header))
			for _, label := range header {
				if !excluded[label] {
					filtered = append(filtered, label)
				}
			}
			lines[0] = strings.Join(filtered, "|")

			_, err := common.GetSqueueQueryModel(lines)
			if err == nil {
				t.Fatal("GetSqueueQueryModel() error = nil, want missing column error")
			}
			for _, label := range tt.missing {
				if !strings.Contains(err.Error(), label) {
					t.Errorf("error = %q, want it to contain %q", err, label)
				}
			}
		})
	}
}

func TestGetSqueueQueryModelColumnCountMismatch(t *testing.T) {
	tests := []struct {
		name      string
		mutateRow func([]string) string
	}{
		{
			name: "short row",
			mutateRow: func(fields []string) string {
				return strings.Join(fields[:5], "|")
			},
		},
		{
			name: "long row",
			mutateRow: func(fields []string) string {
				return strings.Join(fields, "|") + "|extra"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := readSqueueFixture(t, "squeue_sc1_anonymized.txt")[:2]
			lines[1] = tt.mutateRow(strings.Split(lines[1], "|"))
			_, err := common.GetSqueueQueryModel(lines)
			if err == nil {
				t.Fatal("GetSqueueQueryModel() error = nil, want column count error")
			}
			if !strings.Contains(err.Error(), "line 2") || !strings.Contains(err.Error(), "want 52") {
				t.Errorf("error = %q, want line 2 and expected column count", err)
			}
		})
	}
}

func TestGetSqueueQueryModelSkipsBlankLines(t *testing.T) {
	fixture := readSqueueFixture(t, "squeue_sc1_anonymized.txt")
	lines := []string{fixture[0], "", "  ", fixture[1], "\t", ""}
	model, err := common.GetSqueueQueryModel(lines)
	if err != nil {
		t.Fatalf("GetSqueueQueryModel() error = %v", err)
	}
	if model == nil || len(model.Items) != 1 {
		t.Errorf("model = %#v, want one item", model)
	}
}

func TestGetSinfoQueryModelEmptyInput(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
	}{
		{name: "nil", lines: nil},
		{name: "empty slice", lines: []string{}},
		{name: "blank lines", lines: []string{"", "  ", "\t"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := common.GetSinfoQueryModel(tt.lines)
			if err != nil {
				t.Fatalf("GetSinfoQueryModel() error = %v", err)
			}
			if model == nil {
				t.Fatal("GetSinfoQueryModel() returned a nil model")
			}
			if len(model.Items) != 0 {
				t.Errorf("len(Items) = %d, want 0", len(model.Items))
			}
		})
	}
}

func TestGetSinfoQueryModelRejectsMalformedInput(t *testing.T) {
	const header = "PARTITION AVAIL NODES(A/I/O/T) CPUS(A/I/O/T)"
	tests := []struct {
		name    string
		lines   []string
		wantErr string
	}{
		{
			name:    "missing header column",
			lines:   []string{"PARTITION AVAIL NODES(A/I/O/T)"},
			wantErr: "CPUS(A/I/O/T)",
		},
		{
			name:    "short row",
			lines:   []string{header, "serial up 24/0/0/24"},
			wantErr: "line 2",
		},
		{
			name:    "long row",
			lines:   []string{header, "serial up 24/0/0/24 263/505/0/768 extra"},
			wantErr: "line 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := common.GetSinfoQueryModel(tt.lines)
			if err == nil {
				t.Fatal("GetSinfoQueryModel() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestGetSinfoQueryModelSkipsBlankLines(t *testing.T) {
	lines := []string{
		"PARTITION AVAIL NODES(A/I/O/T) CPUS(A/I/O/T)",
		"",
		"  ",
		"serial up 24/0/0/24 263/505/0/768",
		"\t",
	}
	model, err := common.GetSinfoQueryModel(lines)
	if err != nil {
		t.Fatalf("GetSinfoQueryModel() error = %v", err)
	}
	if model == nil || len(model.Items) != 1 {
		t.Errorf("model = %#v, want one item", model)
	}
}

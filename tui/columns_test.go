package tui

import "testing"

func TestGetProp(t *testing.T) {
	item := newTestItem("100123_2", "alice", "gpu", "RUNNING", "/jobs/train model.sh")
	tests := []struct {
		name            string
		propertyID      string
		wantText        string
		wantDisplayName string
	}{
		{name: "job ID string", propertyID: "squeue.job_id", wantText: "100123_2", wantDisplayName: "JOB ID"},
		{name: "state string", propertyID: "squeue.state", wantText: "RUNNING", wantDisplayName: "State"},
		{name: "partition string", propertyID: "squeue.partition", wantText: "gpu", wantDisplayName: "Partition"},
		{name: "user string", propertyID: "squeue.user", wantText: "alice", wantDisplayName: "User"},
		{name: "command string", propertyID: "squeue.command", wantText: "/jobs/train model.sh", wantDisplayName: "Command"},
		{name: "nodes number", propertyID: "squeue.nodes", wantText: "2", wantDisplayName: "NODEs"},
		{name: "submit time datetime", propertyID: "squeue.submit_time", wantText: "2026-10-26 08:09:10", wantDisplayName: "Submit Time"},
		{name: "run time duration", propertyID: "squeue.run_time", wantText: "01:02:03", wantDisplayName: "Time"},
		{name: "unknown property", propertyID: "squeue.unknown", wantText: "", wantDisplayName: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, displayName := GetProp(item, tt.propertyID)
			if text != tt.wantText || displayName != tt.wantDisplayName {
				t.Errorf("GetProp() = (%q, %q), want (%q, %q)", text, displayName, tt.wantText, tt.wantDisplayName)
			}
		})
	}
}

func TestJobItemFilterValue(t *testing.T) {
	job := newTestJob("100123_2", "alice", "gpu", "RUNNING")
	if got, want := job.FilterValue(), "100123_2"; got != want {
		t.Errorf("FilterValue() = %q, want %q", got, want)
	}
}

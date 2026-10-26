package tui

import (
	"reflect"
	"testing"
)

type treeRow struct {
	itemType   treeItemType
	groupKey   string
	pathKey    string
	groupCount int
	isExpanded bool
	jobID      string
	level      int
}

func simplifyTreeRows(items []treeItem) []treeRow {
	rows := make([]treeRow, len(items))
	for i, item := range items {
		rows[i] = treeRow{
			itemType:   item.itemType,
			groupKey:   item.groupKey,
			pathKey:    item.pathKey,
			groupCount: item.groupCount,
			isExpanded: item.isExpanded,
			level:      item.level,
		}
		if item.itemType == treeItemJob {
			rows[i].jobID = item.job.FilterValue()
		}
	}
	return rows
}

func TestBuildTreeItemsSingleLevelSortsGroupsAndCountsJobs(t *testing.T) {
	jobs := []jobItem{
		newTestJob("3", "bob", "cpu", "RUNNING"),
		newTestJob("2", "alice", "cpu", "PENDING"),
		newTestJob("1", "alice", "gpu", "RUNNING"),
	}
	got := simplifyTreeRows(buildTreeItems(jobs, []string{"squeue.user"}, nil))
	want := []treeRow{
		{itemType: treeItemGroup, groupKey: "alice", pathKey: "alice", groupCount: 2, isExpanded: true},
		{itemType: treeItemJob, jobID: "2", level: 1},
		{itemType: treeItemJob, jobID: "1", level: 1},
		{itemType: treeItemGroup, groupKey: "bob", pathKey: "bob", groupCount: 1, isExpanded: true},
		{itemType: treeItemJob, jobID: "3", level: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildTreeItems() = %#v, want %#v", got, want)
	}
}

func TestBuildTreeItemsMultipleLevels(t *testing.T) {
	jobs := []jobItem{
		newTestJob("4", "bob", "cpu", "RUNNING"),
		newTestJob("3", "alice", "cpu", "PENDING"),
		newTestJob("2", "alice", "gpu", "RUNNING"),
		newTestJob("1", "alice", "cpu", "RUNNING"),
	}
	levels := []string{"squeue.user", "squeue.state"}
	got := simplifyTreeRows(buildTreeItems(jobs, levels, nil))
	want := []treeRow{
		{itemType: treeItemGroup, groupKey: "alice", pathKey: "alice", groupCount: 3, isExpanded: true},
		{itemType: treeItemGroup, groupKey: "PENDING", pathKey: "alice\x00PENDING", groupCount: 1, isExpanded: true, level: 1},
		{itemType: treeItemJob, jobID: "3", level: 2},
		{itemType: treeItemGroup, groupKey: "RUNNING", pathKey: "alice\x00RUNNING", groupCount: 2, isExpanded: true, level: 1},
		{itemType: treeItemJob, jobID: "2", level: 2},
		{itemType: treeItemJob, jobID: "1", level: 2},
		{itemType: treeItemGroup, groupKey: "bob", pathKey: "bob", groupCount: 1, isExpanded: true},
		{itemType: treeItemGroup, groupKey: "RUNNING", pathKey: "bob\x00RUNNING", groupCount: 1, isExpanded: true, level: 1},
		{itemType: treeItemJob, jobID: "4", level: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildTreeItems() = %#v, want %#v", got, want)
	}
}

func TestBuildTreeItemsCollapseAndExpand(t *testing.T) {
	jobs := []jobItem{
		newTestJob("1", "alice", "cpu", "RUNNING"),
		newTestJob("2", "alice", "cpu", "PENDING"),
		newTestJob("3", "bob", "cpu", "RUNNING"),
	}
	levels := []string{"squeue.user", "squeue.state"}
	collapsed := map[string]bool{joinPath([]string{"alice", "RUNNING"}): true}

	got := buildTreeItems(jobs, levels, collapsed)
	if len(got) != 7 {
		t.Fatalf("len(buildTreeItems()) = %d, want 7", len(got))
	}
	if got[1].groupKey != "PENDING" || !got[1].isExpanded || got[3].groupKey != "RUNNING" || got[3].isExpanded {
		t.Errorf("nested group expansion state is incorrect: %#v", got)
	}
	for _, item := range got {
		if item.itemType == treeItemJob && item.job.FilterValue() == "1" {
			t.Error("collapsed RUNNING group still contains job 1")
		}
	}

	delete(collapsed, joinPath([]string{"alice", "RUNNING"}))
	collapsed["alice"] = true
	got = buildTreeItems(jobs, levels, collapsed)
	if len(got) != 4 {
		t.Fatalf("len(buildTreeItems()) after collapsing alice = %d, want 4", len(got))
	}
	if got[0].groupKey != "alice" || got[0].isExpanded || got[0].groupCount != 2 {
		t.Errorf("collapsed alice group = %#v, want a collapsed group with two jobs", got[0])
	}

	delete(collapsed, "alice")
	got = buildTreeItems(jobs, levels, collapsed)
	if len(got) != 8 {
		t.Errorf("len(buildTreeItems()) after expanding alice = %d, want 8", len(got))
	}
}

func TestGroupJobsByUsesNAForEmptyValues(t *testing.T) {
	jobs := []jobItem{
		newTestJob("1", "", "cpu", "PENDING"),
		newTestJob("2", "alice", "cpu", "RUNNING"),
	}
	groups := groupJobsBy(jobs, "squeue.user")
	if len(groups) != 2 || len(groups["N/A"]) != 1 || groups["N/A"][0].FilterValue() != "1" {
		t.Errorf("groupJobsBy() = %#v, want empty user grouped under N/A", groups)
	}
}

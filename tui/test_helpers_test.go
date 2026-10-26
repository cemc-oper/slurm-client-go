package tui

import (
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	hpcmodel "github.com/cemc-oper/hpc-model-go"
)

func addTestStringProperty(item *hpcmodel.Item, id, displayName, value string) {
	property := &hpcmodel.StringProperty{}
	property.SetCategory(hpcmodel.QueryCategory{ID: id, DisplayName: displayName})
	property.SetValue(value)
	item.AddProp(property)
}

func newTestItem(id, user, partition, state, command string) hpcmodel.Item {
	var item hpcmodel.Item
	for _, property := range []struct {
		id          string
		displayName string
		value       string
	}{
		{id: "squeue.job_id", displayName: "JOB ID", value: id},
		{id: "squeue.state", displayName: "State", value: state},
		{id: "squeue.partition", displayName: "Partition", value: partition},
		{id: "squeue.user", displayName: "User", value: user},
		{id: "squeue.command", displayName: "Command", value: command},
	} {
		addTestStringProperty(&item, property.id, property.displayName, property.value)
	}

	number := &hpcmodel.NumberProperty{}
	number.SetCategory(hpcmodel.QueryCategory{ID: "squeue.nodes", DisplayName: "NODEs"})
	number.SetValue("2")
	item.AddProp(number)

	date := &hpcmodel.DateTimeProperty{}
	date.SetCategory(hpcmodel.QueryCategory{ID: "squeue.submit_time", DisplayName: "Submit Time"})
	date.SetValue("2026-10-26T08:09:10")
	item.AddProp(date)

	duration := &hpcmodel.TimeStringProperty{}
	duration.SetCategory(hpcmodel.QueryCategory{ID: "squeue.run_time", DisplayName: "Time"})
	duration.SetValue("01:02:03")
	item.AddProp(duration)
	return item
}

func newTestJob(id, user, partition, state string) jobItem {
	return jobItem{item: newTestItem(id, user, partition, state, "/jobs/"+id+".sh")}
}

func testColumns() []ColDef {
	return []ColDef{
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

func newTestModel(jobs []jobItem) model {
	columns := testColumns()
	m := model{
		list:          list.New(nil, jobDelegate{}, 100, 16),
		treeList:      list.New(nil, treeDelegate{}, 100, 16),
		width:         100,
		height:        20,
		columns:       columns,
		displayMode:   modeList,
		groupBy:       []string{"squeue.user"},
		treeCollapsed: make(map[string]bool),
		allItems:      jobs,
	}
	items := make([]hpcmodel.Item, len(jobs))
	for i, job := range jobs {
		items[i] = job.item
	}
	m.colWidths = computeColWidths(columns, items)
	m.rebuildList()
	return m
}

func updateWithKey(t *testing.T, m model, msg tea.KeyPressMsg) model {
	t.Helper()
	updated, _ := m.Update(msg)
	result, ok := updated.(model)
	if !ok {
		t.Fatalf("Update() returned %T, want model", updated)
	}
	return result
}

func pressPrintableKey(t *testing.T, m model, key rune) model {
	t.Helper()
	return updateWithKey(t, m, tea.KeyPressMsg{Code: key, Text: string(key)})
}

func pressSpecialKey(t *testing.T, m model, key rune) model {
	t.Helper()
	return updateWithKey(t, m, tea.KeyPressMsg{Code: key})
}

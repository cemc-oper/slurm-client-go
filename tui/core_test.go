package tui

import (
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestKeyboardTogglesListAndTree(t *testing.T) {
	m := newTestModel([]jobItem{newTestJob("1", "alice", "cpu", "RUNNING")})
	m = pressPrintableKey(t, m, 't')
	if m.displayMode != modeTree {
		t.Fatalf("display mode after first t = %v, want tree", m.displayMode)
	}
	m = pressPrintableKey(t, m, 't')
	if m.displayMode != modeList {
		t.Errorf("display mode after second t = %v, want list", m.displayMode)
	}
}

func TestKeyboardEntersAndLeavesDetail(t *testing.T) {
	m := newTestModel([]jobItem{newTestJob("100123_2", "alice", "cpu", "RUNNING")})
	m.detailOffset = 5
	m = pressSpecialKey(t, m, tea.KeyEnter)
	if m.state != detailView || m.detailItem == nil || m.detailItem.FilterValue() != "100123_2" {
		t.Fatalf("enter did not open the selected job detail: %#v", m.detailItem)
	}
	if m.detailOffset != 0 {
		t.Errorf("detail offset after enter = %d, want 0", m.detailOffset)
	}

	for _, key := range []rune{tea.KeyEscape, tea.KeyLeft} {
		m = pressSpecialKey(t, m, key)
		if m.state != listView || m.detailItem != nil || m.detailOffset != 0 {
			t.Errorf("key %q did not reset detail state", key)
		}
		m = pressSpecialKey(t, m, tea.KeyRight)
		if m.state != detailView {
			t.Fatalf("right did not reopen detail before key %q", key)
		}
	}
}

func TestKeyboardTreeGroupCollapseExpandAndJobDetail(t *testing.T) {
	m := newTestModel([]jobItem{
		newTestJob("1", "alice", "cpu", "RUNNING"),
		newTestJob("2", "bob", "cpu", "PENDING"),
	})
	m = pressPrintableKey(t, m, 't')
	m = pressSpecialKey(t, m, tea.KeyEnter)
	if !m.treeCollapsed["alice"] {
		t.Fatal("enter on an expanded group did not collapse it")
	}
	if len(m.treeList.Items()) != 3 {
		t.Errorf("tree item count after collapse = %d, want 3 visible rows", len(m.treeList.Items()))
	}

	m = pressSpecialKey(t, m, tea.KeyEnter)
	if m.treeCollapsed["alice"] {
		t.Fatal("enter on a collapsed group did not expand it")
	}
	if len(m.treeList.Items()) != 4 {
		t.Errorf("tree item count after expand = %d, want 4", len(m.treeList.Items()))
	}

	m.treeList.Select(1)
	m = pressSpecialKey(t, m, tea.KeyEnter)
	if m.state != detailView || m.detailItem == nil || m.detailItem.FilterValue() != "1" {
		t.Errorf("enter on a tree job did not open its detail: %#v", m.detailItem)
	}
}

func TestKeyboardCyclesGroupByPresets(t *testing.T) {
	m := newTestModel([]jobItem{
		newTestJob("1", "alice", "cpu", "RUNNING"),
		newTestJob("2", "bob", "gpu", "PENDING"),
	})
	m = pressPrintableKey(t, m, 't')
	m.treeCollapsed["stale"] = true

	for i := 1; i <= len(groupByPresets); i++ {
		m = pressPrintableKey(t, m, 'c')
		want := groupByPresets[i%len(groupByPresets)].levels
		if !reflect.DeepEqual(m.groupBy, want) {
			t.Errorf("group levels after c press %d = %v, want %v", i, m.groupBy, want)
		}
		if len(m.treeCollapsed) != 0 {
			t.Errorf("collapsed groups were not reset after c press %d", i)
		}
	}
}

func TestKeyboardCollapsesAndExpandsAllTreeGroups(t *testing.T) {
	m := newTestModel([]jobItem{
		newTestJob("1", "alice", "cpu", "RUNNING"),
		newTestJob("2", "bob", "gpu", "PENDING"),
	})
	m = pressPrintableKey(t, m, 't')
	m = pressPrintableKey(t, m, 'a')
	if len(m.treeList.Items()) != 2 || !m.treeCollapsed["alice"] || !m.treeCollapsed["bob"] {
		t.Errorf("a did not collapse all groups: items=%d collapsed=%v", len(m.treeList.Items()), m.treeCollapsed)
	}
	m = pressPrintableKey(t, m, 'a')
	if len(m.treeList.Items()) != 4 || len(m.treeCollapsed) != 0 {
		t.Errorf("a did not expand all groups: items=%d collapsed=%v", len(m.treeList.Items()), m.treeCollapsed)
	}
}

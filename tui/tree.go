package tui

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// displayMode controls whether jobs are shown as a flat list or a grouped tree.
type displayMode int

const (
	modeList displayMode = iota
	modeTree
)

// groupByPreset defines a multi-level grouping configuration for the tree view.
type groupByPreset struct {
	label  string
	levels []string
}

// groupByPresets is the ordered set of grouping presets the user can cycle through.
var groupByPresets = []groupByPreset{
	{"user", []string{"squeue.user"}},
	{"partition", []string{"squeue.partition"}},
	{"state", []string{"squeue.state"}},
	{"user > state", []string{"squeue.user", "squeue.state"}},
	{"partition > state", []string{"squeue.partition", "squeue.state"}},
}

// treeItemType distinguishes group headers from job rows.
type treeItemType int

const (
	treeItemGroup treeItemType = iota
	treeItemJob
)

// treeItem is a single row in the tree view: either a collapsible group header or a job row.
type treeItem struct {
	itemType   treeItemType
	groupKey   string // display text for this level
	pathKey    string // full path key for collapsed lookup
	groupCount int
	isExpanded bool
	job        jobItem
	level      int
}

// FilterValue implements list.Item.
func (t treeItem) FilterValue() string {
	if t.itemType == treeItemGroup {
		return t.groupKey
	}
	return t.job.FilterValue()
}

// treeDelegate renders tree rows (group headers or indented job rows).
type treeDelegate struct {
	jobDel jobDelegate
	isDark bool
}

func (d treeDelegate) Height() int                         { return 1 }
func (d treeDelegate) Spacing() int                        { return 0 }
func (d treeDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d treeDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	t, ok := listItem.(treeItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	selectedBG := lipgloss.Color("#1e293b")
	if !d.isDark {
		selectedBG = lipgloss.Color("#a1a1aa")
	}

	indent := strings.Repeat("  ", t.level)
	if t.itemType == treeItemGroup {
		icon := "▸"
		if t.isExpanded {
			icon = "▾"
		}
		text := fmt.Sprintf("%s%s %s (%d)", indent, icon, t.groupKey, t.groupCount)
		style := lipgloss.NewStyle().Bold(true)
		if isSelected {
			style = style.Background(selectedBG)
		}
		fmt.Fprint(w, style.Render(text))
	} else {
		// Render the job row through jobDelegate, which already handles columns.
		var buf bytes.Buffer
		d.jobDel.Render(&buf, m, index, t.job)
		fmt.Fprintf(w, "%s%s", indent, buf.String())
	}
}

// groupJobsBy groups jobs by the specified property and returns a map of group key to jobs.
func groupJobsBy(items []jobItem, propID string) map[string][]jobItem {
	groups := make(map[string][]jobItem)
	for _, ji := range items {
		key, _ := GetProp(ji.item, propID)
		if key == "" {
			key = "N/A"
		}
		groups[key] = append(groups[key], ji)
	}
	return groups
}

// joinPath encodes a path slice into a single string key for collapsed lookup.
func joinPath(path []string) string {
	return strings.Join(path, "\x00")
}

// buildTreeItems builds a multi-level tree from jobs grouped by the specified level hierarchy.
func buildTreeItems(items []jobItem, levels []string, collapsed map[string]bool) []treeItem {
	return buildTreeItemsRecursive(items, levels, 0, nil, collapsed)
}

func buildTreeItemsRecursive(items []jobItem, levels []string, depth int, parentPath []string, collapsed map[string]bool) []treeItem {
	if depth >= len(levels) {
		var result []treeItem
		for _, ji := range items {
			result = append(result, treeItem{itemType: treeItemJob, job: ji, level: depth})
		}
		return result
	}

	groups := groupJobsBy(items, levels[depth])
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result []treeItem
	for _, key := range keys {
		path := append(append([]string(nil), parentPath...), key)
		pathStr := joinPath(path)
		isCollapsed := collapsed[pathStr]
		isExpanded := !isCollapsed

		result = append(result, treeItem{
			itemType:   treeItemGroup,
			groupKey:   key,
			pathKey:    pathStr,
			groupCount: len(groups[key]),
			isExpanded: isExpanded,
			level:      depth,
		})

		if isExpanded {
			result = append(result, buildTreeItemsRecursive(groups[key], levels, depth+1, path, collapsed)...)
		}
	}

	return result
}

// treeListView renders the tree view using the common list frame.
func (m model) treeListView() tea.View {
	summary := buildSummaryFromJobs(m.allItems)
	groupLabel := groupByPresets[0].label
	for _, preset := range groupByPresets {
		if len(preset.levels) == len(m.groupBy) {
			match := true
			for i := range preset.levels {
				if preset.levels[i] != m.groupBy[i] {
					match = false
					break
				}
			}
			if match {
				groupLabel = preset.label
				break
			}
		}
	}
	hint := fmt.Sprintf("[r] refresh  [t] list  [c] group: %s  [a] expand/collapse all  [enter/→] expand/detail  [q] quit", groupLabel)
	return m.renderJobList(m.treeList, " Slurm Jobs (Tree) ", hint, summary)
}

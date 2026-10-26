package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  string
	}{
		{name: "empty", text: "", width: 4, want: ""},
		{name: "shorter than width", text: "abc", width: 4, want: "abc"},
		{name: "exact width", text: "abcd", width: 4, want: "abcd"},
		{name: "multiple lines", text: "abcdefgh", width: 3, want: "abc\ndef\ngh"},
		{name: "non-positive width", text: "abcdef", width: 0, want: "abcdef"},
		{name: "wide characters", text: "ab作业cd", width: 4, want: "ab作\n业cd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.width)
			if got != tt.want {
				t.Errorf("wrapText() = %q, want %q", got, tt.want)
			}
			if !utf8.ValidString(got) {
				t.Errorf("wrapText() returned invalid UTF-8: %q", got)
			}
			if tt.width > 0 {
				for _, line := range strings.Split(got, "\n") {
					if width := lipgloss.Width(line); width > tt.width {
						t.Errorf("line %q has display width %d, want at most %d", line, width, tt.width)
					}
				}
			}
		})
	}
}

func TestDetailScrolling(t *testing.T) {
	job := newTestJob("1001", "alice", "cpu", "RUNNING")
	job.item = newTestItem("1001", "alice", "cpu", "RUNNING", strings.Repeat("x", 200))
	m := newTestModel([]jobItem{job})
	m.state = detailView
	m.detailItem = &m.allItems[0]
	m.width = 40
	m.height = 12

	for _, key := range []string{"up", "pgup"} {
		if _, handled := m.handleDetailKey(key); !handled {
			t.Fatalf("handleDetailKey(%q) was not handled", key)
		}
		if m.detailOffset != 0 {
			t.Errorf("offset after %q at top = %d, want 0", key, m.detailOffset)
		}
	}

	m.handleDetailKey("down")
	if m.detailOffset != 1 {
		t.Errorf("offset after down = %d, want 1", m.detailOffset)
	}
	m.handleDetailKey("pgdown")
	if got, want := m.detailOffset, 9; got != want {
		t.Errorf("offset after pgdown = %d, want %d", got, want)
	}
	m.handleDetailKey("pgup")
	if m.detailOffset != 1 {
		t.Errorf("offset after pgup = %d, want 1", m.detailOffset)
	}
	m.handleDetailKey("up")
	if m.detailOffset != 0 {
		t.Errorf("offset after up = %d, want 0", m.detailOffset)
	}
}

func TestDetailViewTopAndBottom(t *testing.T) {
	job := newTestJob("1001", "alice", "cpu", "RUNNING")
	job.item = newTestItem("1001", "alice", "cpu", "RUNNING", "command-tail")
	m := newTestModel([]jobItem{job})
	m.state = detailView
	m.detailItem = &m.allItems[0]
	m.width = 40
	m.height = 10

	top := ansi.Strip(m.detailView().Content)
	if !strings.Contains(top, "Basic Information") || strings.Contains(top, "command-tail") {
		t.Errorf("top view does not contain the expected lines: %q", top)
	}

	m.detailOffset = 1000
	bottom := ansi.Strip(m.detailView().Content)
	if !strings.Contains(bottom, "command-tail") || strings.Contains(bottom, "Basic Information") {
		t.Errorf("bottom view does not clamp to the final lines: %q", bottom)
	}
}

func TestDetailViewAfterWindowHeightChanges(t *testing.T) {
	job := newTestJob("1001", "alice", "cpu", "RUNNING")
	job.item = newTestItem("1001", "alice", "cpu", "RUNNING", "resize-tail")
	m := newTestModel([]jobItem{job})
	m.state = detailView
	m.detailItem = &m.allItems[0]
	m.detailOffset = 1000
	m.width = 40

	for _, height := range []int{8, 20, 6} {
		m.height = height
		content := ansi.Strip(m.detailView().Content)
		if !strings.Contains(content, "resize-tail") {
			t.Errorf("height %d: bottom view lost the command tail: %q", height, content)
		}
	}
}

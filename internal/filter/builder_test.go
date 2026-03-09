package filter

import (
	"testing"

	"github.com/spf13/pflag"
)

func newFlagSet() *pflag.FlagSet {
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	RegisterFlags(f)
	return f
}

// TestParseDate tests conversion of convenience aliases and passthrough of raw values.
func TestParseDate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		// passthrough: already ISO 8601 date
		{"2026-03-01", "2026-03-01"},
		// passthrough: ISO 8601 duration
		{"-P30D", "-P30D"},
		{"P2W", "P2W"},
		{"-P2W1D", "-P2W1D"},
		// convenience aliases
		{"7d", "-P7D"},
		{"30d", "-P30D"},
		{"2w", "-P2W"},
		{"1m", "-P1M"},
		{"4w", "-P4W"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDateTodayYesterday(t *testing.T) {
	t.Parallel()
	// just verify no error and non-empty result (actual date varies)
	got, err := ParseDate("today")
	if err != nil {
		t.Fatalf("ParseDate(today) error: %v", err)
	}
	if len(got) != 10 { // "2026-03-09"
		t.Errorf("ParseDate(today) = %q, want 10-char date", got)
	}

	got, err = ParseDate("yesterday")
	if err != nil {
		t.Fatalf("ParseDate(yesterday) error: %v", err)
	}
	if len(got) != 10 {
		t.Errorf("ParseDate(yesterday) = %q, want 10-char date", got)
	}
}

// TestBuildFromFlags_Empty verifies that no flags set = nil filter.
func TestBuildFromFlags_Empty(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil filter, got %v", got)
	}
}

// TestBuildFromFlags_SingleCreatedAfter tests a single date filter.
func TestBuildFromFlags_SingleCreatedAfter(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("created-after", "7d"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	createdAt, ok := got["createdAt"].(map[string]any)
	if !ok {
		t.Fatalf("createdAt not found or wrong type: %v", got)
	}
	if createdAt["gt"] != "-P7D" {
		t.Errorf("createdAt.gt = %v, want -P7D", createdAt["gt"])
	}
}

// TestBuildFromFlags_DateRange tests createdAt range (after and before merged into same comparator).
func TestBuildFromFlags_DateRange(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("created-after", "2026-01-01"))
	must(t, f.Set("created-before", "2026-03-01"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	createdAt, ok := got["createdAt"].(map[string]any)
	if !ok {
		t.Fatalf("createdAt not found: %v", got)
	}
	if createdAt["gt"] != "2026-01-01" {
		t.Errorf("createdAt.gt = %v, want 2026-01-01", createdAt["gt"])
	}
	if createdAt["lt"] != "2026-03-01" {
		t.Errorf("createdAt.lt = %v, want 2026-03-01", createdAt["lt"])
	}
}

// TestBuildFromFlags_CompoundAnd tests multiple filters combined as AND.
func TestBuildFromFlags_CompoundAnd(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("created-after", "7d"))
	must(t, f.Set("priority-gte", "2"))
	must(t, f.Set("no-assignee", "true"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// all three top-level keys must be present
	if _, ok := got["createdAt"]; !ok {
		t.Error("missing createdAt")
	}
	if _, ok := got["priority"]; !ok {
		t.Error("missing priority")
	}
	if _, ok := got["assignee"]; !ok {
		t.Error("missing assignee")
	}
	// no or: key
	if _, ok := got["or"]; ok {
		t.Error("unexpected or: key in AND mode")
	}
}

// TestBuildFromFlags_CompoundOr tests --or flag wrapping conditions.
func TestBuildFromFlags_CompoundOr(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("created-after", "7d"))
	must(t, f.Set("no-assignee", "true"))
	must(t, f.Set("or", "true"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	orList, ok := got["or"].([]map[string]any)
	if !ok {
		t.Fatalf("or: not found or wrong type: %v", got)
	}
	if len(orList) != 2 {
		t.Errorf("or: len = %d, want 2", len(orList))
	}
	// verify contents
	keys := map[string]bool{}
	for _, item := range orList {
		for k := range item {
			keys[k] = true
		}
	}
	if !keys["createdAt"] {
		t.Error("or: missing createdAt element")
	}
	if !keys["assignee"] {
		t.Error("or: missing assignee element")
	}
}

// TestBuildFromFlags_NullFilters tests no-assignee, no-project, no-cycle.
func TestBuildFromFlags_NullFilters(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("no-assignee", "true"))
	must(t, f.Set("no-project", "true"))
	must(t, f.Set("no-cycle", "true"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkNull := func(field string) {
		t.Helper()
		v, ok := got[field].(map[string]any)
		if !ok {
			t.Errorf("%s: not found or wrong type", field)
			return
		}
		if v["null"] != true {
			t.Errorf("%s.null = %v, want true", field, v["null"])
		}
	}
	checkNull("assignee")
	checkNull("project")
	checkNull("cycle")
}

// TestBuildFromFlags_My tests --my sets assignee.isMe filter.
func TestBuildFromFlags_My(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("my", "true"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assignee, ok := got["assignee"].(map[string]any)
	if !ok {
		t.Fatalf("assignee not found: %v", got)
	}
	isMe, ok := assignee["isMe"].(map[string]any)
	if !ok {
		t.Fatalf("assignee.isMe not found: %v", assignee)
	}
	if isMe["eq"] != true {
		t.Errorf("assignee.isMe.eq = %v, want true", isMe["eq"])
	}
}

// TestBuildFromFlags_PriorityRange tests priority-gte and priority-lte merged.
func TestBuildFromFlags_PriorityRange(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("priority-gte", "2"))
	must(t, f.Set("priority-lte", "3"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	priority, ok := got["priority"].(map[string]any)
	if !ok {
		t.Fatalf("priority not found: %v", got)
	}
	if priority["gte"] != float64(2) {
		t.Errorf("priority.gte = %v, want 2", priority["gte"])
	}
	if priority["lte"] != float64(3) {
		t.Errorf("priority.lte = %v, want 3", priority["lte"])
	}
}

// TestBuildFromFlags_DueDates tests --due-after and --due-before.
func TestBuildFromFlags_DueDates(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("due-after", "2026-01-01"))
	must(t, f.Set("due-before", "2026-06-01"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dueDate, ok := got["dueDate"].(map[string]any)
	if !ok {
		t.Fatalf("dueDate not found: %v", got)
	}
	if dueDate["gt"] != "2026-01-01" {
		t.Errorf("dueDate.gt = %v, want 2026-01-01", dueDate["gt"])
	}
	if dueDate["lt"] != "2026-06-01" {
		t.Errorf("dueDate.lt = %v, want 2026-06-01", dueDate["lt"])
	}
}

// TestBuildFromFlags_CompletedDates tests --completed-after and --completed-before.
func TestBuildFromFlags_CompletedDates(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("completed-after", "2w"))
	must(t, f.Set("completed-before", "1m"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	completedAt, ok := got["completedAt"].(map[string]any)
	if !ok {
		t.Fatalf("completedAt not found: %v", got)
	}
	if completedAt["gt"] != "-P2W" {
		t.Errorf("completedAt.gt = %v, want -P2W", completedAt["gt"])
	}
	if completedAt["lt"] != "-P1M" {
		t.Errorf("completedAt.lt = %v, want -P1M", completedAt["lt"])
	}
}

// TestBuildFromFlags_UpdatedDates tests --updated-after and --updated-before.
func TestBuildFromFlags_UpdatedDates(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("updated-after", "-P7D"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updatedAt, ok := got["updatedAt"].(map[string]any)
	if !ok {
		t.Fatalf("updatedAt not found: %v", got)
	}
	if updatedAt["gt"] != "-P7D" {
		t.Errorf("updatedAt.gt = %v, want -P7D", updatedAt["gt"])
	}
}

// TestBuildFromFlags_MyAndNoAssigneeOrMode verifies --my --no-assignee --or produces valid OR filter.
func TestBuildFromFlags_MyAndNoAssigneeOrMode(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("my", "true"))
	must(t, f.Set("no-assignee", "true"))
	must(t, f.Set("or", "true"))

	got, err := BuildFromFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	orList, ok := got["or"].([]map[string]any)
	if !ok {
		t.Fatalf("or: not found or wrong type: %v", got)
	}
	if len(orList) != 2 {
		t.Errorf("or: len = %d, want 2", len(orList))
	}
}

// TestBuildFromFlags_MyAndNoAssigneeMutualExclusion verifies --my --no-assignee without --or is an error.
func TestBuildFromFlags_MyAndNoAssigneeMutualExclusion(t *testing.T) {
	t.Parallel()
	f := newFlagSet()
	must(t, f.Set("my", "true"))
	must(t, f.Set("no-assignee", "true"))

	_, err := BuildFromFlags(f)
	if err == nil {
		t.Fatal("expected error for --my + --no-assignee without --or")
	}
}

// TestBuildFromFlags_PriorityOutOfRange verifies --priority-gte/--priority-lte > 4 returns error.
func TestBuildFromFlags_PriorityOutOfRange(t *testing.T) {
	t.Parallel()
	tests := []struct{ flag, val string }{
		{"priority-gte", "5"},
		{"priority-gte", "9"},
		{"priority-lte", "5"},
	}
	for _, tt := range tests {
		t.Run(tt.flag+"="+tt.val, func(t *testing.T) {
			t.Parallel()
			f := newFlagSet()
			must(t, f.Set(tt.flag, tt.val))
			_, err := BuildFromFlags(f)
			if err == nil {
				t.Fatalf("expected error for --%s %s", tt.flag, tt.val)
			}
		})
	}
}

// TestBuildFromFlags_PriorityBoundary verifies priority 0 and 4 are accepted.
func TestBuildFromFlags_PriorityBoundary(t *testing.T) {
	t.Parallel()
	for _, val := range []string{"0", "4"} {
		t.Run("priority-gte="+val, func(t *testing.T) {
			t.Parallel()
			f := newFlagSet()
			must(t, f.Set("priority-gte", val))
			_, err := BuildFromFlags(f)
			if err != nil {
				t.Fatalf("unexpected error for priority-gte=%s: %v", val, err)
			}
		})
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

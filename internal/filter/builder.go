package filter

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// RegisterFlags adds all issue filter flags to a pflag.FlagSet.
func RegisterFlags(f *pflag.FlagSet) {
	f.String("created-after", "", "filter by created date (e.g. 7d, 2w, 2026-03-01, -P30D)")
	f.String("created-before", "", "filter by created date")
	f.String("updated-after", "", "filter by updated date")
	f.String("updated-before", "", "filter by updated date")
	f.String("due-after", "", "filter by due date")
	f.String("due-before", "", "filter by due date")
	f.String("completed-after", "", "filter by completion date")
	f.String("completed-before", "", "filter by completion date")
	f.Bool("no-assignee", false, "filter issues with no assignee")
	f.Bool("no-project", false, "filter issues with no project")
	f.Bool("no-cycle", false, "filter issues with no cycle")
	f.Int("priority-gte", -1, "filter by minimum priority (0=none,1=urgent,2=high,3=normal,4=low)")
	f.Int("priority-lte", -1, "filter by maximum priority")
	f.Bool("my", false, "filter issues assigned to me")
	f.Bool("or", false, "combine filters with OR logic (default is AND)")
}

// AddFlags adds issue filter flags to a cobra.Command.
func AddFlags(cmd *cobra.Command) {
	RegisterFlags(cmd.Flags())
}

var durationAlias = regexp.MustCompile(`^(\d+)([dwm])$`)

// ParseDate converts a date string to the format expected by the Linear API.
// Convenience aliases (7d, 2w, 1m, today, yesterday) are expanded;
// ISO 8601 dates and durations are passed through unchanged.
func ParseDate(s string) (string, error) {
	switch s {
	case "today":
		return time.Now().Format("2006-01-02"), nil
	case "yesterday":
		return time.Now().AddDate(0, 0, -1).Format("2006-01-02"), nil
	}
	if m := durationAlias.FindStringSubmatch(s); m != nil {
		n := m[1]
		unit := strings.ToUpper(m[2])
		// d -> D, w -> W, m -> M
		return fmt.Sprintf("-P%s%s", n, unit), nil
	}
	// pass through as-is (ISO 8601 date or duration)
	return s, nil
}

type condition struct {
	field string
	value map[string]any
}

// BuildFromFlags constructs an IssueFilter map from the given pflag.FlagSet.
// Returns nil if no filter flags are set.
func BuildFromFlags(f *pflag.FlagSet) (map[string]any, error) {
	var conds []condition

	addDate := func(field, op, raw string) error {
		d, err := ParseDate(raw)
		if err != nil {
			return fmt.Errorf("parse date %q: %w", raw, err)
		}
		conds = append(conds, condition{field, map[string]any{op: d}})
		return nil
	}

	dateFlags := []struct {
		flag  string
		field string
		op    string
	}{
		{"created-after", "createdAt", "gt"},
		{"created-before", "createdAt", "lt"},
		{"updated-after", "updatedAt", "gt"},
		{"updated-before", "updatedAt", "lt"},
		{"due-after", "dueDate", "gt"},
		{"due-before", "dueDate", "lt"},
		{"completed-after", "completedAt", "gt"},
		{"completed-before", "completedAt", "lt"},
	}
	for _, df := range dateFlags {
		if v, _ := f.GetString(df.flag); v != "" {
			if err := addDate(df.field, df.op, v); err != nil {
				return nil, err
			}
		}
	}

	useOr, _ := f.GetBool("or")
	my, _ := f.GetBool("my")
	noAssignee, _ := f.GetBool("no-assignee")
	if my && noAssignee && !useOr {
		return nil, fmt.Errorf("--my and --no-assignee are mutually exclusive")
	}
	if noAssignee {
		conds = append(conds, condition{"assignee", map[string]any{"null": true}})
	}
	if v, _ := f.GetBool("no-project"); v {
		conds = append(conds, condition{"project", map[string]any{"null": true}})
	}
	if v, _ := f.GetBool("no-cycle"); v {
		conds = append(conds, condition{"cycle", map[string]any{"null": true}})
	}
	if my {
		conds = append(conds, condition{"assignee", map[string]any{"isMe": map[string]any{"eq": true}}})
	}

	intFlags := []struct {
		flag  string
		field string
		op    string
	}{
		{"priority-gte", "priority", "gte"},
		{"priority-lte", "priority", "lte"},
	}
	for _, pf := range intFlags {
		if v, _ := f.GetInt(pf.flag); v >= 0 {
			if v > 4 {
				return nil, fmt.Errorf("--%s: priority must be between 0 and 4", pf.flag)
			}
			conds = append(conds, condition{pf.field, map[string]any{pf.op: float64(v)}})
		}
	}

	if len(conds) == 0 {
		return nil, nil
	}

	if useOr {
		orList := make([]map[string]any, len(conds))
		for i, c := range conds {
			orList[i] = map[string]any{c.field: c.value}
		}
		return map[string]any{"or": orList}, nil
	}

	// AND: merge conditions for same field by combining their comparator maps
	merged := map[string]any{}
	for _, c := range conds {
		if existing, ok := merged[c.field].(map[string]any); ok {
			for k, v := range c.value {
				existing[k] = v
			}
		} else {
			merged[c.field] = c.value
		}
	}
	return merged, nil
}

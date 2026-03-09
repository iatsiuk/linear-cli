package query

import (
	"strings"
	"testing"
)

func TestStateListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "StateList"},
		{"first var", "$first: Int"},
		{"after var", "$after: String"},
		{"filter var", "$filter: WorkflowStateFilter"},
		{"workflowStates call", "workflowStates("},
		{"nodes block", "nodes {"},
		{"pageInfo block", "pageInfo {"},
		{"hasNextPage", "hasNextPage"},
		{"endCursor", "endCursor"},
		{"id field", "id"},
		{"name field", "name"},
		{"color field", "color"},
		{"type field", "type"},
		{"position field", "position"},
		{"team block", "team {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(StateListQuery, c.contain) {
				t.Errorf("StateListQuery missing %q", c.contain)
			}
		})
	}
}

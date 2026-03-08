package query

import (
	"strings"
	"testing"
)

func TestCycleListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "CycleList"},
		{"pagination var first", "$first: Int"},
		{"pagination var after", "$after: String"},
		{"filter var", "$filter: CycleFilter"},
		{"includeArchived var", "$includeArchived: Boolean"},
		{"orderBy var", "$orderBy: PaginationOrderBy"},
		{"nodes block", "nodes {"},
		{"pageInfo block", "pageInfo {"},
		{"hasNextPage", "hasNextPage"},
		{"endCursor", "endCursor"},
		{"team block", "team {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(CycleListQuery, c.contain) {
				t.Errorf("CycleListQuery missing %q", c.contain)
			}
		})
	}
}

func TestCycleGetQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "CycleGet"},
		{"id var", "$id: String!"},
		{"cycle call", "cycle(id: $id)"},
		{"team block", "team {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(CycleGetQuery, c.contain) {
				t.Errorf("CycleGetQuery missing %q", c.contain)
			}
		})
	}
}

func TestCycleFieldsPresence(t *testing.T) {
	t.Parallel()
	fields := []string{
		"id", "name", "number", "description",
		"startsAt", "endsAt", "isActive", "isFuture", "isPast",
		"progress", "createdAt", "updatedAt",
	}
	queries := map[string]string{
		"CycleListQuery": CycleListQuery,
		"CycleGetQuery":  CycleGetQuery,
	}
	for qName, q := range queries {
		for _, f := range fields {
			if !strings.Contains(q, f) {
				t.Errorf("%s missing field %q", qName, f)
			}
		}
	}
}

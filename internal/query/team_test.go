package query

import (
	"strings"
	"testing"
)

func TestTeamListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "TeamList"},
		{"nodes block", "nodes {"},
		{"pageInfo block", "pageInfo {"},
		{"hasNextPage", "hasNextPage"},
		{"endCursor", "endCursor"},
		{"id field", "id"},
		{"name field", "name"},
		{"key field", "key"},
		{"description field", "description"},
		{"cyclesEnabled field", "cyclesEnabled"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(TeamListQuery, c.contain) {
				t.Errorf("TeamListQuery missing %q", c.contain)
			}
		})
	}
}

func TestTeamGetQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "TeamGet"},
		{"id var", "$id: String!"},
		{"team call", "team(id: $id)"},
		{"id field", "id"},
		{"name field", "name"},
		{"key field", "key"},
		{"description field", "description"},
		{"cyclesEnabled field", "cyclesEnabled"},
		{"issueEstimationType field", "issueEstimationType"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(TeamGetQuery, c.contain) {
				t.Errorf("TeamGetQuery missing %q", c.contain)
			}
		})
	}
}

func TestTeamFieldsPresence(t *testing.T) {
	t.Parallel()
	fields := []string{
		"id", "name", "displayName", "key",
		"cyclesEnabled", "issueEstimationType",
		"createdAt", "updatedAt",
	}
	queries := map[string]string{
		"TeamListQuery": TeamListQuery,
		"TeamGetQuery":  TeamGetQuery,
	}
	for qName, q := range queries {
		for _, f := range fields {
			if !strings.Contains(q, f) {
				t.Errorf("%s missing field %q", qName, f)
			}
		}
	}
}

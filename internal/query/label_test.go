package query

import (
	"strings"
	"testing"
)

func TestLabelListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "LabelList"},
		{"first var", "$first: Int"},
		{"after var", "$after: String"},
		{"filter var", "$filter: IssueLabelFilter"},
		{"issueLabels call", "issueLabels("},
		{"nodes block", "nodes {"},
		{"pageInfo block", "pageInfo {"},
		{"hasNextPage", "hasNextPage"},
		{"endCursor", "endCursor"},
		{"id field", "id"},
		{"name field", "name"},
		{"color field", "color"},
		{"isGroup field", "isGroup"},
		{"team block", "team {"},
		{"parent block", "parent {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(LabelListQuery, c.contain) {
				t.Errorf("LabelListQuery missing %q", c.contain)
			}
		})
	}
}

func TestLabelCreateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "LabelCreate"},
		{"input var", "$input: IssueLabelCreateInput!"},
		{"issueLabelCreate call", "issueLabelCreate(input: $input)"},
		{"success field", "success"},
		{"issueLabel block", "issueLabel {"},
		{"name field", "name"},
		{"color field", "color"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(LabelCreateMutation, c.contain) {
				t.Errorf("LabelCreateMutation missing %q", c.contain)
			}
		})
	}
}

func TestLabelUpdateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "LabelUpdate"},
		{"id var", "$id: String!"},
		{"input var", "$input: IssueLabelUpdateInput!"},
		{"issueLabelUpdate call", "issueLabelUpdate(id: $id, input: $input)"},
		{"success field", "success"},
		{"issueLabel block", "issueLabel {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(LabelUpdateMutation, c.contain) {
				t.Errorf("LabelUpdateMutation missing %q", c.contain)
			}
		})
	}
}

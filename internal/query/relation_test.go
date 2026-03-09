package query

import (
	"strings"
	"testing"
)

func TestRelationListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "RelationList"},
		{"issueId var", "$issueId: String!"},
		{"issue call", "issue(id: $issueId)"},
		{"relations block", "relations("},
		{"inverseRelations block", "inverseRelations("},
		{"nodes block", "nodes {"},
		{"id field", "id"},
		{"type field", "type"},
		{"createdAt field", "createdAt"},
		{"issue nested", "issue {"},
		{"relatedIssue nested", "relatedIssue {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(RelationListQuery, c.contain) {
				t.Errorf("RelationListQuery missing %q", c.contain)
			}
		})
	}
}

func TestRelationCreateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "RelationCreate"},
		{"input var", "$input: IssueRelationCreateInput!"},
		{"issueRelationCreate call", "issueRelationCreate(input: $input)"},
		{"success field", "success"},
		{"issueRelation block", "issueRelation {"},
		{"type field", "type"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(RelationCreateMutation, c.contain) {
				t.Errorf("RelationCreateMutation missing %q", c.contain)
			}
		})
	}
}

func TestRelationDeleteMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "RelationDelete"},
		{"id var", "$id: String!"},
		{"issueRelationDelete call", "issueRelationDelete(id: $id)"},
		{"success field", "success"},
		{"entityId field", "entityId"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(RelationDeleteMutation, c.contain) {
				t.Errorf("RelationDeleteMutation missing %q", c.contain)
			}
		})
	}
}

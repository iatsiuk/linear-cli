package query

import (
	"strings"
	"testing"
)

func TestIssueListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "IssueList"},
		{"pagination var first", "$first: Int"},
		{"pagination var after", "$after: String"},
		{"filter var", "$filter: IssueFilter"},
		{"includeArchived var", "$includeArchived: Boolean"},
		{"orderBy var", "$orderBy: PaginationOrderBy"},
		{"nodes block", "nodes {"},
		{"pageInfo block", "pageInfo {"},
		{"hasNextPage", "hasNextPage"},
		{"endCursor", "endCursor"},
		{"identifier field", "identifier"},
		{"state block", "state {"},
		{"assignee block", "assignee {"},
		{"team block", "team {"},
		{"labels block", "labels {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(IssueListQuery, c.contain) {
				t.Errorf("IssueListQuery missing %q", c.contain)
			}
		})
	}
}

func TestIssueGetQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "IssueGet"},
		{"id var", "$id: String!"},
		{"identifier field", "identifier"},
		{"state block", "state {"},
		{"assignee block", "assignee {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(IssueGetQuery, c.contain) {
				t.Errorf("IssueGetQuery missing %q", c.contain)
			}
		})
	}
}

func TestIssueCreateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "IssueCreate"},
		{"input var", "$input: IssueCreateInput!"},
		{"issueCreate call", "issueCreate(input: $input)"},
		{"success field", "success"},
		{"issue block", "issue {"},
		{"identifier field", "identifier"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(IssueCreateMutation, c.contain) {
				t.Errorf("IssueCreateMutation missing %q", c.contain)
			}
		})
	}
}

func TestIssueUpdateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "IssueUpdate"},
		{"id var", "$id: String!"},
		{"input var", "$input: IssueUpdateInput!"},
		{"issueUpdate call", "issueUpdate(id: $id, input: $input)"},
		{"success field", "success"},
		{"issue block", "issue {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(IssueUpdateMutation, c.contain) {
				t.Errorf("IssueUpdateMutation missing %q", c.contain)
			}
		})
	}
}

func TestIssueDeleteMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "IssueDelete"},
		{"id var", "$id: String!"},
		{"issueDelete call", "issueDelete(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(IssueDeleteMutation, c.contain) {
				t.Errorf("IssueDeleteMutation missing %q", c.contain)
			}
		})
	}
}

func TestIssueArchiveMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "IssueArchive"},
		{"id var", "$id: String!"},
		{"issueArchive call", "issueArchive(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(IssueArchiveMutation, c.contain) {
				t.Errorf("IssueArchiveMutation missing %q", c.contain)
			}
		})
	}
}

func TestIssueFieldsContainsParent(t *testing.T) {
	t.Parallel()
	want := "parent { id identifier title }"
	if !strings.Contains(issueListFields, want) {
		t.Errorf("issueListFields missing %q", want)
	}
}

func TestIssueFieldsContainsProject(t *testing.T) {
	t.Parallel()
	want := "project { id name }"
	if !strings.Contains(issueListFields, want) {
		t.Errorf("issueListFields missing %q", want)
	}
}

func TestIssueFieldsPresence(t *testing.T) {
	t.Parallel()
	// all queries must include the common issue fields
	fields := []string{
		"id", "identifier", "title", "description",
		"priority", "priorityLabel", "estimate", "dueDate",
		"url", "createdAt", "updatedAt",
	}
	queries := map[string]string{
		"IssueListQuery":      IssueListQuery,
		"IssueGetQuery":       IssueGetQuery,
		"IssueCreateMutation": IssueCreateMutation,
		"IssueUpdateMutation": IssueUpdateMutation,
	}
	for qName, q := range queries {
		for _, f := range fields {
			if !strings.Contains(q, f) {
				t.Errorf("%s missing field %q", qName, f)
			}
		}
	}
}

func TestIssueListFieldsCompact(t *testing.T) {
	t.Parallel()
	// list fields must contain core fields
	wantPresent := []string{
		"id", "identifier", "title", "description",
		"priority", "priorityLabel", "estimate", "dueDate",
		"url", "createdAt", "updatedAt",
		"state { id name color type }",
		"assignee { id displayName email }",
		"team { id name key }",
		"labels { nodes { id name color } }",
		"parent { id identifier title }",
		"project { id name }",
	}
	for _, f := range wantPresent {
		if !strings.Contains(issueListFields, f) {
			t.Errorf("issueListFields missing %q", f)
		}
	}
	// detail-only fields must NOT appear in list fields
	wantAbsent := []string{
		"number", "branchName", "trashed", "customerTicketCount",
		"archivedAt", "autoArchivedAt", "autoClosedAt", "canceledAt",
		"completedAt", "startedAt", "startedTriageAt", "triagedAt",
		"snoozedUntilAt", "addedToCycleAt", "addedToProjectAt", "addedToTeamAt",
		"slaBreachesAt", "slaHighRiskAt", "slaMediumRiskAt", "slaStartedAt", "slaType",
		"creator {", "cycle {",
	}
	for _, f := range wantAbsent {
		if strings.Contains(issueListFields, f) {
			t.Errorf("issueListFields should not contain %q (detail-only field)", f)
		}
	}
}

func TestIssueDetailFieldsContainsAll(t *testing.T) {
	t.Parallel()
	// detail fields must contain all list fields
	listFields := []string{
		"id", "identifier", "title", "description",
		"priority", "priorityLabel", "estimate", "dueDate",
		"url", "createdAt", "updatedAt",
		"state { id name color type }",
		"assignee { id displayName email }",
		"team { id name key }",
		"labels { nodes { id name color } }",
		"parent { id identifier title }",
		"project { id name }",
	}
	for _, f := range listFields {
		if !strings.Contains(issueDetailFields, f) {
			t.Errorf("issueDetailFields missing list field %q", f)
		}
	}
	// detail fields must also contain detail-only fields
	detailOnly := []string{
		"number", "branchName", "trashed", "customerTicketCount",
		"archivedAt", "autoArchivedAt", "autoClosedAt", "canceledAt",
		"completedAt", "startedAt", "startedTriageAt", "triagedAt",
		"snoozedUntilAt", "addedToCycleAt", "addedToProjectAt", "addedToTeamAt",
		"slaBreachesAt", "slaHighRiskAt", "slaMediumRiskAt", "slaStartedAt", "slaType",
		"creator { id displayName email }",
		"cycle { id name number }",
	}
	for _, f := range detailOnly {
		if !strings.Contains(issueDetailFields, f) {
			t.Errorf("issueDetailFields missing detail field %q", f)
		}
	}
}

package query

import (
	"strings"
	"testing"
)

func TestProjectListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "ProjectList"},
		{"pagination var first", "$first: Int"},
		{"pagination var after", "$after: String"},
		{"filter var", "$filter: ProjectFilter"},
		{"includeArchived var", "$includeArchived: Boolean"},
		{"orderBy var", "$orderBy: PaginationOrderBy"},
		{"nodes block", "nodes {"},
		{"pageInfo block", "pageInfo {"},
		{"hasNextPage", "hasNextPage"},
		{"endCursor", "endCursor"},
		{"status block", "status {"},
		{"teams block", "teams {"},
		{"creator block", "creator {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ProjectListQuery, c.contain) {
				t.Errorf("ProjectListQuery missing %q", c.contain)
			}
		})
	}
}

func TestProjectGetQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "ProjectGet"},
		{"id var", "$id: String!"},
		{"project call", "project(id: $id)"},
		{"status block", "status {"},
		{"teams block", "teams {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ProjectGetQuery, c.contain) {
				t.Errorf("ProjectGetQuery missing %q", c.contain)
			}
		})
	}
}

func TestProjectCreateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "ProjectCreate"},
		{"input var", "$input: ProjectCreateInput!"},
		{"projectCreate call", "projectCreate(input: $input)"},
		{"success field", "success"},
		{"project block", "project {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ProjectCreateMutation, c.contain) {
				t.Errorf("ProjectCreateMutation missing %q", c.contain)
			}
		})
	}
}

func TestProjectUpdateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "ProjectUpdate"},
		{"id var", "$id: String!"},
		{"input var", "$input: ProjectUpdateInput!"},
		{"projectUpdate call", "projectUpdate(id: $id, input: $input)"},
		{"success field", "success"},
		{"project block", "project {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ProjectUpdateMutation, c.contain) {
				t.Errorf("ProjectUpdateMutation missing %q", c.contain)
			}
		})
	}
}

func TestProjectDeleteMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "ProjectDelete"},
		{"id var", "$id: String!"},
		{"projectDelete call", "projectDelete(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ProjectDeleteMutation, c.contain) {
				t.Errorf("ProjectDeleteMutation missing %q", c.contain)
			}
		})
	}
}

func TestProjectFieldsPresence(t *testing.T) {
	t.Parallel()
	fields := []string{
		"id", "name", "description", "color",
		"health", "progress", "startDate", "targetDate",
		"url", "createdAt", "updatedAt",
	}
	queries := map[string]string{
		"ProjectListQuery":      ProjectListQuery,
		"ProjectGetQuery":       ProjectGetQuery,
		"ProjectCreateMutation": ProjectCreateMutation,
		"ProjectUpdateMutation": ProjectUpdateMutation,
	}
	for qName, q := range queries {
		for _, f := range fields {
			if !strings.Contains(q, f) {
				t.Errorf("%s missing field %q", qName, f)
			}
		}
	}
}

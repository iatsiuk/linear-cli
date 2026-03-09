package query

import (
	"strings"
	"testing"
)

func TestProjectUpdateListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "ProjectUpdateList"},
		{"projectId var", "$projectId: String!"},
		{"first var", "$first: Int"},
		{"project call", "project(id: $projectId)"},
		{"projectUpdates call", "projectUpdates("},
		{"nodes block", "nodes {"},
		{"id field", "id"},
		{"body field", "body"},
		{"health field", "health"},
		{"user field", "user {"},
		{"createdAt field", "createdAt"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ProjectUpdateListQuery, c.contain) {
				t.Errorf("ProjectUpdateListQuery missing %q", c.contain)
			}
		})
	}
}

func TestProjectUpdateCreateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "ProjectUpdateCreate"},
		{"input var", "$input: ProjectUpdateCreateInput!"},
		{"projectUpdateCreate call", "projectUpdateCreate(input: $input)"},
		{"success field", "success"},
		{"projectUpdate field", "projectUpdate {"},
		{"body field", "body"},
		{"health field", "health"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ProjectUpdateCreateMutation, c.contain) {
				t.Errorf("ProjectUpdateCreateMutation missing %q", c.contain)
			}
		})
	}
}

func TestProjectUpdateArchiveMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "ProjectUpdateArchive"},
		{"id var", "$id: String!"},
		{"projectUpdateArchive call", "projectUpdateArchive(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ProjectUpdateArchiveMutation, c.contain) {
				t.Errorf("ProjectUpdateArchiveMutation missing %q", c.contain)
			}
		})
	}
}

package query

import (
	"strings"
	"testing"
)

func TestProjectSearchQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "SearchProjects"},
		{"term var", "$term: String!"},
		{"first var", "$first: Int"},
		{"searchProjects call", "searchProjects("},
		{"nodes block", "nodes {"},
		{"id field", "id"},
		{"name field", "name"},
		{"status field", "status"},
		{"url field", "url"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ProjectSearchQuery, c.contain) {
				t.Errorf("ProjectSearchQuery missing %q", c.contain)
			}
		})
	}
}

func TestDocumentSearchQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "SearchDocuments"},
		{"term var", "$term: String!"},
		{"first var", "$first: Int"},
		{"searchDocuments call", "searchDocuments("},
		{"nodes block", "nodes {"},
		{"id field", "id"},
		{"title field", "title"},
		{"url field", "url"},
		{"project field", "project"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(DocumentSearchQuery, c.contain) {
				t.Errorf("DocumentSearchQuery missing %q", c.contain)
			}
		})
	}
}

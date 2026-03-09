package query

import (
	"strings"
	"testing"
)

func TestDocumentListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "DocumentList"},
		{"first var", "$first: Int"},
		{"after var", "$after: String"},
		{"filter var", "$filter: DocumentFilter"},
		{"includeArchived var", "$includeArchived: Boolean"},
		{"documents call", "documents("},
		{"nodes block", "nodes {"},
		{"pageInfo block", "pageInfo {"},
		{"hasNextPage", "hasNextPage"},
		{"endCursor", "endCursor"},
		{"id field", "id"},
		{"title field", "title"},
		{"content field", "content"},
		{"creator block", "creator {"},
		{"project block", "project {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(DocumentListQuery, c.contain) {
				t.Errorf("DocumentListQuery missing %q", c.contain)
			}
		})
	}
}

func TestDocumentGetQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "DocumentGet"},
		{"id var", "$id: String!"},
		{"document call", "document(id: $id)"},
		{"title field", "title"},
		{"content field", "content"},
		{"creator block", "creator {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(DocumentGetQuery, c.contain) {
				t.Errorf("DocumentGetQuery missing %q", c.contain)
			}
		})
	}
}

func TestDocumentCreateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "DocumentCreate"},
		{"input var", "$input: DocumentCreateInput!"},
		{"documentCreate call", "documentCreate(input: $input)"},
		{"success field", "success"},
		{"document block", "document {"},
		{"title field", "title"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(DocumentCreateMutation, c.contain) {
				t.Errorf("DocumentCreateMutation missing %q", c.contain)
			}
		})
	}
}

func TestDocumentUpdateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "DocumentUpdate"},
		{"id var", "$id: String!"},
		{"input var", "$input: DocumentUpdateInput!"},
		{"documentUpdate call", "documentUpdate(id: $id, input: $input)"},
		{"success field", "success"},
		{"document block", "document {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(DocumentUpdateMutation, c.contain) {
				t.Errorf("DocumentUpdateMutation missing %q", c.contain)
			}
		})
	}
}

func TestDocumentDeleteMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "DocumentDelete"},
		{"id var", "$id: String!"},
		{"documentDelete call", "documentDelete(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(DocumentDeleteMutation, c.contain) {
				t.Errorf("DocumentDeleteMutation missing %q", c.contain)
			}
		})
	}
}

func TestDocumentUnarchiveMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "DocumentUnarchive"},
		{"id var", "$id: String!"},
		{"documentUnarchive call", "documentUnarchive(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(DocumentUnarchiveMutation, c.contain) {
				t.Errorf("DocumentUnarchiveMutation missing %q", c.contain)
			}
		})
	}
}

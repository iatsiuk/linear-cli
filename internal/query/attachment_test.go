package query

import (
	"strings"
	"testing"
)

func TestAttachmentListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "AttachmentList"},
		{"issueId var", "$issueId: String!"},
		{"issue field", "issue(id: $issueId)"},
		{"attachments block", "attachments("},
		{"nodes block", "nodes {"},
		{"id field", "id"},
		{"title field", "title"},
		{"url field", "url"},
		{"createdAt field", "createdAt"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(AttachmentListQuery, c.contain) {
				t.Errorf("AttachmentListQuery missing %q", c.contain)
			}
		})
	}
}

func TestAttachmentCreateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "AttachmentCreate"},
		{"input var", "$input: AttachmentCreateInput!"},
		{"attachmentCreate call", "attachmentCreate(input: $input)"},
		{"attachment block", "attachment {"},
		{"title field", "title"},
		{"url field", "url"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(AttachmentCreateMutation, c.contain) {
				t.Errorf("AttachmentCreateMutation missing %q", c.contain)
			}
		})
	}
}

func TestAttachmentShowQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "AttachmentShow"},
		{"id var", "$id: String!"},
		{"attachment call", "attachment(id: $id)"},
		{"url field", "url"},
		{"title field", "title"},
		{"creator field", "creator {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(AttachmentShowQuery, c.contain) {
				t.Errorf("AttachmentShowQuery missing %q", c.contain)
			}
		})
	}
}

func TestAttachmentDeleteMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "AttachmentDelete"},
		{"id var", "$id: String!"},
		{"attachmentDelete call", "attachmentDelete(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(AttachmentDeleteMutation, c.contain) {
				t.Errorf("AttachmentDeleteMutation missing %q", c.contain)
			}
		})
	}
}

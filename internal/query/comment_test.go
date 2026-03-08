package query

import (
	"strings"
	"testing"
)

func TestCommentListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "CommentList"},
		{"issueId var", "$issueId: String!"},
		{"first var", "$first: Int"},
		{"after var", "$after: String"},
		{"issue call", "issue(id: $issueId)"},
		{"comments block", "comments("},
		{"nodes block", "nodes {"},
		{"pageInfo block", "pageInfo {"},
		{"hasNextPage", "hasNextPage"},
		{"endCursor", "endCursor"},
		{"id field", "id"},
		{"body field", "body"},
		{"createdAt field", "createdAt"},
		{"user block", "user {"},
		{"parent block", "parent {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(CommentListQuery, c.contain) {
				t.Errorf("CommentListQuery missing %q", c.contain)
			}
		})
	}
}

func TestCommentCreateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "CommentCreate"},
		{"input var", "$input: CommentCreateInput!"},
		{"commentCreate call", "commentCreate(input: $input)"},
		{"success field", "success"},
		{"comment block", "comment {"},
		{"body field", "body"},
		{"user block", "user {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(CommentCreateMutation, c.contain) {
				t.Errorf("CommentCreateMutation missing %q", c.contain)
			}
		})
	}
}

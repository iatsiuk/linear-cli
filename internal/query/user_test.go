package query

import (
	"strings"
	"testing"
)

func TestUserListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "UserList"},
		{"includeDisabled var", "$includeDisabled: Boolean"},
		{"nodes block", "nodes {"},
		{"pageInfo block", "pageInfo {"},
		{"hasNextPage", "hasNextPage"},
		{"endCursor", "endCursor"},
		{"id field", "id"},
		{"email field", "email"},
		{"displayName field", "displayName"},
		{"active field", "active"},
		{"admin field", "admin"},
		{"guest field", "guest"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(UserListQuery, c.contain) {
				t.Errorf("UserListQuery missing %q", c.contain)
			}
		})
	}
}

func TestUserGetQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "UserGet"},
		{"id var", "$id: String!"},
		{"user call", "user(id: $id)"},
		{"id field", "id"},
		{"email field", "email"},
		{"displayName field", "displayName"},
		{"active field", "active"},
		{"admin field", "admin"},
		{"guest field", "guest"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(UserGetQuery, c.contain) {
				t.Errorf("UserGetQuery missing %q", c.contain)
			}
		})
	}
}

func TestViewerQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "Viewer"},
		{"viewer call", "viewer {"},
		{"id field", "id"},
		{"email field", "email"},
		{"displayName field", "displayName"},
		{"active field", "active"},
		{"admin field", "admin"},
		{"guest field", "guest"},
		{"teams block", "teams {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ViewerQuery, c.contain) {
				t.Errorf("ViewerQuery missing %q", c.contain)
			}
		})
	}
}

func TestViewerAssignedIssuesQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "ViewerAssignedIssues"},
		{"viewer call", "viewer {"},
		{"assignedIssues block", "assignedIssues {"},
		{"nodes block", "nodes {"},
		{"identifier field", "identifier"},
		{"title field", "title"},
		{"state block", "state {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ViewerAssignedIssuesQuery, c.contain) {
				t.Errorf("ViewerAssignedIssuesQuery missing %q", c.contain)
			}
		})
	}
}

func TestViewerCreatedIssuesQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "ViewerCreatedIssues"},
		{"viewer call", "viewer {"},
		{"createdIssues block", "createdIssues {"},
		{"nodes block", "nodes {"},
		{"identifier field", "identifier"},
		{"title field", "title"},
		{"state block", "state {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(ViewerCreatedIssuesQuery, c.contain) {
				t.Errorf("ViewerCreatedIssuesQuery missing %q", c.contain)
			}
		})
	}
}

func TestUserFieldsPresence(t *testing.T) {
	t.Parallel()
	fields := []string{
		"id", "email", "displayName",
		"active", "admin", "guest",
		"createdAt", "updatedAt",
	}
	queries := map[string]string{
		"UserListQuery": UserListQuery,
		"UserGetQuery":  UserGetQuery,
	}
	for qName, q := range queries {
		for _, f := range fields {
			if !strings.Contains(q, f) {
				t.Errorf("%s missing field %q", qName, f)
			}
		}
	}
}

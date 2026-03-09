package query

import (
	"strings"
	"testing"
)

func TestTeamMemberListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "TeamMemberList"},
		{"teamId var", "$teamId: String!"},
		{"first var", "$first: Int"},
		{"team call", "team(id: $teamId)"},
		{"memberships call", "memberships("},
		{"nodes block", "nodes {"},
		{"id field", "id"},
		{"owner field", "owner"},
		{"user block", "user {"},
		{"displayName field", "displayName"},
		{"email field", "email"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(TeamMemberListQuery, c.contain) {
				t.Errorf("TeamMemberListQuery missing %q", c.contain)
			}
		})
	}
}

func TestTeamMemberAddMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "TeamMemberAdd"},
		{"input var", "$input: TeamMembershipCreateInput!"},
		{"teamMembershipCreate call", "teamMembershipCreate(input: $input)"},
		{"success field", "success"},
		{"teamMembership field", "teamMembership {"},
		{"user block", "user {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(TeamMemberAddMutation, c.contain) {
				t.Errorf("TeamMemberAddMutation missing %q", c.contain)
			}
		})
	}
}

func TestTeamMemberRemoveMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "TeamMemberRemove"},
		{"id var", "$id: String!"},
		{"teamMembershipDelete call", "teamMembershipDelete(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(TeamMemberRemoveMutation, c.contain) {
				t.Errorf("TeamMemberRemoveMutation missing %q", c.contain)
			}
		})
	}
}

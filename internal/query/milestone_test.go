package query

import (
	"strings"
	"testing"
)

func TestMilestoneListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "MilestoneList"},
		{"projectId var", "$projectId: String!"},
		{"first var", "$first: Int"},
		{"project call", "project(id: $projectId)"},
		{"projectMilestones call", "projectMilestones("},
		{"nodes block", "nodes {"},
		{"id field", "id"},
		{"name field", "name"},
		{"targetDate field", "targetDate"},
		{"status field", "status"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(MilestoneListQuery, c.contain) {
				t.Errorf("MilestoneListQuery missing %q", c.contain)
			}
		})
	}
}

func TestMilestoneCreateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "MilestoneCreate"},
		{"input var", "$input: ProjectMilestoneCreateInput!"},
		{"projectMilestoneCreate call", "projectMilestoneCreate(input: $input)"},
		{"success field", "success"},
		{"projectMilestone field", "projectMilestone {"},
		{"name field", "name"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(MilestoneCreateMutation, c.contain) {
				t.Errorf("MilestoneCreateMutation missing %q", c.contain)
			}
		})
	}
}

func TestMilestoneUpdateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "MilestoneUpdate"},
		{"id var", "$id: String!"},
		{"input var", "$input: ProjectMilestoneUpdateInput!"},
		{"projectMilestoneUpdate call", "projectMilestoneUpdate(id: $id, input: $input)"},
		{"success field", "success"},
		{"projectMilestone field", "projectMilestone {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(MilestoneUpdateMutation, c.contain) {
				t.Errorf("MilestoneUpdateMutation missing %q", c.contain)
			}
		})
	}
}

func TestMilestoneDeleteMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "MilestoneDelete"},
		{"id var", "$id: String!"},
		{"projectMilestoneDelete call", "projectMilestoneDelete(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(MilestoneDeleteMutation, c.contain) {
				t.Errorf("MilestoneDeleteMutation missing %q", c.contain)
			}
		})
	}
}

package query

import (
	"strings"
	"testing"
)

func TestInitiativeListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "InitiativeList"},
		{"first var", "$first: Int"},
		{"initiatives call", "initiatives("},
		{"nodes block", "nodes {"},
		{"id field", "id"},
		{"name field", "name"},
		{"status field", "status"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(InitiativeListQuery, c.contain) {
				t.Errorf("InitiativeListQuery missing %q", c.contain)
			}
		})
	}
}

func TestInitiativeShowQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "InitiativeShow"},
		{"id var", "$id: String!"},
		{"initiative call", "initiative(id: $id)"},
		{"id field", "id"},
		{"name field", "name"},
		{"status field", "status"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(InitiativeShowQuery, c.contain) {
				t.Errorf("InitiativeShowQuery missing %q", c.contain)
			}
		})
	}
}

func TestInitiativeCreateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "InitiativeCreate"},
		{"input var", "$input: InitiativeCreateInput!"},
		{"initiativeCreate call", "initiativeCreate(input: $input)"},
		{"success field", "success"},
		{"initiative field", "initiative {"},
		{"name field", "name"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(InitiativeCreateMutation, c.contain) {
				t.Errorf("InitiativeCreateMutation missing %q", c.contain)
			}
		})
	}
}

func TestInitiativeUpdateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "InitiativeUpdate"},
		{"id var", "$id: String!"},
		{"input var", "$input: InitiativeUpdateInput!"},
		{"initiativeUpdate call", "initiativeUpdate(id: $id, input: $input)"},
		{"success field", "success"},
		{"initiative field", "initiative {"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(InitiativeUpdateMutation, c.contain) {
				t.Errorf("InitiativeUpdateMutation missing %q", c.contain)
			}
		})
	}
}

func TestInitiativeDeleteMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "InitiativeDelete"},
		{"id var", "$id: String!"},
		{"initiativeDelete call", "initiativeDelete(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(InitiativeDeleteMutation, c.contain) {
				t.Errorf("InitiativeDeleteMutation missing %q", c.contain)
			}
		})
	}
}

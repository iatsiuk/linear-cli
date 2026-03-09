package query

import (
	"strings"
	"testing"
)

func TestOrganizationQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "Organization"},
		{"organization call", "organization {"},
		{"id field", "id"},
		{"name field", "name"},
		{"urlKey field", "urlKey"},
		{"logoUrl field", "logoUrl"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(OrganizationQuery, c.contain) {
				t.Errorf("OrganizationQuery missing %q", c.contain)
			}
		})
	}
}

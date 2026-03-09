package query

import (
	"strings"
	"testing"
)

func TestTemplateListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "TemplateList"},
		{"templates call", "templates {"},
		{"id field", "id"},
		{"name field", "name"},
		{"type field", "type"},
		{"description field", "description"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(TemplateListQuery, c.contain) {
				t.Errorf("TemplateListQuery missing %q", c.contain)
			}
		})
	}
}

func TestTemplateShowQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "TemplateShow"},
		{"id var", "$id: String!"},
		{"template call", "template(id: $id)"},
		{"id field", "id"},
		{"name field", "name"},
		{"type field", "type"},
		{"templateData field", "templateData"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(TemplateShowQuery, c.contain) {
				t.Errorf("TemplateShowQuery missing %q", c.contain)
			}
		})
	}
}

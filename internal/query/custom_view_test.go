package query

import (
	"strings"
	"testing"
)

func TestCustomViewListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "CustomViewList"},
		{"first var", "$first: Int"},
		{"customViews call", "customViews("},
		{"nodes block", "nodes {"},
		{"id field", "id"},
		{"name field", "name"},
		{"shared field", "shared"},
		{"modelName field", "modelName"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(CustomViewListQuery, c.contain) {
				t.Errorf("CustomViewListQuery missing %q", c.contain)
			}
		})
	}
}

func TestCustomViewShowQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "CustomViewShow"},
		{"id var", "$id: String!"},
		{"customView call", "customView(id: $id)"},
		{"id field", "id"},
		{"name field", "name"},
		{"shared field", "shared"},
		{"modelName field", "modelName"},
		{"filterData field", "filterData"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(CustomViewShowQuery, c.contain) {
				t.Errorf("CustomViewShowQuery missing %q", c.contain)
			}
		})
	}
}

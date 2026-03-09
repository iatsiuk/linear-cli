package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func orgResponse(org map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"organization": org,
		},
	}
}

func makeOrg(id, name, urlKey string, logoURL *string) map[string]any {
	m := map[string]any{
		"id":     id,
		"name":   name,
		"urlKey": urlKey,
	}
	if logoURL != nil {
		m["logoUrl"] = *logoURL
	} else {
		m["logoUrl"] = nil
	}
	return m
}

func TestOrgCommand_TableOutput(t *testing.T) {
	logo := "https://example.com/logo.png"
	org := makeOrg("org-1", "Acme Corp", "acme", &logo)

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, orgResponse(org))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"org"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, want := range []string{"Acme Corp", "acme", logo} {
		if !strings.Contains(result, want) {
			t.Errorf("output should contain %q, got:\n%s", want, result)
		}
	}
}

func TestOrgCommand_JSONOutput(t *testing.T) {
	org := makeOrg("org-1", "Acme Corp", "acme", nil)

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, orgResponse(org))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "org"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "Acme Corp" {
		t.Errorf("expected name Acme Corp, got %v", decoded["name"])
	}
	if decoded["urlKey"] != "acme" {
		t.Errorf("expected urlKey acme, got %v", decoded["urlKey"])
	}
}

func TestOrgCommand_NoLogo(t *testing.T) {
	org := makeOrg("org-1", "Acme Corp", "acme", nil)

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, orgResponse(org))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"org"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if strings.Contains(result, "Logo") {
		t.Errorf("output should not contain Logo when logoUrl is nil, got:\n%s", result)
	}
	if !strings.Contains(result, "Acme Corp") {
		t.Errorf("output should contain org name, got:\n%s", result)
	}
}

package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func makeTeam(id, name, key, description string, cyclesEnabled bool) map[string]any {
	return map[string]any{
		"id":                  id,
		"name":                name,
		"displayName":         name,
		"description":         description,
		"icon":                nil,
		"color":               "#000000",
		"key":                 key,
		"cyclesEnabled":       cyclesEnabled,
		"issueEstimationType": "notUsed",
		"createdAt":           "2026-01-01T00:00:00Z",
		"updatedAt":           "2026-01-02T00:00:00Z",
	}
}

func teamListResponse(teams []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"teams": map[string]any{
				"nodes":    teams,
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			},
		},
	}
}

func teamGetResponse(team map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"team": team,
		},
	}
}

func TestTeamListCommand_TableOutput(t *testing.T) {
	teams := []map[string]any{
		makeTeam("t1", "Engineering", "ENG", "Core engineering team", true),
		makeTeam("t2", "Design", "DES", "Design team", false),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, teamListResponse(teams))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Engineering") {
		t.Errorf("output should contain Engineering, got:\n%s", result)
	}
	if !strings.Contains(result, "ENG") {
		t.Errorf("output should contain ENG, got:\n%s", result)
	}
	if !strings.Contains(result, "Design") {
		t.Errorf("output should contain Design, got:\n%s", result)
	}
	if !strings.Contains(result, "DES") {
		t.Errorf("output should contain DES, got:\n%s", result)
	}
}

func TestTeamListCommand_TableHeaders(t *testing.T) {
	teams := []map[string]any{
		makeTeam("t1", "Engineering", "ENG", "Core team", true),
	}
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, teamListResponse(teams))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, col := range []string{"NAME", "KEY", "DESCRIPTION", "CYCLES"} {
		if !strings.Contains(result, col) {
			t.Errorf("output should contain %s column header, got:\n%s", col, result)
		}
	}
}

func TestTeamListCommand_JSONOutput(t *testing.T) {
	teams := []map[string]any{
		makeTeam("t1", "Engineering", "ENG", "Core engineering team", true),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, teamListResponse(teams))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "team", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 team, got %d", len(decoded))
	}
	if decoded[0]["name"] != "Engineering" {
		t.Errorf("expected name Engineering, got %v", decoded[0]["name"])
	}
	if decoded[0]["key"] != "ENG" {
		t.Errorf("expected key ENG, got %v", decoded[0]["key"])
	}
}

func TestTeamShowCommand_TableOutput(t *testing.T) {
	desc := "Core engineering team"
	team := makeTeam("t1", "Engineering", "ENG", desc, true)

	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if strings.Contains(body.Query, "ResolveTeam") {
			writeJSONResponse(w, map[string]any{
				"data": map[string]any{
					"teams": map[string]any{
						"nodes": []map[string]any{{"id": "t1"}},
					},
				},
			})
			return
		}
		writeJSONResponse(w, teamGetResponse(team))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "show", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Engineering") {
		t.Errorf("output should contain name, got:\n%s", result)
	}
	if !strings.Contains(result, "ENG") {
		t.Errorf("output should contain key, got:\n%s", result)
	}
	if !strings.Contains(result, "Core engineering team") {
		t.Errorf("output should contain description, got:\n%s", result)
	}
	if !strings.Contains(result, "true") {
		t.Errorf("output should contain cycles enabled=true, got:\n%s", result)
	}
}

func TestTeamShowCommand_NotFound(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{"data": map[string]any{"team": nil}})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "show", "MISSING"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing team, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestTeamShowCommand_JSONOutput(t *testing.T) {
	team := makeTeam("t1", "Engineering", "ENG", "Core engineering team", true)

	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if strings.Contains(body.Query, "ResolveTeam") {
			writeJSONResponse(w, map[string]any{
				"data": map[string]any{
					"teams": map[string]any{
						"nodes": []map[string]any{{"id": "t1"}},
					},
				},
			})
			return
		}
		writeJSONResponse(w, teamGetResponse(team))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "team", "show", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "Engineering" {
		t.Errorf("expected name Engineering, got %v", decoded["name"])
	}
	if decoded["key"] != "ENG" {
		t.Errorf("expected key ENG, got %v", decoded["key"])
	}
}

func TestTeamShowCommand_RequiresArg(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "show"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no arg provided")
	}
}

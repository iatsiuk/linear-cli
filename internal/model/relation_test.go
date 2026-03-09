package model

import (
	"encoding/json"
	"testing"
)

func TestIssueRelationDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "rel-1",
		"type": "blocks",
		"createdAt": "2026-01-01T10:00:00.000Z",
		"updatedAt": "2026-01-01T10:00:00.000Z",
		"issue": {
			"id": "iss-1",
			"identifier": "ENG-1",
			"title": "Issue 1",
			"priority": 1,
			"priorityLabel": "Urgent",
			"url": "https://linear.app/issue/ENG-1",
			"createdAt": "2026-01-01T00:00:00.000Z",
			"updatedAt": "2026-01-01T00:00:00.000Z",
			"state": {"id": "s1", "name": "In Progress", "color": "#aaa", "type": "started"},
			"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
			"labels": {"nodes": []}
		},
		"relatedIssue": {
			"id": "iss-2",
			"identifier": "ENG-2",
			"title": "Issue 2",
			"priority": 0,
			"priorityLabel": "No priority",
			"url": "https://linear.app/issue/ENG-2",
			"createdAt": "2026-01-01T00:00:00.000Z",
			"updatedAt": "2026-01-01T00:00:00.000Z",
			"state": {"id": "s2", "name": "Backlog", "color": "#ccc", "type": "backlog"},
			"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
			"labels": {"nodes": []}
		}
	}`

	var rel IssueRelation
	if err := json.Unmarshal([]byte(raw), &rel); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if rel.ID != "rel-1" {
		t.Errorf("ID: got %q, want %q", rel.ID, "rel-1")
	}
	if rel.Type != "blocks" {
		t.Errorf("Type: got %q, want %q", rel.Type, "blocks")
	}
	if rel.CreatedAt != "2026-01-01T10:00:00.000Z" {
		t.Errorf("CreatedAt: got %q", rel.CreatedAt)
	}
	if rel.Issue.Identifier != "ENG-1" {
		t.Errorf("Issue.Identifier: got %q, want %q", rel.Issue.Identifier, "ENG-1")
	}
	if rel.RelatedIssue.Identifier != "ENG-2" {
		t.Errorf("RelatedIssue.Identifier: got %q, want %q", rel.RelatedIssue.Identifier, "ENG-2")
	}
}

func TestIssueRelationConnectionDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"nodes": [
			{
				"id": "rel-1",
				"type": "related",
				"createdAt": "2026-01-01T00:00:00.000Z",
				"updatedAt": "2026-01-01T00:00:00.000Z",
				"issue": {
					"id": "iss-1", "identifier": "ENG-1", "title": "A",
					"priority": 0, "priorityLabel": "No priority",
					"url": "https://linear.app/issue/ENG-1",
					"createdAt": "2026-01-01T00:00:00.000Z",
					"updatedAt": "2026-01-01T00:00:00.000Z",
					"state": {"id": "s1", "name": "Backlog", "color": "#ccc", "type": "backlog"},
					"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
					"labels": {"nodes": []}
				},
				"relatedIssue": {
					"id": "iss-2", "identifier": "ENG-2", "title": "B",
					"priority": 0, "priorityLabel": "No priority",
					"url": "https://linear.app/issue/ENG-2",
					"createdAt": "2026-01-01T00:00:00.000Z",
					"updatedAt": "2026-01-01T00:00:00.000Z",
					"state": {"id": "s2", "name": "Backlog", "color": "#ccc", "type": "backlog"},
					"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
					"labels": {"nodes": []}
				}
			}
		]
	}`

	var conn IssueRelationConnection
	if err := json.Unmarshal([]byte(raw), &conn); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(conn.Nodes) != 1 {
		t.Fatalf("Nodes: got %d, want 1", len(conn.Nodes))
	}
	if conn.Nodes[0].Type != "related" {
		t.Errorf("Nodes[0].Type: got %q, want %q", conn.Nodes[0].Type, "related")
	}
}

package model

import (
	"encoding/json"
	"testing"
)

func TestTeamMembershipDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "tm-1",
		"owner": true,
		"sortOrder": 1.0,
		"user": {"id": "u-1", "displayName": "Alice", "email": "alice@example.com", "active": true, "admin": false, "guest": false, "isMe": false, "createdAt": "2025-01-01", "updatedAt": "2025-01-01"},
		"team": {"id": "t-1", "name": "Engineering", "displayName": "Engineering", "key": "ENG", "cyclesEnabled": false, "issueEstimationType": "points", "createdAt": "2025-01-01", "updatedAt": "2025-01-01"},
		"createdAt": "2025-01-01",
		"updatedAt": "2025-01-01"
	}`

	var tm TeamMembership
	if err := json.Unmarshal([]byte(raw), &tm); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if tm.ID != "tm-1" {
		t.Errorf("ID: got %q, want tm-1", tm.ID)
	}
	if !tm.Owner {
		t.Errorf("Owner: got false, want true")
	}
	if tm.User.DisplayName != "Alice" {
		t.Errorf("User.DisplayName: got %q, want Alice", tm.User.DisplayName)
	}
	if tm.User.Email != "alice@example.com" {
		t.Errorf("User.Email: got %q, want alice@example.com", tm.User.Email)
	}
	if tm.Team.Key != "ENG" {
		t.Errorf("Team.Key: got %q, want ENG", tm.Team.Key)
	}
}

func TestTeamMembershipDeserialization_Member(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "tm-2",
		"owner": false,
		"sortOrder": 2.0,
		"user": {"id": "u-2", "displayName": "Bob", "email": "bob@example.com", "active": true, "admin": false, "guest": false, "isMe": false, "createdAt": "2025-01-01", "updatedAt": "2025-01-01"},
		"team": {"id": "t-1", "name": "Engineering", "displayName": "Engineering", "key": "ENG", "cyclesEnabled": false, "issueEstimationType": "points", "createdAt": "2025-01-01", "updatedAt": "2025-01-01"},
		"createdAt": "2025-01-02",
		"updatedAt": "2025-01-02"
	}`

	var tm TeamMembership
	if err := json.Unmarshal([]byte(raw), &tm); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if tm.Owner {
		t.Errorf("Owner: got true, want false")
	}
	if tm.User.DisplayName != "Bob" {
		t.Errorf("User.DisplayName: got %q, want Bob", tm.User.DisplayName)
	}
}

func TestTeamMembershipConnectionDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"nodes": [
			{"id": "tm-1", "owner": true, "sortOrder": 1.0,
			 "user": {"id": "u-1", "displayName": "Alice", "email": "alice@example.com", "active": true, "admin": false, "guest": false, "isMe": false, "createdAt": "2025-01-01", "updatedAt": "2025-01-01"},
			 "team": {"id": "t-1", "name": "Eng", "displayName": "Eng", "key": "ENG", "cyclesEnabled": false, "issueEstimationType": "points", "createdAt": "2025-01-01", "updatedAt": "2025-01-01"},
			 "createdAt": "2025-01-01", "updatedAt": "2025-01-01"},
			{"id": "tm-2", "owner": false, "sortOrder": 2.0,
			 "user": {"id": "u-2", "displayName": "Bob", "email": "bob@example.com", "active": true, "admin": false, "guest": false, "isMe": false, "createdAt": "2025-01-01", "updatedAt": "2025-01-01"},
			 "team": {"id": "t-1", "name": "Eng", "displayName": "Eng", "key": "ENG", "cyclesEnabled": false, "issueEstimationType": "points", "createdAt": "2025-01-01", "updatedAt": "2025-01-01"},
			 "createdAt": "2025-01-02", "updatedAt": "2025-01-02"}
		]
	}`

	var conn TeamMembershipConnection
	if err := json.Unmarshal([]byte(raw), &conn); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(conn.Nodes) != 2 {
		t.Fatalf("Nodes: got %d, want 2", len(conn.Nodes))
	}
	if conn.Nodes[0].User.DisplayName != "Alice" {
		t.Errorf("Nodes[0].User.DisplayName: got %q, want Alice", conn.Nodes[0].User.DisplayName)
	}
	if conn.Nodes[1].Owner {
		t.Errorf("Nodes[1].Owner: got true, want false")
	}
}

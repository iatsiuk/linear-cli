package model

import (
	"encoding/json"
	"testing"
)

func TestUserDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "u1",
		"email": "alice@example.com",
		"displayName": "Alice",
		"avatarUrl": "https://cdn.example.com/alice.png",
		"active": true,
		"admin": false,
		"guest": false,
		"isMe": true,
		"createdAt": "2025-01-01T00:00:00.000Z",
		"updatedAt": "2025-06-01T00:00:00.000Z"
	}`

	var user User
	if err := json.Unmarshal([]byte(raw), &user); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if user.ID != "u1" {
		t.Errorf("ID: got %q, want %q", user.ID, "u1")
	}
	if user.Email != "alice@example.com" {
		t.Errorf("Email: got %q", user.Email)
	}
	if user.DisplayName != "Alice" {
		t.Errorf("DisplayName: got %q", user.DisplayName)
	}
	if user.AvatarURL == nil || *user.AvatarURL != "https://cdn.example.com/alice.png" {
		t.Errorf("AvatarURL: unexpected value")
	}
	if !user.Active {
		t.Errorf("Active: got false, want true")
	}
	if user.Admin {
		t.Errorf("Admin: got true, want false")
	}
	if user.Guest {
		t.Errorf("Guest: got true, want false")
	}
	if !user.IsMe {
		t.Errorf("IsMe: got false, want true")
	}
	if user.CreatedAt != "2025-01-01T00:00:00.000Z" {
		t.Errorf("CreatedAt: got %q", user.CreatedAt)
	}
	if user.UpdatedAt != "2025-06-01T00:00:00.000Z" {
		t.Errorf("UpdatedAt: got %q", user.UpdatedAt)
	}
}

func TestUserNullableFields(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "u2",
		"email": "bob@example.com",
		"displayName": "Bob",
		"active": false,
		"admin": true,
		"guest": false,
		"isMe": false,
		"createdAt": "2025-01-01T00:00:00.000Z",
		"updatedAt": "2025-01-01T00:00:00.000Z"
	}`

	var user User
	if err := json.Unmarshal([]byte(raw), &user); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if user.AvatarURL != nil {
		t.Errorf("AvatarURL should be nil, got %v", user.AvatarURL)
	}
	if user.Active {
		t.Errorf("Active: got true, want false")
	}
	if !user.Admin {
		t.Errorf("Admin: got false, want true")
	}
}

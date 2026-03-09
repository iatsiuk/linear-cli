package query

import (
	"strings"
	"testing"
)

func TestNotificationListQuery(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "NotificationList"},
		{"first var", "$first: Int"},
		{"notifications call", "notifications("},
		{"nodes block", "nodes {"},
		{"id field", "id"},
		{"type field", "type"},
		{"readAt field", "readAt"},
		{"archivedAt field", "archivedAt"},
		{"createdAt field", "createdAt"},
		{"title field", "title"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(NotificationListQuery, c.contain) {
				t.Errorf("NotificationListQuery missing %q", c.contain)
			}
		})
	}
}

func TestNotificationUpdateMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "NotificationUpdate"},
		{"id var", "$id: String!"},
		{"input var", "$input: NotificationUpdateInput!"},
		{"notificationUpdate call", "notificationUpdate(id: $id, input: $input)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(NotificationUpdateMutation, c.contain) {
				t.Errorf("NotificationUpdateMutation missing %q", c.contain)
			}
		})
	}
}

func TestNotificationMarkReadAllMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "NotificationMarkReadAll"},
		{"input var", "$input: NotificationEntityInput!"},
		{"readAt var", "$readAt: DateTime!"},
		{"notificationMarkReadAll call", "notificationMarkReadAll("},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(NotificationMarkReadAllMutation, c.contain) {
				t.Errorf("NotificationMarkReadAllMutation missing %q", c.contain)
			}
		})
	}
}

func TestNotificationArchiveAllMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "NotificationArchiveAll"},
		{"input var", "$input: NotificationEntityInput!"},
		{"notificationArchiveAll call", "notificationArchiveAll(input: $input)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(NotificationArchiveAllMutation, c.contain) {
				t.Errorf("NotificationArchiveAllMutation missing %q", c.contain)
			}
		})
	}
}

func TestNotificationArchiveMutation(t *testing.T) {
	t.Parallel()
	checks := []struct {
		name    string
		contain string
	}{
		{"operation name", "NotificationArchive"},
		{"id var", "$id: String!"},
		{"notificationArchive call", "notificationArchive(id: $id)"},
		{"success field", "success"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(NotificationArchiveMutation, c.contain) {
				t.Errorf("NotificationArchiveMutation missing %q", c.contain)
			}
		})
	}
}

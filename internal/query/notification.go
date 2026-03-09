package query

// notificationFields is the common field selection for Notification.
const notificationFields = `
	id
	type
	readAt
	archivedAt
	createdAt
	updatedAt
	title
	subtitle
	url
`

// NotificationListQuery fetches notifications for the current user.
const NotificationListQuery = `
query NotificationList($first: Int) {
	notifications(first: $first) {
		nodes {` + notificationFields + `}
	}
}
`

// NotificationUpdateMutation updates a single notification (e.g. mark as read).
const NotificationUpdateMutation = `
mutation NotificationUpdate($id: String!, $input: NotificationUpdateInput!) {
	notificationUpdate(id: $id, input: $input) {
		success
		notification {` + notificationFields + `}
	}
}
`

// NotificationMarkReadAllMutation marks all notifications for an entity as read.
const NotificationMarkReadAllMutation = `
mutation NotificationMarkReadAll($input: NotificationEntityInput!, $readAt: DateTime!) {
	notificationMarkReadAll(input: $input, readAt: $readAt) {
		success
	}
}
`

// NotificationArchiveAllMutation archives all notifications for an entity.
const NotificationArchiveAllMutation = `
mutation NotificationArchiveAll($input: NotificationEntityInput!) {
	notificationArchiveAll(input: $input) {
		success
	}
}
`

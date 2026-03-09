package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type notificationListResult struct {
	Notifications model.NotificationConnection `json:"notifications"`
}

type notificationUpdateResult struct {
	NotificationUpdate struct {
		Success      bool                `json:"success"`
		Notification *model.Notification `json:"notification"`
	} `json:"notificationUpdate"`
}

type notificationArchiveResult struct {
	NotificationArchive struct {
		Success bool `json:"success"`
	} `json:"notificationArchive"`
}

type notificationRow struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Created string `json:"created"`
	Read    string `json:"read"`
}

func newNotificationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notification",
		Short: "Manage notifications",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newNotificationListCommand())
	cmd.AddCommand(newNotificationReadCommand())
	cmd.AddCommand(newNotificationArchiveCommand())
	return cmd
}

func newNotificationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List notifications",
		RunE:  runNotificationList,
	}
	f := cmd.Flags()
	f.Bool("unread", false, "show only unread notifications")
	f.Int("limit", 50, "maximum number of notifications to fetch")
	return cmd
}

func runNotificationList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	f := cmd.Flags()
	limit, _ := f.GetInt("limit")
	unreadOnly, _ := f.GetBool("unread")

	vars := map[string]any{"first": limit}
	var result notificationListResult
	if err := client.Do(context.Background(), query.NotificationListQuery, vars, &result); err != nil {
		return fmt.Errorf("list notifications: %w", err)
	}

	notifications := result.Notifications.Nodes
	if unreadOnly {
		var filtered []model.Notification
		for _, n := range notifications {
			if n.ReadAt == nil {
				filtered = append(filtered, n)
			}
		}
		notifications = filtered
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), notifications)
	}

	rows := make([]notificationRow, len(notifications))
	for i, n := range notifications {
		readStr := ""
		if n.ReadAt != nil {
			readStr = *n.ReadAt
		}
		rows[i] = notificationRow{
			ID:      n.ID,
			Type:    n.Type,
			Created: n.CreatedAt,
			Read:    readStr,
		}
	}
	return output.NewFormatter(false).Format(cmd.OutOrStdout(), rows)
}

func newNotificationReadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read [id]",
		Short: "Mark notification(s) as read",
		RunE:  runNotificationRead,
	}
	cmd.Flags().Bool("all", false, "mark all notifications as read")
	return cmd
}

func runNotificationRead(cmd *cobra.Command, args []string) error {
	allFlag, _ := cmd.Flags().GetBool("all")

	if !allFlag && len(args) == 0 {
		return fmt.Errorf("notification id is required, or use --all")
	}

	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	if allFlag {
		type markReadAllResult struct {
			NotificationMarkReadAll struct {
				Success bool `json:"success"`
			} `json:"notificationMarkReadAll"`
		}
		vars := map[string]any{
			"input":  map[string]any{},
			"readAt": time.Now().UTC().Format(time.RFC3339),
		}
		var result markReadAllResult
		if err := client.Do(ctx, query.NotificationMarkReadAllMutation, vars, &result); err != nil {
			return fmt.Errorf("mark all read: %w", err)
		}
		if !result.NotificationMarkReadAll.Success {
			return fmt.Errorf("mark all read: mutation returned success=false")
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "All notifications marked as read.")
		return err
	}

	id := args[0]
	vars := map[string]any{
		"id": id,
		"input": map[string]any{
			"readAt": time.Now().UTC().Format(time.RFC3339),
		},
	}
	var result notificationUpdateResult
	if err := client.Do(ctx, query.NotificationUpdateMutation, vars, &result); err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	if !result.NotificationUpdate.Success {
		return fmt.Errorf("mark read: mutation returned success=false")
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Notification %s marked as read.\n", id)
	return err
}

func newNotificationArchiveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive [id]",
		Short: "Archive notification(s)",
		RunE:  runNotificationArchive,
	}
	cmd.Flags().Bool("all", false, "archive all notifications")
	return cmd
}

func runNotificationArchive(cmd *cobra.Command, args []string) error {
	allFlag, _ := cmd.Flags().GetBool("all")

	if !allFlag && len(args) == 0 {
		return fmt.Errorf("notification id is required, or use --all")
	}

	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	if allFlag {
		type archiveAllResult struct {
			NotificationArchiveAll struct {
				Success bool `json:"success"`
			} `json:"notificationArchiveAll"`
		}
		vars := map[string]any{"input": map[string]any{}}
		var result archiveAllResult
		if err := client.Do(ctx, query.NotificationArchiveAllMutation, vars, &result); err != nil {
			return fmt.Errorf("archive all: %w", err)
		}
		if !result.NotificationArchiveAll.Success {
			return fmt.Errorf("archive all: mutation returned success=false")
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "All notifications archived.")
		return err
	}

	id := args[0]
	var result notificationArchiveResult
	if err := client.Do(ctx, query.NotificationArchiveMutation, map[string]any{"id": id}, &result); err != nil {
		return fmt.Errorf("archive notification: %w", err)
	}
	if !result.NotificationArchive.Success {
		return fmt.Errorf("archive notification: mutation returned success=false")
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Notification %s archived.\n", id)
	return err
}

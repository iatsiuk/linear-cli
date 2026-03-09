package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"linear-cli/internal/api"
	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type attachmentListResult struct {
	Issue *struct {
		Attachments struct {
			Nodes    []model.Attachment `json:"nodes"`
			PageInfo api.PageInfo       `json:"pageInfo"`
		} `json:"attachments"`
	} `json:"issue"`
}

type attachmentCreateResult struct {
	AttachmentCreate struct {
		Attachment *model.Attachment `json:"attachment"`
	} `json:"attachmentCreate"`
}

type attachmentDeleteResult struct {
	AttachmentDelete struct {
		Success bool `json:"success"`
	} `json:"attachmentDelete"`
}

// AttachmentRow is a display row for the attachment list table.
type AttachmentRow struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Created string `json:"created"`
}

func newAttachmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment",
		Short: "Manage Linear issue attachments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newAttachmentListCommand())
	cmd.AddCommand(newAttachmentCreateCommand())
	cmd.AddCommand(newAttachmentDeleteCommand())
	return cmd
}

func newAttachmentListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <issue-identifier>",
		Short: "List attachments for an issue",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("issue identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runAttachmentList,
	}
}

func runAttachmentList(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	identifier := args[0]
	ctx := context.Background()

	attachments, err := api.PaginateAll(ctx, func(ctx context.Context, _ *string, _ int) (api.Connection[model.Attachment], error) {
		vars := map[string]any{"issueId": identifier}
		var result attachmentListResult
		if err := client.Do(ctx, query.AttachmentListQuery, vars, &result); err != nil {
			return api.Connection[model.Attachment]{}, err
		}
		if result.Issue == nil {
			return api.Connection[model.Attachment]{}, fmt.Errorf("issue %q not found", identifier)
		}
		return api.Connection[model.Attachment]{
			Nodes:    result.Issue.Attachments.Nodes,
			PageInfo: result.Issue.Attachments.PageInfo,
		}, nil
	}, 50, 0)
	if err != nil {
		return fmt.Errorf("list attachments: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), attachments)
	}

	rows := make([]AttachmentRow, len(attachments))
	for i, a := range attachments {
		rows[i] = AttachmentRow{
			Title:   truncate(a.Title, 40),
			URL:     truncate(a.URL, 50),
			Created: a.CreatedAt,
		}
	}
	return output.NewFormatter(false).Format(cmd.OutOrStdout(), rows)
}

func newAttachmentCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <issue-identifier>",
		Short: "Create an attachment for an issue",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("issue identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runAttachmentCreate,
	}
	f := cmd.Flags()
	f.String("url", "", "attachment URL (required)")
	f.String("title", "", "attachment title (required)")
	_ = cmd.MarkFlagRequired("url")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func runAttachmentCreate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	identifier := args[0]
	f := cmd.Flags()
	url, _ := f.GetString("url")
	title, _ := f.GetString("title")

	input := map[string]any{
		"issueId": identifier,
		"url":     url,
		"title":   title,
	}

	vars := map[string]any{"input": input}
	var result attachmentCreateResult
	if err := client.Do(ctx, query.AttachmentCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create attachment: %w", err)
	}
	if result.AttachmentCreate.Attachment == nil {
		return fmt.Errorf("create attachment: no attachment in response")
	}

	a := result.AttachmentCreate.Attachment
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), a)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", a.ID)
	return err
}

func newAttachmentDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an attachment",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("attachment id is required")
			}
			return nil
		},
		RunE: runAttachmentDelete,
	}
	f := cmd.Flags()
	f.Bool("yes", false, "skip confirmation prompt")
	return cmd
}

func runAttachmentDelete(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	attID := args[0]
	yes, _ := cmd.Flags().GetBool("yes")

	if !yes {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete attachment %s? [y/N] ", attID)
		scanner := bufio.NewScanner(cmd.InOrStdin())
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())
		if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
			return fmt.Errorf("aborted")
		}
	}

	var result attachmentDeleteResult
	if err := client.Do(ctx, query.AttachmentDeleteMutation, map[string]any{"id": attID}, &result); err != nil {
		return fmt.Errorf("delete attachment: %w", err)
	}
	if !result.AttachmentDelete.Success {
		return fmt.Errorf("delete attachment: mutation returned success=false")
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Attachment %s deleted.\n", attID)
	return nil
}

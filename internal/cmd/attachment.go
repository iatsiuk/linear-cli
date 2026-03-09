package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type attachmentListResult struct {
	Issue *struct {
		Attachments struct {
			Nodes    []model.Attachment `json:"nodes"`
			PageInfo struct {
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"attachments"`
	} `json:"issue"`
}

type attachmentCreateResult struct {
	AttachmentCreate struct {
		Success    bool              `json:"success"`
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

	vars := map[string]any{"issueId": identifier}
	var listResult attachmentListResult
	if err := client.Do(ctx, query.AttachmentListQuery, vars, &listResult); err != nil {
		return fmt.Errorf("list attachments: %w", err)
	}
	if listResult.Issue == nil {
		return fmt.Errorf("issue %q not found", identifier)
	}
	attachments := listResult.Issue.Attachments.Nodes
	if listResult.Issue.Attachments.PageInfo.HasNextPage {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: showing first 50 attachments; more exist\n")
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
	f.String("url", "", "attachment URL (mutually exclusive with --file)")
	f.String("title", "", "attachment title (required)")
	f.String("file", "", "local file to upload and attach")
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
	urlFlag, _ := f.GetString("url")
	title, _ := f.GetString("title")
	fileFlag, _ := f.GetString("file")

	if urlFlag == "" && fileFlag == "" {
		return fmt.Errorf("one of --url or --file is required")
	}
	if urlFlag != "" && fileFlag != "" {
		return fmt.Errorf("--url and --file are mutually exclusive")
	}

	attachmentURL := urlFlag
	if fileFlag != "" {
		assetURL, uploadErr := client.Upload(ctx, fileFlag)
		if uploadErr != nil {
			return fmt.Errorf("upload file: %w", uploadErr)
		}
		attachmentURL = assetURL
	}

	input := map[string]any{
		"issueId": identifier,
		"url":     attachmentURL,
		"title":   title,
	}

	vars := map[string]any{"input": input}
	var result attachmentCreateResult
	if err := client.Do(ctx, query.AttachmentCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create attachment: %w", err)
	}
	if !result.AttachmentCreate.Success {
		return fmt.Errorf("create attachment: mutation returned success=false")
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
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Are you sure you want to delete attachment %s? [y/N] ", attID)
		scanner := bufio.NewScanner(cmd.InOrStdin())
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("read confirmation: %w", err)
		}
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

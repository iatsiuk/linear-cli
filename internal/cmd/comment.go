package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type commentListResult struct {
	Issue *struct {
		Comments struct {
			Nodes []model.Comment `json:"nodes"`
		} `json:"comments"`
	} `json:"issue"`
}

type commentCreateResult struct {
	CommentCreate struct {
		Success bool           `json:"success"`
		Comment *model.Comment `json:"comment"`
	} `json:"commentCreate"`
}

// CommentRow is a display row for the comment list table.
type CommentRow struct {
	Author string `json:"author"`
	Date   string `json:"date"`
	Body   string `json:"body"`
}

func newCommentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Manage Linear comments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newCommentListCommand())
	cmd.AddCommand(newCommentCreateCommand())
	return cmd
}

func newCommentListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <issue-identifier>",
		Short: "List comments for an issue",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("issue identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runCommentList,
	}
}

func runCommentList(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	identifier := args[0]
	ctx := context.Background()

	vars := map[string]any{"issueId": identifier, "first": 250}
	var result commentListResult
	if err := client.Do(ctx, query.CommentListQuery, vars, &result); err != nil {
		return fmt.Errorf("list comments: %w", err)
	}
	if result.Issue == nil {
		return fmt.Errorf("issue %q not found", identifier)
	}

	comments := result.Issue.Comments.Nodes
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")

	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), comments)
	}

	rows := make([]CommentRow, len(comments))
	for i, c := range comments {
		author := ""
		if c.User != nil {
			author = c.User.DisplayName
		}
		prefix := ""
		if c.Parent != nil {
			prefix = "> "
		}
		rows[i] = CommentRow{
			Author: author,
			Date:   c.CreatedAt,
			Body:   prefix + truncate(c.Body, 60),
		}
	}
	return output.NewFormatter(false).Format(cmd.OutOrStdout(), rows)
}

func newCommentCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <issue-identifier>",
		Short: "Create a comment on an issue",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("issue identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runCommentCreate,
	}
	f := cmd.Flags()
	f.String("body", "", "comment body in markdown (required)")
	f.String("parent", "", "parent comment ID for threading")
	_ = cmd.MarkFlagRequired("body")
	return cmd
}

func runCommentCreate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	identifier := args[0]

	// fetch issue to get its UUID
	var getResult issueGetResult
	if err := client.Do(ctx, query.IssueGetQuery, map[string]any{"id": identifier}, &getResult); err != nil {
		return fmt.Errorf("get issue: %w", err)
	}
	if getResult.Issue == nil {
		return fmt.Errorf("issue %q not found", identifier)
	}

	f := cmd.Flags()
	body, _ := f.GetString("body")
	parentID, _ := f.GetString("parent")

	input := map[string]any{
		"issueId": getResult.Issue.ID,
		"body":    body,
	}
	if parentID != "" {
		input["parentId"] = parentID
	}

	vars := map[string]any{"input": input}
	var result commentCreateResult
	if err := client.Do(ctx, query.CommentCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	if !result.CommentCreate.Success {
		return fmt.Errorf("create comment: mutation returned success=false")
	}
	if result.CommentCreate.Comment == nil {
		return fmt.Errorf("create comment: no comment in response")
	}

	c := result.CommentCreate.Comment
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), c)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", c.ID)
	return err
}

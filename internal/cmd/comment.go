package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iatsiuk/linear-cli/internal/api"
	"github.com/iatsiuk/linear-cli/internal/model"
	"github.com/iatsiuk/linear-cli/internal/output"
	"github.com/iatsiuk/linear-cli/internal/query"
)

type commentListResult struct {
	Issue *struct {
		Comments struct {
			Nodes    []model.Comment `json:"nodes"`
			PageInfo api.PageInfo    `json:"pageInfo"`
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

	comments, err := api.PaginateAll(ctx, func(ctx context.Context, after *string, first int) (api.Connection[model.Comment], error) {
		vars := map[string]any{"issueId": identifier, "first": first}
		if after != nil {
			vars["after"] = *after
		}
		var result commentListResult
		if err := client.Do(ctx, query.CommentListQuery, vars, &result); err != nil {
			return api.Connection[model.Comment]{}, err
		}
		if result.Issue == nil {
			return api.Connection[model.Comment]{}, fmt.Errorf("issue %q not found", identifier)
		}
		return api.Connection[model.Comment]{
			Nodes:    result.Issue.Comments.Nodes,
			PageInfo: result.Issue.Comments.PageInfo,
		}, nil
	}, 50, 0)
	if err != nil {
		return fmt.Errorf("list comments: %w", err)
	}
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

	f := cmd.Flags()
	body, _ := f.GetString("body")
	parentID, _ := f.GetString("parent")

	input := map[string]any{
		"issueId": identifier,
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

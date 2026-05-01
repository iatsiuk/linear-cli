package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/iatsiuk/linear-cli/internal/api"
	"github.com/iatsiuk/linear-cli/internal/model"
	"github.com/iatsiuk/linear-cli/internal/output"
	"github.com/iatsiuk/linear-cli/internal/query"
)

// readBody returns the comment body from --body or --body-file. When
// --body-file is "-" it reads from cmd's stdin, otherwise from the file.
func readBody(cmd *cobra.Command) (string, error) {
	f := cmd.Flags()
	if f.Changed("body-file") {
		path, _ := f.GetString("body-file")
		if path == "-" {
			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return "", fmt.Errorf("read body file %q: %w", path, err)
			}
			return string(data), nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read body file %q: %w", path, err)
		}
		return string(data), nil
	}
	body, _ := f.GetString("body")
	return body, nil
}

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

type commentUpdateResult struct {
	CommentUpdate struct {
		Success bool           `json:"success"`
		Comment *model.Comment `json:"comment"`
	} `json:"commentUpdate"`
}

type commentDeleteResult struct {
	CommentDelete struct {
		Success bool `json:"success"`
	} `json:"commentDelete"`
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
	cmd.AddCommand(newCommentUpdateCmd())
	cmd.AddCommand(newCommentDeleteCmd())
	return cmd
}

func newCommentUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <comment-id>",
		Short: "Update a comment",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("comment ID is required")
			}
			return nil
		},
		RunE: runCommentUpdate,
	}
	f := cmd.Flags()
	f.String("body", "", "new comment body in markdown")
	f.String("body-file", "", "read comment body from file ('-' for stdin)")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	cmd.MarkFlagsOneRequired("body", "body-file")
	return cmd
}

func runCommentUpdate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := cmd.Context()

	id := args[0]
	body, err := readBody(cmd)
	if err != nil {
		return err
	}

	vars := map[string]any{
		"id":    id,
		"input": map[string]any{"body": body},
	}
	var result commentUpdateResult
	if err := client.Do(ctx, query.CommentUpdateMutation, vars, &result); err != nil {
		return fmt.Errorf("update comment: %w", err)
	}
	if !result.CommentUpdate.Success {
		return fmt.Errorf("update comment: mutation returned success=false")
	}
	if result.CommentUpdate.Comment == nil {
		return fmt.Errorf("update comment: no comment in response")
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), result.CommentUpdate.Comment)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Comment %s updated.\n", id)
	return err
}

func newCommentDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <comment-id>",
		Short: "Delete a comment",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("comment ID is required")
			}
			return nil
		},
		RunE: runCommentDelete,
	}
	cmd.Flags().Bool("yes", false, "skip confirmation prompt")
	return cmd
}

func runCommentDelete(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := cmd.Context()

	id := args[0]
	yes, _ := cmd.Flags().GetBool("yes")

	if !yes {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Delete comment %s? [y/N] ", id)
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

	var result commentDeleteResult
	if err := client.Do(ctx, query.CommentDeleteMutation, map[string]any{"id": id}, &result); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if !result.CommentDelete.Success {
		return fmt.Errorf("delete comment: mutation returned success=false")
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Comment %s deleted.\n", id)
	return err
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
	ctx := cmd.Context()

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
	f.String("body", "", "comment body in markdown")
	f.String("body-file", "", "read comment body from file ('-' for stdin)")
	f.String("parent", "", "parent comment ID for threading")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	cmd.MarkFlagsOneRequired("body", "body-file")
	return cmd
}

func runCommentCreate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := cmd.Context()

	identifier := args[0]

	body, err := readBody(cmd)
	if err != nil {
		return err
	}
	parentID, _ := cmd.Flags().GetString("parent")

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

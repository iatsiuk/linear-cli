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

type customViewListResult struct {
	CustomViews model.CustomViewConnection `json:"customViews"`
}

type customViewShowResult struct {
	CustomView *model.CustomView `json:"customView"`
}

type customViewIssuesResult struct {
	CustomView *struct {
		Issues struct {
			Nodes []model.Issue `json:"nodes"`
		} `json:"issues"`
	} `json:"customView"`
}

// CustomViewRow is a display row for the custom view list table.
type CustomViewRow struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Shared string `json:"shared"`
}

func newViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Manage custom views",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newViewListCommand())
	cmd.AddCommand(newViewShowCommand())
	cmd.AddCommand(newViewIssuesCommand())
	return cmd
}

func newViewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom views",
		RunE:  runViewList,
	}
	cmd.Flags().Int("limit", 50, "maximum number of views to return")
	return cmd
}

func runViewList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	vars := map[string]any{"first": limit}
	var result customViewListResult
	if err := client.Do(context.Background(), query.CustomViewListQuery, vars, &result); err != nil {
		return fmt.Errorf("list custom views: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.CustomViews.Nodes)
	}

	rows := make([]CustomViewRow, len(result.CustomViews.Nodes))
	for i, v := range result.CustomViews.Nodes {
		shared := "no"
		if v.Shared {
			shared = "yes"
		}
		rows[i] = CustomViewRow{
			Name:   truncate(v.Name, 50),
			Type:   v.ModelName,
			Shared: shared,
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newViewShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <view>",
		Short: "Show custom view details (accepts name, UUID, or URL slug)",
		Long:  "Show details for a custom view. Accepts a name, UUID, or URL slug.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("view id is required")
			}
			return nil
		},
		RunE: runViewShow,
	}
}

func runViewShow(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()
	id, err := api.ResolveCustomViewID(ctx, client, args[0])
	if err != nil {
		return err
	}

	vars := map[string]any{"id": id}
	var result customViewShowResult
	if err := client.Do(ctx, query.CustomViewShowQuery, vars, &result); err != nil {
		return fmt.Errorf("show custom view: %w", err)
	}
	if result.CustomView == nil {
		return fmt.Errorf("custom view not found: %s", args[0])
	}

	v := result.CustomView
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), v)
	}

	w := cmd.OutOrStdout()
	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-14s %s\n", label+":", value)
		return err
	}

	if err := writeLine("Name", v.Name); err != nil {
		return err
	}
	if err := writeLine("Type", v.ModelName); err != nil {
		return err
	}
	shared := "no"
	if v.Shared {
		shared = "yes"
	}
	if err := writeLine("Shared", shared); err != nil {
		return err
	}
	if v.Description != nil {
		if err := writeLine("Description", *v.Description); err != nil {
			return err
		}
	}
	if len(v.FilterData) > 0 && string(v.FilterData) != "null" {
		if err := writeLine("Filters", string(v.FilterData)); err != nil {
			return err
		}
	}
	return nil
}

func newViewIssuesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issues <view>",
		Short: "List issues in a custom view (accepts name, UUID, or URL slug)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("view id is required")
			}
			return nil
		},
		RunE: runViewIssues,
	}
	cmd.Flags().Int("limit", 50, "maximum number of issues to return")
	cmd.Flags().String("order-by", "updatedAt", "sort order (createdAt|updatedAt)")
	cmd.Flags().Bool("include-archived", false, "include archived issues")
	return cmd
}

func runViewIssues(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	f := cmd.Flags()
	limit, _ := f.GetInt("limit")
	if limit <= 0 {
		return fmt.Errorf("--limit must be greater than 0")
	}
	orderBy, _ := f.GetString("order-by")
	includeArchived, _ := f.GetBool("include-archived")

	ctx := context.Background()
	id, err := api.ResolveCustomViewID(ctx, client, args[0])
	if err != nil {
		return err
	}

	vars := map[string]any{
		"id":    id,
		"first": limit,
	}
	if orderBy != "" {
		vars["orderBy"] = orderBy
	}
	if includeArchived {
		vars["includeArchived"] = true
	}

	var result customViewIssuesResult
	if err := client.Do(ctx, query.ViewIssuesQuery, vars, &result); err != nil {
		return fmt.Errorf("view issues: %w", err)
	}
	if result.CustomView == nil {
		return fmt.Errorf("custom view not found: %s", args[0])
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	issues := result.CustomView.Issues.Nodes
	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), issues)
	}

	rows := make([]IssueRow, len(issues))
	for i, issue := range issues {
		rows[i] = IssueRow{
			ID:       issue.Identifier,
			Title:    truncate(issue.Title, 40),
			Status:   issue.State.Name,
			Priority: issue.PriorityLabel,
		}
		if issue.Assignee != nil {
			rows[i].Assignee = issue.Assignee.DisplayName
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

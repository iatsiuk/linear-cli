package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"linear-cli/internal/api"
	"linear-cli/internal/config"
	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

func newIssueCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage Linear issues",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newIssueListCommand())
	return cmd
}

// IssueRow is a display row for the issue list table.
type IssueRow struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
	Assignee string `json:"assignee"`
}

type issueListResult struct {
	Issues struct {
		Nodes []model.Issue `json:"nodes"`
	} `json:"issues"`
}

func newIssueListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		RunE:  runIssueList,
	}
	f := cmd.Flags()
	f.String("team", "", "filter by team key")
	f.String("assignee", "", "filter by assignee display name")
	f.String("state", "", "filter by state name")
	f.Int("priority", -1, "filter by priority (0-4)")
	f.Int("limit", 50, "maximum number of issues to return")
	f.Bool("include-archived", false, "include archived issues")
	f.String("order-by", "updatedAt", "sort order (createdAt|updatedAt)")
	return cmd
}

func runIssueList(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("not authenticated: run 'linear auth' first")
	}

	var opts []api.Option
	if ep := os.Getenv("LINEAR_API_ENDPOINT"); ep != "" {
		opts = append(opts, api.WithEndpoint(ep))
	}
	client := api.NewClient(cfg.APIKey, opts...)

	f := cmd.Flags()
	limit, _ := f.GetInt("limit")
	includeArchived, _ := f.GetBool("include-archived")
	orderBy, _ := f.GetString("order-by")
	teamKey, _ := f.GetString("team")
	assignee, _ := f.GetString("assignee")
	stateName, _ := f.GetString("state")
	priority, _ := f.GetInt("priority")

	vars := map[string]any{"first": limit}
	if includeArchived {
		vars["includeArchived"] = true
	}
	if orderBy != "" {
		vars["orderBy"] = orderBy
	}

	filter := map[string]any{}
	if teamKey != "" {
		filter["team"] = map[string]any{"key": map[string]any{"eq": teamKey}}
	}
	if assignee != "" {
		filter["assignee"] = map[string]any{"displayName": map[string]any{"eq": assignee}}
	}
	if stateName != "" {
		filter["state"] = map[string]any{"name": map[string]any{"eq": stateName}}
	}
	if priority >= 0 {
		filter["priority"] = map[string]any{"eq": float64(priority)}
	}
	if len(filter) > 0 {
		vars["filter"] = filter
	}

	var result issueListResult
	if err := client.Do(context.Background(), query.IssueListQuery, vars, &result); err != nil {
		return fmt.Errorf("list issues: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.Issues.Nodes)
	}

	rows := make([]IssueRow, len(result.Issues.Nodes))
	for i, issue := range result.Issues.Nodes {
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

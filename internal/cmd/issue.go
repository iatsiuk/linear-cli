package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"linear-cli/internal/api"
	"linear-cli/internal/config"
	"linear-cli/internal/filter"
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
	cmd.AddCommand(newIssueShowCommand())
	cmd.AddCommand(newIssueCreateCommand())
	cmd.AddCommand(newIssueUpdateCommand())
	cmd.AddCommand(newIssueDeleteCommand())
	cmd.AddCommand(newIssueBatchCommand())
	cmd.AddCommand(newRelationCommand())
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
	filter.AddFlags(cmd)
	return cmd
}

func runIssueList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	f := cmd.Flags()
	limit, _ := f.GetInt("limit")
	includeArchived, _ := f.GetBool("include-archived")
	orderBy, _ := f.GetString("order-by")
	teamKey, _ := f.GetString("team")
	assignee, _ := f.GetString("assignee")
	stateName, _ := f.GetString("state")
	priority, _ := f.GetInt("priority")

	if limit <= 0 {
		return fmt.Errorf("--limit must be greater than 0")
	}
	vars := map[string]any{"first": limit}
	if includeArchived {
		vars["includeArchived"] = true
	}
	if orderBy != "" {
		vars["orderBy"] = orderBy
	}

	useOr, _ := f.GetBool("or")
	issueFilter := map[string]any{}
	if teamKey != "" {
		issueFilter["team"] = map[string]any{"key": map[string]any{"eq": teamKey}}
	}
	if assignee != "" {
		if !useOr {
			if my, _ := f.GetBool("my"); my {
				return fmt.Errorf("--assignee and --my are mutually exclusive")
			}
			if noAssignee, _ := f.GetBool("no-assignee"); noAssignee {
				return fmt.Errorf("--assignee and --no-assignee are mutually exclusive")
			}
		}
		issueFilter["assignee"] = map[string]any{"displayName": map[string]any{"eq": assignee}}
	}
	if stateName != "" {
		issueFilter["state"] = map[string]any{"name": map[string]any{"eq": stateName}}
	}
	if priority >= 0 {
		if !useOr {
			if gte, _ := f.GetInt("priority-gte"); gte >= 0 {
				return fmt.Errorf("--priority and --priority-gte are mutually exclusive")
			}
			if lte, _ := f.GetInt("priority-lte"); lte >= 0 {
				return fmt.Errorf("--priority and --priority-lte are mutually exclusive")
			}
		}
		issueFilter["priority"] = map[string]any{"eq": float64(priority)}
	}

	advancedFilter, err := filter.BuildFromFlags(f)
	if err != nil {
		return fmt.Errorf("build filter: %w", err)
	}

	if useOr {
		// collect all conditions (base + advanced) into a single OR list
		var orList []map[string]any
		for k, v := range issueFilter {
			orList = append(orList, map[string]any{k: v})
		}
		if adOr, ok := advancedFilter["or"].([]map[string]any); ok {
			orList = append(orList, adOr...)
		}
		if len(orList) > 0 {
			vars["filter"] = map[string]any{"or": orList}
		}
	} else {
		for k, v := range advancedFilter {
			if existing, ok := issueFilter[k].(map[string]any); ok {
				if srcMap, ok2 := v.(map[string]any); ok2 {
					for sk, sv := range srcMap {
						existing[sk] = sv
					}
					issueFilter[k] = existing
					continue
				}
			}
			issueFilter[k] = v
		}
		if len(issueFilter) > 0 {
			vars["filter"] = issueFilter
		}
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
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-3]) + "..."
}

func newClientFromConfig() (*api.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("not authenticated: run 'linear auth' first")
	}
	var opts []api.Option
	if ep := os.Getenv("LINEAR_API_ENDPOINT"); ep != "" {
		opts = append(opts, api.WithEndpoint(ep))
	}
	return api.NewClient(cfg.APIKey, opts...), nil
}

func printIssueRow(cmd *cobra.Command, issue *model.Issue) error {
	rows := []IssueRow{{
		ID:       issue.Identifier,
		Title:    truncate(issue.Title, 40),
		Status:   issue.State.Name,
		Priority: issue.PriorityLabel,
	}}
	if issue.Assignee != nil {
		rows[0].Assignee = issue.Assignee.DisplayName
	}
	return output.NewFormatter(false).Format(cmd.OutOrStdout(), rows)
}

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"linear-cli/internal/api"
	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

// SearchRow is a display row for the search results table.
type SearchRow struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Team   string `json:"team"`
}

type searchResult struct {
	SearchIssues struct {
		Nodes []model.Issue `json:"nodes"`
	} `json:"searchIssues"`
}

func newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search issues by full-text query",
		Args:  cobra.ExactArgs(1),
		RunE:  runSearch,
	}
	f := cmd.Flags()
	f.String("team", "", "boost results for team key (team UUID used as hint)")
	f.Int("limit", 25, "maximum number of results to return")
	return cmd
}

func runSearch(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	f := cmd.Flags()
	limit, _ := f.GetInt("limit")
	teamKey, _ := f.GetString("team")

	vars := map[string]any{
		"term":  args[0],
		"first": limit,
	}

	ctx := context.Background()
	if teamKey != "" {
		teamID, err := api.ResolveTeamID(ctx, client, teamKey)
		if err != nil {
			return fmt.Errorf("resolve team: %w", err)
		}
		vars["teamId"] = teamID
	}

	var result searchResult
	if err := client.Do(ctx, query.IssueSearchQuery, vars, &result); err != nil {
		return fmt.Errorf("search issues: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.SearchIssues.Nodes)
	}

	rows := make([]SearchRow, len(result.SearchIssues.Nodes))
	for i, issue := range result.SearchIssues.Nodes {
		rows[i] = SearchRow{
			ID:     issue.Identifier,
			Title:  truncate(issue.Title, 40),
			Status: issue.State.Name,
			Team:   issue.Team.Key,
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

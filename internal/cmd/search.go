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

// SearchRow is a display row for the search results table.
type SearchRow struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Team   string `json:"team"`
}

// ProjectSearchRow is a display row for project search results.
type ProjectSearchRow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// DocumentSearchRow is a display row for document search results.
type DocumentSearchRow struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Project string `json:"project"`
}

type issueSearchResult struct {
	SearchIssues struct {
		Nodes []model.Issue `json:"nodes"`
	} `json:"searchIssues"`
}

type projectSearchResult struct {
	SearchProjects struct {
		Nodes []model.Project `json:"nodes"`
	} `json:"searchProjects"`
}

type documentSearchResult struct {
	SearchDocuments struct {
		Nodes []model.Document `json:"nodes"`
	} `json:"searchDocuments"`
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
	f.String("type", "issue", "search type: issue, project, or document")
	return cmd
}

func runSearch(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	f := cmd.Flags()
	limit, _ := f.GetInt("limit")
	if limit <= 0 {
		return fmt.Errorf("--limit must be greater than 0")
	}
	teamKey, _ := f.GetString("team")
	searchType, _ := f.GetString("type")

	switch searchType {
	case "issue", "project", "document":
	default:
		return fmt.Errorf("--type must be one of: issue, project, document")
	}

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

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	switch searchType {
	case "project":
		var result projectSearchResult
		if err := client.Do(ctx, query.ProjectSearchQuery, vars, &result); err != nil {
			return fmt.Errorf("search projects: %w", err)
		}
		if jsonMode {
			return formatter.Format(cmd.OutOrStdout(), result.SearchProjects.Nodes)
		}
		rows := make([]ProjectSearchRow, len(result.SearchProjects.Nodes))
		for i, p := range result.SearchProjects.Nodes {
			rows[i] = ProjectSearchRow{
				ID:          p.ID,
				Name:        truncate(p.Name, 40),
				Status:      p.Status.Name,
				Description: truncate(p.Description, 50),
			}
		}
		return formatter.Format(cmd.OutOrStdout(), rows)

	case "document":
		var result documentSearchResult
		if err := client.Do(ctx, query.DocumentSearchQuery, vars, &result); err != nil {
			return fmt.Errorf("search documents: %w", err)
		}
		if jsonMode {
			return formatter.Format(cmd.OutOrStdout(), result.SearchDocuments.Nodes)
		}
		rows := make([]DocumentSearchRow, len(result.SearchDocuments.Nodes))
		for i, d := range result.SearchDocuments.Nodes {
			projectName := ""
			if d.Project != nil {
				projectName = d.Project.Name
			}
			rows[i] = DocumentSearchRow{
				ID:      d.ID,
				Title:   truncate(d.Title, 50),
				Project: projectName,
			}
		}
		return formatter.Format(cmd.OutOrStdout(), rows)

	default:
		var result issueSearchResult
		if err := client.Do(ctx, query.IssueSearchQuery, vars, &result); err != nil {
			return fmt.Errorf("search issues: %w", err)
		}
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
}

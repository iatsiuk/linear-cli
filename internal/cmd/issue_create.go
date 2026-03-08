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

type issueCreateResult struct {
	IssueCreate struct {
		Success bool         `json:"success"`
		Issue   *model.Issue `json:"issue"`
	} `json:"issueCreate"`
}

func newIssueCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		RunE:  runIssueCreate,
	}
	f := cmd.Flags()
	f.String("title", "", "issue title (required)")
	f.String("team", "", "team key or ID (required)")
	f.String("description", "", "issue description in markdown")
	f.String("assignee", "", "assignee name, email, UUID, or \"me\"")
	f.String("state", "", "workflow state name or ID")
	f.Int("priority", -1, "priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")
	f.StringArray("label", []string{}, "label name or ID (repeatable)")
	f.String("due-date", "", "due date (YYYY-MM-DD)")
	f.Int("estimate", -1, "complexity estimate (integer)")
	f.String("cycle", "", "cycle ID")
	f.String("project", "", "project name or ID")
	f.String("parent", "", "parent issue identifier or ID")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("team")
	return cmd
}

func runIssueCreate(cmd *cobra.Command, _ []string) error {
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
	ctx := context.Background()

	f := cmd.Flags()
	title, _ := f.GetString("title")
	teamKey, _ := f.GetString("team")
	description, _ := f.GetString("description")
	assignee, _ := f.GetString("assignee")
	stateName, _ := f.GetString("state")
	priority, _ := f.GetInt("priority")
	labels, _ := f.GetStringArray("label")
	dueDate, _ := f.GetString("due-date")
	estimate, _ := f.GetInt("estimate")
	cycle, _ := f.GetString("cycle")
	project, _ := f.GetString("project")
	parent, _ := f.GetString("parent")

	if priority != -1 && (priority < 0 || priority > 4) {
		return fmt.Errorf("priority must be 0-4, got %d", priority)
	}

	teamID, err := api.ResolveTeamID(ctx, client, teamKey)
	if err != nil {
		return err
	}

	input := map[string]any{
		"teamId": teamID,
		"title":  title,
	}

	if description != "" {
		input["description"] = description
	}

	if assignee != "" {
		var assigneeID string
		if assignee == "me" {
			assigneeID, err = api.ResolveViewerID(ctx, client)
		} else {
			assigneeID, err = api.ResolveUserID(ctx, client, assignee)
		}
		if err != nil {
			return err
		}
		input["assigneeId"] = assigneeID
	}

	if stateName != "" {
		stateID, err := api.ResolveStateID(ctx, client, stateName, teamID)
		if err != nil {
			return err
		}
		input["stateId"] = stateID
	}

	if priority >= 0 {
		input["priority"] = priority
	}

	if len(labels) > 0 {
		labelIDs := make([]string, len(labels))
		for i, l := range labels {
			id, err := api.ResolveLabelID(ctx, client, l, teamID)
			if err != nil {
				return err
			}
			labelIDs[i] = id
		}
		input["labelIds"] = labelIDs
	}

	if dueDate != "" {
		input["dueDate"] = dueDate
	}

	if estimate >= 0 {
		input["estimate"] = estimate
	}

	if cycle != "" {
		input["cycleId"] = cycle
	}

	if project != "" {
		projectID, err := api.ResolveProjectID(ctx, client, project)
		if err != nil {
			return err
		}
		input["projectId"] = projectID
	}

	if parent != "" {
		input["parentId"] = parent
	}

	vars := map[string]any{"input": input}
	var result issueCreateResult
	if err := client.Do(ctx, query.IssueCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create issue: %w", err)
	}
	if !result.IssueCreate.Success {
		return fmt.Errorf("create issue: mutation returned success=false")
	}
	if result.IssueCreate.Issue == nil {
		return fmt.Errorf("create issue: no issue in response")
	}

	issue := result.IssueCreate.Issue
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), issue)
	}

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

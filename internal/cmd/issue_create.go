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
	f.String("description-file", "", "read issue description from file ('-' for stdin)")
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
	cmd.MarkFlagsMutuallyExclusive("description", "description-file")
	return cmd
}

func runIssueCreate(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	f := cmd.Flags()
	title, _ := f.GetString("title")
	teamKey, _ := f.GetString("team")
	assignee, _ := f.GetString("assignee")
	stateName, _ := f.GetString("state")
	priority, _ := f.GetInt("priority")
	labels, _ := f.GetStringArray("label")
	dueDate, _ := f.GetString("due-date")
	estimate, _ := f.GetInt("estimate")
	cycle, _ := f.GetString("cycle")
	project, _ := f.GetString("project")
	parent, _ := f.GetString("parent")

	if f.Changed("priority") && (priority < 0 || priority > 4) {
		return fmt.Errorf("priority must be 0-4, got %d", priority)
	}

	description, hasDesc, err := readDescription(cmd)
	if err != nil {
		return err
	}

	teamID, err := api.ResolveTeamID(ctx, client, teamKey)
	if err != nil {
		return err
	}

	input := map[string]any{
		"teamId": teamID,
		"title":  title,
	}

	if hasDesc {
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

	if f.Changed("priority") {
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

	if f.Changed("estimate") {
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
		var parentResult issueGetResult
		if err := client.Do(ctx, query.IssueGetQuery, map[string]any{"id": parent}, &parentResult); err != nil {
			return fmt.Errorf("resolve parent: %w", err)
		}
		if parentResult.Issue == nil {
			return fmt.Errorf("parent issue %q not found", parent)
		}
		input["parentId"] = parentResult.Issue.ID
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
	return printIssueRow(cmd, issue)
}

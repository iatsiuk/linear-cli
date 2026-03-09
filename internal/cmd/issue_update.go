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

type issueUpdateResult struct {
	IssueUpdate struct {
		Success bool         `json:"success"`
		Issue   *model.Issue `json:"issue"`
	} `json:"issueUpdate"`
}

func newIssueUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <identifier>",
		Short: "Update an issue",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runIssueUpdate,
	}
	f := cmd.Flags()
	f.String("title", "", "issue title")
	f.String("description", "", "issue description in markdown")
	f.String("assignee", "", "assignee name, email, UUID, or \"me\"")
	f.String("state", "", "workflow state name or ID")
	f.Int("priority", -1, "priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")
	f.StringArray("label", []string{}, "set labels (replaces all existing labels)")
	f.StringArray("add-label", []string{}, "add label by name or ID (repeatable)")
	f.StringArray("remove-label", []string{}, "remove label by name or ID (repeatable)")
	f.String("due-date", "", "due date (YYYY-MM-DD)")
	f.Int("estimate", -1, "complexity estimate (integer)")
	f.String("cycle", "", "cycle ID")
	f.String("project", "", "project name or ID")
	f.String("parent", "", "parent issue identifier or ID")
	return cmd
}

func runIssueUpdate(cmd *cobra.Command, args []string) error {
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
	issueID := getResult.Issue.ID

	f := cmd.Flags()

	if f.Changed("label") && (f.Changed("add-label") || f.Changed("remove-label")) {
		return fmt.Errorf("--label cannot be combined with --add-label or --remove-label")
	}

	// resolve team ID for state/label resolution (use team from fetched issue)
	teamID := getResult.Issue.Team.ID

	input := map[string]any{}

	if f.Changed("title") {
		title, _ := f.GetString("title")
		input["title"] = title
	}

	if f.Changed("description") {
		desc, _ := f.GetString("description")
		input["description"] = desc
	}

	if f.Changed("assignee") {
		assignee, _ := f.GetString("assignee")
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

	if f.Changed("state") {
		stateName, _ := f.GetString("state")
		stateID, err := api.ResolveStateID(ctx, client, stateName, teamID)
		if err != nil {
			return err
		}
		input["stateId"] = stateID
	}

	if f.Changed("priority") {
		priority, _ := f.GetInt("priority")
		if priority < 0 || priority > 4 {
			return fmt.Errorf("priority must be 0-4, got %d", priority)
		}
		input["priority"] = priority
	}

	if f.Changed("label") {
		labels, _ := f.GetStringArray("label")
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

	if f.Changed("add-label") {
		addLabels, _ := f.GetStringArray("add-label")
		addedIDs := make([]string, len(addLabels))
		for i, l := range addLabels {
			id, err := api.ResolveLabelID(ctx, client, l, teamID)
			if err != nil {
				return err
			}
			addedIDs[i] = id
		}
		input["addedLabelIds"] = addedIDs
	}

	if f.Changed("remove-label") {
		removeLabels, _ := f.GetStringArray("remove-label")
		removedIDs := make([]string, len(removeLabels))
		for i, l := range removeLabels {
			id, err := api.ResolveLabelID(ctx, client, l, teamID)
			if err != nil {
				return err
			}
			removedIDs[i] = id
		}
		input["removedLabelIds"] = removedIDs
	}

	if f.Changed("due-date") {
		dueDate, _ := f.GetString("due-date")
		input["dueDate"] = dueDate
	}

	if f.Changed("estimate") {
		estimate, _ := f.GetInt("estimate")
		input["estimate"] = estimate
	}

	if f.Changed("cycle") {
		cycle, _ := f.GetString("cycle")
		input["cycleId"] = cycle
	}

	if f.Changed("project") {
		project, _ := f.GetString("project")
		projectID, err := api.ResolveProjectID(ctx, client, project)
		if err != nil {
			return err
		}
		input["projectId"] = projectID
	}

	if f.Changed("parent") {
		parent, _ := f.GetString("parent")
		var parentResult issueGetResult
		if err := client.Do(ctx, query.IssueGetQuery, map[string]any{"id": parent}, &parentResult); err != nil {
			return fmt.Errorf("resolve parent: %w", err)
		}
		if parentResult.Issue == nil {
			return fmt.Errorf("parent issue %q not found", parent)
		}
		input["parentId"] = parentResult.Issue.ID
	}

	if len(input) == 0 {
		return fmt.Errorf("no fields to update: specify at least one flag")
	}

	vars := map[string]any{"id": issueID, "input": input}
	var result issueUpdateResult
	if err := client.Do(ctx, query.IssueUpdateMutation, vars, &result); err != nil {
		return fmt.Errorf("update issue: %w", err)
	}
	if !result.IssueUpdate.Success {
		return fmt.Errorf("update issue: mutation returned success=false")
	}
	if result.IssueUpdate.Issue == nil {
		return fmt.Errorf("update issue: no issue in response")
	}

	issue := result.IssueUpdate.Issue
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), issue)
	}
	return printIssueRow(cmd, issue)
}

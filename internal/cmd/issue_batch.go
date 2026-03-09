package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"linear-cli/internal/api"
	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

const maxBatchSize = 50

type issueBatchUpdateResult struct {
	IssueBatchUpdate struct {
		Issues []model.Issue `json:"issues"`
	} `json:"issueBatchUpdate"`
}

func newIssueBatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch operations on issues",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newIssueBatchUpdateCommand())
	return cmd
}

func newIssueBatchUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [<id1> <id2> ...]",
		Short: "Update multiple issues at once",
		RunE:  runIssueBatchUpdate,
	}
	f := cmd.Flags()
	f.String("assignee", "", "assignee name, email, UUID, or \"me\"")
	f.String("state", "", "workflow state name or ID")
	f.Int("priority", -1, "priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")
	f.StringArray("label", []string{}, "set labels (replaces all existing labels)")
	f.StringArray("add-label", []string{}, "add label by name or ID (repeatable)")
	f.StringArray("remove-label", []string{}, "remove label by name or ID (repeatable)")
	f.String("project", "", "project name or ID")
	f.String("cycle", "", "cycle ID")
	return cmd
}

func runIssueBatchUpdate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	// collect identifiers from args; if none provided, read from stdin
	identifiers := args
	if len(identifiers) == 0 {
		scanner := bufio.NewScanner(cmd.InOrStdin())
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				identifiers = append(identifiers, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
	}

	if len(identifiers) == 0 {
		return fmt.Errorf("at least one identifier is required")
	}
	if len(identifiers) > maxBatchSize {
		return fmt.Errorf("too many identifiers: %d (max %d)", len(identifiers), maxBatchSize)
	}

	f := cmd.Flags()
	if f.Changed("label") && (f.Changed("add-label") || f.Changed("remove-label")) {
		return fmt.Errorf("--label cannot be combined with --add-label or --remove-label")
	}

	input := map[string]any{}

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
		stateID, err := api.ResolveStateID(ctx, client, stateName, "")
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
			id, err := api.ResolveLabelID(ctx, client, l, "")
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
			id, err := api.ResolveLabelID(ctx, client, l, "")
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
			id, err := api.ResolveLabelID(ctx, client, l, "")
			if err != nil {
				return err
			}
			removedIDs[i] = id
		}
		input["removedLabelIds"] = removedIDs
	}

	if f.Changed("project") {
		project, _ := f.GetString("project")
		projectID, err := api.ResolveProjectID(ctx, client, project)
		if err != nil {
			return err
		}
		input["projectId"] = projectID
	}

	if f.Changed("cycle") {
		cycle, _ := f.GetString("cycle")
		if cycle == "" {
			return fmt.Errorf("--cycle value cannot be empty")
		}
		input["cycleId"] = cycle
	}

	if len(input) == 0 {
		return fmt.Errorf("no fields to update: specify at least one flag")
	}

	// resolve each identifier to UUID via issue query
	ids := make([]string, len(identifiers))
	for i, ident := range identifiers {
		var getResult issueGetResult
		if err := client.Do(ctx, query.IssueGetQuery, map[string]any{"id": ident}, &getResult); err != nil {
			return fmt.Errorf("resolve %q: %w", ident, err)
		}
		if getResult.Issue == nil {
			return fmt.Errorf("issue %q not found", ident)
		}
		ids[i] = getResult.Issue.ID
	}

	vars := map[string]any{"ids": ids, "input": input}
	var result issueBatchUpdateResult
	if err := client.Do(ctx, query.IssueBatchUpdateMutation, vars, &result); err != nil {
		return fmt.Errorf("batch update: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)
	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.IssueBatchUpdate.Issues)
	}
	rows := make([]IssueRow, len(result.IssueBatchUpdate.Issues))
	for i, issue := range result.IssueBatchUpdate.Issues {
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

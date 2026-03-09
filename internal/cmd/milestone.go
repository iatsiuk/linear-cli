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

type milestoneListResult struct {
	Project struct {
		ProjectMilestones model.ProjectMilestoneConnection `json:"projectMilestones"`
	} `json:"project"`
}

type milestonePayloadResult struct {
	ProjectMilestoneCreate struct {
		Success          bool                    `json:"success"`
		ProjectMilestone *model.ProjectMilestone `json:"projectMilestone"`
	} `json:"projectMilestoneCreate"`
}

type milestoneUpdatePayloadResult struct {
	ProjectMilestoneUpdate struct {
		Success          bool                    `json:"success"`
		ProjectMilestone *model.ProjectMilestone `json:"projectMilestone"`
	} `json:"projectMilestoneUpdate"`
}

type milestoneDeleteResult struct {
	ProjectMilestoneDelete struct {
		Success bool `json:"success"`
	} `json:"projectMilestoneDelete"`
}

// MilestoneRow is a display row for the milestone list table.
type MilestoneRow struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	TargetDate  string `json:"target_date"`
	Description string `json:"description"`
}

func newProjectMilestoneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "milestone",
		Short: "Manage project milestones",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newMilestoneListCommand())
	cmd.AddCommand(newMilestoneCreateCommand())
	cmd.AddCommand(newMilestoneUpdateCommand())
	cmd.AddCommand(newMilestoneDeleteCommand())
	return cmd
}

func newMilestoneListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project-id>",
		Short: "List milestones for a project",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("project id is required")
			}
			return nil
		},
		RunE: runMilestoneList,
	}
	cmd.Flags().Int("limit", 50, "maximum number of milestones to return")
	return cmd
}

func runMilestoneList(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	projectID, err := api.ResolveProjectID(ctx, client, args[0])
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	vars := map[string]any{"projectId": projectID, "first": limit}

	var result milestoneListResult
	if err := client.Do(ctx, query.MilestoneListQuery, vars, &result); err != nil {
		return fmt.Errorf("list milestones: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.Project.ProjectMilestones.Nodes)
	}

	rows := make([]MilestoneRow, len(result.Project.ProjectMilestones.Nodes))
	for i, m := range result.Project.ProjectMilestones.Nodes {
		rows[i] = MilestoneRow{
			Name:   truncate(m.Name, 40),
			Status: m.Status,
		}
		if m.TargetDate != nil {
			rows[i].TargetDate = *m.TargetDate
		}
		if m.Description != nil {
			rows[i].Description = truncate(*m.Description, 60)
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newMilestoneCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <project-id>",
		Short: "Create a milestone for a project",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("project id is required")
			}
			return nil
		},
		RunE: runMilestoneCreate,
	}
	f := cmd.Flags()
	f.String("name", "", "milestone name (required)")
	f.String("description", "", "milestone description")
	f.String("target-date", "", "target date (YYYY-MM-DD)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runMilestoneCreate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	projectID, err := api.ResolveProjectID(ctx, client, args[0])
	if err != nil {
		return err
	}

	f := cmd.Flags()
	name, _ := f.GetString("name")

	input := map[string]any{
		"projectId": projectID,
		"name":      name,
	}
	if f.Changed("description") {
		v, _ := f.GetString("description")
		input["description"] = v
	}
	if f.Changed("target-date") {
		v, _ := f.GetString("target-date")
		input["targetDate"] = v
	}

	vars := map[string]any{"input": input}
	var result milestonePayloadResult
	if err := client.Do(ctx, query.MilestoneCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create milestone: %w", err)
	}
	if !result.ProjectMilestoneCreate.Success {
		return fmt.Errorf("create milestone: mutation returned success=false")
	}
	if result.ProjectMilestoneCreate.ProjectMilestone == nil {
		return fmt.Errorf("create milestone: no milestone in response")
	}

	m := result.ProjectMilestoneCreate.ProjectMilestone
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), m)
	}
	targetDate := ""
	if m.TargetDate != nil {
		targetDate = *m.TargetDate
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", m.ID, m.Name, targetDate)
	return err
}

func newMilestoneUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project milestone",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("milestone id is required")
			}
			return nil
		},
		RunE: runMilestoneUpdate,
	}
	f := cmd.Flags()
	f.String("name", "", "milestone name")
	f.String("description", "", "milestone description")
	f.String("target-date", "", "target date (YYYY-MM-DD)")
	return cmd
}

func runMilestoneUpdate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	f := cmd.Flags()
	input := map[string]any{}

	if f.Changed("name") {
		v, _ := f.GetString("name")
		input["name"] = v
	}
	if f.Changed("description") {
		v, _ := f.GetString("description")
		input["description"] = v
	}
	if f.Changed("target-date") {
		v, _ := f.GetString("target-date")
		input["targetDate"] = v
	}
	if len(input) == 0 {
		return fmt.Errorf("no fields to update: specify at least one flag")
	}

	vars := map[string]any{"id": args[0], "input": input}
	var result milestoneUpdatePayloadResult
	if err := client.Do(ctx, query.MilestoneUpdateMutation, vars, &result); err != nil {
		return fmt.Errorf("update milestone: %w", err)
	}
	if !result.ProjectMilestoneUpdate.Success {
		return fmt.Errorf("update milestone: mutation returned success=false")
	}
	if result.ProjectMilestoneUpdate.ProjectMilestone == nil {
		return fmt.Errorf("update milestone: no milestone in response")
	}

	m := result.ProjectMilestoneUpdate.ProjectMilestone
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), m)
	}
	targetDate := ""
	if m.TargetDate != nil {
		targetDate = *m.TargetDate
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", m.ID, m.Name, targetDate)
	return err
}

func newMilestoneDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project milestone",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("milestone id is required")
			}
			return nil
		},
		RunE: runMilestoneDelete,
	}
	cmd.Flags().Bool("yes", false, "skip confirmation prompt")
	return cmd
}

func runMilestoneDelete(cmd *cobra.Command, args []string) error {
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Delete milestone %s? [y/N] ", args[0])
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

	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	vars := map[string]any{"id": args[0]}
	var result milestoneDeleteResult
	if err := client.Do(ctx, query.MilestoneDeleteMutation, vars, &result); err != nil {
		return fmt.Errorf("delete milestone: %w", err)
	}
	if !result.ProjectMilestoneDelete.Success {
		return fmt.Errorf("delete milestone: mutation returned success=false")
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Milestone %s deleted.\n", args[0])
	return err
}

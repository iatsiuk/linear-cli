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

type projectUpdateCheckinListResult struct {
	Project *struct {
		ProjectUpdates model.ProjectUpdateConnection `json:"projectUpdates"`
	} `json:"project"`
}

type projectUpdateCheckinCreateResult struct {
	ProjectUpdateCreate struct {
		Success       bool                 `json:"success"`
		ProjectUpdate *model.ProjectUpdate `json:"projectUpdate"`
	} `json:"projectUpdateCreate"`
}

type projectUpdateCheckinArchiveResult struct {
	ProjectUpdateArchive struct {
		Success bool `json:"success"`
	} `json:"projectUpdateArchive"`
}

// ProjectUpdateCheckinRow is a display row for the check-in list table.
type ProjectUpdateCheckinRow struct {
	Health string `json:"health"`
	Author string `json:"author"`
	Date   string `json:"date"`
	Body   string `json:"body"`
}

func newProjectUpdateListCheckinCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project-id>",
		Short: "List status check-ins for a project",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("project id is required")
			}
			return nil
		},
		RunE: runProjectUpdateCheckinList,
	}
	cmd.Flags().Int("limit", 25, "maximum number of check-ins to return")
	return cmd
}

func runProjectUpdateCheckinList(cmd *cobra.Command, args []string) error {
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

	var result projectUpdateCheckinListResult
	if err := client.Do(ctx, query.ProjectUpdateListQuery, vars, &result); err != nil {
		return fmt.Errorf("list project updates: %w", err)
	}
	if result.Project == nil {
		return fmt.Errorf("project not found: %s", projectID)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.Project.ProjectUpdates.Nodes)
	}

	rows := make([]ProjectUpdateCheckinRow, len(result.Project.ProjectUpdates.Nodes))
	for i, u := range result.Project.ProjectUpdates.Nodes {
		date := u.CreatedAt
		if len(date) > 10 {
			date = date[:10]
		}
		rows[i] = ProjectUpdateCheckinRow{
			Health: u.Health,
			Author: u.User.DisplayName,
			Date:   date,
			Body:   truncate(u.Body, 60),
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newProjectUpdateCreateCheckinCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <project-id>",
		Short: "Create a status check-in for a project",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("project id is required")
			}
			return nil
		},
		RunE: runProjectUpdateCheckinCreate,
	}
	f := cmd.Flags()
	f.String("body", "", "check-in body text (required)")
	f.String("health", "", "project health (onTrack|atRisk|offTrack)")
	_ = cmd.MarkFlagRequired("body")
	return cmd
}

func runProjectUpdateCheckinCreate(cmd *cobra.Command, args []string) error {
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
	body, _ := f.GetString("body")

	input := map[string]any{
		"projectId": projectID,
		"body":      body,
	}
	if f.Changed("health") {
		health, _ := f.GetString("health")
		switch health {
		case "onTrack", "atRisk", "offTrack":
		default:
			return fmt.Errorf("invalid health %q: must be onTrack, atRisk, or offTrack", health)
		}
		input["health"] = health
	}

	vars := map[string]any{"input": input}
	var result projectUpdateCheckinCreateResult
	if err := client.Do(ctx, query.ProjectUpdateCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create project update: %w", err)
	}
	if !result.ProjectUpdateCreate.Success {
		return fmt.Errorf("create project update: mutation returned success=false")
	}
	if result.ProjectUpdateCreate.ProjectUpdate == nil {
		return fmt.Errorf("create project update: no project update in response")
	}

	pu := result.ProjectUpdateCreate.ProjectUpdate
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), pu)
	}
	date := pu.CreatedAt
	if len(date) > 10 {
		date = date[:10]
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", pu.ID, pu.Health, date)
	return err
}

func newProjectUpdateArchiveCheckinCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a status check-in",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("check-in id is required")
			}
			return nil
		},
		RunE: runProjectUpdateCheckinArchive,
	}
}

func runProjectUpdateCheckinArchive(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	vars := map[string]any{"id": args[0]}
	var result projectUpdateCheckinArchiveResult
	if err := client.Do(ctx, query.ProjectUpdateArchiveMutation, vars, &result); err != nil {
		return fmt.Errorf("archive project update: %w", err)
	}
	if !result.ProjectUpdateArchive.Success {
		return fmt.Errorf("archive project update: mutation returned success=false")
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Check-in %s archived.\n", args[0])
	return err
}

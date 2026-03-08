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

type projectUpdateResult struct {
	ProjectUpdate struct {
		Success bool           `json:"success"`
		Project *model.Project `json:"project"`
	} `json:"projectUpdate"`
}

func newProjectUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("project id is required")
			}
			return nil
		},
		RunE: runProjectUpdate,
	}
	f := cmd.Flags()
	f.String("name", "", "project name")
	f.String("description", "", "project description")
	f.String("state", "", "project state type or UUID (backlog|planned|started|paused|completed|canceled)")
	f.String("target-date", "", "target date (YYYY-MM-DD)")
	f.String("start-date", "", "start date (YYYY-MM-DD)")
	return cmd
}

func runProjectUpdate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	id := args[0]
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
	if f.Changed("state") {
		v, _ := f.GetString("state")
		statusID, err := api.ResolveProjectStatusID(ctx, client, v)
		if err != nil {
			return err
		}
		input["statusId"] = statusID
	}
	if f.Changed("target-date") {
		v, _ := f.GetString("target-date")
		input["targetDate"] = v
	}
	if f.Changed("start-date") {
		v, _ := f.GetString("start-date")
		input["startDate"] = v
	}
	if len(input) == 0 {
		return fmt.Errorf("no fields to update: specify at least one flag")
	}

	vars := map[string]any{"id": id, "input": input}
	var result projectUpdateResult
	if err := client.Do(ctx, query.ProjectUpdateMutation, vars, &result); err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	if !result.ProjectUpdate.Success {
		return fmt.Errorf("update project: mutation returned success=false")
	}
	if result.ProjectUpdate.Project == nil {
		return fmt.Errorf("update project: no project in response")
	}

	p := result.ProjectUpdate.Project
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), p)
	}
	return printProjectDetail(cmd, p)
}

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

type projectCreateResult struct {
	ProjectCreate struct {
		Success bool           `json:"success"`
		Project *model.Project `json:"project"`
	} `json:"projectCreate"`
}

func newProjectCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		RunE:  runProjectCreate,
	}
	f := cmd.Flags()
	f.String("name", "", "project name (required)")
	f.StringArray("team", []string{}, "team key or ID (repeatable, required)")
	f.String("description", "", "project description")
	f.String("color", "", "project color (hex)")
	f.String("target-date", "", "target date (YYYY-MM-DD)")
	f.String("start-date", "", "start date (YYYY-MM-DD)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("team")
	return cmd
}

func runProjectCreate(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	f := cmd.Flags()
	name, _ := f.GetString("name")
	teamKeys, _ := f.GetStringArray("team")
	description, _ := f.GetString("description")
	color, _ := f.GetString("color")
	targetDate, _ := f.GetString("target-date")
	startDate, _ := f.GetString("start-date")

	teamIDs := make([]string, len(teamKeys))
	for i, key := range teamKeys {
		id, err := api.ResolveTeamID(ctx, client, key)
		if err != nil {
			return err
		}
		teamIDs[i] = id
	}

	input := map[string]any{
		"name":    name,
		"teamIds": teamIDs,
	}
	if description != "" {
		input["description"] = description
	}
	if color != "" {
		input["color"] = color
	}
	if targetDate != "" {
		input["targetDate"] = targetDate
	}
	if startDate != "" {
		input["startDate"] = startDate
	}

	vars := map[string]any{"input": input}
	var result projectCreateResult
	if err := client.Do(ctx, query.ProjectCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	if !result.ProjectCreate.Success {
		return fmt.Errorf("create project: mutation returned success=false")
	}
	if result.ProjectCreate.Project == nil {
		return fmt.Errorf("create project: no project in response")
	}

	p := result.ProjectCreate.Project
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), p)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", p.ID, p.Name)
	return err
}

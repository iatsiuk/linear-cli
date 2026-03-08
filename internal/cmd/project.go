package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

func newProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage Linear projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newProjectListCommand())
	cmd.AddCommand(newProjectShowCommand())
	return cmd
}

// ProjectRow is a display row for the project list table.
type ProjectRow struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Health     string `json:"health"`
	Progress   string `json:"progress"`
	TargetDate string `json:"target_date"`
}

type projectListResult struct {
	Projects struct {
		Nodes []model.Project `json:"nodes"`
	} `json:"projects"`
}

type projectGetResult struct {
	Project *model.Project `json:"project"`
}

func newProjectListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE:  runProjectList,
	}
	f := cmd.Flags()
	f.String("team", "", "filter by team key")
	f.String("status", "", "filter by status type (backlog|planned|started|paused|completed|canceled)")
	f.String("health", "", "filter by health (onTrack|atRisk|offTrack)")
	f.Int("limit", 50, "maximum number of projects to return")
	f.Bool("include-archived", false, "include archived projects")
	f.String("order-by", "updatedAt", "sort order (createdAt|updatedAt)")
	return cmd
}

func runProjectList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	f := cmd.Flags()
	limit, _ := f.GetInt("limit")
	includeArchived, _ := f.GetBool("include-archived")
	orderBy, _ := f.GetString("order-by")
	teamKey, _ := f.GetString("team")
	statusType, _ := f.GetString("status")
	health, _ := f.GetString("health")

	vars := map[string]any{"first": limit}
	if includeArchived {
		vars["includeArchived"] = true
	}
	if orderBy != "" {
		vars["orderBy"] = orderBy
	}

	filter := map[string]any{}
	if teamKey != "" {
		filter["accessibleTeams"] = map[string]any{
			"some": map[string]any{"key": map[string]any{"eq": teamKey}},
		}
	}
	if statusType != "" {
		filter["status"] = map[string]any{
			"type": map[string]any{"eq": statusType},
		}
	}
	if health != "" {
		filter["health"] = map[string]any{"eq": health}
	}
	if len(filter) > 0 {
		vars["filter"] = filter
	}

	var result projectListResult
	if err := client.Do(context.Background(), query.ProjectListQuery, vars, &result); err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.Projects.Nodes)
	}

	rows := make([]ProjectRow, len(result.Projects.Nodes))
	for i, p := range result.Projects.Nodes {
		rows[i] = ProjectRow{
			Name:     truncate(p.Name, 40),
			Status:   p.Status.Type,
			Progress: fmt.Sprintf("%.0f%%", p.Progress*100),
		}
		if p.Health != nil {
			rows[i].Health = *p.Health
		}
		if p.TargetDate != nil {
			rows[i].TargetDate = *p.TargetDate
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newProjectShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show details of a project",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("project id is required")
			}
			return nil
		},
		RunE: runProjectShow,
	}
}

func runProjectShow(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	id := args[0]
	vars := map[string]any{"id": id}

	var result projectGetResult
	if err := client.Do(context.Background(), query.ProjectGetQuery, vars, &result); err != nil {
		return fmt.Errorf("get project: %w", err)
	}
	if result.Project == nil {
		return fmt.Errorf("project %q not found", id)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), result.Project)
	}

	return printProjectDetail(cmd, result.Project)
}

func printProjectDetail(cmd *cobra.Command, p *model.Project) error {
	w := cmd.OutOrStdout()

	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-14s %s\n", label+":", value)
		return err
	}

	if err := writeLine("Name", p.Name); err != nil {
		return err
	}
	if err := writeLine("Status", p.Status.Type); err != nil {
		return err
	}

	health := ""
	if p.Health != nil {
		health = *p.Health
	}
	if err := writeLine("Health", health); err != nil {
		return err
	}

	if err := writeLine("Progress", fmt.Sprintf("%.0f%%", p.Progress*100)); err != nil {
		return err
	}

	// teams
	teamNames := make([]string, len(p.Teams.Nodes))
	for i, t := range p.Teams.Nodes {
		teamNames[i] = t.Name
	}
	if err := writeLine("Teams", strings.Join(teamNames, ", ")); err != nil {
		return err
	}

	if p.Creator != nil {
		if err := writeLine("Creator", p.Creator.DisplayName); err != nil {
			return err
		}
	}

	if p.StartDate != nil {
		if err := writeLine("Start Date", *p.StartDate); err != nil {
			return err
		}
	}
	if p.TargetDate != nil {
		if err := writeLine("Target Date", *p.TargetDate); err != nil {
			return err
		}
	}

	for _, f := range []struct{ label, value string }{
		{"URL", p.URL},
		{"Created", p.CreatedAt},
		{"Updated", p.UpdatedAt},
	} {
		if err := writeLine(f.label, f.value); err != nil {
			return err
		}
	}

	if p.Description != "" {
		_, err := fmt.Fprintf(w, "\n%s\n", p.Description)
		return err
	}

	return nil
}

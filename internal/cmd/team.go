package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"linear-cli/internal/api"
	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type teamListResult struct {
	Teams struct {
		Nodes    []model.Team `json:"nodes"`
		PageInfo api.PageInfo `json:"pageInfo"`
	} `json:"teams"`
}

type teamGetResult struct {
	Team *model.Team `json:"team"`
}

// TeamRow is a display row for the team list table.
type TeamRow struct {
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
	Cycles      string `json:"cycles"`
}

func newTeamCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage Linear teams",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newTeamListCommand())
	cmd.AddCommand(newTeamShowCommand())
	return cmd
}

func newTeamListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List teams",
		RunE:  runTeamList,
	}
}

func runTeamList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()
	teams, err := api.PaginateAll(ctx, func(ctx context.Context, after *string, first int) (api.Connection[model.Team], error) {
		vars := map[string]any{"first": first}
		if after != nil {
			vars["after"] = *after
		}
		var result teamListResult
		if err := client.Do(ctx, query.TeamListQuery, vars, &result); err != nil {
			return api.Connection[model.Team]{}, err
		}
		return api.Connection[model.Team]{Nodes: result.Teams.Nodes, PageInfo: result.Teams.PageInfo}, nil
	}, 50, 0)
	if err != nil {
		return fmt.Errorf("list teams: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), teams)
	}

	rows := make([]TeamRow, len(teams))
	for i, t := range teams {
		desc := ""
		if t.Description != nil {
			desc = truncate(*t.Description, 40)
		}
		rows[i] = TeamRow{
			Name:        t.Name,
			Key:         t.Key,
			Description: desc,
			Cycles:      strconv.FormatBool(t.CyclesEnabled),
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newTeamShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id|key>",
		Short: "Show team details",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("exactly one team id or key is required")
			}
			return nil
		},
		RunE: runTeamShow,
	}
}

func runTeamShow(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	teamID, err := api.ResolveTeamID(context.Background(), client, args[0])
	if err != nil {
		return err
	}

	vars := map[string]any{"id": teamID}

	var result teamGetResult
	if err := client.Do(context.Background(), query.TeamGetQuery, vars, &result); err != nil {
		return fmt.Errorf("get team: %w", err)
	}
	if result.Team == nil {
		return fmt.Errorf("team %q not found", args[0])
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), result.Team)
	}

	return printTeamDetail(cmd, result.Team)
}

func printTeamDetail(cmd *cobra.Command, t *model.Team) error {
	w := cmd.OutOrStdout()

	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-14s %s\n", label+":", value)
		return err
	}

	fields := []struct{ label, value string }{
		{"Name", t.Name},
		{"Key", t.Key},
		{"Cycles", strconv.FormatBool(t.CyclesEnabled)},
		{"Estimation", t.IssueEstimationType},
		{"Created", t.CreatedAt},
		{"Updated", t.UpdatedAt},
	}
	for _, f := range fields {
		if err := writeLine(f.label, f.value); err != nil {
			return err
		}
	}

	if t.Description != nil && *t.Description != "" {
		if err := writeLine("Description", *t.Description); err != nil {
			return err
		}
	}

	return nil
}

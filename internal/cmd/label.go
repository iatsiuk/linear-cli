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

type labelListResult struct {
	IssueLabels struct {
		Nodes    []model.IssueLabel `json:"nodes"`
		PageInfo api.PageInfo       `json:"pageInfo"`
	} `json:"issueLabels"`
}

type labelCreateResult struct {
	IssueLabelCreate struct {
		Success    bool              `json:"success"`
		IssueLabel *model.IssueLabel `json:"issueLabel"`
	} `json:"issueLabelCreate"`
}

type labelUpdateResult struct {
	IssueLabelUpdate struct {
		Success    bool              `json:"success"`
		IssueLabel *model.IssueLabel `json:"issueLabel"`
	} `json:"issueLabelUpdate"`
}

// LabelRow is a display row for the label list table.
type LabelRow struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
	Team        string `json:"team"`
	Group       string `json:"group"`
}

func newLabelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label",
		Short: "Manage Linear labels",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newLabelListCommand())
	cmd.AddCommand(newLabelCreateCommand())
	cmd.AddCommand(newLabelUpdateCommand())
	return cmd
}

func newLabelListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issue labels",
		RunE:  runLabelList,
	}
	cmd.Flags().String("team", "", "filter by team key")
	return cmd
}

func runLabelList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	teamKey, _ := cmd.Flags().GetString("team")

	ctx := context.Background()
	labels, err := api.PaginateAll(ctx, func(ctx context.Context, after *string, first int) (api.Connection[model.IssueLabel], error) {
		vars := map[string]any{"first": first}
		if after != nil {
			vars["after"] = *after
		}
		if teamKey != "" {
			vars["filter"] = map[string]any{
				"team": map[string]any{
					"key": map[string]any{"eq": teamKey},
				},
			}
		}
		var result labelListResult
		if err := client.Do(ctx, query.LabelListQuery, vars, &result); err != nil {
			return api.Connection[model.IssueLabel]{}, err
		}
		return api.Connection[model.IssueLabel]{
			Nodes:    result.IssueLabels.Nodes,
			PageInfo: result.IssueLabels.PageInfo,
		}, nil
	}, 50, 0)
	if err != nil {
		return fmt.Errorf("list labels: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), labels)
	}

	rows := make([]LabelRow, len(labels))
	for i, l := range labels {
		desc := ""
		if l.Description != nil {
			desc = truncate(*l.Description, 40)
		}
		teamName := ""
		if l.Team != nil {
			teamName = l.Team.Key
		}
		isGroup := ""
		if l.IsGroup {
			isGroup = "yes"
		}
		rows[i] = LabelRow{
			Name:        l.Name,
			Color:       l.Color,
			Description: desc,
			Team:        teamName,
			Group:       isGroup,
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newLabelCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue label",
		RunE:  runLabelCreate,
	}
	f := cmd.Flags()
	f.String("name", "", "label name (required)")
	f.String("color", "", "label color hex (required)")
	f.String("team", "", "team key or ID")
	f.String("description", "", "label description")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("color")
	return cmd
}

func runLabelCreate(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	f := cmd.Flags()
	name, _ := f.GetString("name")
	color, _ := f.GetString("color")
	teamKey, _ := f.GetString("team")
	description, _ := f.GetString("description")

	input := map[string]any{
		"name":  name,
		"color": color,
	}
	if teamKey != "" {
		teamID, err := api.ResolveTeamID(ctx, client, teamKey)
		if err != nil {
			return err
		}
		input["teamId"] = teamID
	}
	if description != "" {
		input["description"] = description
	}

	vars := map[string]any{"input": input}
	var result labelCreateResult
	if err := client.Do(ctx, query.LabelCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create label: %w", err)
	}
	if !result.IssueLabelCreate.Success {
		return fmt.Errorf("create label: mutation returned success=false")
	}
	if result.IssueLabelCreate.IssueLabel == nil {
		return fmt.Errorf("create label: no label in response")
	}

	l := result.IssueLabelCreate.IssueLabel
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), l)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", l.ID, l.Name, l.Color)
	return err
}

func newLabelUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an issue label",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("label id is required")
			}
			return nil
		},
		RunE: runLabelUpdate,
	}
	f := cmd.Flags()
	f.String("name", "", "label name")
	f.String("color", "", "label color hex")
	f.String("description", "", "label description")
	return cmd
}

func runLabelUpdate(cmd *cobra.Command, args []string) error {
	f := cmd.Flags()
	input := map[string]any{}

	if f.Changed("name") {
		v, _ := f.GetString("name")
		input["name"] = v
	}
	if f.Changed("color") {
		v, _ := f.GetString("color")
		input["color"] = v
	}
	if f.Changed("description") {
		v, _ := f.GetString("description")
		input["description"] = v
	}
	if len(input) == 0 {
		return fmt.Errorf("no fields to update: specify at least one flag")
	}

	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	id, err := api.ResolveLabelID(ctx, client, args[0], "")
	if err != nil {
		return err
	}

	vars := map[string]any{"id": id, "input": input}
	var result labelUpdateResult
	if err := client.Do(ctx, query.LabelUpdateMutation, vars, &result); err != nil {
		return fmt.Errorf("update label: %w", err)
	}
	if !result.IssueLabelUpdate.Success {
		return fmt.Errorf("update label: mutation returned success=false")
	}
	if result.IssueLabelUpdate.IssueLabel == nil {
		return fmt.Errorf("update label: no label in response")
	}

	l := result.IssueLabelUpdate.IssueLabel
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), l)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", l.ID, l.Name, l.Color)
	return err
}

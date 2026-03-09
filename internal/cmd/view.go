package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type customViewListResult struct {
	CustomViews model.CustomViewConnection `json:"customViews"`
}

type customViewShowResult struct {
	CustomView *model.CustomView `json:"customView"`
}

// CustomViewRow is a display row for the custom view list table.
type CustomViewRow struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Shared string `json:"shared"`
}

func newViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Manage custom views",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newViewListCommand())
	cmd.AddCommand(newViewShowCommand())
	return cmd
}

func newViewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom views",
		RunE:  runViewList,
	}
	cmd.Flags().Int("limit", 50, "maximum number of views to return")
	return cmd
}

func runViewList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	vars := map[string]any{"first": limit}
	var result customViewListResult
	if err := client.Do(context.Background(), query.CustomViewListQuery, vars, &result); err != nil {
		return fmt.Errorf("list custom views: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.CustomViews.Nodes)
	}

	rows := make([]CustomViewRow, len(result.CustomViews.Nodes))
	for i, v := range result.CustomViews.Nodes {
		shared := "no"
		if v.Shared {
			shared = "yes"
		}
		rows[i] = CustomViewRow{
			Name:   truncate(v.Name, 50),
			Type:   v.ModelName,
			Shared: shared,
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newViewShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show custom view details",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("view id is required")
			}
			return nil
		},
		RunE: runViewShow,
	}
}

func runViewShow(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	vars := map[string]any{"id": args[0]}
	var result customViewShowResult
	if err := client.Do(context.Background(), query.CustomViewShowQuery, vars, &result); err != nil {
		return fmt.Errorf("show custom view: %w", err)
	}
	if result.CustomView == nil {
		return fmt.Errorf("custom view not found: %s", args[0])
	}

	v := result.CustomView
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), v)
	}

	w := cmd.OutOrStdout()
	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-14s %s\n", label+":", value)
		return err
	}

	if err := writeLine("Name", v.Name); err != nil {
		return err
	}
	if err := writeLine("Type", v.ModelName); err != nil {
		return err
	}
	shared := "no"
	if v.Shared {
		shared = "yes"
	}
	if err := writeLine("Shared", shared); err != nil {
		return err
	}
	if v.Description != nil {
		if err := writeLine("Description", *v.Description); err != nil {
			return err
		}
	}
	if len(v.FilterData) > 0 && string(v.FilterData) != "null" {
		if err := writeLine("Filters", string(v.FilterData)); err != nil {
			return err
		}
	}
	return nil
}

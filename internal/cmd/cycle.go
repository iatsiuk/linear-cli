package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

func newCycleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cycle",
		Short: "Manage Linear cycles",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newCycleListCommand())
	cmd.AddCommand(newCycleShowCommand())
	cmd.AddCommand(newCycleActiveCommand())
	return cmd
}

// CycleRow is a display row for the cycle list table.
type CycleRow struct {
	Number   string `json:"number"`
	Name     string `json:"name"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Progress string `json:"progress"`
	Status   string `json:"status"`
}

type cycleListResult struct {
	Cycles struct {
		Nodes []model.Cycle `json:"nodes"`
	} `json:"cycles"`
}

type cycleGetResult struct {
	Cycle *model.Cycle `json:"cycle"`
}

func newCycleListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cycles",
		RunE:  runCycleList,
	}
	f := cmd.Flags()
	f.String("team", "", "filter by team key")
	if err := cmd.MarkFlagRequired("team"); err != nil {
		panic(err)
	}
	f.Int("limit", 50, "maximum number of cycles to return")
	f.Bool("include-archived", false, "include archived cycles")
	f.String("order-by", "updatedAt", "sort order (createdAt|updatedAt)")
	return cmd
}

func runCycleList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	f := cmd.Flags()
	limit, _ := f.GetInt("limit")
	includeArchived, _ := f.GetBool("include-archived")
	orderBy, _ := f.GetString("order-by")
	teamKey, _ := f.GetString("team")

	vars := map[string]any{"first": limit}
	if includeArchived {
		vars["includeArchived"] = true
	}
	if orderBy != "" {
		vars["orderBy"] = orderBy
	}

	if teamKey != "" {
		vars["filter"] = map[string]any{
			"team": map[string]any{
				"key": map[string]any{"eq": teamKey},
			},
		}
	}

	var result cycleListResult
	if err := client.Do(context.Background(), query.CycleListQuery, vars, &result); err != nil {
		return fmt.Errorf("list cycles: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.Cycles.Nodes)
	}

	rows := make([]CycleRow, len(result.Cycles.Nodes))
	for i, c := range result.Cycles.Nodes {
		rows[i] = buildCycleRow(c)
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newCycleShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show details of a cycle",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("cycle id is required")
			}
			return nil
		},
		RunE: runCycleShow,
	}
}

func runCycleShow(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	id := args[0]
	vars := map[string]any{"id": id}

	var result cycleGetResult
	if err := client.Do(context.Background(), query.CycleGetQuery, vars, &result); err != nil {
		return fmt.Errorf("get cycle: %w", err)
	}
	if result.Cycle == nil {
		return fmt.Errorf("cycle %q not found", id)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), result.Cycle)
	}

	return printCycleDetail(cmd, result.Cycle)
}

func newCycleActiveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active",
		Short: "Show the active cycle for a team",
		RunE:  runCycleActive,
	}
	cmd.Flags().String("team", "", "team key (required)")
	if err := cmd.MarkFlagRequired("team"); err != nil {
		panic(err)
	}
	return cmd
}

func runCycleActive(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	teamKey, _ := cmd.Flags().GetString("team")
	vars := map[string]any{
		"first": 1,
		"filter": map[string]any{
			"isActive": map[string]any{"eq": true},
			"team":     map[string]any{"key": map[string]any{"eq": teamKey}},
		},
	}

	var result cycleListResult
	if err := client.Do(context.Background(), query.CycleListQuery, vars, &result); err != nil {
		return fmt.Errorf("get active cycle: %w", err)
	}

	if len(result.Cycles.Nodes) == 0 {
		return fmt.Errorf("no active cycle found for team %q", teamKey)
	}

	cycle := &result.Cycles.Nodes[0]

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), cycle)
	}

	return printCycleDetail(cmd, cycle)
}

func printCycleDetail(cmd *cobra.Command, c *model.Cycle) error {
	w := cmd.OutOrStdout()

	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-12s %s\n", label+":", value)
		return err
	}

	name := ""
	if c.Name != nil {
		name = *c.Name
	}
	if err := writeLine("Number", fmt.Sprintf("%.0f", c.Number)); err != nil {
		return err
	}
	if err := writeLine("Name", name); err != nil {
		return err
	}
	if err := writeLine("Status", cycleStatus(c)); err != nil {
		return err
	}
	if err := writeLine("Progress", fmt.Sprintf("%.0f%%", c.Progress*100)); err != nil {
		return err
	}
	if err := writeLine("Start", c.StartsAt); err != nil {
		return err
	}
	if err := writeLine("End", c.EndsAt); err != nil {
		return err
	}
	if err := writeLine("Team", c.Team.Name); err != nil {
		return err
	}

	if c.Description != nil && *c.Description != "" {
		_, err := fmt.Fprintf(w, "\n%s\n", *c.Description)
		return err
	}

	return nil
}

func buildCycleRow(c model.Cycle) CycleRow {
	name := ""
	if c.Name != nil {
		name = *c.Name
	}
	return CycleRow{
		Number:   fmt.Sprintf("%.0f", c.Number),
		Name:     truncate(name, 40),
		Start:    c.StartsAt,
		End:      c.EndsAt,
		Progress: fmt.Sprintf("%.0f%%", c.Progress*100),
		Status:   cycleStatus(&c),
	}
}

func cycleStatus(c *model.Cycle) string {
	switch {
	case c.IsActive:
		return "Active"
	case c.IsFuture:
		return "Future"
	case c.IsPast:
		return "Past"
	default:
		return ""
	}
}

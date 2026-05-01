package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/iatsiuk/linear-cli/internal/api"
	"github.com/iatsiuk/linear-cli/internal/model"
	"github.com/iatsiuk/linear-cli/internal/output"
	"github.com/iatsiuk/linear-cli/internal/query"
)

type stateListResult struct {
	WorkflowStates struct {
		Nodes    []model.WorkflowState `json:"nodes"`
		PageInfo api.PageInfo          `json:"pageInfo"`
	} `json:"workflowStates"`
}

// stateTypeOrder defines the display order for workflow state types.
var stateTypeOrder = []string{"triage", "backlog", "unstarted", "started", "completed", "canceled"}

// StateRow is a display row for the state list table.
type StateRow struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Color string `json:"color"`
}

func newStateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Manage Linear workflow states",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newStateListCommand())
	return cmd
}

func newStateListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflow states for a team",
		RunE:  runStateList,
	}
	f := cmd.Flags()
	f.String("team", "", "team key (required)")
	if err := cmd.MarkFlagRequired("team"); err != nil {
		panic(err)
	}
	return cmd
}

func runStateList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	teamKey, _ := cmd.Flags().GetString("team")
	ctx := context.Background()

	states, err := api.PaginateAll(ctx, func(ctx context.Context, after *string, first int) (api.Connection[model.WorkflowState], error) {
		vars := map[string]any{
			"first": first,
			"filter": map[string]any{
				"team": map[string]any{
					"key": map[string]any{"eq": teamKey},
				},
			},
		}
		if after != nil {
			vars["after"] = *after
		}
		var result stateListResult
		if err := client.Do(ctx, query.StateListQuery, vars, &result); err != nil {
			return api.Connection[model.WorkflowState]{}, err
		}
		return api.Connection[model.WorkflowState]{
			Nodes:    result.WorkflowStates.Nodes,
			PageInfo: result.WorkflowStates.PageInfo,
		}, nil
	}, 50, 0)
	if err != nil {
		return fmt.Errorf("list states: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), states)
	}

	w := cmd.OutOrStdout()
	if len(states) == 0 {
		_, err := fmt.Fprintln(w, "(no results)")
		return err
	}

	// group states by type
	grouped := make(map[string][]model.WorkflowState)
	for _, s := range states {
		grouped[s.Type] = append(grouped[s.Type], s)
	}
	first := true
	for _, typeKey := range stateTypeOrder {
		group := grouped[typeKey]
		if len(group) == 0 {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			return group[i].Position < group[j].Position
		})
		label := strings.ToUpper(typeKey[:1]) + typeKey[1:]
		if !first {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		first = false
		if _, err := fmt.Fprintf(w, "%s\n", label); err != nil {
			return err
		}
		rows := make([]StateRow, len(group))
		for i, s := range group {
			rows[i] = StateRow{
				Name:  s.Name,
				Type:  s.Type,
				Color: s.Color,
			}
		}
		if err := output.NewFormatter(false).Format(w, rows); err != nil {
			return err
		}
	}

	return nil
}

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/iatsiuk/linear-cli/internal/model"
	"github.com/iatsiuk/linear-cli/internal/output"
	"github.com/iatsiuk/linear-cli/internal/query"
)

type initiativeListResult struct {
	Initiatives model.InitiativeConnection `json:"initiatives"`
}

type initiativeShowResult struct {
	Initiative *model.Initiative `json:"initiative"`
}

type initiativePayloadResult struct {
	InitiativeCreate struct {
		Success    bool              `json:"success"`
		Initiative *model.Initiative `json:"initiative"`
	} `json:"initiativeCreate"`
}

type initiativeUpdatePayloadResult struct {
	InitiativeUpdate struct {
		Success    bool              `json:"success"`
		Initiative *model.Initiative `json:"initiative"`
	} `json:"initiativeUpdate"`
}

type initiativeDeleteResult struct {
	InitiativeDelete struct {
		Success bool `json:"success"`
	} `json:"initiativeDelete"`
}

// InitiativeRow is a display row for the initiative list table.
type InitiativeRow struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func newInitiativeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initiative",
		Short: "Manage initiatives",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newInitiativeListCommand())
	cmd.AddCommand(newInitiativeShowCommand())
	cmd.AddCommand(newInitiativeCreateCommand())
	cmd.AddCommand(newInitiativeUpdateCommand())
	cmd.AddCommand(newInitiativeDeleteCommand())
	return cmd
}

func newInitiativeListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List initiatives",
		RunE:  runInitiativeList,
	}
	cmd.Flags().Int("limit", 50, "maximum number of initiatives to return")
	return cmd
}

func runInitiativeList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	vars := map[string]any{"first": limit}

	var result initiativeListResult
	if err := client.Do(context.Background(), query.InitiativeListQuery, vars, &result); err != nil {
		return fmt.Errorf("list initiatives: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.Initiatives.Nodes)
	}

	rows := make([]InitiativeRow, len(result.Initiatives.Nodes))
	for i, ini := range result.Initiatives.Nodes {
		rows[i] = InitiativeRow{
			Name:   truncate(ini.Name, 50),
			Status: ini.Status,
		}
		if ini.Description != nil {
			rows[i].Description = truncate(*ini.Description, 60)
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newInitiativeShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show initiative details",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("initiative id is required")
			}
			return nil
		},
		RunE: runInitiativeShow,
	}
}

func runInitiativeShow(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	vars := map[string]any{"id": args[0]}
	var result initiativeShowResult
	if err := client.Do(context.Background(), query.InitiativeShowQuery, vars, &result); err != nil {
		return fmt.Errorf("show initiative: %w", err)
	}
	if result.Initiative == nil {
		return fmt.Errorf("initiative not found: %s", args[0])
	}

	ini := result.Initiative
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), ini)
	}

	w := cmd.OutOrStdout()
	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-14s %s\n", label+":", value)
		return err
	}

	if err := writeLine("Name", ini.Name); err != nil {
		return err
	}
	if err := writeLine("Status", ini.Status); err != nil {
		return err
	}
	if ini.Description != nil {
		if err := writeLine("Description", *ini.Description); err != nil {
			return err
		}
	}
	return nil
}

func newInitiativeCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an initiative",
		RunE:  runInitiativeCreate,
	}
	f := cmd.Flags()
	f.String("name", "", "initiative name (required)")
	f.String("description", "", "initiative description")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runInitiativeCreate(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	f := cmd.Flags()
	name, _ := f.GetString("name")

	input := map[string]any{"name": name}
	if f.Changed("description") {
		v, _ := f.GetString("description")
		input["description"] = v
	}

	vars := map[string]any{"input": input}
	var result initiativePayloadResult
	if err := client.Do(context.Background(), query.InitiativeCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create initiative: %w", err)
	}
	if !result.InitiativeCreate.Success {
		return fmt.Errorf("create initiative: mutation returned success=false")
	}
	if result.InitiativeCreate.Initiative == nil {
		return fmt.Errorf("create initiative: no initiative in response")
	}

	ini := result.InitiativeCreate.Initiative
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), ini)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", ini.ID, ini.Name)
	return err
}

func newInitiativeUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an initiative",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("initiative id is required")
			}
			return nil
		},
		RunE: runInitiativeUpdate,
	}
	f := cmd.Flags()
	f.String("name", "", "initiative name")
	f.String("description", "", "initiative description")
	f.String("status", "", "initiative status (Active, Completed, Planned)")
	return cmd
}

func runInitiativeUpdate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

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
	if f.Changed("status") {
		v, _ := f.GetString("status")
		switch v {
		case "Active", "Completed", "Planned":
		default:
			return fmt.Errorf("invalid status %q: must be Active, Completed, or Planned", v)
		}
		input["status"] = v
	}
	if len(input) == 0 {
		return fmt.Errorf("no fields to update: specify at least one flag")
	}

	vars := map[string]any{"id": args[0], "input": input}
	var result initiativeUpdatePayloadResult
	if err := client.Do(context.Background(), query.InitiativeUpdateMutation, vars, &result); err != nil {
		return fmt.Errorf("update initiative: %w", err)
	}
	if !result.InitiativeUpdate.Success {
		return fmt.Errorf("update initiative: mutation returned success=false")
	}
	if result.InitiativeUpdate.Initiative == nil {
		return fmt.Errorf("update initiative: no initiative in response")
	}

	ini := result.InitiativeUpdate.Initiative
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), ini)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", ini.ID, ini.Name)
	return err
}

func newInitiativeDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an initiative",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("initiative id is required")
			}
			return nil
		},
		RunE: runInitiativeDelete,
	}
	cmd.Flags().Bool("yes", false, "skip confirmation prompt")
	return cmd
}

func runInitiativeDelete(cmd *cobra.Command, args []string) error {
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Delete initiative %s? [y/N] ", args[0])
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

	vars := map[string]any{"id": args[0]}
	var result initiativeDeleteResult
	if err := client.Do(context.Background(), query.InitiativeDeleteMutation, vars, &result); err != nil {
		return fmt.Errorf("delete initiative: %w", err)
	}
	if !result.InitiativeDelete.Success {
		return fmt.Errorf("delete initiative: mutation returned success=false")
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Initiative %s deleted.\n", args[0])
	return err
}

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type templateListResult struct {
	Templates []model.Template `json:"templates"`
}

type templateShowResult struct {
	Template *model.Template `json:"template"`
}

// TemplateRow is a display row for the template list table.
type TemplateRow struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func newTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage templates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newTemplateListCommand())
	cmd.AddCommand(newTemplateShowCommand())
	return cmd
}

func newTemplateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all templates",
		RunE:  runTemplateList,
	}
}

func runTemplateList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	var result templateListResult
	if err := client.Do(context.Background(), query.TemplateListQuery, nil, &result); err != nil {
		return fmt.Errorf("list templates: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.Templates)
	}

	rows := make([]TemplateRow, len(result.Templates))
	for i, t := range result.Templates {
		rows[i] = TemplateRow{
			Name: truncate(t.Name, 50),
			Type: t.Type,
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newTemplateShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show template details",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("template id is required")
			}
			return nil
		},
		RunE: runTemplateShow,
	}
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	vars := map[string]any{"id": args[0]}
	var result templateShowResult
	if err := client.Do(context.Background(), query.TemplateShowQuery, vars, &result); err != nil {
		return fmt.Errorf("show template: %w", err)
	}
	if result.Template == nil {
		return fmt.Errorf("template not found: %s", args[0])
	}

	t := result.Template
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), t)
	}

	w := cmd.OutOrStdout()
	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-14s %s\n", label+":", value)
		return err
	}

	if err := writeLine("Name", t.Name); err != nil {
		return err
	}
	if err := writeLine("Type", t.Type); err != nil {
		return err
	}
	if t.Description != nil {
		if err := writeLine("Description", *t.Description); err != nil {
			return err
		}
	}
	if len(t.TemplateData) > 0 {
		_, err = fmt.Fprintf(w, "%-14s %s\n", "TemplateData:", string(t.TemplateData))
		return err
	}
	return nil
}

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type docListResult struct {
	Documents struct {
		Nodes []model.Document `json:"nodes"`
	} `json:"documents"`
}

type docGetResult struct {
	Document *model.Document `json:"document"`
}

// DocRow is a display row for the doc list table.
type DocRow struct {
	Title   string `json:"title"`
	Project string `json:"project"`
	Creator string `json:"creator"`
	Updated string `json:"updated"`
}

func newDocCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc",
		Short: "Manage Linear documents",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newDocListCommand())
	cmd.AddCommand(newDocShowCommand())
	return cmd
}

func newDocListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List documents",
		RunE:  runDocList,
	}
	f := cmd.Flags()
	f.String("project", "", "filter by project ID or name")
	f.Int("limit", 50, "maximum number of documents to return")
	f.Bool("include-archived", false, "include archived documents")
	return cmd
}

func runDocList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	f := cmd.Flags()
	limit, _ := f.GetInt("limit")
	includeArchived, _ := f.GetBool("include-archived")
	projectID, _ := f.GetString("project")

	vars := map[string]any{"first": limit}
	if includeArchived {
		vars["includeArchived"] = true
	}

	if projectID != "" {
		vars["filter"] = map[string]any{
			"project": map[string]any{
				"id": map[string]any{"eq": projectID},
			},
		}
	}

	var result docListResult
	if err := client.Do(context.Background(), query.DocumentListQuery, vars, &result); err != nil {
		return fmt.Errorf("list documents: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.Documents.Nodes)
	}

	rows := make([]DocRow, len(result.Documents.Nodes))
	for i, d := range result.Documents.Nodes {
		rows[i] = DocRow{
			Title:   truncate(d.Title, 40),
			Updated: d.UpdatedAt,
		}
		if d.Project != nil {
			rows[i].Project = d.Project.Name
		}
		if d.Creator != nil {
			rows[i].Creator = d.Creator.DisplayName
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newDocShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show details of a document",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("document id is required")
			}
			return nil
		},
		RunE: runDocShow,
	}
}

func runDocShow(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	vars := map[string]any{"id": args[0]}
	var result docGetResult
	if err := client.Do(context.Background(), query.DocumentGetQuery, vars, &result); err != nil {
		return fmt.Errorf("get document: %w", err)
	}
	if result.Document == nil {
		return fmt.Errorf("document %q not found", args[0])
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), result.Document)
	}

	return printDocDetail(cmd, result.Document)
}

func printDocDetail(cmd *cobra.Command, d *model.Document) error {
	w := cmd.OutOrStdout()

	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-10s %s\n", label+":", value)
		return err
	}

	if err := writeLine("Title", d.Title); err != nil {
		return err
	}
	if d.Project != nil {
		if err := writeLine("Project", d.Project.Name); err != nil {
			return err
		}
	}
	if d.Creator != nil {
		if err := writeLine("Creator", d.Creator.DisplayName); err != nil {
			return err
		}
	}
	if err := writeLine("Created", d.CreatedAt); err != nil {
		return err
	}
	if err := writeLine("Updated", d.UpdatedAt); err != nil {
		return err
	}
	if d.URL != "" {
		if err := writeLine("URL", d.URL); err != nil {
			return err
		}
	}

	if d.Content != nil && *d.Content != "" {
		_, err := fmt.Fprintf(w, "\n%s\n", *d.Content)
		return err
	}
	return nil
}

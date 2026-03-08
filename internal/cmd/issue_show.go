package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"linear-cli/internal/api"
	"linear-cli/internal/config"
	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type issueGetResult struct {
	Issue *model.Issue `json:"issue"`
}

func newIssueShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <identifier>",
		Short: "Show details of an issue",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runIssueShow,
	}
}

func runIssueShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("not authenticated: run 'linear auth' first")
	}

	var opts []api.Option
	if ep := os.Getenv("LINEAR_API_ENDPOINT"); ep != "" {
		opts = append(opts, api.WithEndpoint(ep))
	}
	client := api.NewClient(cfg.APIKey, opts...)

	identifier := args[0]
	vars := map[string]any{"id": identifier}

	var result issueGetResult
	if err := client.Do(context.Background(), query.IssueGetQuery, vars, &result); err != nil {
		return fmt.Errorf("get issue: %w", err)
	}
	if result.Issue == nil {
		return fmt.Errorf("issue %q not found", identifier)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		formatter := output.NewFormatter(true)
		return formatter.Format(cmd.OutOrStdout(), result.Issue)
	}

	return printIssueDetail(cmd, result.Issue)
}

func printIssueDetail(cmd *cobra.Command, issue *model.Issue) error {
	w := cmd.OutOrStdout()

	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-14s %s\n", label+":", value)
		return err
	}

	fields := []struct{ label, value string }{
		{"Identifier", issue.Identifier},
		{"Title", issue.Title},
		{"Status", issue.State.Name},
		{"Priority", issue.PriorityLabel},
		{"Team", issue.Team.Name},
	}
	for _, f := range fields {
		if err := writeLine(f.label, f.value); err != nil {
			return err
		}
	}

	assignee := ""
	if issue.Assignee != nil {
		assignee = issue.Assignee.DisplayName
	}
	if err := writeLine("Assignee", assignee); err != nil {
		return err
	}

	if issue.DueDate != nil {
		if err := writeLine("Due Date", *issue.DueDate); err != nil {
			return err
		}
	}

	if issue.Estimate != nil {
		if err := writeLine("Estimate", fmt.Sprintf("%.0f", *issue.Estimate)); err != nil {
			return err
		}
	}

	labels := make([]string, len(issue.Labels.Nodes))
	for i, l := range issue.Labels.Nodes {
		labels[i] = l.Name
	}
	if len(labels) > 0 {
		if err := writeLine("Labels", strings.Join(labels, ", ")); err != nil {
			return err
		}
	}

	for _, f := range []struct{ label, value string }{
		{"URL", issue.URL},
		{"Created", issue.CreatedAt},
		{"Updated", issue.UpdatedAt},
	} {
		if err := writeLine(f.label, f.value); err != nil {
			return err
		}
	}

	if issue.Description != nil && *issue.Description != "" {
		_, err := fmt.Fprintf(w, "\n%s\n", *issue.Description)
		return err
	}

	return nil
}

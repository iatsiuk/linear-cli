package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/iatsiuk/linear-cli/internal/model"
	"github.com/iatsiuk/linear-cli/internal/output"
	"github.com/iatsiuk/linear-cli/internal/query"
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
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

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

	if issue.Parent != nil {
		if err := writeLine("Parent", issue.Parent.Identifier+": "+issue.Parent.Title); err != nil {
			return err
		}
	}

	if issue.Project != nil {
		if err := writeLine("Project", issue.Project.Name); err != nil {
			return err
		}
	}

	if issue.Cycle != nil {
		name := fmt.Sprintf("#%.0f", issue.Cycle.Number)
		if issue.Cycle.Name != nil && *issue.Cycle.Name != "" {
			name += " " + *issue.Cycle.Name
		}
		if err := writeLine("Cycle", name); err != nil {
			return err
		}
	}

	if issue.Creator != nil {
		if err := writeLine("Creator", issue.Creator.DisplayName); err != nil {
			return err
		}
	}

	if issue.BranchName != "" {
		if err := writeLine("Branch", issue.BranchName); err != nil {
			return err
		}
	}

	if issue.Number != 0 {
		if err := writeLine("Number", fmt.Sprintf("%.0f", issue.Number)); err != nil {
			return err
		}
	}

	if issue.CustomerTicketCount != 0 {
		if err := writeLine("Tickets", fmt.Sprintf("%d", issue.CustomerTicketCount)); err != nil {
			return err
		}
	}

	if issue.Trashed != nil && *issue.Trashed {
		if err := writeLine("Trashed", "yes"); err != nil {
			return err
		}
	}

	for _, f := range []struct {
		label string
		value *string
	}{
		{"Started", issue.StartedAt},
		{"Completed", issue.CompletedAt},
		{"Canceled", issue.CanceledAt},
		{"Triaged", issue.TriagedAt},
		{"Triage Start", issue.StartedTriageAt},
		{"Snoozed Until", issue.SnoozedUntilAt},
		{"Archived", issue.ArchivedAt},
		{"AutoArchived", issue.AutoArchivedAt},
		{"AutoClosed", issue.AutoClosedAt},
		{"Added to Cycle", issue.AddedToCycleAt},
		{"Added to Proj", issue.AddedToProjectAt},
		{"Added to Team", issue.AddedToTeamAt},
		{"SLA Breach", issue.SlaBreachesAt},
		{"SLA Started", issue.SlaStartedAt},
		{"SLA High Risk", issue.SlaHighRiskAt},
		{"SLA Med Risk", issue.SlaMediumRiskAt},
	} {
		if f.value != nil {
			if err := writeLine(f.label, *f.value); err != nil {
				return err
			}
		}
	}

	if issue.SlaType != nil {
		if err := writeLine("SLA Type", *issue.SlaType); err != nil {
			return err
		}
	}

	if issue.Description != nil && *issue.Description != "" {
		_, err := fmt.Fprintf(w, "\n%s\n", *issue.Description)
		return err
	}

	return nil
}

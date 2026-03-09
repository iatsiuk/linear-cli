package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type relationListResult struct {
	Issue *struct {
		Relations        model.IssueRelationConnection `json:"relations"`
		InverseRelations model.IssueRelationConnection `json:"inverseRelations"`
	} `json:"issue"`
}

type relationCreateResult struct {
	IssueRelationCreate struct {
		Success       bool                 `json:"success"`
		IssueRelation *model.IssueRelation `json:"issueRelation"`
	} `json:"issueRelationCreate"`
}

type relationDeleteResult struct {
	IssueRelationDelete struct {
		Success  bool   `json:"success"`
		EntityID string `json:"entityId"`
	} `json:"issueRelationDelete"`
}

// RelationRow is a display row for the relation list table.
type RelationRow struct {
	Type         string `json:"type"`
	Direction    string `json:"direction"`
	RelatedIssue string `json:"related_issue"`
	Title        string `json:"title"`
}

func newRelationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relation",
		Short: "Manage issue relations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newRelationListCommand())
	cmd.AddCommand(newRelationCreateCommand())
	cmd.AddCommand(newRelationDeleteCommand())
	return cmd
}

func newRelationListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <issue-identifier>",
		Short: "List relations for an issue",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("issue identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runRelationList,
	}
}

func runRelationList(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	identifier := args[0]
	var result relationListResult
	if err := client.Do(context.Background(), query.RelationListQuery, map[string]any{"issueId": identifier}, &result); err != nil {
		return fmt.Errorf("list relations: %w", err)
	}
	if result.Issue == nil {
		return fmt.Errorf("issue %q not found", identifier)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")

	type relEntry struct {
		model.IssueRelation
		Direction string
	}

	var entries []relEntry
	for _, r := range result.Issue.Relations.Nodes {
		entries = append(entries, relEntry{r, "outgoing"})
	}
	for _, r := range result.Issue.InverseRelations.Nodes {
		entries = append(entries, relEntry{r, "incoming"})
	}

	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), entries)
	}

	rows := make([]RelationRow, len(entries))
	for i, e := range entries {
		var relIdentifier, relTitle string
		if e.Direction == "outgoing" {
			relIdentifier = e.RelatedIssue.Identifier
			relTitle = e.RelatedIssue.Title
		} else {
			relIdentifier = e.Issue.Identifier
			relTitle = e.Issue.Title
		}
		rows[i] = RelationRow{
			Type:         e.Type,
			Direction:    e.Direction,
			RelatedIssue: relIdentifier,
			Title:        truncate(relTitle, 40),
		}
	}
	return output.NewFormatter(false).Format(cmd.OutOrStdout(), rows)
}

func newRelationCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <issue-identifier>",
		Short: "Create a relation between issues",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("issue identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runRelationCreate,
	}
	f := cmd.Flags()
	f.String("related", "", "identifier of the related issue (required)")
	f.String("type", "related", "relation type: blocks|duplicate|related|similar")
	_ = cmd.MarkFlagRequired("related")
	return cmd
}

func runRelationCreate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	identifier := args[0]
	f := cmd.Flags()
	related, _ := f.GetString("related")
	relType, _ := f.GetString("type")

	input := map[string]any{
		"issueId":        identifier,
		"relatedIssueId": related,
		"type":           relType,
	}
	vars := map[string]any{"input": input}

	var result relationCreateResult
	if err := client.Do(ctx, query.RelationCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create relation: %w", err)
	}
	if !result.IssueRelationCreate.Success {
		return fmt.Errorf("create relation: mutation returned success=false")
	}
	if result.IssueRelationCreate.IssueRelation == nil {
		return fmt.Errorf("create relation: no relation in response")
	}

	rel := result.IssueRelationCreate.IssueRelation
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), rel)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", rel.ID)
	return err
}

func newRelationDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <relation-id>",
		Short: "Delete an issue relation",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("relation ID is required")
			}
			return nil
		},
		RunE: runRelationDelete,
	}
	cmd.Flags().Bool("yes", false, "skip confirmation prompt")
	return cmd
}

func runRelationDelete(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	relationID := args[0]
	yes, _ := cmd.Flags().GetBool("yes")

	if !yes {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete relation %s? [y/N] ", relationID)
		scanner := bufio.NewScanner(cmd.InOrStdin())
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())
		if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
			return fmt.Errorf("aborted")
		}
	}

	var result relationDeleteResult
	if err := client.Do(ctx, query.RelationDeleteMutation, map[string]any{"id": relationID}, &result); err != nil {
		return fmt.Errorf("delete relation: %w", err)
	}
	if !result.IssueRelationDelete.Success {
		return fmt.Errorf("delete relation: mutation returned success=false")
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Relation %s deleted.\n", result.IssueRelationDelete.EntityID)
	return err
}

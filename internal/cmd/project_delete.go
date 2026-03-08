package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"linear-cli/internal/api"
	"linear-cli/internal/query"
)

type projectDeleteResult struct {
	ProjectDelete struct {
		Success bool `json:"success"`
	} `json:"projectDelete"`
}

func newProjectDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("exactly one project id required")
			}
			return nil
		},
		RunE: runProjectDelete,
	}
	f := cmd.Flags()
	f.Bool("yes", false, "skip confirmation prompt")
	return cmd
}

func runProjectDelete(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	id, err := api.ResolveProjectID(ctx, client, args[0])
	if err != nil {
		return err
	}

	// fetch project to confirm it exists
	var getResult projectGetResult
	if err := client.Do(ctx, query.ProjectGetQuery, map[string]any{"id": id}, &getResult); err != nil {
		return fmt.Errorf("get project: %w", err)
	}
	if getResult.Project == nil {
		return fmt.Errorf("project %q not found", id)
	}

	f := cmd.Flags()
	yes, _ := f.GetBool("yes")

	if !yes {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete project %q? [y/N] ", getResult.Project.Name)
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

	var result projectDeleteResult
	if err := client.Do(ctx, query.ProjectDeleteMutation, map[string]any{"id": id}, &result); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if !result.ProjectDelete.Success {
		return fmt.Errorf("delete project: mutation returned success=false")
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project %q deleted.\n", getResult.Project.Name)
	return nil
}

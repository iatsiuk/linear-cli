package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"linear-cli/internal/api"
	"linear-cli/internal/config"
	"linear-cli/internal/query"
)

type issueDeleteResult struct {
	IssueDelete struct {
		Success bool `json:"success"`
	} `json:"issueDelete"`
}

type issueArchiveResult struct {
	IssueArchive struct {
		Success bool `json:"success"`
	} `json:"issueArchive"`
}

func newIssueDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <identifier>",
		Short: "Delete (trash) an issue",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runIssueDelete,
	}
	f := cmd.Flags()
	f.Bool("archive", false, "archive instead of trash (use issueArchive)")
	f.Bool("yes", false, "skip confirmation prompt")
	return cmd
}

func runIssueDelete(cmd *cobra.Command, args []string) error {
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
	ctx := context.Background()

	identifier := args[0]

	// fetch issue to get its UUID
	var getResult issueGetResult
	if err := client.Do(ctx, query.IssueGetQuery, map[string]any{"id": identifier}, &getResult); err != nil {
		return fmt.Errorf("get issue: %w", err)
	}
	if getResult.Issue == nil {
		return fmt.Errorf("issue %q not found", identifier)
	}
	issueID := getResult.Issue.ID

	f := cmd.Flags()
	archive, _ := f.GetBool("archive")
	yes, _ := f.GetBool("yes")

	if !yes {
		action := "trash"
		if archive {
			action = "archive"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to %s issue %s? [y/N] ", action, identifier)
		scanner := bufio.NewScanner(cmd.InOrStdin())
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())
		if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
			return fmt.Errorf("aborted")
		}
	}

	if archive {
		var result issueArchiveResult
		if err := client.Do(ctx, query.IssueArchiveMutation, map[string]any{"id": issueID}, &result); err != nil {
			return fmt.Errorf("archive issue: %w", err)
		}
		if !result.IssueArchive.Success {
			return fmt.Errorf("archive issue: mutation returned success=false")
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Issue %s archived.\n", identifier)
		return nil
	}

	var result issueDeleteResult
	if err := client.Do(ctx, query.IssueDeleteMutation, map[string]any{"id": issueID}, &result); err != nil {
		return fmt.Errorf("delete issue: %w", err)
	}
	if !result.IssueDelete.Success {
		return fmt.Errorf("delete issue: mutation returned success=false")
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Issue %s deleted.\n", identifier)
	return nil
}

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type issueBranchResult struct {
	Issue *model.Issue `json:"issueVcsBranchSearch"`
}

func newIssueBranchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "branch [branch-name]",
		Short: "Show issue linked to a git branch",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runIssueBranch,
	}
}

func runIssueBranch(cmd *cobra.Command, args []string) error {
	branchName, err := resolveBranchName(args)
	if err != nil {
		return err
	}

	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	vars := map[string]any{"branchName": branchName}

	var result issueBranchResult
	if err := client.Do(context.Background(), query.IssueBranchQuery, vars, &result); err != nil {
		return fmt.Errorf("branch lookup: %w", err)
	}
	if result.Issue == nil {
		return fmt.Errorf("no issue found for branch %q", branchName)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		formatter := output.NewFormatter(true)
		return formatter.Format(cmd.OutOrStdout(), result.Issue)
	}

	return printIssueDetail(cmd, result.Issue)
}

func resolveBranchName(args []string) (string, error) {
	if len(args) > 0 {
		branch := strings.TrimSpace(args[0])
		if branch == "" {
			return "", fmt.Errorf("branch name cannot be empty")
		}
		return branch, nil
	}
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("get current git branch: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("get current git branch: %w", err)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" || branch == "HEAD" {
		return "", fmt.Errorf("could not determine current git branch")
	}
	return branch, nil
}

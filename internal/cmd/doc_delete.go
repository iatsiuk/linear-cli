package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/iatsiuk/linear-cli/internal/query"
)

type docDeleteResult struct {
	DocumentDelete struct {
		Success bool `json:"success"`
	} `json:"documentDelete"`
}

type docUnarchiveResult struct {
	DocumentUnarchive struct {
		Success bool `json:"success"`
	} `json:"documentUnarchive"`
}

func newDocDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Move a document to trash (30-day grace period)",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("document id is required")
			}
			return nil
		},
		RunE: runDocDelete,
	}
	f := cmd.Flags()
	f.Bool("restore", false, "restore document from trash instead of deleting")
	f.Bool("yes", false, "skip confirmation prompt")
	return cmd
}

func runDocDelete(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	docID := args[0]
	f := cmd.Flags()
	restore, _ := f.GetBool("restore")
	yes, _ := f.GetBool("yes")

	if !yes {
		action := "trash"
		if restore {
			action = "restore"
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Are you sure you want to %s document %s? [y/N] ", action, docID)
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

	if restore {
		var result docUnarchiveResult
		if err := client.Do(ctx, query.DocumentUnarchiveMutation, map[string]any{"id": docID}, &result); err != nil {
			return fmt.Errorf("restore document: %w", err)
		}
		if !result.DocumentUnarchive.Success {
			return fmt.Errorf("restore document: mutation returned success=false")
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Document %s restored.\n", docID)
		return nil
	}

	var result docDeleteResult
	if err := client.Do(ctx, query.DocumentDeleteMutation, map[string]any{"id": docID}, &result); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	if !result.DocumentDelete.Success {
		return fmt.Errorf("delete document: mutation returned success=false")
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Document %s moved to trash.\n", docID)
	return nil
}

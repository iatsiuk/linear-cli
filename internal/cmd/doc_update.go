package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type docUpdateResult struct {
	DocumentUpdate struct {
		Success  bool            `json:"success"`
		Document *model.Document `json:"document"`
	} `json:"documentUpdate"`
}

func newDocUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a document",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("document id is required")
			}
			return nil
		},
		RunE: runDocUpdate,
	}
	f := cmd.Flags()
	f.String("title", "", "new document title")
	f.String("content", "", "new document content in markdown")
	f.String("content-file", "", "read content from file")
	return cmd
}

func runDocUpdate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	docID := args[0]
	f := cmd.Flags()

	content, _ := f.GetString("content")
	contentFile, _ := f.GetString("content-file")

	if contentFile != "" && content != "" {
		return fmt.Errorf("--content and --content-file are mutually exclusive")
	}

	if contentFile != "" {
		data, err := os.ReadFile(contentFile)
		if err != nil {
			return fmt.Errorf("read content file: %w", err)
		}
		if len(data) == 0 {
			return fmt.Errorf("content file is empty")
		}
		content = string(data)
	}

	input := map[string]any{}
	if f.Changed("title") {
		title, _ := f.GetString("title")
		input["title"] = title
	}
	if f.Changed("content") || contentFile != "" {
		input["content"] = content
	}

	if len(input) == 0 {
		return fmt.Errorf("no fields to update: specify at least one flag")
	}

	vars := map[string]any{"id": docID, "input": input}
	var result docUpdateResult
	if err := client.Do(ctx, query.DocumentUpdateMutation, vars, &result); err != nil {
		return fmt.Errorf("update document: %w", err)
	}
	if !result.DocumentUpdate.Success {
		return fmt.Errorf("update document: mutation returned success=false")
	}
	if result.DocumentUpdate.Document == nil {
		return fmt.Errorf("update document: no document in response")
	}

	doc := result.DocumentUpdate.Document
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), doc)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Updated document: %s\n", doc.Title)
	return err
}

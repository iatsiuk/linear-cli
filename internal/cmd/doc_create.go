package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"linear-cli/internal/api"
	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type docCreateResult struct {
	DocumentCreate struct {
		Success  bool            `json:"success"`
		Document *model.Document `json:"document"`
	} `json:"documentCreate"`
}

func newDocCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new document",
		RunE:  runDocCreate,
	}
	f := cmd.Flags()
	f.String("title", "", "document title (required)")
	f.String("content", "", "document content in markdown")
	f.String("content-file", "", "read content from file")
	f.String("project", "", "project name or ID")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func runDocCreate(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	f := cmd.Flags()
	title, _ := f.GetString("title")
	content, _ := f.GetString("content")
	contentFile, _ := f.GetString("content-file")
	project, _ := f.GetString("project")

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

	input := map[string]any{"title": title}
	if content != "" {
		input["content"] = content
	}
	if project != "" {
		projectID, err := api.ResolveProjectID(ctx, client, project)
		if err != nil {
			return err
		}
		input["projectId"] = projectID
	}

	vars := map[string]any{"input": input}
	var result docCreateResult
	if err := client.Do(ctx, query.DocumentCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create document: %w", err)
	}
	if !result.DocumentCreate.Success {
		return fmt.Errorf("create document: mutation returned success=false")
	}
	if result.DocumentCreate.Document == nil {
		return fmt.Errorf("create document: no document in response")
	}

	doc := result.DocumentCreate.Document
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), doc)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created document: %s\n", doc.Title)
	return err
}

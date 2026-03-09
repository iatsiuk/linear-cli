package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/iatsiuk/linear-cli/internal/model"
	"github.com/iatsiuk/linear-cli/internal/output"
	"github.com/iatsiuk/linear-cli/internal/query"
)

type attachmentShowResult struct {
	Attachment *model.Attachment `json:"attachment"`
}

type attachmentListResult struct {
	Issue *struct {
		Attachments struct {
			Nodes    []model.Attachment `json:"nodes"`
			PageInfo struct {
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"attachments"`
	} `json:"issue"`
}

type attachmentCreateResult struct {
	AttachmentCreate struct {
		Success    bool              `json:"success"`
		Attachment *model.Attachment `json:"attachment"`
	} `json:"attachmentCreate"`
}

type attachmentDeleteResult struct {
	AttachmentDelete struct {
		Success bool `json:"success"`
	} `json:"attachmentDelete"`
}

// AttachmentRow is a display row for the attachment list table.
type AttachmentRow struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Created string `json:"created"`
}

func newAttachmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment",
		Short: "Manage Linear issue attachments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newAttachmentListCommand())
	cmd.AddCommand(newAttachmentShowCommand())
	cmd.AddCommand(newAttachmentCreateCommand())
	cmd.AddCommand(newAttachmentDeleteCommand())
	cmd.AddCommand(newAttachmentDownloadCommand())
	return cmd
}

func newAttachmentShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show attachment metadata",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("attachment id is required")
			}
			return nil
		},
		RunE: runAttachmentShow,
	}
}

func runAttachmentShow(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	vars := map[string]any{"id": args[0]}
	var result attachmentShowResult
	if err := client.Do(cmd.Context(), query.AttachmentShowQuery, vars, &result); err != nil {
		return fmt.Errorf("show attachment: %w", err)
	}
	if result.Attachment == nil {
		return fmt.Errorf("attachment %q not found", args[0])
	}

	a := result.Attachment
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), a)
	}

	w := cmd.OutOrStdout()
	writeLine := func(label, value string) error {
		_, e := fmt.Fprintf(w, "%-10s %s\n", label+":", value)
		return e
	}

	if err := writeLine("Title", a.Title); err != nil {
		return err
	}
	if err := writeLine("URL", a.URL); err != nil {
		return err
	}
	if a.Creator != nil {
		if err := writeLine("Creator", a.Creator.DisplayName); err != nil {
			return err
		}
	}
	if err := writeLine("Created", a.CreatedAt); err != nil {
		return err
	}
	return writeLine("Updated", a.UpdatedAt)
}

func newAttachmentDownloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <id>",
		Short: "Download an attachment file",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("attachment id is required")
			}
			return nil
		},
		RunE: runAttachmentDownload,
	}
	cmd.Flags().StringP("output", "o", "", "destination path ('-' for stdout)")
	return cmd
}

func runAttachmentDownload(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	attID := args[0]
	vars := map[string]any{"id": attID}
	var result attachmentShowResult
	if err := client.Do(cmd.Context(), query.AttachmentShowQuery, vars, &result); err != nil {
		return fmt.Errorf("show attachment: %w", err)
	}
	if result.Attachment == nil {
		return fmt.Errorf("attachment %q not found", attID)
	}

	fileURL := result.Attachment.URL
	outputFlag, _ := cmd.Flags().GetString("output")

	// determine destination: stdout, explicit path, or filename from URL
	toStdout := outputFlag == "-"
	dest := outputFlag
	if !toStdout && dest == "" {
		dest = filenameFromURL(fileURL)
		if dest == "" {
			dest = "attachment-" + attID
		}
	}

	// download file
	dlClient := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, fileURL, nil)
	if err != nil {
		return fmt.Errorf("build download request: %w", err)
	}
	resp, err := dlClient.Do(req) //nolint:gosec // URL comes from Linear API response
	if err != nil {
		return fmt.Errorf("download attachment: %w", err)
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download attachment: unexpected status %d", resp.StatusCode)
	}

	if toStdout {
		_, err = io.Copy(cmd.OutOrStdout(), resp.Body)
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	n, copyErr := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(dest)
		return fmt.Errorf("write file: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close file: %w", closeErr)
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Downloaded: %s (%d bytes)\n", dest, n)
	return err
}

// filenameFromURL returns the last non-empty path segment of rawURL, or "".
func filenameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	base := path.Base(u.Path)
	if base == "." || base == "/" {
		return ""
	}
	return base
}

func newAttachmentListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <issue-identifier>",
		Short: "List attachments for an issue",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("issue identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runAttachmentList,
	}
}

func runAttachmentList(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	identifier := args[0]
	ctx := context.Background()

	vars := map[string]any{"issueId": identifier}
	var listResult attachmentListResult
	if err := client.Do(ctx, query.AttachmentListQuery, vars, &listResult); err != nil {
		return fmt.Errorf("list attachments: %w", err)
	}
	if listResult.Issue == nil {
		return fmt.Errorf("issue %q not found", identifier)
	}
	attachments := listResult.Issue.Attachments.Nodes
	if listResult.Issue.Attachments.PageInfo.HasNextPage {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: showing first 50 attachments; more exist\n")
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), attachments)
	}

	rows := make([]AttachmentRow, len(attachments))
	for i, a := range attachments {
		rows[i] = AttachmentRow{
			Title:   truncate(a.Title, 40),
			URL:     truncate(a.URL, 50),
			Created: a.CreatedAt,
		}
	}
	return output.NewFormatter(false).Format(cmd.OutOrStdout(), rows)
}

func newAttachmentCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <issue-identifier>",
		Short: "Create an attachment for an issue",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("issue identifier is required (e.g. ENG-42)")
			}
			return nil
		},
		RunE: runAttachmentCreate,
	}
	f := cmd.Flags()
	f.String("url", "", "attachment URL (mutually exclusive with --file)")
	f.String("title", "", "attachment title (required)")
	f.String("file", "", "local file to upload and attach")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func runAttachmentCreate(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	identifier := args[0]
	f := cmd.Flags()
	urlFlag, _ := f.GetString("url")
	title, _ := f.GetString("title")
	fileFlag, _ := f.GetString("file")

	if urlFlag == "" && fileFlag == "" {
		return fmt.Errorf("one of --url or --file is required")
	}
	if urlFlag != "" && fileFlag != "" {
		return fmt.Errorf("--url and --file are mutually exclusive")
	}

	attachmentURL := urlFlag
	if fileFlag != "" {
		assetURL, uploadErr := client.Upload(ctx, fileFlag)
		if uploadErr != nil {
			return fmt.Errorf("upload file: %w", uploadErr)
		}
		attachmentURL = assetURL
	}

	input := map[string]any{
		"issueId": identifier,
		"url":     attachmentURL,
		"title":   title,
	}

	vars := map[string]any{"input": input}
	var result attachmentCreateResult
	if err := client.Do(ctx, query.AttachmentCreateMutation, vars, &result); err != nil {
		return fmt.Errorf("create attachment: %w", err)
	}
	if !result.AttachmentCreate.Success {
		return fmt.Errorf("create attachment: mutation returned success=false")
	}
	if result.AttachmentCreate.Attachment == nil {
		return fmt.Errorf("create attachment: no attachment in response")
	}

	a := result.AttachmentCreate.Attachment
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), a)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", a.ID)
	return err
}

func newAttachmentDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an attachment",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("attachment id is required")
			}
			return nil
		},
		RunE: runAttachmentDelete,
	}
	f := cmd.Flags()
	f.Bool("yes", false, "skip confirmation prompt")
	return cmd
}

func runAttachmentDelete(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	attID := args[0]
	yes, _ := cmd.Flags().GetBool("yes")

	if !yes {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Are you sure you want to delete attachment %s? [y/N] ", attID)
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

	var result attachmentDeleteResult
	if err := client.Do(ctx, query.AttachmentDeleteMutation, map[string]any{"id": attID}, &result); err != nil {
		return fmt.Errorf("delete attachment: %w", err)
	}
	if !result.AttachmentDelete.Success {
		return fmt.Errorf("delete attachment: mutation returned success=false")
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Attachment %s deleted.\n", attID)
	return err
}

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type orgInfo struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	URLKey  string  `json:"urlKey"`
	LogoURL *string `json:"logoUrl,omitempty"`
}

type orgResult struct {
	Organization orgInfo `json:"organization"`
}

func newOrgCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "org",
		Short: "Show organization info",
		RunE:  runOrg,
	}
}

func runOrg(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	var result orgResult
	if err := client.Do(context.Background(), query.OrganizationQuery, nil, &result); err != nil {
		return fmt.Errorf("organization query: %w", err)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), result.Organization)
	}

	return printOrgDetail(cmd, &result.Organization)
}

func printOrgDetail(cmd *cobra.Command, org *orgInfo) error {
	w := cmd.OutOrStdout()

	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-8s %s\n", label+":", value)
		return err
	}

	fields := []struct{ label, value string }{
		{"Name", org.Name},
		{"URL key", org.URLKey},
	}
	if org.LogoURL != nil {
		fields = append(fields, struct{ label, value string }{"Logo", *org.LogoURL})
	}

	for _, f := range fields {
		if err := writeLine(f.label, f.value); err != nil {
			return err
		}
	}
	return nil
}

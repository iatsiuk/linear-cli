package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iatsiuk/linear-cli/internal/api"
	"github.com/iatsiuk/linear-cli/internal/model"
	"github.com/iatsiuk/linear-cli/internal/output"
	"github.com/iatsiuk/linear-cli/internal/query"
)

type viewerTeam struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type viewerUser struct {
	model.User
	Teams struct {
		Nodes []viewerTeam `json:"nodes"`
	} `json:"teams"`
}

type viewerResult struct {
	Viewer viewerUser `json:"viewer"`
}

type viewerIssueNode struct {
	ID         string              `json:"id"`
	Identifier string              `json:"identifier"`
	Title      string              `json:"title"`
	State      model.WorkflowState `json:"state"`
	Team       viewerTeam          `json:"team"`
}

type viewerIssuesResult struct {
	Viewer struct {
		AssignedIssues *struct {
			Nodes []viewerIssueNode `json:"nodes"`
		} `json:"assignedIssues,omitempty"`
		CreatedIssues *struct {
			Nodes []viewerIssueNode `json:"nodes"`
		} `json:"createdIssues,omitempty"`
	} `json:"viewer"`
}

func newMeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Show current authenticated user",
		RunE:  runMe,
	}
	cmd.Flags().Bool("assigned", false, "show issues assigned to me")
	cmd.Flags().Bool("created", false, "show issues created by me")
	return cmd
}

func runMe(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	assigned, _ := cmd.Flags().GetBool("assigned")
	created, _ := cmd.Flags().GetBool("created")
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")

	if assigned && created {
		return fmt.Errorf("--assigned and --created are mutually exclusive")
	}
	if assigned || created {
		return runMeIssues(cmd, client, assigned, jsonMode)
	}

	var result viewerResult
	if err := client.Do(context.Background(), query.ViewerQuery, nil, &result); err != nil {
		return fmt.Errorf("viewer query: %w", err)
	}

	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), result.Viewer)
	}

	return printViewerDetail(cmd, &result.Viewer)
}

func runMeIssues(cmd *cobra.Command, client *api.Client, assigned bool, jsonMode bool) error {
	q := query.ViewerCreatedIssuesQuery
	if assigned {
		q = query.ViewerAssignedIssuesQuery
	}

	var result viewerIssuesResult
	if err := client.Do(context.Background(), q, nil, &result); err != nil {
		return fmt.Errorf("viewer issues query: %w", err)
	}

	var nodes []viewerIssueNode
	if assigned && result.Viewer.AssignedIssues != nil {
		nodes = result.Viewer.AssignedIssues.Nodes
	} else if !assigned && result.Viewer.CreatedIssues != nil {
		nodes = result.Viewer.CreatedIssues.Nodes
	}

	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), nodes)
	}

	type meIssueRow struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Status string `json:"status"`
	}
	rows := make([]meIssueRow, len(nodes))
	for i, n := range nodes {
		rows[i] = meIssueRow{
			ID:     n.Identifier,
			Title:  truncate(n.Title, 40),
			Status: n.State.Name,
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func printViewerDetail(cmd *cobra.Command, u *viewerUser) error {
	w := cmd.OutOrStdout()

	role := userRole(u.User)

	active := "yes"
	if !u.Active {
		active = "no"
	}

	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-10s %s\n", label+":", value)
		return err
	}

	for _, f := range []struct{ label, value string }{
		{"Name", u.DisplayName},
		{"Email", u.Email},
		{"Role", role},
		{"Active", active},
	} {
		if err := writeLine(f.label, f.value); err != nil {
			return err
		}
	}

	for _, t := range u.Teams.Nodes {
		if _, err := fmt.Fprintf(w, "  team: %s (%s)\n", t.Name, t.Key); err != nil {
			return err
		}
	}

	return nil
}

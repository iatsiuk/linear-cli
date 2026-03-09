package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/iatsiuk/linear-cli/internal/api"
	"github.com/iatsiuk/linear-cli/internal/model"
	"github.com/iatsiuk/linear-cli/internal/output"
	"github.com/iatsiuk/linear-cli/internal/query"
)

type teamMemberListResult struct {
	Team *struct {
		Memberships model.TeamMembershipConnection `json:"memberships"`
	} `json:"team"`
}

type teamMemberAddResult struct {
	TeamMembershipCreate struct {
		Success        bool                  `json:"success"`
		TeamMembership *model.TeamMembership `json:"teamMembership"`
	} `json:"teamMembershipCreate"`
}

type teamMemberRemoveResult struct {
	TeamMembershipDelete struct {
		Success bool `json:"success"`
	} `json:"teamMembershipDelete"`
}

// TeamMemberRow is a display row for the team member list table.
type TeamMemberRow struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func newTeamMemberCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage team members",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newTeamMemberListCommand())
	cmd.AddCommand(newTeamMemberAddCommand())
	cmd.AddCommand(newTeamMemberRemoveCommand())
	return cmd
}

func newTeamMemberListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <team-key>",
		Short: "List members of a team",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("team key or id is required")
			}
			return nil
		},
		RunE: runTeamMemberList,
	}
	cmd.Flags().Int("limit", 50, "maximum number of members to return")
	return cmd
}

func runTeamMemberList(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	teamID, err := api.ResolveTeamID(ctx, client, args[0])
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	vars := map[string]any{"teamId": teamID, "first": limit}

	var result teamMemberListResult
	if err := client.Do(ctx, query.TeamMemberListQuery, vars, &result); err != nil {
		return fmt.Errorf("list team members: %w", err)
	}
	if result.Team == nil {
		return fmt.Errorf("team not found: %s", teamID)
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), result.Team.Memberships.Nodes)
	}

	rows := make([]TeamMemberRow, len(result.Team.Memberships.Nodes))
	for i, m := range result.Team.Memberships.Nodes {
		role := "Member"
		if m.Owner {
			role = "Owner"
		}
		rows[i] = TeamMemberRow{
			Name:  m.User.DisplayName,
			Email: m.User.Email,
			Role:  role,
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newTeamMemberAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <team-key> <user>",
		Short: "Add a user to a team",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("team key and user are required")
			}
			return nil
		},
		RunE: runTeamMemberAdd,
	}
}

func runTeamMemberAdd(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	teamID, err := api.ResolveTeamID(ctx, client, args[0])
	if err != nil {
		return err
	}

	userID, err := api.ResolveUserID(ctx, client, args[1])
	if err != nil {
		return err
	}

	input := map[string]any{
		"teamId": teamID,
		"userId": userID,
	}
	vars := map[string]any{"input": input}

	var result teamMemberAddResult
	if err := client.Do(ctx, query.TeamMemberAddMutation, vars, &result); err != nil {
		return fmt.Errorf("add team member: %w", err)
	}
	if !result.TeamMembershipCreate.Success {
		return fmt.Errorf("add team member: mutation returned success=false")
	}
	if result.TeamMembershipCreate.TeamMembership == nil {
		return fmt.Errorf("add team member: no membership in response")
	}

	m := result.TeamMembershipCreate.TeamMembership
	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), m)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Added %s to team.\n", m.User.DisplayName)
	return err
}

func newTeamMemberRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <team-key> <user>",
		Short: "Remove a user from a team",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("team key and user are required")
			}
			return nil
		},
		RunE: runTeamMemberRemove,
	}
	cmd.Flags().Bool("yes", false, "skip confirmation prompt")
	return cmd
}

func runTeamMemberRemove(cmd *cobra.Command, args []string) error {
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Remove user %q from team %q? [y/N] ", args[1], args[0])
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

	client, err := newClientFromConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	teamID, err := api.ResolveTeamID(ctx, client, args[0])
	if err != nil {
		return err
	}

	userID, err := api.ResolveUserID(ctx, client, args[1])
	if err != nil {
		return err
	}

	// fetch memberships with pagination to find the membership ID for this user
	membershipID := ""
	var cursor *string
	for {
		vars := map[string]any{"teamId": teamID, "first": 250}
		if cursor != nil {
			vars["after"] = *cursor
		}
		var listResult teamMemberListResult
		if err := client.Do(ctx, query.TeamMemberListQuery, vars, &listResult); err != nil {
			return fmt.Errorf("list team members: %w", err)
		}
		if listResult.Team == nil {
			return fmt.Errorf("team not found: %s", args[0])
		}
		for _, m := range listResult.Team.Memberships.Nodes {
			if m.User.ID == userID {
				membershipID = m.ID
				break
			}
		}
		if membershipID != "" || !listResult.Team.Memberships.PageInfo.HasNextPage {
			break
		}
		end := listResult.Team.Memberships.PageInfo.EndCursor
		if end == "" {
			break
		}
		cursor = &end
	}
	if membershipID == "" {
		return fmt.Errorf("user %q is not a member of team %q", args[1], args[0])
	}

	deleteVars := map[string]any{"id": membershipID}
	var delResult teamMemberRemoveResult
	if err := client.Do(ctx, query.TeamMemberRemoveMutation, deleteVars, &delResult); err != nil {
		return fmt.Errorf("remove team member: %w", err)
	}
	if !delResult.TeamMembershipDelete.Success {
		return fmt.Errorf("remove team member: mutation returned success=false")
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed user from team.\n")
	return err
}

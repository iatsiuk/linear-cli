package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"linear-cli/internal/model"
	"linear-cli/internal/output"
	"linear-cli/internal/query"
)

type userListResult struct {
	Users struct {
		Nodes []model.User `json:"nodes"`
	} `json:"users"`
}

type userGetResult struct {
	User *model.User `json:"user"`
}

// UserRow is a display row for the user list table.
type UserRow struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Active string `json:"active"`
}

func userRole(u model.User) string {
	switch {
	case u.Admin:
		return "Admin"
	case u.Guest:
		return "Guest"
	default:
		return "Member"
	}
}

func newUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage Linear users",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newUserListCommand())
	cmd.AddCommand(newUserShowCommand())
	return cmd
}

func newUserListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		RunE:  runUserList,
	}
	cmd.Flags().Bool("include-guests", false, "include guest users")
	cmd.Flags().Bool("include-disabled", false, "include disabled users")
	return cmd
}

func runUserList(cmd *cobra.Command, _ []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	includeDisabled, _ := cmd.Flags().GetBool("include-disabled")
	includeGuests, _ := cmd.Flags().GetBool("include-guests")

	vars := map[string]any{}
	if includeDisabled {
		vars["includeDisabled"] = true
	}

	var result userListResult
	if err := client.Do(context.Background(), query.UserListQuery, vars, &result); err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	users := result.Users.Nodes
	if !includeGuests {
		filtered := users[:0]
		for _, u := range users {
			if !u.Guest {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	formatter := output.NewFormatter(jsonMode)

	if jsonMode {
		return formatter.Format(cmd.OutOrStdout(), users)
	}

	rows := make([]UserRow, len(users))
	for i, u := range users {
		rows[i] = UserRow{
			Name:   u.DisplayName,
			Email:  u.Email,
			Role:   userRole(u),
			Active: fmt.Sprintf("%v", u.Active),
		}
	}
	return formatter.Format(cmd.OutOrStdout(), rows)
}

func newUserShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show user details",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("user id is required")
			}
			return nil
		},
		RunE: runUserShow,
	}
}

func runUserShow(cmd *cobra.Command, args []string) error {
	client, err := newClientFromConfig()
	if err != nil {
		return err
	}

	vars := map[string]any{"id": args[0]}

	var result userGetResult
	if err := client.Do(context.Background(), query.UserGetQuery, vars, &result); err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if result.User == nil {
		return fmt.Errorf("user %q not found", args[0])
	}

	jsonMode, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonMode {
		return output.NewFormatter(true).Format(cmd.OutOrStdout(), result.User)
	}

	return printUserDetail(cmd, result.User)
}

func printUserDetail(cmd *cobra.Command, u *model.User) error {
	w := cmd.OutOrStdout()

	writeLine := func(label, value string) error {
		_, err := fmt.Fprintf(w, "%-14s %s\n", label+":", value)
		return err
	}

	fields := []struct{ label, value string }{
		{"Name", u.DisplayName},
		{"Email", u.Email},
		{"Role", userRole(*u)},
		{"Active", fmt.Sprintf("%v", u.Active)},
		{"Created", u.CreatedAt},
		{"Updated", u.UpdatedAt},
	}
	for _, f := range fields {
		if err := writeLine(f.label, f.value); err != nil {
			return err
		}
	}
	return nil
}

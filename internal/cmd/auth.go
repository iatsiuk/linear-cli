package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"linear-cli/internal/config"
)

func newAuthCommand() *cobra.Command {
	auth := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		RunE:  runAuth,
	}
	auth.AddCommand(newAuthStatusCommand())
	return auth
}

func runAuth(cmd *cobra.Command, _ []string) error {
	_, _ = fmt.Fprint(cmd.OutOrStdout(), "Enter your Linear API key: ")
	reader := bufio.NewReader(cmd.InOrStdin())
	key, err := reader.ReadString('\n')
	if err != nil && (!errors.Is(err, io.EOF) || key == "") {
		return fmt.Errorf("read api key: %w", err)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("api key cannot be empty")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.APIKey = key
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "API key saved.")
	return nil
}

func newAuthStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE:  runAuthStatus,
	}
}

func runAuthStatus(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if cfg.APIKey == "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "not configured")
		return nil
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "authenticated: %s\n", maskKey(cfg.APIKey))
	return nil
}

// maskKey shows the first 12 chars and replaces the rest with asterisks.
func maskKey(key string) string {
	const visible = 12
	if len(key) <= visible {
		return key
	}
	return key[:visible] + strings.Repeat("*", len(key)-visible)
}

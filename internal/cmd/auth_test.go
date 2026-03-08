package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func TestAuthCommand_SavesAPIKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("lin_api_testkey123\n"))
	root.SetArgs([]string{"auth"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if !strings.Contains(string(data), "lin_api_testkey123") {
		t.Errorf("config file should contain api key, got: %s", string(data))
	}
}

func TestAuthCommand_EmptyKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("\n"))
	root.SetArgs([]string{"auth"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for empty key, got nil")
	}
}

func TestAuthStatusCommand_ShowsMaskedKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("api_key: lin_api_secretkey\n"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"auth", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if strings.Contains(result, "secretkey") {
		t.Errorf("output should not contain full key, got: %s", result)
	}
	if !strings.Contains(result, "lin_api") {
		t.Errorf("output should contain key prefix, got: %s", result)
	}
}

func TestAuthStatusCommand_NotConfigured(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"auth", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "not configured") {
		t.Errorf("output should say 'not configured', got: %s", result)
	}
}

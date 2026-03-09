package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func TestCompletionCommand_Registered(t *testing.T) {
	t.Parallel()

	root := cmd.NewRootCommand("test")
	var found bool
	for _, c := range root.Commands() {
		if c.Use == "completion [bash|zsh|fish|powershell]" {
			found = true
			break
		}
	}
	if !found {
		t.Error("completion command not registered in root")
	}
}

func TestCompletionCommand_Bash(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"completion", "bash"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "bash") {
		t.Errorf("bash completion output should contain 'bash', got:\n%s", result)
	}
}

func TestCompletionCommand_Zsh(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"completion", "zsh"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if len(result) == 0 {
		t.Error("zsh completion output should not be empty")
	}
}

func TestCompletionCommand_Fish(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"completion", "fish"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if len(result) == 0 {
		t.Error("fish completion output should not be empty")
	}
}

func TestCompletionCommand_PowerShell(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"completion", "powershell"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if len(result) == 0 {
		t.Error("powershell completion output should not be empty")
	}
}

func TestCompletionCommand_NoArgs(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"completion"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no shell specified")
	}
}

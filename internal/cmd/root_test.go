package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func TestRootCommand_Help(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "linear") {
		t.Errorf("help output should contain 'linear', got: %s", out)
	}
}

func TestRootCommand_Version(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	root := cmd.NewRootCommand("1.2.3")
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--version"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "1.2.3") {
		t.Errorf("version output should contain '1.2.3', got: %s", out)
	}
}

func TestRootCommand_NoArgs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected no error running with no args, got %v", err)
	}
}

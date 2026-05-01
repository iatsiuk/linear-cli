package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func TestFormatExecError_JSON(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
	}{
		{name: "simple", err: errors.New("team \"NONEXISTENT\" not found")},
		{name: "newline", err: errors.New("line one\nline two")},
		{name: "quotes", err: errors.New(`he said "hi"`)},
		{name: "backslash", err: errors.New(`path\to\file`)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			formatExecError(tc.err, true, &buf)

			out := buf.Bytes()
			if len(out) == 0 || out[len(out)-1] != '\n' {
				t.Fatalf("expected trailing newline, got %q", out)
			}

			var parsed map[string]string
			if err := json.Unmarshal(out[:len(out)-1], &parsed); err != nil {
				t.Fatalf("output is not valid JSON: %v, raw=%q", err, out)
			}

			got, ok := parsed["error"]
			if !ok {
				t.Fatalf("missing \"error\" key, got %v", parsed)
			}
			if got != tc.err.Error() {
				t.Fatalf("error mismatch: want %q, got %q", tc.err.Error(), got)
			}
			if len(parsed) != 1 {
				t.Fatalf("expected single key, got %v", parsed)
			}
		})
	}
}

func TestFormatExecError_Plain(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatExecError(errors.New("boom"), false, &buf)

	got := buf.String()
	want := "boom\n"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestFormatExecError_NilError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatExecError(nil, true, &buf)
	if buf.Len() != 0 {
		t.Fatalf("expected no output for nil error, got %q", buf.String())
	}

	buf.Reset()
	formatExecError(nil, false, &buf)
	if buf.Len() != 0 {
		t.Fatalf("expected no output for nil error, got %q", buf.String())
	}
}

func TestHasJSONFlag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		args []string
		want bool
	}{
		{[]string{"--json", "issue", "list"}, true},
		{[]string{"issue", "--json", "list"}, true},
		{[]string{"--json=true"}, true},
		{[]string{"--json=1"}, true},
		{[]string{"--json=false"}, false},
		{[]string{"--json=0"}, false},
		{[]string{"--json=True"}, true},
		{[]string{"--json=False"}, false},
		{[]string{"--json=TRUE"}, true},
		{[]string{"--json", "--json=false"}, false},
		{[]string{"--json=false", "--json"}, true},
		{[]string{"--", "--json"}, false},
		{[]string{"issue", "list"}, false},
		{[]string{}, false},
	}

	for _, tc := range cases {
		got := hasJSONFlag(tc.args)
		if got != tc.want {
			t.Errorf("hasJSONFlag(%v) = %v, want %v", tc.args, got, tc.want)
		}
	}
}

// TestUnknownCommand_JSONError verifies that unknown-command errors are
// wrapped in a JSON envelope when --json appears before the bad arg.
func TestUnknownCommand_JSONError(t *testing.T) {
	t.Parallel()

	var errBuf bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--json", "nosuchcommand"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}

	// simulate what main() does
	var outBuf bytes.Buffer
	formatExecError(err, hasJSONFlag([]string{"--json", "nosuchcommand"}), &outBuf)

	out := outBuf.Bytes()
	if len(out) == 0 || out[len(out)-1] != '\n' {
		t.Fatalf("expected output with trailing newline, got %q", out)
	}
	var parsed map[string]string
	if err := json.Unmarshal(out[:len(out)-1], &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v, raw=%q", err, out)
	}
	if _, ok := parsed["error"]; !ok {
		t.Fatalf("missing \"error\" key, got %v", parsed)
	}
}

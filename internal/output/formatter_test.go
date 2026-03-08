package output_test

import (
	"bytes"
	"strings"
	"testing"

	"linear-cli/internal/output"
)

type row struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func TestTableFormatter_EmptyData(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	f := output.NewFormatter(false)
	if err := f.Format(&buf, []row{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for empty slice, got %q", buf.String())
	}
}

func TestTableFormatter_RendersHeaders(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	f := output.NewFormatter(false)
	data := []row{
		{Name: "foo", Value: "bar"},
	}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected header NAME in output, got:\n%s", out)
	}
	if !strings.Contains(out, "VALUE") {
		t.Errorf("expected header VALUE in output, got:\n%s", out)
	}
	if !strings.Contains(out, "foo") {
		t.Errorf("expected value foo in output, got:\n%s", out)
	}
	if !strings.Contains(out, "bar") {
		t.Errorf("expected value bar in output, got:\n%s", out)
	}
}

func TestTableFormatter_MultipleRows(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	f := output.NewFormatter(false)
	data := []row{
		{Name: "alpha", Value: "1"},
		{Name: "beta", Value: "2"},
	}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// header + 2 data rows
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (header + 2 rows), got %d:\n%s", len(lines), out)
	}
}

func TestTableFormatter_ColumnAlignment(t *testing.T) {
	t.Parallel()
	type wide struct {
		Short string `json:"short"`
		Long  string `json:"long"`
	}
	var buf bytes.Buffer
	f := output.NewFormatter(false)
	data := []wide{
		{Short: "a", Long: "very long value here"},
		{Short: "bb", Long: "x"},
	}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// all lines should have the same length (padded)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines")
	}
	firstLen := len(lines[0])
	for i, l := range lines[1:] {
		if len(l) != firstLen {
			t.Errorf("line %d length %d != header length %d", i+1, len(l), firstLen)
		}
	}
}

func TestJSONFormatter_ValidJSON(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	f := output.NewFormatter(true)
	data := []row{
		{Name: "hello", Value: "world"},
	}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"name"`) {
		t.Errorf("expected json key name, got:\n%s", out)
	}
	if !strings.Contains(out, `"hello"`) {
		t.Errorf("expected json value hello, got:\n%s", out)
	}
	// should be indented (pretty)
	if !strings.Contains(out, "\n") {
		t.Errorf("expected indented JSON output, got:\n%s", out)
	}
}

func TestJSONFormatter_EmptySlice(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	f := output.NewFormatter(true)
	if err := f.Format(&buf, []row{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "[]" {
		t.Errorf("expected [], got %q", out)
	}
}

func TestNewFormatter_TableByDefault(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	f := output.NewFormatter(false)
	data := []row{{Name: "x", Value: "y"}}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	// table formatter uppercases JSON tag names as headers
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected table output with NAME header, got: %s", out)
	}
}

func TestTableFormatter_PointerFields(t *testing.T) {
	t.Parallel()
	type withPtr struct {
		Name string  `json:"name"`
		Desc *string `json:"desc"`
	}
	s := "some description"
	var buf bytes.Buffer
	f := output.NewFormatter(false)
	data := []withPtr{
		{Name: "a", Desc: &s},
		{Name: "b", Desc: nil},
	}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "some description") {
		t.Errorf("expected dereferenced pointer value, got:\n%s", out)
	}
	if strings.Contains(out, "0x") {
		t.Errorf("expected no pointer addresses in output, got:\n%s", out)
	}
}

func TestNewFormatter_JSON(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	f := output.NewFormatter(true)
	data := []row{{Name: "x", Value: "y"}}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	// json formatter produces JSON with lowercase keys
	if !strings.Contains(out, `"name"`) {
		t.Errorf("expected JSON output with name key, got: %s", out)
	}
}

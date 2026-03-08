package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Formatter writes data to a writer in a specific format.
type Formatter interface {
	Format(w io.Writer, data any) error
}

// NewFormatter returns a JSON formatter if json is true, otherwise a table formatter.
func NewFormatter(jsonMode bool) Formatter {
	if jsonMode {
		return &JSONFormatter{}
	}
	return &TableFormatter{}
}

// JSONFormatter renders data as indented JSON.
type JSONFormatter struct{}

func (f *JSONFormatter) Format(w io.Writer, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	b = append(b, '\n')
	_, err = w.Write(b)
	return err
}

// TableFormatter renders a slice of structs as an aligned text table.
// Column headers are derived from JSON tags (uppercased).
type TableFormatter struct{}

func (f *TableFormatter) Format(w io.Writer, data any) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("table formatter requires a slice, got %T", data)
	}
	if v.Len() == 0 {
		return nil
	}

	// resolve element type (handle pointer elements)
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("table formatter requires a slice of structs, got %T", data)
	}

	// collect field paths and header names from JSON tags,
	// expanding anonymous (embedded) struct fields recursively
	type col struct {
		index []int // field path for FieldByIndex
		name  string
	}
	var collectFields func(t reflect.Type, prefix []int) []col
	collectFields = func(t reflect.Type, prefix []int) []col {
		var result []col
		for i := range t.NumField() {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			path := append(append([]int{}, prefix...), i)
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				result = append(result, collectFields(field.Type, path)...)
				continue
			}
			tag := field.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}
			name := strings.Split(tag, ",")[0]
			if name == "-" {
				continue
			}
			result = append(result, col{index: path, name: strings.ToUpper(name)})
		}
		return result
	}
	cols := collectFields(elemType, nil)
	if len(cols) == 0 {
		return nil
	}

	// build rows as strings
	rows := make([][]string, v.Len())
	for i := range v.Len() {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				rows[i] = make([]string, len(cols))
				continue
			}
			elem = elem.Elem()
		}
		rows[i] = make([]string, len(cols))
		for j, c := range cols {
			field := elem.FieldByIndex(c.index)
			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					rows[i][j] = ""
				} else {
					rows[i][j] = fmt.Sprintf("%v", field.Elem().Interface())
				}
			} else {
				rows[i][j] = fmt.Sprintf("%v", field.Interface())
			}
		}
	}

	// calculate column widths
	widths := make([]int, len(cols))
	for j, c := range cols {
		widths[j] = len(c.name)
	}
	for _, row := range rows {
		for j, cell := range row {
			if len(cell) > widths[j] {
				widths[j] = len(cell)
			}
		}
	}

	// build format string: each column padded to width, separated by two spaces
	printRow := func(cells []string) error {
		parts := make([]string, len(cols))
		for j, cell := range cells {
			parts[j] = fmt.Sprintf("%-*s", widths[j], cell)
		}
		_, err := fmt.Fprintln(w, strings.Join(parts, "  "))
		return err
	}

	// header
	headers := make([]string, len(cols))
	for j, c := range cols {
		headers[j] = c.name
	}
	if err := printRow(headers); err != nil {
		return err
	}

	// data rows
	for _, row := range rows {
		if err := printRow(row); err != nil {
			return err
		}
	}

	return nil
}

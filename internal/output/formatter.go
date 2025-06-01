package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table" // For table formatting
	"gopkg.in/yaml.v3"                     // For YAML formatting
)

// Formatter defines an interface for writing structured data in different formats.
type Formatter interface {
	Write(data interface{}) error
}

// NewFormatter returns a formatter based on the specified format string.
// writer is where the output will be written (e.g., os.Stdout).
// query is currently a placeholder for future JMESPath/query support.
func NewFormatter(format string, query string, writer io.Writer) (Formatter, error) {
	// TODO: Implement query wrapping if query string is provided and format is json/yaml

	switch strings.ToLower(format) {
	case "json":
		return &JSONFormatter{Writer: writer, PrettyPrint: true}, nil
	case "yaml":
		return &YAMLFormatter{Writer: writer}, nil
	case "text":
		return &TextFormatter{Writer: writer}, nil
	case "table":
		return &TableFormatter{Writer: writer}, nil
	default:
		// Fallback to text or return error
		// fmt.Fprintf(os.Stderr, "Warning: unknown output format '%s', defaulting to text.\n", format)
		return &TextFormatter{Writer: writer}, nil
		// return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

// JSONFormatter writes data as JSON.
type JSONFormatter struct {
	Writer      io.Writer
	PrettyPrint bool
}

func (f *JSONFormatter) Write(data interface{}) error {
	encoder := json.NewEncoder(f.Writer)
	if f.PrettyPrint {
		encoder.SetIndent("", "  ") // Standard 2-space indent
	}
	return encoder.Encode(data)
}

// YAMLFormatter writes data as YAML.
type YAMLFormatter struct {
	Writer io.Writer
}

func (f *YAMLFormatter) Write(data interface{}) error {
	encoder := yaml.NewEncoder(f.Writer)
	// go-yaml.v3 has options for indentation if needed, e.g. encoder.SetIndent(2)
	return encoder.Encode(data)
}

// TextFormatter writes data in a simple text format.
// This is a basic implementation; specific commands might provide their own templates.
type TextFormatter struct {
	Writer io.Writer
}

func (f *TextFormatter) Write(data interface{}) error {
	// Basic text output: just print using fmt.Fprintln
	// More sophisticated text formatting would involve type assertions and custom rendering.
	// For lists/slices, iterate and print each item.
	// For maps/structs, print key-value pairs.

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			_, err := fmt.Fprintln(f.Writer, val.Index(i).Interface())
			if err != nil {
				return err
			}
		}
		return nil
	}
	_, err := fmt.Fprintln(f.Writer, data)
	return err
}

// TableFormatter writes data as a table.
type TableFormatter struct {
	Writer io.Writer
}

func (f *TableFormatter) Write(data interface{}) error {
	t := table.NewWriter()
	t.SetOutputMirror(f.Writer)
	// t.SetStyle(table.StyleLight) // Example style

	val := reflect.ValueOf(data)
	kind := val.Kind()

	if kind == reflect.Slice {
		if val.Len() == 0 {
			// fmt.Fprintln(f.Writer, "(no results)") // Or let table print empty
			t.Render() // Render empty table (just headers if any, or nothing)
			return nil
		}

		// Use the first element to determine headers for structs/maps
		firstElem := val.Index(0)
		if firstElem.Kind() == reflect.Struct {
			var headers []interface{}
			typ := firstElem.Type()
			for i := 0; i < typ.NumField(); i++ {
				// Use field name as header. Could use struct tags for custom names.
				headers = append(headers, typ.Field(i).Name)
			}
			t.AppendHeader(headers)

			for i := 0; i < val.Len(); i++ {
				var row []interface{}
				elem := val.Index(i)
				for j := 0; j < elem.NumField(); j++ {
					row = append(row, elem.Field(j).Interface())
				}
				t.AppendRow(row)
			}
		} else if firstElem.Kind() == reflect.Map && firstElem.Type().Key().Kind() == reflect.String {
			// Assuming map[string]interface{} or similar
			// Headers from keys of the first map (order might not be guaranteed)
			var headers []interface{}
			var headerKeys []string // To keep order for rows
			iter := firstElem.MapRange()
			for iter.Next() {
				headers = append(headers, iter.Key().String())
				headerKeys = append(headerKeys, iter.Key().String())
			}
			t.AppendHeader(headers)

			for i := 0; i < val.Len(); i++ {
				var row []interface{}
				elem := val.Index(i)
				for _, key := range headerKeys {
					row = append(row, elem.MapIndex(reflect.ValueOf(key)).Interface())
				}
				t.AppendRow(row)
			}
		} else {
			// Slice of simple types, just print as a single column
			t.AppendHeader(table.Row{"Value"})
			for i := 0; i < val.Len(); i++ {
				t.AppendRow(table.Row{val.Index(i).Interface()})
			}
		}
	} else if kind == reflect.Struct {
        var headers []interface{}
        var row []interface{}
        typ := val.Type()
        for i := 0; i < typ.NumField(); i++ {
            headers = append(headers, typ.Field(i).Name)
            row = append(row, val.Field(i).Interface())
        }
        t.AppendHeader(headers)
        t.AppendRow(row)

    } else if kind == reflect.Map && val.Type().Key().Kind() == reflect.String {
        var headers []interface{}
        var row []interface{}
        iter := val.MapRange()
        for iter.Next() {
            headers = append(headers, iter.Key().String())
            row = append(row, iter.Value().Interface())
        }
        t.AppendHeader(headers)
        t.AppendRow(row)
    } else {
		// Not a slice, map, or struct; fallback to simple text print for single items
		// Or, could make TableFormatter return an error for unsupported types.
		_, err := fmt.Fprintln(f.Writer, data)
		return err
	}

	t.Render()
	return nil
}

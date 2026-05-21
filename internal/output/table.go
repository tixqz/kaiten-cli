package output

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
)

// formatTable writes data as a formatted table using text/tabwriter.
// It expects a slice of structs/maps or a single struct/map.
func formatTable(w io.Writer, data any) error {
	rows, err := toRows(data)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	// Determine column order: if the first row is a map, use sorted keys.
	columns := columnOrder(rows[0])

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Header.
	for i, col := range columns {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, strings.ToUpper(col))
	}
	fmt.Fprintln(tw)

	// Separator.
	for i, col := range columns {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, strings.Repeat("-", len(col)))
	}
	fmt.Fprintln(tw)

	// Rows.
	for _, row := range rows {
		for i, col := range columns {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			val := row[col]
			fmt.Fprint(tw, fmt.Sprintf("%v", val))
		}
		fmt.Fprintln(tw)
	}

	return tw.Flush()
}

// toRows converts data into a slice of map[string]any for tabular output.
// Supports: []struct, []map, single struct, single map.
func toRows(data any) ([]map[string]any, error) {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil, nil
	}
	v = indirect(v)

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, nil
		}
		rows := make([]map[string]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			row, err := valueToMap(v.Index(i))
			if err != nil {
				return nil, err
			}
			rows[i] = row
		}
		return rows, nil

	case reflect.Map:
		row := make(map[string]any, v.Len())
		for _, k := range v.MapKeys() {
			row[fmt.Sprintf("%v", k.Interface())] = v.MapIndex(k).Interface()
		}
		return []map[string]any{row}, nil

	case reflect.Struct:
		row, err := valueToMap(v)
		if err != nil {
			return nil, err
		}
		return []map[string]any{row}, nil

	default:
		return nil, fmt.Errorf("tabular output requires a struct, map, or slice of structs/maps, got %T", data)
	}
}

// valueToMap converts a struct or map value to map[string]any.
func valueToMap(v reflect.Value) (map[string]any, error) {
	v = indirect(v)
	if !v.IsValid() {
		return map[string]any{}, nil
	}

	switch v.Kind() {
	case reflect.Map:
		m := make(map[string]any, v.Len())
		for _, k := range v.MapKeys() {
			m[fmt.Sprintf("%v", k.Interface())] = v.MapIndex(k).Interface()
		}
		return m, nil

	case reflect.Struct:
		t := v.Type()
		m := make(map[string]any, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)
			if !ft.IsExported() {
				continue
			}
			name := jsonTagName(ft)
			if name == "-" {
				continue
			}
			m[name] = v.Field(i).Interface()
		}
		return m, nil

	default:
		return nil, fmt.Errorf("cannot convert %T to map", v.Interface())
	}
}

// columnOrder returns a sorted slice of column names from a row map.
func columnOrder(row map[string]any) []string {
	cols := make([]string, 0, len(row))
	for k := range row {
		cols = append(cols, k)
	}
	sort.Strings(cols)
	return cols
}

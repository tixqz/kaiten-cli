// Package output provides reusable formatters for CLI output.
// Supported formats: json, jsonl, yaml, table, csv.
package output

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Options controls output formatting behavior.
type Options struct {
	// Fields restricts output to only these field names.
	// Names are matched against JSON tag names, falling back to
	// lowercased field names.
	Fields []string

	// NoDescriptions removes "description" fields unless explicitly
	// listed in Fields.
	NoDescriptions bool

	// Quiet prints only object IDs (one per line for slices).
	Quiet bool
}

// Format writes data to w using the specified format.
// Supported formats: json, jsonl, yaml, table, csv.
// If format is empty, defaults to "json".
func Format(w io.Writer, data any, format string, opts Options) error {
	// Quiet mode short-circuits all formatters.
	if opts.Quiet {
		return writeQuiet(w, data)
	}

	format = strings.TrimSpace(strings.ToLower(format))
	if format == "" {
		format = "json"
	}

	// Apply field filtering (Fields + NoDescriptions).
	filtered, err := filterData(data, opts)
	if err != nil {
		return err
	}

	switch format {
	case "json":
		return formatJSON(w, filtered)
	case "jsonl":
		return formatJSONL(w, filtered)
	case "yaml":
		return formatYAML(w, filtered)
	case "table":
		return formatTable(w, filtered)
	case "csv":
		return formatCSV(w, filtered)
	default:
		return fmt.Errorf("unsupported output format: %q", format)
	}
}

// writeQuiet extracts IDs from data and writes one per line.
func writeQuiet(w io.Writer, data any) error {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil
	}

	v = indirect(v)

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := writeID(w, v.Index(i)); err != nil {
				return err
			}
		}
		return nil
	default:
		return writeID(w, v)
	}
}

func writeID(w io.Writer, v reflect.Value) error {
	v = indirect(v)

	var id string
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)
			if !ft.IsExported() {
				continue
			}
			tag := ft.Tag.Get("json")
			tagName := strings.Split(tag, ",")[0]
			if tagName == "id" || strings.EqualFold(ft.Name, "id") {
				id = fmt.Sprintf("%v", v.Field(i).Interface())
				break
			}
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			ks := fmt.Sprintf("%v", k.Interface())
			if strings.EqualFold(ks, "id") {
				id = fmt.Sprintf("%v", v.MapIndex(k).Interface())
				break
			}
		}
	}

	if id != "" {
		_, err := fmt.Fprintln(w, id)
		return err
	}
	return nil
}

// indirect dereferences pointer values.
func indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

// filterData applies Fields and NoDescriptions to data.
// If no filtering is required, returns the original data unchanged.
func filterData(data any, opts Options) (any, error) {
	if len(opts.Fields) == 0 && !opts.NoDescriptions {
		return data, nil
	}

	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return data, nil
	}
	v = indirect(v)

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			item, err := filterSingle(v.Index(i), opts.Fields, opts.NoDescriptions)
			if err != nil {
				return nil, err
			}
			result[i] = item
		}
		return result, nil
	case reflect.Struct, reflect.Map:
		return filterSingle(v, opts.Fields, opts.NoDescriptions)
	default:
		return data, nil
	}
}

func filterSingle(v reflect.Value, fields []string, noDesc bool) (any, error) {
	v = indirect(v)
	if !v.IsValid() {
		return nil, nil
	}

	switch v.Kind() {
	case reflect.Struct:
		return filterStruct(v, fields, noDesc)
	case reflect.Map:
		return filterMap(v, fields, noDesc)
	default:
		return v.Interface(), nil
	}
}

func filterStruct(v reflect.Value, fields []string, noDesc bool) (map[string]any, error) {
	t := v.Type()
	result := make(map[string]any, t.NumField())
	fieldSet := makeLowerSet(fields)
	hasFields := len(fields) > 0

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if !ft.IsExported() {
			continue
		}

		tagName := jsonTagName(ft)
		if tagName == "" || tagName == "-" {
			continue
		}

		// Field whitelist.
		if hasFields && !fieldSet[tagName] {
			continue
		}

		// NoDescriptions: skip "description" unless explicitly listed in Fields.
		if noDesc && strings.EqualFold(tagName, "description") {
			if !hasFields || !fieldSet[tagName] {
				continue
			}
		}

		result[tagName] = v.Field(i).Interface()
	}

	return result, nil
}

func filterMap(v reflect.Value, fields []string, noDesc bool) (map[string]any, error) {
	result := make(map[string]any, v.Len())
	fieldSet := makeLowerSet(fields)
	hasFields := len(fields) > 0

	for _, k := range v.MapKeys() {
		ks := fmt.Sprintf("%v", k.Interface())

		if hasFields && !fieldSet[ks] {
			continue
		}

		if noDesc && strings.EqualFold(ks, "description") {
			if !hasFields || !fieldSet[ks] {
				continue
			}
		}

		result[ks] = v.MapIndex(k).Interface()
	}

	return result, nil
}

// jsonTagName returns the JSON field name from a struct field's json tag.
// Falls back to strings.ToLower of the field name.
func jsonTagName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return strings.ToLower(f.Name)
	}
	name := strings.Split(tag, ",")[0]
	if name == "" {
		return strings.ToLower(f.Name)
	}
	return name
}

func makeLowerSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[strings.ToLower(item)] = true
	}
	return s
}

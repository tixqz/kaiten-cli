package output

import (
	"encoding/json"
	"io"
	"reflect"
)

// formatJSONL writes data as JSON Lines.
// For slices/arrays, each element is written as a separate JSON line.
// For a single object, it is written as one JSON line.
func formatJSONL(w io.Writer, data any) error {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil
	}

	v = indirect(v)

	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		enc := json.NewEncoder(w)
		for i := 0; i < v.Len(); i++ {
			if err := enc.Encode(v.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	}

	enc := json.NewEncoder(w)
	return enc.Encode(data)
}

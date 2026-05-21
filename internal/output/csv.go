package output

import (
	"encoding/csv"
	"fmt"
	"io"
)

// formatCSV writes data as CSV with a header row.
// It expects a slice of structs/maps or a single struct/map.
func formatCSV(w io.Writer, data any) error {
	rows, err := toRows(data)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	columns := columnOrder(rows[0])
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header.
	if err := writer.Write(columns); err != nil {
		return err
	}

	// Data rows.
	for _, row := range rows {
		record := make([]string, len(columns))
		for i, col := range columns {
			val := row[col]
			record[i] = fmt.Sprintf("%v", val)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Error()
}

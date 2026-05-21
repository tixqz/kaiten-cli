package output

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// --- Test data types ---

type testItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	BoardID     int    `json:"board_id"`
}

type testNested struct {
	ID    int      `json:"id"`
	Label string   `json:"label"`
	Item  testItem `json:"item"`
}

// --- JSON tests ---

func TestFormatJSON(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 1, Title: "hello", Description: "desc", BoardID: 5}
	err := Format(&buf, data, "json", Options{})
	if err != nil {
		t.Fatalf("Format json: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 1`) {
		t.Errorf("json output missing id: %s", out)
	}
	if !strings.Contains(out, `"title": "hello"`) {
		t.Errorf("json output missing title: %s", out)
	}
}

func TestFormatJSONSlice(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 1, Title: "a", BoardID: 10},
		{ID: 2, Title: "b", BoardID: 20},
	}
	err := Format(&buf, data, "json", Options{})
	if err != nil {
		t.Fatalf("Format json slice: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 1`) || !strings.Contains(out, `"id": 2`) {
		t.Errorf("json slice missing items: %s", out)
	}
}

// --- JSONL tests ---

func TestFormatJSONLSlice(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 1, Title: "a", BoardID: 10},
		{ID: 2, Title: "b", BoardID: 20},
	}
	err := Format(&buf, data, "jsonl", Options{})
	if err != nil {
		t.Fatalf("Format jsonl: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], `"id":1`) {
		t.Errorf("line 1 missing id: %s", lines[0])
	}
	if !strings.Contains(lines[1], `"id":2`) {
		t.Errorf("line 2 missing id: %s", lines[1])
	}
}

func TestFormatJSONLSingle(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 42, Title: "single", BoardID: 7}
	err := Format(&buf, data, "jsonl", Options{})
	if err != nil {
		t.Fatalf("Format jsonl single: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], `"id":42`) {
		t.Errorf("missing id 42: %s", lines[0])
	}
}

// --- YAML tests ---

func TestFormatYAML(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 10, Title: "yaml-test", Description: "yay", BoardID: 3}
	err := Format(&buf, data, "yaml", Options{})
	if err != nil {
		t.Fatalf("Format yaml: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "id: 10") {
		t.Errorf("yaml output missing id: %s", out)
	}
	if !strings.Contains(out, "title: yaml-test") {
		t.Errorf("yaml output missing title: %s", out)
	}
}

func TestFormatYAMLSlice(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 1, Title: "a"},
		{ID: 2, Title: "b"},
	}
	err := Format(&buf, data, "yaml", Options{})
	if err != nil {
		t.Fatalf("Format yaml slice: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "id: 1") || !strings.Contains(out, "id: 2") {
		t.Errorf("yaml slice missing items: %s", out)
	}
}

// --- Table tests ---

func TestFormatTableSlice(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 1, Title: "foo", BoardID: 100},
		{ID: 2, Title: "bar", BoardID: 200},
	}
	err := Format(&buf, data, "table", Options{})
	if err != nil {
		t.Fatalf("Format table: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "TITLE") {
		t.Errorf("table missing headers: %s", out)
	}
	if !strings.Contains(out, "foo") || !strings.Contains(out, "bar") {
		t.Errorf("table missing data: %s", out)
	}
}

func TestFormatTableSingleStruct(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 99, Title: "single-row", BoardID: 1}
	err := Format(&buf, data, "table", Options{})
	if err != nil {
		t.Fatalf("Format table single: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "99") {
		t.Errorf("table single missing value: %s", out)
	}
}

func TestFormatTableMap(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"id": 7, "title": "map-item", "board_id": 3}
	err := Format(&buf, data, "table", Options{})
	if err != nil {
		t.Fatalf("Format table map: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "7") || !strings.Contains(out, "map-item") {
		t.Errorf("table map missing data: %s", out)
	}
}

func TestFormatTableUnsupportedType(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, "just a string", "table", Options{})
	if err == nil {
		t.Fatal("expected error for unsupported table data type")
	}
	if !strings.Contains(err.Error(), "tabular") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- CSV tests ---

func TestFormatCSV(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 1, Title: "csv1", BoardID: 10},
		{ID: 2, Title: "csv2", BoardID: 20},
	}
	err := Format(&buf, data, "csv", Options{})
	if err != nil {
		t.Fatalf("Format csv: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (header+2 rows), got %d", len(lines))
	}
	// header
	if !strings.Contains(lines[0], "id") || !strings.Contains(lines[0], "title") {
		t.Errorf("csv header missing columns: %s", lines[0])
	}
	// data
	if !strings.Contains(lines[1], "csv1") {
		t.Errorf("csv missing row data: %s", lines[1])
	}
}

func TestFormatCSVUnsupportedType(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, 42, "csv", Options{})
	if err == nil {
		t.Fatal("expected error for unsupported csv data type")
	}
}

func TestFormatCSVEmptySlice(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, []testItem{}, "csv", Options{})
	if err != nil {
		t.Fatalf("Format csv empty slice: %v", err)
	}
	if buf.Len() > 0 {
		t.Errorf("expected empty output for empty slice, got: %s", buf.String())
	}
}

// --- Fields filtering ---

func TestFieldsFilterStruct(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 1, Title: "hello", Description: "desc", BoardID: 5}
	err := Format(&buf, data, "json", Options{Fields: []string{"id", "title"}})
	if err != nil {
		t.Fatalf("Format with fields: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id"`) || !strings.Contains(out, `"title"`) {
		t.Errorf("fields should include id and title: %s", out)
	}
	if strings.Contains(out, `"board_id"`) {
		t.Errorf("fields should exclude board_id: %s", out)
	}
}

func TestFieldsFilterSlice(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 1, Title: "a", BoardID: 10},
		{ID: 2, Title: "b", BoardID: 20},
	}
	err := Format(&buf, data, "json", Options{Fields: []string{"id", "title"}})
	if err != nil {
		t.Fatalf("Format slice with fields: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, `"board_id"`) {
		t.Errorf("fields should exclude board_id: %s", out)
	}
}

func TestFieldsFilterTable(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 1, Title: "a", BoardID: 10},
		{ID: 2, Title: "b", BoardID: 20},
	}
	err := Format(&buf, data, "table", Options{Fields: []string{"id", "title"}})
	if err != nil {
		t.Fatalf("Format table with fields: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "BOARD_ID") {
		t.Errorf("table should exclude board_id: %s", out)
	}
}

func TestFieldsFilterYAML(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 42, Title: "test", Description: "desc", BoardID: 7}
	err := Format(&buf, data, "yaml", Options{Fields: []string{"id", "title"}})
	if err != nil {
		t.Fatalf("Format yaml with fields: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "board_id") {
		t.Errorf("yaml should exclude board_id: %s", out)
	}
}

func TestFieldsFilterCSV(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 1, Title: "a", BoardID: 10},
	}
	err := Format(&buf, data, "csv", Options{Fields: []string{"id"}})
	if err != nil {
		t.Fatalf("Format csv with fields: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
	// Only "id" column should be present
	if lines[0] != "id" {
		t.Errorf("csv header should be just 'id', got: %s", lines[0])
	}
}

// --- NoDescriptions ---

func TestNoDescriptionsStruct(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 1, Title: "x", Description: "should-be-hidden", BoardID: 5}
	err := Format(&buf, data, "json", Options{NoDescriptions: true})
	if err != nil {
		t.Fatalf("Format NoDescriptions: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "description") {
		t.Errorf("NoDescriptions should remove description: %s", out)
	}
}

func TestNoDescriptionsTable(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 1, Title: "x", Description: "hide", BoardID: 5},
	}
	err := Format(&buf, data, "table", Options{NoDescriptions: true})
	if err != nil {
		t.Fatalf("Format table NoDescriptions: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "DESCRIPTION") {
		t.Errorf("table should not have DESCRIPTION column: %s", out)
	}
	if !strings.Contains(out, "TITLE") {
		t.Errorf("table should still have TITLE: %s", out)
	}
}

func TestNoDescriptionsWithExplicitField(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 1, Title: "x", Description: "keep-me", BoardID: 5}
	err := Format(&buf, data, "json", Options{
		NoDescriptions: true,
		Fields:         []string{"id", "description"},
	})
	if err != nil {
		t.Fatalf("Format NoDescriptions+explicit field: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "keep-me") {
		t.Errorf("description should be present when explicitly requested: %s", out)
	}
	if strings.Contains(out, "board_id") {
		t.Errorf("board_id should be absent: %s", out)
	}
}

// --- Quiet mode ---

func TestQuietSlice(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 101, Title: "a"},
		{ID: 202, Title: "b"},
		{ID: 303, Title: "c"},
	}
	err := Format(&buf, data, "json", Options{Quiet: true})
	if err != nil {
		t.Fatalf("Format quiet slice: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 IDs, got %d: %s", len(lines), buf.String())
	}
	if lines[0] != "101" || lines[1] != "202" || lines[2] != "303" {
		t.Errorf("unexpected IDs: %s", buf.String())
	}
}

func TestQuietSingle(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 42, Title: "lonely"}
	err := Format(&buf, data, "yaml", Options{Quiet: true})
	if err != nil {
		t.Fatalf("Format quiet single: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "42" {
		t.Errorf("expected '42', got %q", out)
	}
}

func TestQuietMap(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"id": 77, "title": "map-item"}
	err := Format(&buf, data, "jsonl", Options{Quiet: true})
	if err != nil {
		t.Fatalf("Format quiet map: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "77" {
		t.Errorf("expected '77', got %q", out)
	}
}

func TestQuietEmptySlice(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, []testItem{}, "csv", Options{Quiet: true})
	if err != nil {
		t.Fatalf("Format quiet empty: %v", err)
	}
	if buf.Len() > 0 {
		t.Errorf("expected empty output: %s", buf.String())
	}
}

// --- Default format ---

func TestFormatDefaultsToJSON(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 1, Title: "default"}
	err := Format(&buf, data, "", Options{})
	if err != nil {
		t.Fatalf("Format empty format: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id"`) {
		t.Errorf("expected json output: %s", out)
	}
}

// --- Unsupported format ---

func TestUnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, "data", "xml", Options{})
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- JSONL whitespace / trailing newlines ---

func TestJSONLNoTrailingEmptyLine(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{{ID: 1, Title: "a"}}
	err := Format(&buf, data, "jsonl", Options{})
	if err != nil {
		t.Fatalf("Format jsonl: %v", err)
	}
	out := buf.String()
	// Should end with newline (Encode adds one) but only have 1 line
	lines := strings.Split(out, "\n")
	// Encode adds \n, so strings.Split gives ["{...}", ""]
	if len(lines) != 2 || lines[1] != "" {
		t.Errorf("unexpected line count: %q", out)
	}
}

// --- Positional: table column order ---

func TestTableColumnOrderSortsKeys(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 5, Title: "z", BoardID: 1, Description: "d"},
	}
	err := Format(&buf, data, "table", Options{})
	if err != nil {
		t.Fatalf("Format table columns: %v", err)
	}
	out := buf.String()
	// Columns sorted alphabetically: board_id, description, id, title
	firstLine := strings.Split(out, "\n")[0]
	if !strings.HasPrefix(firstLine, "BOARD_ID") {
		t.Errorf("expected BOARD_ID first, got: %s", firstLine)
	}
}

// --- Empty data ---

func TestFormatNilData(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, nil, "json", Options{})
	if err != nil {
		t.Fatalf("Format nil: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "null" {
		t.Errorf("expected 'null', got %q", out)
	}
}

func TestFormatJSONLEmptySlice(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, []testItem{}, "jsonl", Options{})
	if err != nil {
		t.Fatalf("Format jsonl empty: %v", err)
	}
	if buf.Len() > 0 {
		t.Errorf("expected empty output: %s", buf.String())
	}
}

// --- Pointer / nil pointer types ---

type testStruct struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestFormatStructPointer(t *testing.T) {
	var buf bytes.Buffer
	data := &testStruct{ID: 7, Name: "ptr"}
	err := Format(&buf, data, "json", Options{})
	if err != nil {
		t.Fatalf("Format ptr: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 7`) {
		t.Errorf("pointer struct missing id: %s", out)
	}
}

func TestQuietStructPointer(t *testing.T) {
	var buf bytes.Buffer
	data := &testStruct{ID: 99, Name: "ptr"}
	err := Format(&buf, data, "table", Options{Quiet: true})
	if err != nil {
		t.Fatalf("Quiet ptr: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "99" {
		t.Errorf("expected '99', got %q", out)
	}
}

// --- Edge: single-element slice with quiet ---

func TestQuietSingleElementSlice(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{{ID: 7, Title: "x"}}
	err := Format(&buf, data, "json", Options{Quiet: true})
	if err != nil {
		t.Fatalf("Format quiet single element slice: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "7" {
		t.Errorf("expected '7', got %q", out)
	}
}

// --- Field matching: non-JSON-tagged fields use lowercased name ---

type noTagStruct struct {
	ID        int
	Title     string
	SecretKey string
}

func TestFieldsNoJSONTags(t *testing.T) {
	var buf bytes.Buffer
	data := noTagStruct{ID: 1, Title: "x", SecretKey: "hidden"}
	err := Format(&buf, data, "json", Options{Fields: []string{"id", "title"}})
	if err != nil {
		t.Fatalf("Format no-tag fields: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "secretkey") && strings.Contains(out, "secret_key") && strings.Contains(out, "SecretKey") {
		t.Errorf("secretkey should be filtered out: %s", out)
	}
	if !strings.Contains(out, "title") {
		t.Errorf("title should be present: %s", out)
	}
}

// --- Combination: Fields + NoDescriptions ---

func TestFieldsAndNoDescriptions(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 1, Title: "x", Description: "desc", BoardID: 5}
	err := Format(&buf, data, "json", Options{
		Fields:         []string{"id", "title", "board_id"},
		NoDescriptions: true,
	})
	if err != nil {
		t.Fatalf("Fields+NoDescriptions: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "description") {
		t.Errorf("description should be hidden: %s", out)
	}
	if !strings.Contains(out, "board_id") {
		t.Errorf("board_id should be present: %s", out)
	}
}

// --- Table with empty Fields should show all columns ---

func TestTableEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{{ID: 1, Title: "x", BoardID: 5, Description: "d"}}
	err := Format(&buf, data, "table", Options{Fields: []string{}})
	if err != nil {
		t.Fatalf("table empty fields: %v", err)
	}
	out := buf.String()
	// Should have all columns when fields is empty but non-nil
	if !strings.Contains(out, "DESCRIPTION") {
		t.Errorf("should show all columns: %s", out)
	}
}

// Verify all outputs are valid via round-trip parsing for json/jsonl
func TestJSONRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 10, Title: "rt1", BoardID: 1},
		{ID: 20, Title: "rt2", BoardID: 2},
	}
	err := Format(&buf, data, "json", Options{})
	if err != nil {
		t.Fatalf("json roundtrip: %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "[") {
		t.Errorf("expected json array: %s", out)
	}
	if !strings.Contains(out, "10") || !strings.Contains(out, "20") {
		t.Errorf("missing data: %s", out)
	}
}

// --- Table/nil description field (omitempty field missing) ---

func TestTableFieldOmitempty(t *testing.T) {
	var buf bytes.Buffer
	data := []testItem{
		{ID: 1, Title: "no-desc"},
	}
	err := Format(&buf, data, "table", Options{})
	if err != nil {
		t.Fatalf("table omitempty: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1") {
		t.Errorf("missing id: %s", out)
	}
}

// Test that JSON output uses indentation (matching current outputJSON behavior)
func TestJSONIndentation(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 1, Title: "indent-test", BoardID: 3}
	err := Format(&buf, data, "json", Options{})
	if err != nil {
		t.Fatalf("json indent: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "  ") {
		t.Errorf("expected indented json: %s", out)
	}
}

// --- Nested structure with quiet ---
func TestQuietNestedStruct(t *testing.T) {
	var buf bytes.Buffer
	data := testNested{ID: 5, Label: "nested", Item: testItem{ID: 99, Title: "inner"}}
	err := Format(&buf, data, "json", Options{Quiet: true})
	if err != nil {
		t.Fatalf("quiet nested: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	// Should find the top-level ID
	if out != "5" {
		t.Errorf("expected '5', got %q", out)
	}
}

// --- CSV with special characters ---
func TestCSVSpecialChars(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{
		{"id": 1, "title": "contains, comma"},
		{"id": 2, "title": `has "quotes"`},
	}
	err := Format(&buf, data, "csv", Options{})
	if err != nil {
		t.Fatalf("csv special chars: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"contains, comma"`) {
		t.Errorf("expected quoted comma value: %s", out)
	}
	if !strings.Contains(out, `"has ""quotes"""`) {
		t.Errorf("expected escaped quotes: %s", out)
	}
}

// --- CSV with map that has no "id" field uses first row column order ---
func TestCSVMapData(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{
		{"name": "alice", "age": 30},
		{"name": "bob", "age": 25},
	}
	err := Format(&buf, data, "csv", Options{})
	if err != nil {
		t.Fatalf("csv map: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	// First line is header
	if lines[0] != "age,name" && lines[0] != "name,age" {
		t.Errorf("unexpected header: %s", lines[0])
	}
}

// --- Quiet with map having ID (uppercase) ---
func TestQuietMapUppercaseID(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"ID": 555, "value": "test"}
	err := Format(&buf, data, "json", Options{Quiet: true})
	if err != nil {
		t.Fatalf("quiet map uppercase: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "555" {
		t.Errorf("expected '555', got %q", out)
	}
}

// --- Table with map input ---
func TestTableMapInput(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{
		{"id": 1, "val": "a"},
		{"id": 2, "val": "b"},
	}
	err := Format(&buf, data, "table", Options{})
	if err != nil {
		t.Fatalf("table map input: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "VAL") {
		t.Errorf("missing headers: %s", out)
	}
}

// --- YAML v3 encoder handles maps ---
func TestYAMLMap(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"id": 1, "title": "map-yaml"}
	err := Format(&buf, data, "yaml", Options{})
	if err != nil {
		t.Fatalf("yaml map: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "id: 1") {
		t.Errorf("missing id: %s", out)
	}
}

// --- The table separator line ---
func TestTableSeparator(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 1, Title: "sep", BoardID: 5}
	err := Format(&buf, data, "table", Options{})
	if err != nil {
		t.Fatalf("table sep: %v", err)
	}
	out := buf.String()
	lines := strings.Split(out, "\n")
	// line 0: header
	// line 1: separator (dashes)
	// line 2: data row
	if len(lines) < 3 {
		t.Fatalf("not enough lines: %s", out)
	}
	if !strings.Contains(lines[1], "---") {
		t.Errorf("expected separator line with dashes: %s", lines[1])
	}
}

// --- Verify JSONL with non-slice data ---
func TestJSONLNonSlice(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"id": 1, "name": "single"}
	err := Format(&buf, data, "jsonl", Options{})
	if err != nil {
		t.Fatalf("jsonl non-slice: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], `"id":1`) {
		t.Errorf("missing data: %s", lines[0])
	}
}

// --- Ensure fields are lowercased matched ---
func TestFieldsCaseInsensitive(t *testing.T) {
	var buf bytes.Buffer
	data := testItem{ID: 1, Title: "x", Description: "d", BoardID: 5}
	err := Format(&buf, data, "json", Options{Fields: []string{"ID", "TITLE"}})
	if err != nil {
		t.Fatalf("fields case insensitive: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id"`) || !strings.Contains(out, `"title"`) {
		t.Errorf("fields should match case-insensitively: %s", out)
	}
	if strings.Contains(out, `"board_id"`) {
		t.Errorf("board_id should be excluded: %s", out)
	}
}

// --- Table with no rows ---
func TestTableEmptySlice(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, []testItem{}, "table", Options{})
	if err != nil {
		t.Fatalf("table empty: %v", err)
	}
	if buf.Len() > 0 {
		t.Errorf("expected empty: %s", buf.String())
	}
}

// --- Format with nil pointer ---
func TestFormatNilPointer(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, (*testStruct)(nil), "json", Options{})
	if err != nil {
		t.Fatalf("format nil ptr: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "null" {
		t.Errorf("expected 'null', got %q", out)
	}
}

// Helper to run a quick benchmark / example output for visual check.
func ExampleFormat_json() {
	var buf bytes.Buffer
	data := testItem{ID: 1, Title: "example", Description: "some desc", BoardID: 10}
	_ = Format(&buf, data, "json", Options{})
	fmt.Print(buf.String())
	// Output: {
	//   "id": 1,
	//   "title": "example",
	//   "description": "some desc",
	//   "board_id": 10
	// }
}

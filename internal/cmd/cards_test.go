package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.life-pay.ru/ai/kaiten-cli/internal/api"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/config"
)

// ---------------------------------------------------------------------------
//  Helpers
// ---------------------------------------------------------------------------

// captureStdout runs fn and returns everything written to os.Stdout.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old
	return buf.String()
}

// execCardsList creates a mock server and directly calls cardsListCmd.RunE
// with the given extra CLI flags. This avoids cobra global state bleeding.
func execCardsList(t *testing.T, extraArgs []string, handler http.HandlerFunc) (string, error) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	// Create a test client pointed at the mock server with rate limiting disabled.
	client := api.NewClient(&config.Config{
		Token:  "test-token",
		URL:    server.URL,
		DBPath: filepath.Join(t.TempDir(), "kaiten.db"),
	})
	client.DisableRateLimit()
	client.SetRetryConfig(1, 0, 0) // no retries

	// Set the global apiClient so cardsListCmd.RunE can use it.
	apiClient = client

	// Build args: first is the "list" subcommand name (which cobra strips),
	// then the board-id flag, then the extra args.
	args := []string{"--board-id", "7"}
	args = append(args, extraArgs...)

	// Reset flags on the cardsListCmd to defaults before setting new values.
	// This prevents flags from previous tests bleeding into this one.
	cardsListCmd.Flags().Set("board-id", "0")
	cardsListCmd.Flags().Set("archived", "false")
	cardsListCmd.Flags().Set("all", "false")
	cardsListCmd.Flags().Set("limit", "0")
	cardsListCmd.Flags().Set("offset", "0")
	cardsListCmd.Flags().Set("all-pages", "false")
	cardsListCmd.Flags().Set("output", "json")
	cardsListCmd.Flags().Set("fields", "")
	cardsListCmd.Flags().Set("no-descriptions", "false")
	cardsListCmd.Flags().Set("quiet", "false")

	// Parse the args onto cardsListCmd.
	cardsListCmd.ParseFlags(args)

	// Execute the RunE directly.
	var out string
	var err error
	out = captureStdout(func() {
		err = cardsListCmd.RunE(cardsListCmd, []string{})
	})
	return out, err
}

// ---------------------------------------------------------------------------
//  Tests
// ---------------------------------------------------------------------------

func TestCardsList_SendsLimitOffset(t *testing.T) {
	var reqURI string
	handler := func(w http.ResponseWriter, r *http.Request) {
		reqURI = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"title":"Card","board_id":7,"column_id":1}]`))
	}
	out, err := execCardsList(t, []string{"--limit", "25", "--offset", "10"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(reqURI, "limit=25") {
		t.Errorf("expected limit=25 in request, got: %s", reqURI)
	}
	if !strings.Contains(reqURI, "offset=10") {
		t.Errorf("expected offset=10 in request, got: %s", reqURI)
	}
	if !strings.Contains(out, `"id": 1`) {
		t.Errorf("output should contain card data: %s", out)
	}
}

func TestCardsList_AllPagesFetchesMultiplePages(t *testing.T) {
	callCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		switch callCount {
		case 1:
			_, _ = w.Write([]byte(`[{"id":1,"title":"A","board_id":7,"column_id":1},{"id":2,"title":"B","board_id":7,"column_id":1}]`))
		case 2:
			_, _ = w.Write([]byte(`[{"id":3,"title":"C","board_id":7,"column_id":1}]`))
		default:
			t.Errorf("unexpected call #%d", callCount)
		}
	}
	out, err := execCardsList(t, []string{"--all-pages", "--limit", "2"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
	if !strings.Contains(out, `"id": 1`) || !strings.Contains(out, `"id": 2`) || !strings.Contains(out, `"id": 3`) {
		t.Errorf("output should contain all 3 cards: %s", out)
	}
}

func TestCardsList_Fields(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"title":"Card","board_id":7,"column_id":1,"description":"desc"}]`))
	}
	out, err := execCardsList(t, []string{"--fields", "id,title"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"id"`) {
		t.Errorf("output should contain id: %s", out)
	}
	if !strings.Contains(out, `"title"`) {
		t.Errorf("output should contain title: %s", out)
	}
	if strings.Contains(out, `"board_id"`) {
		t.Errorf("output should NOT contain board_id: %s", out)
	}
	if strings.Contains(out, `"column_id"`) {
		t.Errorf("output should NOT contain column_id: %s", out)
	}
}

func TestCardsList_Quiet(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":42,"title":"Card","board_id":7,"column_id":1}]`))
	}
	out, err := execCardsList(t, []string{"--quiet"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 || lines[0] != "42" {
		t.Errorf("expected '42', got %q", out)
	}
}

func TestCardsList_QuietShortFlag(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":99,"title":"Quiet","board_id":7,"column_id":1}]`))
	}
	out, err := execCardsList(t, []string{"-q"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 || lines[0] != "99" {
		t.Errorf("expected '99', got %q", out)
	}
}

func TestCardsList_OutputJSONL(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"title":"A","board_id":7,"column_id":1},{"id":2,"title":"B","board_id":7,"column_id":1}]`))
	}
	out, err := execCardsList(t, []string{"--output", "jsonl"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 JSONL lines, got %d: %q", len(lines), out)
	}
	if !strings.Contains(lines[0], `"id":1`) {
		t.Errorf("first line should have id 1: %s", lines[0])
	}
	if !strings.Contains(lines[1], `"id":2`) {
		t.Errorf("second line should have id 2: %s", lines[1])
	}
}

func TestCardsList_NoDescriptions(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"title":"Card","board_id":7,"column_id":1,"description":"should-be-hidden"}]`))
	}
	out, err := execCardsList(t, []string{"--no-descriptions"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "should-be-hidden") || strings.Contains(out, `"description"`) {
		t.Errorf("output should not contain description: %s", out)
	}
}

func TestCardsList_DefaultOutputIsJSON(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"title":"Test","board_id":7,"column_id":1}]`))
	}
	out, err := execCardsList(t, []string{}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Errorf("expected JSON array as default output: %s", out)
	}
	var cards []map[string]any
	if err := json.Unmarshal([]byte(out), &cards); err != nil {
		t.Errorf("default output should be valid JSON: %v\noutput: %s", err, out)
	}
}

// --- Existing archived/all condition behavior ---

func TestCardsList_ArchivedCondition(t *testing.T) {
	var reqURI string
	handler := func(w http.ResponseWriter, r *http.Request) {
		reqURI = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"title":"Archived","board_id":7,"column_id":1}]`))
	}
	_, err := execCardsList(t, []string{"--archived"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(reqURI, "condition=2") {
		t.Errorf("expected condition=2 for archived, got: %s", reqURI)
	}
}

func TestCardsList_AllCondition(t *testing.T) {
	var reqURI string
	handler := func(w http.ResponseWriter, r *http.Request) {
		reqURI = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"title":"All","board_id":7,"column_id":1}]`))
	}
	_, err := execCardsList(t, []string{"--all"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(reqURI, "condition") {
		t.Errorf("condition should be omitted for --all (condition=0), got: %s", reqURI)
	}
}

func TestCardsList_DefaultActiveCondition(t *testing.T) {
	var reqURI string
	handler := func(w http.ResponseWriter, r *http.Request) {
		reqURI = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"title":"Active","board_id":7,"column_id":1}]`))
	}
	_, err := execCardsList(t, []string{}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(reqURI, "condition=1") {
		t.Errorf("expected condition=1 for active cards by default, got: %s", reqURI)
	}
}

// --- Output table format ---

func TestCardsList_OutputTable(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"title":"Card A","board_id":7,"column_id":3}]`))
	}
	out, err := execCardsList(t, []string{"--output", "table"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ID") {
		t.Errorf("table output should have ID header: %s", out)
	}
	if !strings.Contains(out, "TITLE") {
		t.Errorf("table output should have TITLE header: %s", out)
	}
	if !strings.Contains(out, "Card A") {
		t.Errorf("table output should contain 'Card A': %s", out)
	}
}

// --- Multiple cards with quiet ---

func TestCardsList_QuietMultipleCards(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":10,"title":"X","board_id":7,"column_id":1},{"id":20,"title":"Y","board_id":7,"column_id":1}]`))
	}
	out, err := execCardsList(t, []string{"-q"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 || lines[0] != "10" || lines[1] != "20" {
		t.Errorf("expected '10\\n20', got %q", out)
	}
}

// --- All-pages without limit uses default page size 100 ---

func TestCardsList_AllPagesDefaultLimit(t *testing.T) {
	var capturedLimit string
	handler := func(w http.ResponseWriter, r *http.Request) {
		capturedLimit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}
	_, err := execCardsList(t, []string{"--all-pages"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != "100" {
		t.Errorf("expected limit=100 (default), got: %s", capturedLimit)
	}
}

// --- Error: missing board-id ---

func TestCardsList_MissingBoardID(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "kaiten.db"))

	// Ensure all flags are at defaults before running.
	cardsListCmd.Flags().Set("board-id", "0")
	cardsListCmd.Flags().Set("all-pages", "false")
	cardsListCmd.Flags().Set("limit", "0")
	cardsListCmd.Flags().Set("output", "json")

	rootCmd.SetArgs([]string{"cards", "list"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --board-id")
	}
	if !strings.Contains(err.Error(), "required flag") && !strings.Contains(err.Error(), "board-id") {
		t.Errorf("unexpected error: %v", err)
	}
}

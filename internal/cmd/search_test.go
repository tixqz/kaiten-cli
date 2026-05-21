package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/db"
)

// ---------------------------------------------------------------------------
//  Helpers
// ---------------------------------------------------------------------------

// resetSearchCardsFlags resets all flags on searchCardsCmd to their default
// values and clears the Changed flag. This prevents cross-test contamination
// since cobra flag values persist across invocations of the same command.
func resetSearchCardsFlags() {
	searchCardsCmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Value.Set(f.DefValue)
		f.Changed = false
	})
}

// ---------------------------------------------------------------------------
//  Missing DB
// ---------------------------------------------------------------------------

func TestSearchCards_MissingDB(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "nonexistent", "kaiten.db"))

	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing database, got nil")
	}
	if !strings.Contains(err.Error(), "local database not found") {
		t.Errorf("expected 'local database not found' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "kaiten sync") {
		t.Errorf("expected sync hint in error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
//  Helper: seed a temp DB with test cards
// ---------------------------------------------------------------------------

func seedTestDB(t *testing.T, dbPath string) {
	t.Helper()
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	defer database.Close()

	cards := []db.Card{
		{
			ID: 1, Title: "Fix login bug", Description: "Users cannot log in with SSO",
			BoardID: 10, SpaceID: 100, ColumnID: 5, Condition: 1, OwnerID: 42,
			Created: "2025-01-15T10:00:00Z", Updated: "2025-01-20T08:00:00Z",
		},
		{
			ID: 2, Title: "Add dark mode", Description: "Implement dark mode toggle",
			BoardID: 10, SpaceID: 100, ColumnID: 6, Condition: 1, OwnerID: 99,
			Created: "2025-02-01T12:00:00Z", Updated: "2025-02-10T14:00:00Z",
		},
		{
			ID: 3, Title: "Archived card", Description: "Old task that is done",
			BoardID: 20, SpaceID: 200, ColumnID: 7, Condition: 2, OwnerID: 42,
			Created: "2024-12-01T09:00:00Z", Updated: "2024-12-15T16:00:00Z",
			CompletedAt: "2024-12-14T10:00:00Z",
		},
		{
			ID: 4, Title: "Completed task", Description: "This is done",
			BoardID: 10, SpaceID: 100, ColumnID: 5, Condition: 3, OwnerID: 42,
			Created: "2024-11-01T08:00:00Z", Updated: "2024-11-10T12:00:00Z",
			CompletedAt: "2024-11-09T16:00:00Z",
		},
	}

	if err := database.BulkUpsertCards(cards); err != nil {
		t.Fatalf("seeding cards: %v", err)
	}

	// Add members to card 1.
	if err := database.SetCardMembers(1, []int{100, 200}); err != nil {
		t.Fatalf("SetCardMembers: %v", err)
	}

	// Add tags to card 1.
	if err := database.SetCardTags(1, []int{1000, 2000}); err != nil {
		t.Fatalf("SetCardTags: %v", err)
	}
}

// ---------------------------------------------------------------------------
//  Basic search
// ---------------------------------------------------------------------------

func TestSearchCards_Basic(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards should not error: %v", err)
	}
}

func TestSearchCards_DoesNotRequireAPICredentials(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "")
	t.Setenv("KAITEN_URL", "")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--quiet"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards should not require API credentials: %v", err)
	}
}

// ---------------------------------------------------------------------------
//  Filter: board-id
// ---------------------------------------------------------------------------

func TestSearchCards_BoardFilter(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	// Capture output.
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--board-id", "10"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --board-id 10: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 1`) {
		t.Errorf("expected card 1 in output, got: %s", out)
	}
	if strings.Contains(out, `"id": 3`) {
		t.Errorf("card 3 (board 20) should not appear, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Filter: condition alias
// ---------------------------------------------------------------------------

func TestSearchCards_ConditionAlias(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--condition", "archived"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --condition archived: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 3`) {
		t.Errorf("expected card 3 (archived) in output, got: %s", out)
	}
	if strings.Contains(out, `"id": 1`) {
		t.Errorf("card 1 (active) should not appear, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Filter: condition numeric
// ---------------------------------------------------------------------------

func TestSearchCards_ConditionNumeric(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--condition", "3"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --condition 3: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 4`) {
		t.Errorf("expected card 4 (done/condition=3) in output, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Filter: owner-id
// ---------------------------------------------------------------------------

func TestSearchCards_OwnerFilter(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--owner-id", "99"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --owner-id 99: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 2`) {
		t.Errorf("expected card 2 (owner=99) in output, got: %s", out)
	}
	if strings.Contains(out, `"id": 1`) {
		t.Errorf("card 1 (owner=42) should not appear, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Filter: member-id
// ---------------------------------------------------------------------------

func TestSearchCards_MemberFilter(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--member-id", "100"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --member-id 100: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 1`) {
		t.Errorf("expected card 1 (member 100) in output, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Filter: tag-id
// ---------------------------------------------------------------------------

func TestSearchCards_TagFilter(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--tag-id", "1000"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --tag-id 1000: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 1`) {
		t.Errorf("expected card 1 (tag 1000) in output, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Filter: date range (created-from / created-to with YYYY-MM-DD)
// ---------------------------------------------------------------------------

func TestSearchCards_DateRange(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{
		"search", "cards",
		"--created-from", "2025-01-01",
		"--created-to", "2025-02-15",
	})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards date range: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 1`) {
		t.Errorf("expected card 1 (created 2025-01-15) in output, got: %s", out)
	}
	if !strings.Contains(out, `"id": 2`) {
		t.Errorf("expected card 2 (created 2025-02-01) in output, got: %s", out)
	}
	if strings.Contains(out, `"id": 3`) {
		t.Errorf("card 3 (created 2024-12) should not appear, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Filter: text search (FTS)
// ---------------------------------------------------------------------------

func TestSearchCards_TextSearch(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--text", "login"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --text login: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id": 1`) {
		t.Errorf("expected card 1 (title contains login) in output, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Output controls: quiet (-q)
// ---------------------------------------------------------------------------

func TestSearchCards_Quiet(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--board-id", "10", "-q"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards -q: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// Should have card 1, 2, 4 in order (board 10), one per line.
	if len(lines) != 3 {
		t.Errorf("expected 3 lines for quiet output, got %d: %s", len(lines), out)
	}
	if lines[0] != "1" && lines[1] != "2" && lines[2] != "4" {
		// Order depends on sort; just verify all three IDs appear.
		if !strings.Contains(out, "1") || !strings.Contains(out, "2") || !strings.Contains(out, "4") {
			t.Errorf("expected IDs 1,2,4 in quiet output, got: %s", out)
		}
	}
}

// ---------------------------------------------------------------------------
//  Output controls: fields
// ---------------------------------------------------------------------------

func TestSearchCards_Fields(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--board-id", "10", "--fields", "id,title"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --fields: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id":`) {
		t.Errorf("expected id field in output, got: %s", out)
	}
	if !strings.Contains(out, `"title":`) {
		t.Errorf("expected title field in output, got: %s", out)
	}
	if strings.Contains(out, `"description":`) {
		t.Errorf("description should be excluded with --fields, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Output controls: jsonl
// ---------------------------------------------------------------------------

func TestSearchCards_OutputJSONL(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--board-id", "10", "--output", "jsonl"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --output jsonl: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 jsonl lines, got %d", len(lines))
	}
	// Each line should be valid JSON.
	for i, line := range lines {
		if !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
			t.Errorf("line %d does not look like JSON: %q", i, line)
		}
	}
}

// ---------------------------------------------------------------------------
//  Output controls: no-descriptions
// ---------------------------------------------------------------------------

func TestSearchCards_NoDescriptions(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--board-id", "10", "--no-descriptions"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --no-descriptions: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, `"description":`) {
		t.Errorf("description should be excluded with --no-descriptions, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Invalid condition returns error
// ---------------------------------------------------------------------------

func TestSearchCards_InvalidCondition(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	// Create an empty DB so the command gets past the "file not found" check.
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	database.Close()

	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--condition", "bogus"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid condition")
	}
	if !strings.Contains(err.Error(), "invalid condition") {
		t.Errorf("expected invalid condition error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
//  Invalid date returns error
// ---------------------------------------------------------------------------

func TestSearchCards_InvalidDate(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	// Create an empty DB so the command gets past the "file not found" check.
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	database.Close()

	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--created-from", "not-a-date"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
	if !strings.Contains(err.Error(), "invalid date") {
		t.Errorf("expected invalid date error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
//  Pagination: limit/offset
// ---------------------------------------------------------------------------

func TestSearchCards_LimitOffset(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--limit", "2", "--offset", "1", "--sort", "id"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards --limit 2 --offset 1: %v", err)
	}
	out := buf.String()
	// With sort=id, offset=1, limit=2, we should get cards 2 and 3.
	if !strings.Contains(out, `"id": 2`) {
		t.Errorf("expected card 2 in output, got: %s", out)
	}
	if !strings.Contains(out, `"id": 3`) {
		t.Errorf("expected card 3 in output, got: %s", out)
	}
	if strings.Contains(out, `"id": 1`) {
		t.Errorf("card 1 should be skipped by offset, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
//  Search with no results
// ---------------------------------------------------------------------------

func TestSearchCards_NoResults(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)
	seedTestDB(t, dbPath)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	resetSearchCardsFlags()
	rootCmd.SetArgs([]string{"search", "cards", "--board-id", "999"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("search cards no results: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[]") {
		t.Errorf("expected empty JSON array for no results, got: %s", out)
	}
}

package db

import (
	"fmt"
	"testing"
)

func setupDB(t *testing.T) *DB {
	t.Helper()
	db, err := OpenInMemory()
	if err != nil {
		t.Fatalf("OpenInMemory: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// --- Schema / DB lifecycle ---

func TestSchemaInit(t *testing.T) {
	db := setupDB(t)

	// Verify all expected tables exist.
	tables := []string{
		"schema_meta", "spaces", "boards", "cards", "users",
		"card_members", "tags", "card_tags", "sync_state",
	}
	for _, name := range tables {
		var cnt int
		err := db.SQL.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", name,
		).Scan(&cnt)
		if err != nil {
			t.Fatalf("check table %s: %v", name, err)
		}
		if cnt == 0 {
			t.Errorf("table %s does not exist", name)
		}
	}

	// Verify FTS table exists.
	var ftsCnt int
	err := db.SQL.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='cards_fts'",
	).Scan(&ftsCnt)
	if err != nil {
		t.Fatalf("check cards_fts: %v", err)
	}
	if ftsCnt == 0 {
		t.Error("cards_fts table does not exist")
	}

	// Verify schema version.
	var version int
	err = db.SQL.QueryRow("SELECT version FROM schema_meta").Scan(&version)
	if err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != 1 {
		t.Errorf("expected schema version 1, got %d", version)
	}
}

func TestSchemaInitIdempotent(t *testing.T) {
	db := setupDB(t)
	// Second migrate should not error.
	if err := db.migrate(); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	// Version should still be 1.
	var count int
	err := db.SQL.QueryRow("SELECT COUNT(*) FROM schema_meta").Scan(&count)
	if err != nil {
		t.Fatalf("count schema_meta: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 schema_meta row, got %d", count)
	}
}

// --- Cards ---

func TestUpsertCard(t *testing.T) {
	db := setupDB(t)

	card := Card{
		ID: 1, Title: "Test Card", Description: "A description",
		BoardID: 10, SpaceID: 100, ColumnID: 5, Condition: 1,
		Created: "2025-01-01T00:00:00Z", Updated: "2025-01-02T00:00:00Z",
	}
	if err := db.UpsertCard(card); err != nil {
		t.Fatalf("UpsertCard: %v", err)
	}

	got, err := db.GetCard(1)
	if err != nil {
		t.Fatalf("GetCard: %v", err)
	}
	if got == nil {
		t.Fatal("card not found")
	}
	if got.Title != "Test Card" || got.Description != "A description" {
		t.Errorf("unexpected card: %+v", got)
	}
	if got.BoardID != 10 || got.SpaceID != 100 {
		t.Errorf("unexpected board/space: %+v", got)
	}
}

func TestUpsertCardUpdate(t *testing.T) {
	db := setupDB(t)

	card := Card{ID: 1, Title: "Original", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"}
	if err := db.UpsertCard(card); err != nil {
		t.Fatalf("UpsertCard: %v", err)
	}

	card.Title = "Updated"
	card.Updated = "2025-01-02T00:00:00Z"
	if err := db.UpsertCard(card); err != nil {
		t.Fatalf("UpsertCard update: %v", err)
	}

	got, err := db.GetCard(1)
	if err != nil {
		t.Fatalf("GetCard: %v", err)
	}
	if got.Title != "Updated" {
		t.Errorf("expected 'Updated', got %q", got.Title)
	}
	if got.Updated != "2025-01-02T00:00:00.000000000Z" {
		t.Errorf("expected updated time, got %q", got.Updated)
	}
}

func TestBulkUpsertCards(t *testing.T) {
	db := setupDB(t)

	cards := []Card{
		{ID: 1, Title: "Card A", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "Card B", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 3, Title: "Card C", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
	}
	if err := db.BulkUpsertCards(cards); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	for _, c := range cards {
		got, err := db.GetCard(c.ID)
		if err != nil {
			t.Fatalf("GetCard(%d): %v", c.ID, err)
		}
		if got == nil || got.Title != c.Title {
			t.Errorf("card %d: expected %q, got %+v", c.ID, c.Title, got)
		}
	}
}

// --- FTS Search ---

func TestSearchCardsFTS(t *testing.T) {
	db := setupDB(t)

	cards := []Card{
		{ID: 1, Title: "Alpha release", Description: "First release of the product", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "Beta testing", Description: "Testing phase begins", Created: "2025-01-02T00:00:00Z", Updated: "2025-01-02T00:00:00Z"},
		{ID: 3, Title: "Gamma feature", Description: "New feature release", Created: "2025-01-03T00:00:00Z", Updated: "2025-01-03T00:00:00Z"},
	}
	if err := db.BulkUpsertCards(cards); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	// Search by title keyword.
	results, err := db.SearchCards(SearchQuery{Text: "release"})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected FTS results for 'release'")
	}
	found := false
	for _, c := range results {
		if c.ID == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected card 1 in FTS results for 'release', got %+v", results)
	}

	// Search by description keyword.
	results, err = db.SearchCards(SearchQuery{Text: "testing"})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 1 || results[0].ID != 2 {
		t.Errorf("expected card 2 for 'testing', got %+v", results)
	}
}

// --- Filters ---

func TestSearchCardsFilterByBoardID(t *testing.T) {
	db := setupDB(t)

	cards := []Card{
		{ID: 1, Title: "A", BoardID: 10, Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "B", BoardID: 20, Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 3, Title: "C", BoardID: 10, Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
	}
	if err := db.BulkUpsertCards(cards); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	b10 := 10
	results, err := db.SearchCards(SearchQuery{BoardID: &b10})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 cards for board 10, got %d", len(results))
	}
}

func TestSearchCardsFilterBySpaceID(t *testing.T) {
	db := setupDB(t)

	cards := []Card{
		{ID: 1, Title: "A", SpaceID: 100, Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "B", SpaceID: 200, Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
	}
	if err := db.BulkUpsertCards(cards); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	s100 := 100
	results, err := db.SearchCards(SearchQuery{SpaceID: &s100})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 1 || results[0].ID != 1 {
		t.Errorf("expected card 1, got %+v", results)
	}
}

func TestSearchCardsFilterByCondition(t *testing.T) {
	db := setupDB(t)

	cards := []Card{
		{ID: 1, Title: "A", Condition: 1, Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "B", Condition: 2, Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
	}
	if err := db.BulkUpsertCards(cards); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	cond1 := 1
	results, err := db.SearchCards(SearchQuery{Condition: &cond1})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 1 || results[0].ID != 1 {
		t.Errorf("expected card 1, got %+v", results)
	}
}

func TestSearchCardsFilterByOwnerID(t *testing.T) {
	db := setupDB(t)

	cards := []Card{
		{ID: 1, Title: "A", OwnerID: 42, Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "B", OwnerID: 99, Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
	}
	if err := db.BulkUpsertCards(cards); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	o42 := 42
	results, err := db.SearchCards(SearchQuery{OwnerID: &o42})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 1 || results[0].ID != 1 {
		t.Errorf("expected card 1, got %+v", results)
	}
}

func TestSearchCardsFilterByMemberID(t *testing.T) {
	db := setupDB(t)

	if err := db.BulkUpsertCards([]Card{
		{ID: 1, Title: "Card with members", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "Card without", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
	}); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}
	if err := db.SetCardMembers(1, []int{100, 200}); err != nil {
		t.Fatalf("SetCardMembers: %v", err)
	}

	m100 := 100
	results, err := db.SearchCards(SearchQuery{MemberID: &m100})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 1 || results[0].ID != 1 {
		t.Errorf("expected card 1, got %+v", results)
	}
}

func TestSearchCardsFilterByTagID(t *testing.T) {
	db := setupDB(t)

	if err := db.BulkUpsertCards([]Card{
		{ID: 1, Title: "Card with tags", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "Card without", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
	}); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}
	if err := db.SetCardTags(1, []int{10, 20}); err != nil {
		t.Fatalf("SetCardTags: %v", err)
	}

	tag10 := 10
	results, err := db.SearchCards(SearchQuery{TagID: &tag10})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 1 || results[0].ID != 1 {
		t.Errorf("expected card 1, got %+v", results)
	}
}

// --- Date range filters ---

func TestSearchCardsCreatedRange(t *testing.T) {
	db := setupDB(t)

	if err := db.BulkUpsertCards([]Card{
		{ID: 1, Title: "Old", Created: "2024-12-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "Mid", Created: "2025-01-15T00:00:00Z", Updated: "2025-01-15T00:00:00Z"},
		{ID: 3, Title: "New", Created: "2025-02-01T00:00:00Z", Updated: "2025-02-01T00:00:00Z"},
	}); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	from := "2025-01-01T00:00:00Z"
	to := "2025-01-31T23:59:59Z"
	results, err := db.SearchCards(SearchQuery{CreatedFrom: &from, CreatedTo: &to})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 1 || results[0].ID != 2 {
		t.Errorf("expected card 2, got %+v", results)
	}
}

func TestSearchCardsUpdatedRange(t *testing.T) {
	db := setupDB(t)

	if err := db.BulkUpsertCards([]Card{
		{ID: 1, Title: "A", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-10T00:00:00Z"},
		{ID: 2, Title: "B", Created: "2025-01-01T00:00:00Z", Updated: "2025-02-10T00:00:00Z"},
	}); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	from := "2025-02-01T00:00:00Z"
	results, err := db.SearchCards(SearchQuery{UpdatedFrom: &from})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 1 || results[0].ID != 2 {
		t.Errorf("expected card 2, got %+v", results)
	}
}

func TestSearchCardsCompletedRange(t *testing.T) {
	db := setupDB(t)

	if err := db.BulkUpsertCards([]Card{
		{ID: 1, Title: "A", CompletedAt: "2025-01-15T00:00:00Z", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "B", CompletedAt: "2025-02-15T00:00:00Z", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 3, Title: "Incomplete", CompletedAt: "", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
	}); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	to := "2025-02-01T00:00:00Z"
	results, err := db.SearchCards(SearchQuery{CompletedTo: &to})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 1 || results[0].ID != 1 {
		t.Errorf("expected card 1, got %+v", results)
	}

	count, err := db.CountCards(SearchQuery{CompletedTo: &to})
	if err != nil {
		t.Fatalf("CountCards: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1 excluding incomplete cards, got %d", count)
	}
}

func TestSearchCardsDateRangeNormalizesOffsetsAndFractions(t *testing.T) {
	db := setupDB(t)

	if err := db.BulkUpsertCards([]Card{
		{ID: 1, Title: "Boundary", CompletedAt: "2025-12-31T23:30:00.123456789+03:00", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
		{ID: 2, Title: "Outside", CompletedAt: "2026-01-01T00:30:00.000000001Z", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"},
	}); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	from := "2025-01-01T00:00:00Z"
	to := "2025-12-31T23:59:59.999999999Z"
	results, err := db.SearchCards(SearchQuery{CompletedFrom: &from, CompletedTo: &to})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 1 || results[0].ID != 1 {
		t.Errorf("expected boundary card only, got %+v", results)
	}
}

// --- Pagination ---

func TestSearchCardsPagination(t *testing.T) {
	db := setupDB(t)

	cards := make([]Card, 10)
	for i := 0; i < 10; i++ {
		cards[i] = Card{
			ID: i + 1, Title: fmt.Sprintf("Card %d", i+1),
			Created: fmt.Sprintf("2025-01-%02dT00:00:00Z", i+1),
			Updated: fmt.Sprintf("2025-01-%02dT00:00:00Z", i+1),
		}
	}
	if err := db.BulkUpsertCards(cards); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	// Limit + Offset
	results, err := db.SearchCards(SearchQuery{Limit: 3, Offset: 2, Sort: "id"})
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	if results[0].ID != 3 || results[1].ID != 4 || results[2].ID != 5 {
		t.Errorf("expected IDs [3,4,5], got %+v", idsFromCards(results))
	}
}

// --- Sort ---

func TestSearchCardsSort(t *testing.T) {
	db := setupDB(t)

	cards := []Card{
		{ID: 1, Title: "C", Created: "2025-01-03T00:00:00Z", Updated: "2025-01-03T00:00:00Z", CompletedAt: "2025-01-03T00:00:00Z"},
		{ID: 2, Title: "A", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z", CompletedAt: "2025-01-01T00:00:00Z"},
		{ID: 3, Title: "B", Created: "2025-01-02T00:00:00Z", Updated: "2025-01-02T00:00:00Z", CompletedAt: "2025-01-02T00:00:00Z"},
	}
	if err := db.BulkUpsertCards(cards); err != nil {
		t.Fatalf("BulkUpsertCards: %v", err)
	}

	tests := []struct {
		sort string
		want []int
	}{
		{"created", []int{2, 3, 1}},
		{"created_desc", []int{1, 3, 2}},
		{"updated", []int{2, 3, 1}},
		{"updated_desc", []int{1, 3, 2}},
		{"completed", []int{2, 3, 1}},
		{"completed_desc", []int{1, 3, 2}},
		{"id", []int{1, 2, 3}},
		{"id_desc", []int{3, 2, 1}},
	}
	for _, tc := range tests {
		t.Run(tc.sort, func(t *testing.T) {
			results, err := db.SearchCards(SearchQuery{Sort: tc.sort})
			if err != nil {
				t.Fatalf("SearchCards: %v", err)
			}
			got := idsFromCards(results)
			if !intSliceEqual(got, tc.want) {
				t.Errorf("sort=%s: got %v, want %v", tc.sort, got, tc.want)
			}
		})
	}
}

// --- Users ---

func TestUpsertUser(t *testing.T) {
	db := setupDB(t)

	user := User{ID: 1, FullName: "Alice", Email: "alice@example.com"}
	if err := db.UpsertUser(user); err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}

	got, err := db.GetUser(1)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got == nil {
		t.Fatal("user not found")
	}
	if got.FullName != "Alice" || got.Email != "alice@example.com" {
		t.Errorf("unexpected user: %+v", got)
	}
}

func TestUpsertUserUpdate(t *testing.T) {
	db := setupDB(t)

	if err := db.UpsertUser(User{ID: 1, FullName: "Alice"}); err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}
	if err := db.UpsertUser(User{ID: 1, FullName: "Alice Updated", Email: "alice@new.com"}); err != nil {
		t.Fatalf("UpsertUser update: %v", err)
	}

	got, err := db.GetUser(1)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got.FullName != "Alice Updated" || got.Email != "alice@new.com" {
		t.Errorf("expected updated user, got %+v", got)
	}
}

func TestSetCardMembers(t *testing.T) {
	db := setupDB(t)

	if err := db.UpsertCard(Card{ID: 1, Title: "Test", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"}); err != nil {
		t.Fatalf("UpsertCard: %v", err)
	}

	if err := db.SetCardMembers(1, []int{10, 20, 30}); err != nil {
		t.Fatalf("SetCardMembers: %v", err)
	}

	members, err := db.GetCardMembers(1)
	if err != nil {
		t.Fatalf("GetCardMembers: %v", err)
	}
	if !intSliceEqual(members, []int{10, 20, 30}) {
		t.Errorf("expected [10,20,30], got %v", members)
	}

	// Replace members.
	if err := db.SetCardMembers(1, []int{99}); err != nil {
		t.Fatalf("SetCardMembers replace: %v", err)
	}
	members, err = db.GetCardMembers(1)
	if err != nil {
		t.Fatalf("GetCardMembers: %v", err)
	}
	if !intSliceEqual(members, []int{99}) {
		t.Errorf("expected [99], got %v", members)
	}
}

// --- Tags ---

func TestUpsertTag(t *testing.T) {
	db := setupDB(t)

	tag := Tag{ID: 1, Name: "bug", Color: 16711680}
	if err := db.UpsertTag(tag); err != nil {
		t.Fatalf("UpsertTag: %v", err)
	}

	got, err := db.GetTag(1)
	if err != nil {
		t.Fatalf("GetTag: %v", err)
	}
	if got == nil {
		t.Fatal("tag not found")
	}
	if got.Name != "bug" || got.Color != 16711680 {
		t.Errorf("unexpected tag: %+v", got)
	}
}

func TestUpsertTagUpdate(t *testing.T) {
	db := setupDB(t)

	if err := db.UpsertTag(Tag{ID: 1, Name: "bug"}); err != nil {
		t.Fatalf("UpsertTag: %v", err)
	}
	if err := db.UpsertTag(Tag{ID: 1, Name: "feature", Color: 65280}); err != nil {
		t.Fatalf("UpsertTag update: %v", err)
	}

	got, err := db.GetTag(1)
	if err != nil {
		t.Fatalf("GetTag: %v", err)
	}
	if got.Name != "feature" || got.Color != 65280 {
		t.Errorf("expected updated tag, got %+v", got)
	}
}

func TestSetCardTags(t *testing.T) {
	db := setupDB(t)

	if err := db.UpsertCard(Card{ID: 1, Title: "Test", Created: "2025-01-01T00:00:00Z", Updated: "2025-01-01T00:00:00Z"}); err != nil {
		t.Fatalf("UpsertCard: %v", err)
	}

	if err := db.SetCardTags(1, []int{10, 20}); err != nil {
		t.Fatalf("SetCardTags: %v", err)
	}

	tags, err := db.GetCardTags(1)
	if err != nil {
		t.Fatalf("GetCardTags: %v", err)
	}
	if !intSliceEqual(tags, []int{10, 20}) {
		t.Errorf("expected [10,20], got %v", tags)
	}

	// Replace tags.
	if err := db.SetCardTags(1, []int{30}); err != nil {
		t.Fatalf("SetCardTags replace: %v", err)
	}
	tags, err = db.GetCardTags(1)
	if err != nil {
		t.Fatalf("GetCardTags: %v", err)
	}
	if !intSliceEqual(tags, []int{30}) {
		t.Errorf("expected [30], got %v", tags)
	}
}

// --- Spaces / Boards ---

func TestUpsertSpace(t *testing.T) {
	db := setupDB(t)

	space := Space{ID: 1, Title: "My Space", Name: "my-space"}
	if err := db.UpsertSpace(space); err != nil {
		t.Fatalf("UpsertSpace: %v", err)
	}

	got, err := db.GetSpace(1)
	if err != nil {
		t.Fatalf("GetSpace: %v", err)
	}
	if got == nil {
		t.Fatal("space not found")
	}
	if got.Title != "My Space" || got.Name != "my-space" {
		t.Errorf("unexpected space: %+v", got)
	}
}

func TestListSpaces(t *testing.T) {
	db := setupDB(t)

	for i := 1; i <= 3; i++ {
		if err := db.UpsertSpace(Space{ID: i, Title: fmt.Sprintf("Space %d", i)}); err != nil {
			t.Fatalf("UpsertSpace: %v", err)
		}
	}

	spaces, err := db.ListSpaces()
	if err != nil {
		t.Fatalf("ListSpaces: %v", err)
	}
	if len(spaces) != 3 {
		t.Errorf("expected 3 spaces, got %d", len(spaces))
	}
}

func TestUpsertBoard(t *testing.T) {
	db := setupDB(t)

	board := Board{ID: 1, SpaceID: 10, Title: "Kanban Board"}
	if err := db.UpsertBoard(board); err != nil {
		t.Fatalf("UpsertBoard: %v", err)
	}

	got, err := db.GetBoard(1)
	if err != nil {
		t.Fatalf("GetBoard: %v", err)
	}
	if got == nil {
		t.Fatal("board not found")
	}
	if got.Title != "Kanban Board" || got.SpaceID != 10 {
		t.Errorf("unexpected board: %+v", got)
	}
}

func TestListBoardsBySpace(t *testing.T) {
	db := setupDB(t)

	boards := []Board{
		{ID: 1, SpaceID: 100, Title: "Board A"},
		{ID: 2, SpaceID: 100, Title: "Board B"},
		{ID: 3, SpaceID: 200, Title: "Board C"},
	}
	for _, b := range boards {
		if err := db.UpsertBoard(b); err != nil {
			t.Fatalf("UpsertBoard: %v", err)
		}
	}

	results, err := db.ListBoardsBySpace(100)
	if err != nil {
		t.Fatalf("ListBoardsBySpace: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 boards for space 100, got %d", len(results))
	}
}

// --- Sync State ---

func TestSyncStateSetGet(t *testing.T) {
	db := setupDB(t)

	state := SyncState{
		ScopeType:   "all",
		ScopeID:     0,
		Status:      "completed",
		CardsSynced: 150,
		StartedAt:   "2025-01-01T00:00:00Z",
		FinishedAt:  "2025-01-01T01:00:00Z",
	}
	if err := db.SetSyncState(state); err != nil {
		t.Fatalf("SetSyncState: %v", err)
	}

	got, err := db.GetSyncState("all", 0)
	if err != nil {
		t.Fatalf("GetSyncState: %v", err)
	}
	if got == nil {
		t.Fatal("sync state not found")
	}
	if got.Status != "completed" || got.CardsSynced != 150 {
		t.Errorf("unexpected sync state: %+v", got)
	}
}

func TestSyncStateUpdate(t *testing.T) {
	db := setupDB(t)

	initial := SyncState{
		ScopeType:   "board",
		ScopeID:     42,
		Status:      "in_progress",
		CardsSynced: 0,
	}
	if err := db.SetSyncState(initial); err != nil {
		t.Fatalf("SetSyncState: %v", err)
	}

	updated := SyncState{
		ScopeType:   "board",
		ScopeID:     42,
		Status:      "completed",
		CardsSynced: 100,
		FinishedAt:  "2025-01-01T01:00:00Z",
	}
	if err := db.SetSyncState(updated); err != nil {
		t.Fatalf("SetSyncState update: %v", err)
	}

	got, err := db.GetSyncState("board", 42)
	if err != nil {
		t.Fatalf("GetSyncState: %v", err)
	}
	if got.Status != "completed" || got.CardsSynced != 100 {
		t.Errorf("expected updated state, got %+v", got)
	}
}

func TestSyncStateNotFound(t *testing.T) {
	db := setupDB(t)

	got, err := db.GetSyncState("nonexistent", 0)
	if err != nil {
		t.Fatalf("GetSyncState: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing sync state, got %+v", got)
	}
}

func TestSyncStateReset(t *testing.T) {
	db := setupDB(t)

	if err := db.SetSyncState(SyncState{ScopeType: "all", ScopeID: 0, Status: "done"}); err != nil {
		t.Fatalf("SetSyncState: %v", err)
	}
	if err := db.SetSyncState(SyncState{ScopeType: "board", ScopeID: 1, Status: "done"}); err != nil {
		t.Fatalf("SetSyncState: %v", err)
	}

	if err := db.ResetSync(); err != nil {
		t.Fatalf("ResetSync: %v", err)
	}

	got, err := db.GetSyncState("all", 0)
	if err != nil {
		t.Fatalf("GetSyncState: %v", err)
	}
	if got != nil {
		t.Error("expected nil after reset")
	}
}

func TestStatusString(t *testing.T) {
	db := setupDB(t)

	status, err := db.StatusString()
	if err != nil {
		t.Fatalf("StatusString: %v", err)
	}
	if status != "no sync history" {
		t.Errorf("expected no sync history, got %q", status)
	}

	db.SetSyncState(SyncState{ScopeType: "board", ScopeID: 1, Status: "completed", CardsSynced: 50})
	status, err = db.StatusString()
	if err != nil {
		t.Fatalf("StatusString: %v", err)
	}
	if status == "no sync history" {
		t.Error("expected non-empty status")
	}
}

// --- Vacuum ---

func TestVacuum(t *testing.T) {
	db := setupDB(t)
	if err := db.Vacuum(); err != nil {
		t.Fatalf("Vacuum: %v", err)
	}
}

// --- Helpers ---

func idsFromCards(cards []Card) []int {
	ids := make([]int, len(cards))
	for i, c := range cards {
		ids[i] = c.ID
	}
	return ids
}

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Prevent unused import error for fmt in pagination test.
var _ = fmt.Sprintf("")

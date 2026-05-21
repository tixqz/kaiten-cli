package syncer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"gitlab.life-pay.ru/ai/kaiten-cli/internal/api"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/db"
)

// --------------------------------------------------------------------------
//  Test helpers
// --------------------------------------------------------------------------

func setupTest(t *testing.T) (*Syncer, *db.DB, *httptest.Server) {
	t.Helper()

	database, err := db.OpenInMemory()
	if err != nil {
		t.Fatalf("OpenInMemory: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default 404 - tests register specific handlers on the mux.
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	client := &api.Client{}
	client.SetBaseURL(server.URL)
	client.SetToken("test-token")
	client.SetHTTPClient(server.Client())

	syncer := New(client, database)
	return syncer, database, server
}

// stub registers a handler that asserts method and path, returning fixed JSON.
func stub(t *testing.T, method, path string, status int, body string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			t.Fatalf("%s: expected %s, got %s", path, method, r.Method)
		}
		if r.URL.Path != path {
			t.Fatalf("expected path %s, got %s", path, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

// stubJSON is like stub but auto-encodes the payload.
func stubJSON(t *testing.T, method, path string, status int, payload interface{}) http.HandlerFunc {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return stub(t, method, path, status, string(data))
}

// --------------------------------------------------------------------------
//  Board sync
// --------------------------------------------------------------------------

func TestSyncBoard_StoresCardsMembersTagsAndSyncState(t *testing.T) {
	syncer, database, server := setupTest(t)

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/boards/1":
			stubJSON(t, "GET", "/boards/1", 200, api.Board{ID: 1, Title: "Board A", SpaceID: 10})(w, r)
		case "/cards":
			q := r.URL.Query()
			if q.Get("board_id") != "1" {
				t.Fatalf("expected board_id=1, got %s", q.Get("board_id"))
			}
			stubJSON(t, "GET", "/cards", 200, []api.Card{
				{
					ID: 101, Title: "Card 1", BoardID: 1, ColumnID: 5,
					CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-02T00:00:00Z",
					OwnerID: intPtr(10),
				},
				{
					ID: 102, Title: "Card 2", BoardID: 1, ColumnID: 6,
					CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-02T00:00:00Z",
				},
			})(w, r)
		case "/cards/101/members":
			stubJSON(t, "GET", "/cards/101/members", 200, []api.Member{
				{ID: 10, FullName: "Alice", Email: "alice@example.com", Username: "alice"},
				{ID: 20, FullName: "Bob", Email: "bob@example.com", Username: "bob"},
			})(w, r)
		case "/cards/101/tags":
			stubJSON(t, "GET", "/cards/101/tags", 200, []api.CardTag{
				{ID: 1, Name: "bug", Color: 16711680, CardID: 101, TagID: 1001},
			})(w, r)
		case "/cards/102/members":
			stubJSON(t, "GET", "/cards/102/members", 200, []api.Member{})(w, r)
		case "/cards/102/tags":
			stubJSON(t, "GET", "/cards/102/tags", 200, []api.CardTag{})(w, r)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})

	result, err := syncer.SyncBoard(1)
	if err != nil {
		t.Fatalf("SyncBoard returned error: %v", err)
	}

	// Verify result fields.
	if result.ScopeType != "board" || result.ScopeID != 1 {
		t.Errorf("unexpected scope: %+v", result)
	}
	if result.Status != "completed" {
		t.Errorf("expected completed, got %s", result.Status)
	}
	if result.Error != "" {
		t.Errorf("expected no error, got %s", result.Error)
	}
	if result.CardsUpserted != 2 {
		t.Errorf("expected 2 cards upserted, got %d", result.CardsUpserted)
	}
	if result.Boards != 1 {
		t.Errorf("expected 1 board synced, got %d", result.Boards)
	}

	// Verify boards table.
	board, err := database.GetBoard(1)
	if err != nil {
		t.Fatalf("GetBoard: %v", err)
	}
	if board == nil || board.Title != "Board A" {
		t.Errorf("unexpected board: %+v", board)
	}

	// Verify cards stored.
	card1, err := database.GetCard(101)
	if err != nil {
		t.Fatalf("GetCard(101): %v", err)
	}
	if card1 == nil || card1.Title != "Card 1" || card1.OwnerID != 10 {
		t.Errorf("unexpected card1: %+v", card1)
	}
	if card1.SpaceID != 10 {
		t.Errorf("card1 SpaceID = %d, want 10", card1.SpaceID)
	}
	if card1.RawJSON == "" {
		t.Error("card1 RawJSON should not be empty")
	}

	card2, err := database.GetCard(102)
	if err != nil {
		t.Fatalf("GetCard(102): %v", err)
	}
	if card2 == nil || card2.Title != "Card 2" {
		t.Errorf("unexpected card2: %+v", card2)
	}

	// Verify members stored.
	memberIDs, err := database.GetCardMembers(101)
	if err != nil {
		t.Fatalf("GetCardMembers(101): %v", err)
	}
	if len(memberIDs) != 2 || memberIDs[0] != 10 || memberIDs[1] != 20 {
		t.Errorf("unexpected members for card 101: %v", memberIDs)
	}

	memberIDs2, err := database.GetCardMembers(102)
	if err != nil {
		t.Fatalf("GetCardMembers(102): %v", err)
	}
	if len(memberIDs2) != 0 {
		t.Errorf("expected no members for card 102, got %v", memberIDs2)
	}

	// Verify users upserted.
	user, err := database.GetUser(10)
	if err != nil {
		t.Fatalf("GetUser(10): %v", err)
	}
	if user == nil || user.FullName != "Alice" || user.Name != "alice" {
		t.Errorf("unexpected user 10: %+v", user)
	}

	// Verify tags stored.
	tagIDs, err := database.GetCardTags(101)
	if err != nil {
		t.Fatalf("GetCardTags(101): %v", err)
	}
	if len(tagIDs) != 1 || tagIDs[0] != 1001 {
		t.Errorf("unexpected tags for card 101: %v", tagIDs)
	}

	tag, err := database.GetTag(1001)
	if err != nil {
		t.Fatalf("GetTag(1001): %v", err)
	}
	if tag == nil || tag.Name != "bug" || tag.Color != 16711680 {
		t.Errorf("unexpected tag: %+v", tag)
	}

	// Verify sync_state.
	state, err := database.GetSyncState("board", 1)
	if err != nil {
		t.Fatalf("GetSyncState: %v", err)
	}
	if state == nil || state.Status != "completed" || state.CardsSynced != 2 {
		t.Errorf("unexpected sync state: %+v", state)
	}
}

// --------------------------------------------------------------------------
//  Space sync
// --------------------------------------------------------------------------

func TestSyncSpace_ListsBoardsAndStoresCards(t *testing.T) {
	syncer, database, server := setupTest(t)

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/spaces/5":
			stubJSON(t, "GET", "/spaces/5", 200, api.Space{ID: 5, Title: "Space Five"})(w, r)
		case "/spaces/5/boards":
			stubJSON(t, "GET", "/spaces/5/boards", 200, []api.Board{
				{ID: 1, Title: "Board 1", SpaceID: 5},
				{ID: 2, Title: "Board 2", SpaceID: 5},
			})(w, r)
		case "/boards/1":
			stubJSON(t, "GET", "/boards/1", 200, api.Board{ID: 1, Title: "Board 1", SpaceID: 5})(w, r)
		case "/boards/2":
			stubJSON(t, "GET", "/boards/2", 200, api.Board{ID: 2, Title: "Board 2", SpaceID: 5})(w, r)
		case "/cards":
			q := r.URL.Query()
			boardID := q.Get("board_id")
			switch boardID {
			case "1":
				stubJSON(t, "GET", "/cards", 200, []api.Card{
					{ID: 101, Title: "Card A", BoardID: 1, ColumnID: 3, CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-02T00:00:00Z"},
				})(w, r)
			case "2":
				stubJSON(t, "GET", "/cards", 200, []api.Card{
					{ID: 102, Title: "Card B", BoardID: 2, ColumnID: 4, CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-02T00:00:00Z"},
				})(w, r)
			default:
				t.Fatalf("unexpected board_id: %s", boardID)
			}
		case "/cards/101/members":
			stubJSON(t, "GET", "/cards/101/members", 200, []api.Member{})(w, r)
		case "/cards/101/tags":
			stubJSON(t, "GET", "/cards/101/tags", 200, []api.CardTag{})(w, r)
		case "/cards/102/members":
			stubJSON(t, "GET", "/cards/102/members", 200, []api.Member{})(w, r)
		case "/cards/102/tags":
			stubJSON(t, "GET", "/cards/102/tags", 200, []api.CardTag{})(w, r)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})

	result, err := syncer.SyncSpace(5)
	if err != nil {
		t.Fatalf("SyncSpace returned error: %v", err)
	}

	if result.ScopeType != "space" || result.ScopeID != 5 {
		t.Errorf("unexpected scope: %+v", result)
	}
	if result.Status != "completed" {
		t.Errorf("expected completed, got %s", result.Status)
	}
	if result.Boards != 2 {
		t.Errorf("expected 2 boards, got %d", result.Boards)
	}
	if result.CardsUpserted != 2 {
		t.Errorf("expected 2 cards upserted, got %d", result.CardsUpserted)
	}

	// Verify space stored.
	space, err := database.GetSpace(5)
	if err != nil {
		t.Fatalf("GetSpace: %v", err)
	}
	if space == nil || space.Title != "Space Five" {
		t.Errorf("unexpected space: %+v", space)
	}

	// Verify boards stored.
	b1, err := database.GetBoard(1)
	if err != nil {
		t.Fatalf("GetBoard(1): %v", err)
	}
	if b1 == nil || b1.Title != "Board 1" || b1.SpaceID != 5 {
		t.Errorf("unexpected board 1: %+v", b1)
	}

	b2, err := database.GetBoard(2)
	if err != nil {
		t.Fatalf("GetBoard(2): %v", err)
	}
	if b2 == nil || b2.Title != "Board 2" {
		t.Errorf("unexpected board 2: %+v", b2)
	}

	// Verify cards stored.
	card, err := database.GetCard(101)
	if err != nil {
		t.Fatalf("GetCard(101): %v", err)
	}
	if card == nil || card.Title != "Card A" {
		t.Errorf("unexpected card: %+v", card)
	}

	// Verify space-level sync_state.
	state, err := database.GetSyncState("space", 5)
	if err != nil {
		t.Fatalf("GetSyncState: %v", err)
	}
	if state == nil || state.Status != "completed" {
		t.Errorf("unexpected space sync state: %+v", state)
	}
}

// --------------------------------------------------------------------------
//  All sync
// --------------------------------------------------------------------------

func TestSyncAll_ListsSpacesAndBoards(t *testing.T) {
	syncer, database, server := setupTest(t)

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/spaces":
			stubJSON(t, "GET", "/spaces", 200, []api.Space{
				{ID: 1, Title: "Space 1"},
			})(w, r)
		case "/spaces/1/boards":
			stubJSON(t, "GET", "/spaces/1/boards", 200, []api.Board{
				{ID: 10, Title: "Board 10", SpaceID: 1},
			})(w, r)
		case "/boards/10":
			stubJSON(t, "GET", "/boards/10", 200, api.Board{ID: 10, Title: "Board 10", SpaceID: 1})(w, r)
		case "/cards":
			stubJSON(t, "GET", "/cards", 200, []api.Card{
				{ID: 1001, Title: "Card X", BoardID: 10, ColumnID: 2, CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-02T00:00:00Z"},
			})(w, r)
		case "/cards/1001/members":
			stubJSON(t, "GET", "/cards/1001/members", 200, []api.Member{})(w, r)
		case "/cards/1001/tags":
			stubJSON(t, "GET", "/cards/1001/tags", 200, []api.CardTag{})(w, r)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})

	result, err := syncer.SyncAll()
	if err != nil {
		t.Fatalf("SyncAll returned error: %v", err)
	}

	if result.ScopeType != "all" || result.ScopeID != 0 {
		t.Errorf("unexpected scope: %+v", result)
	}
	if result.Status != "completed" {
		t.Errorf("expected completed, got %s", result.Status)
	}
	if result.Boards != 1 {
		t.Errorf("expected 1 board, got %d", result.Boards)
	}
	if result.CardsUpserted != 1 {
		t.Errorf("expected 1 card upserted, got %d", result.CardsUpserted)
	}

	// Verify space stored.
	space, err := database.GetSpace(1)
	if err != nil {
		t.Fatalf("GetSpace: %v", err)
	}
	if space == nil || space.Title != "Space 1" {
		t.Errorf("unexpected space: %+v", space)
	}

	// Verify board stored.
	board, err := database.GetBoard(10)
	if err != nil {
		t.Fatalf("GetBoard: %v", err)
	}
	if board == nil || board.Title != "Board 10" {
		t.Errorf("unexpected board: %+v", board)
	}

	// Verify card stored.
	card, err := database.GetCard(1001)
	if err != nil {
		t.Fatalf("GetCard: %v", err)
	}
	if card == nil || card.Title != "Card X" {
		t.Errorf("unexpected card: %+v", card)
	}

	// Verify all-level sync_state.
	state, err := database.GetSyncState("all", 0)
	if err != nil {
		t.Fatalf("GetSyncState: %v", err)
	}
	if state == nil || state.Status != "completed" || state.CardsSynced != 1 {
		t.Errorf("unexpected sync state: %+v", state)
	}
}

// --------------------------------------------------------------------------
//  Pagination
// --------------------------------------------------------------------------

func TestSyncBoard_UsesPagination(t *testing.T) {
	syncer, database, server := setupTest(t)

	callCount := 0
	pageSize := 100
	// Build a page of exactly 100 cards.
	firstPage := make([]api.Card, pageSize)
	for i := 0; i < pageSize; i++ {
		firstPage[i] = api.Card{
			ID: i + 1, Title: fmt.Sprintf("Card %d", i+1), BoardID: 99, ColumnID: 1,
			CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-01T00:00:00Z",
		}
	}

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/boards/99":
			stubJSON(t, "GET", "/boards/99", 200, api.Board{ID: 99, Title: "Board Paginated", SpaceID: 1})(w, r)
		case "/cards":
			callCount++
			q := r.URL.Query()
			if q.Get("board_id") != "99" {
				t.Fatalf("expected board_id=99, got %s", q.Get("board_id"))
			}
			offset, _ := strconv.Atoi(q.Get("offset"))
			w.Header().Set("Content-Type", "application/json")
			switch callCount {
			case 1:
				if offset != 0 {
					t.Fatalf("expected offset=0 on first call, got %d", offset)
				}
				data, _ := json.Marshal(firstPage)
				_, _ = w.Write(data)
			case 2:
				if offset != pageSize {
					t.Fatalf("expected offset=%d on second call, got %d", pageSize, offset)
				}
				_, _ = w.Write([]byte(`[]`))
			default:
				t.Fatalf("unexpected call #%d", callCount)
			}
		case "/cards/1/members", "/cards/1/tags",
			"/cards/2/members", "/cards/2/tags":
			// Catch-all for member/tag endpoints.
			stubJSON(t, "GET", r.URL.Path, 200, []interface{}{})(w, r)
		default:
			// For all other card member/tag paths, return empty.
			if strings.HasPrefix(r.URL.Path, "/cards/") && (strings.HasSuffix(r.URL.Path, "/members") || strings.HasSuffix(r.URL.Path, "/tags")) {
				stubJSON(t, "GET", r.URL.Path, 200, []interface{}{})(w, r)
				return
			}
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})

	result, err := syncer.SyncBoard(99)
	if err != nil {
		t.Fatalf("SyncBoard returned error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
	if result.CardsUpserted != pageSize {
		t.Errorf("expected %d cards, got %d", pageSize, result.CardsUpserted)
	}

	// Verify first and last cards exist.
	for _, id := range []int{1, pageSize} {
		card, err := database.GetCard(id)
		if err != nil {
			t.Fatalf("GetCard(%d): %v", id, err)
		}
		if card == nil {
			t.Errorf("card %d not found", id)
		}
	}
}

func TestSyncBoard_UsesDefaultPageSize100(t *testing.T) {
	syncer, _, server := setupTest(t)

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/boards/1":
			stubJSON(t, "GET", "/boards/1", 200, api.Board{ID: 1, Title: "B", SpaceID: 1})(w, r)
		case "/cards":
			q := r.URL.Query()
			if q.Get("limit") != "100" {
				t.Errorf("expected limit=100, got %s", q.Get("limit"))
			}
			stubJSON(t, "GET", "/cards", 200, []api.Card{})(w, r)
		case "/cards/1/members", "/cards/1/tags", "/cards/2/members", "/cards/2/tags":
			stubJSON(t, "GET", r.URL.Path, 200, []api.CardTag{})(w, r)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})

	_, err := syncer.SyncBoard(1)
	if err != nil {
		t.Fatalf("SyncBoard returned error: %v", err)
	}
}

// --------------------------------------------------------------------------
//  Error handling
// --------------------------------------------------------------------------

func TestSyncBoard_FailureUpdatesSyncStateWithError(t *testing.T) {
	syncer, database, server := setupTest(t)

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/boards/1" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`internal error`))
			return
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
	})

	_, err := syncer.SyncBoard(1)
	if err == nil {
		t.Fatal("expected error from SyncBoard, got nil")
	}

	// Verify sync_state has error status.
	state, err := database.GetSyncState("board", 1)
	if err != nil {
		t.Fatalf("GetSyncState: %v", err)
	}
	if state == nil {
		t.Fatal("sync_state should exist even on failure")
	}
	if state.Status != "error" {
		t.Errorf("expected status 'error', got %q", state.Status)
	}
	if state.Error == "" {
		t.Error("expected non-empty error text in sync_state")
	}
}

func TestSyncBoard_CardAPIErrorUpdatesSyncState(t *testing.T) {
	syncer, database, server := setupTest(t)

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/boards/2":
			stubJSON(t, "GET", "/boards/2", 200, api.Board{ID: 2, Title: "Board B", SpaceID: 1})(w, r)
		case "/cards":
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`service down`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	})

	_, err := syncer.SyncBoard(2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	state, err := database.GetSyncState("board", 2)
	if err != nil {
		t.Fatalf("GetSyncState: %v", err)
	}
	if state == nil {
		t.Fatal("sync_state should exist on failure")
	}
	if state.Status != "error" {
		t.Errorf("expected 'error', got %q", state.Status)
	}
	if !strings.Contains(state.Error, "service down") && !strings.Contains(state.Error, "503") {
		t.Errorf("expected error mentioning service down or 503, got: %s", state.Error)
	}
}

func TestSyncAll_FailureUpdatesSyncStateWithError(t *testing.T) {
	syncer, database, server := setupTest(t)

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/spaces" {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`forbidden`))
			return
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
	})

	_, err := syncer.SyncAll()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	state, err := database.GetSyncState("all", 0)
	if err != nil {
		t.Fatalf("GetSyncState: %v", err)
	}
	if state == nil {
		t.Fatal("sync_state should exist on failure")
	}
	if state.Status != "error" {
		t.Errorf("expected 'error', got %q", state.Status)
	}
	if !strings.Contains(state.Error, "forbidden") && !strings.Contains(state.Error, "403") {
		t.Errorf("expected error mentioning forbidden or 403, got: %s", state.Error)
	}
}

// --------------------------------------------------------------------------
//  Helpers
// --------------------------------------------------------------------------

func intPtr(n int) *int {
	return &n
}

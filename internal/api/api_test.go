package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func newTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := &Client{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: server.Client(),
	}
	return client, server
}

func TestSpaces(t *testing.T) {
	t.Run("ListSpaces", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/spaces" {
				t.Fatalf("expected path /spaces, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":1,"title":"Space 1"},{"id":2,"title":"Space 2"}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		spaces, err := client.ListSpaces()
		if err != nil {
			t.Fatalf("ListSpaces returned error: %v", err)
		}
		if len(spaces) != 2 {
			t.Fatalf("expected 2 spaces, got %d", len(spaces))
		}
		if spaces[0].ID != 1 || spaces[0].Title != "Space 1" {
			t.Fatalf("unexpected first space: %+v", spaces[0])
		}
	})

	t.Run("GetSpace", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/spaces/1" {
				t.Fatalf("expected path /spaces/1, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":1,"title":"Space 1","description":"desc"}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		space, err := client.GetSpace(1)
		if err != nil {
			t.Fatalf("GetSpace returned error: %v", err)
		}
		if space.ID != 1 || space.Title != "Space 1" || space.Description != "desc" {
			t.Fatalf("unexpected space: %+v", space)
		}
	})
}

func TestBoards(t *testing.T) {
	t.Run("ListBoards", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/spaces/5/boards" {
				t.Fatalf("expected path /spaces/5/boards, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"id":1,"title":"Board 1","space_id":5},{"id":2,"title":"Board 2","space_id":5}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		boards, err := client.ListBoards(5)
		if err != nil {
			t.Fatalf("ListBoards returned error: %v", err)
		}
		if len(boards) != 2 || boards[0].SpaceID != 5 {
			t.Fatalf("unexpected boards: %+v", boards)
		}
	})

	t.Run("GetBoard", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/boards/10" {
				t.Fatalf("expected path /boards/10, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"id":10,"title":"Board","space_id":3}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		board, err := client.GetBoard(10)
		if err != nil {
			t.Fatalf("GetBoard returned error: %v", err)
		}
		if board.ID != 10 || board.SpaceID != 3 {
			t.Fatalf("unexpected board: %+v", board)
		}
	})
}

func TestColumns(t *testing.T) {
	t.Run("ListColumns", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/boards/7/columns" {
				t.Fatalf("expected path /boards/7/columns, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"id":1,"title":"Column","board_id":7,"sort":1}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		columns, err := client.ListColumns(7)
		if err != nil {
			t.Fatalf("ListColumns returned error: %v", err)
		}
		if len(columns) != 1 || columns[0].Sort != 1 {
			t.Fatalf("unexpected columns: %+v", columns)
		}
	})
}

func TestLanes(t *testing.T) {
	t.Run("ListLanes", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/boards/7/lanes" {
				t.Fatalf("expected path /boards/7/lanes, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"id":1,"title":"Lane","board_id":7,"sort":2}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		lanes, err := client.ListLanes(7)
		if err != nil {
			t.Fatalf("ListLanes returned error: %v", err)
		}
		if len(lanes) != 1 || lanes[0].Sort != 2 {
			t.Fatalf("unexpected lanes: %+v", lanes)
		}
	})
}

func TestCards(t *testing.T) {
	t.Run("ListCards", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.RequestURI() != "/cards?board_id=3" {
				t.Fatalf("expected path /cards?board_id=3, got %s", r.URL.RequestURI())
			}
			_, _ = w.Write([]byte(`[{"id":1,"title":"Card","board_id":3,"column_id":1}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		cards, err := client.ListCards(3, 0)
		if err != nil {
			t.Fatalf("ListCards returned error: %v", err)
		}
		if len(cards) != 1 || cards[0].BoardID != 3 {
			t.Fatalf("unexpected cards: %+v", cards)
		}
	})

	t.Run("GetCard", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42" {
				t.Fatalf("expected path /cards/42, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"id":42,"title":"Card","board_id":3,"column_id":4}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		card, err := client.GetCard(42)
		if err != nil {
			t.Fatalf("GetCard returned error: %v", err)
		}
		if card.ID != 42 || card.ColumnID != 4 {
			t.Fatalf("unexpected card: %+v", card)
		}
	})

	t.Run("CreateCard", func(t *testing.T) {
		reqBody := &CreateCardRequest{Title: "New Card", BoardID: 1, ColumnID: 2}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/cards" {
				t.Fatalf("expected path /cards, got %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}
			if payload["title"] != "New Card" || int(payload["board_id"].(float64)) != 1 || int(payload["column_id"].(float64)) != 2 {
				t.Fatalf("unexpected payload: %v", payload)
			}
			_, _ = w.Write([]byte(`{"id":99,"title":"New Card","board_id":1,"column_id":2}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		card, err := client.CreateCard(reqBody)
		if err != nil {
			t.Fatalf("CreateCard returned error: %v", err)
		}
		if card.ID != 99 || card.Title != "New Card" {
			t.Fatalf("unexpected card: %+v", card)
		}
	})

	t.Run("UpdateCard", func(t *testing.T) {
		title := "Updated"
		columnID := 7
		reqBody := &UpdateCardRequest{Title: &title, ColumnID: &columnID}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				t.Fatalf("expected PATCH, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42" {
				t.Fatalf("expected path /cards/42, got %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}
			if len(payload) != 2 {
				t.Fatalf("expected 2 fields, got %d", len(payload))
			}
			if payload["title"] != "Updated" || int(payload["column_id"].(float64)) != 7 {
				t.Fatalf("unexpected payload: %v", payload)
			}
			_, _ = w.Write([]byte(`{"id":42,"title":"Updated","column_id":7}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		card, err := client.UpdateCard(42, reqBody)
		if err != nil {
			t.Fatalf("UpdateCard returned error: %v", err)
		}
		if card.Title != "Updated" || card.ColumnID != 7 {
			t.Fatalf("unexpected card: %+v", card)
		}
	})

	t.Run("DeleteCard", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Fatalf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42" {
				t.Fatalf("expected path /cards/42, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}
		client, server := newTestClient(handler)
		defer server.Close()

		if err := client.DeleteCard(42); err != nil {
			t.Fatalf("DeleteCard returned error: %v", err)
		}
	})
}

func TestComments(t *testing.T) {
	t.Run("ListComments", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/comments" {
				t.Fatalf("expected path /cards/42/comments, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"id":1,"text":"hello","type":1,"card_id":42,"author_id":5}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		comments, err := client.ListComments(42)
		if err != nil {
			t.Fatalf("ListComments returned error: %v", err)
		}
		if len(comments) != 1 {
			t.Fatalf("expected 1 comment, got %d", len(comments))
		}
		if comments[0].Text != "hello" {
			t.Fatalf("unexpected comment: %+v", comments[0])
		}
	})

	t.Run("AddComment", func(t *testing.T) {
		reqBody := &CreateCommentRequest{Text: "test comment", Type: 1}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/comments" {
				t.Fatalf("expected path /cards/42/comments, got %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}
			if payload["text"] != "test comment" {
				t.Fatalf("unexpected payload: %v", payload)
			}
			_, _ = w.Write([]byte(`{"id":2,"text":"test comment","type":1,"card_id":42,"author_id":5}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		comment, err := client.AddComment(42, reqBody)
		if err != nil {
			t.Fatalf("AddComment returned error: %v", err)
		}
		if comment.ID != 2 || comment.Text != "test comment" {
			t.Fatalf("unexpected comment: %+v", comment)
		}
	})
}

func TestBlockers(t *testing.T) {
	t.Run("BlockCard", func(t *testing.T) {
		reqBody := &BlockCardRequest{Reason: "Need info"}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/cards/1/blockers" {
				t.Fatalf("expected path /cards/1/blockers, got %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}
			if payload["reason"] != "Need info" {
				t.Fatalf("unexpected payload: %v", payload)
			}
			_, _ = w.Write([]byte("{\"id\":5,\"reason\":\"Need info\",\"card_id\":1,\"blocker_id\":2,\"released\":false,\"created\":\"now\",\"updated\":\"now\"}"))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		blocker, err := client.BlockCard(1, reqBody)
		if err != nil {
			t.Fatalf("BlockCard returned error: %v", err)
		}
		if blocker.ID != 5 || blocker.Reason != "Need info" {
			t.Fatalf("unexpected blocker: %+v", blocker)
		}
	})

	t.Run("UpdateBlocker", func(t *testing.T) {
		released := true
		reqBody := &UpdateBlockerRequest{Reason: "Updated reason", Released: &released}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				t.Fatalf("expected PATCH, got %s", r.Method)
			}
			if r.URL.Path != "/cards/1/blockers/5" {
				t.Fatalf("expected path /cards/1/blockers/5, got %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}
			if payload["reason"] != "Updated reason" || payload["released"] != true {
				t.Fatalf("unexpected payload: %v", payload)
			}
			_, _ = w.Write([]byte("{\"id\":5,\"reason\":\"Updated reason\",\"card_id\":1,\"blocker_id\":2,\"released\":true,\"created\":\"now\",\"updated\":\"now\"}"))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		blocker, err := client.UpdateBlocker(1, 5, reqBody)
		if err != nil {
			t.Fatalf("UpdateBlocker returned error: %v", err)
		}
		if !blocker.Released || blocker.Reason != "Updated reason" {
			t.Fatalf("unexpected blocker: %+v", blocker)
		}
	})

	t.Run("DeleteBlocker", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Fatalf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/cards/1/blockers/5" {
				t.Fatalf("expected path /cards/1/blockers/5, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}
		client, server := newTestClient(handler)
		defer server.Close()

		if err := client.DeleteBlocker(1, 5); err != nil {
			t.Fatalf("DeleteBlocker returned error: %v", err)
		}
	})
}

func TestTags(t *testing.T) {
	t.Run("ListTags", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/tags" {
				t.Fatalf("expected path /tags, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"id":1,"name":"bug","color":1},{"id":2,"name":"feature","color":2}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		tags, err := client.ListTags()
		if err != nil {
			t.Fatalf("ListTags returned error: %v", err)
		}
		if len(tags) != 2 {
			t.Fatalf("expected 2 tags, got %d", len(tags))
		}
		if tags[0].Name != "bug" {
			t.Fatalf("unexpected tag: %+v", tags[0])
		}
	})

	t.Run("ListCardTags", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/tags" {
				t.Fatalf("expected path /cards/42/tags, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"id":1,"name":"bug","color":1,"card_id":42,"tag_id":10}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		tags, err := client.ListCardTags(42)
		if err != nil {
			t.Fatalf("ListCardTags returned error: %v", err)
		}
		if len(tags) != 1 || tags[0].TagID != 10 {
			t.Fatalf("unexpected card tags: %+v", tags)
		}
	})

	t.Run("AddCardTag", func(t *testing.T) {
		reqBody := &AddTagRequest{Name: "bug"}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/tags" {
				t.Fatalf("expected path /cards/42/tags, got %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}
			if payload["name"] != "bug" {
				t.Fatalf("unexpected payload: %v", payload)
			}
			_, _ = w.Write([]byte(`{"id":1,"name":"bug","color":1,"company_id":5}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		tag, err := client.AddCardTag(42, reqBody)
		if err != nil {
			t.Fatalf("AddCardTag returned error: %v", err)
		}
		if tag.ID != 1 || tag.Name != "bug" {
			t.Fatalf("unexpected tag: %+v", tag)
		}
	})

	t.Run("RemoveCardTag", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Fatalf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/tags/10" {
				t.Fatalf("expected path /cards/42/tags/10, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}
		client, server := newTestClient(handler)
		defer server.Close()

		if err := client.RemoveCardTag(42, 10); err != nil {
			t.Fatalf("RemoveCardTag returned error: %v", err)
		}
	})
}

func TestMembers(t *testing.T) {
	t.Run("ListCardMembers", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/members" {
				t.Fatalf("expected path /cards/42/members, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`[{"id":1,"full_name":"John Doe","email":"john@example.com","username":"johndoe","type":1}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		members, err := client.ListCardMembers(42)
		if err != nil {
			t.Fatalf("ListCardMembers returned error: %v", err)
		}
		if len(members) != 1 {
			t.Fatalf("expected 1 member, got %d", len(members))
		}
		if members[0].FullName != "John Doe" || members[0].Username != "johndoe" {
			t.Fatalf("unexpected member: %+v", members[0])
		}
	})
}

func TestListCardsWithOptions(t *testing.T) {
	t.Run("sends query params", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			q := r.URL.Query()
			if q.Get("board_id") != "7" {
				t.Fatalf("expected board_id=7, got %s", q.Get("board_id"))
			}
			if q.Get("condition") != "1" {
				t.Fatalf("expected condition=1, got %s", q.Get("condition"))
			}
			if q.Get("limit") != "50" {
				t.Fatalf("expected limit=50, got %s", q.Get("limit"))
			}
			if q.Get("offset") != "10" {
				t.Fatalf("expected offset=10, got %s", q.Get("offset"))
			}
			if q.Get("owner_id") != "42" {
				t.Fatalf("expected owner_id=42, got %s", q.Get("owner_id"))
			}
			if q.Get("member_id") != "99" {
				t.Fatalf("expected member_id=99, got %s", q.Get("member_id"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":1,"title":"Card","board_id":7}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		opts := ListCardsOptions{BoardID: 7, Condition: 1, Limit: 50, Offset: 10, OwnerID: 42, MemberID: 99}
		cards, err := client.ListCardsWithOptions(opts)
		if err != nil {
			t.Fatalf("ListCardsWithOptions returned error: %v", err)
		}
		if len(cards) != 1 || cards[0].BoardID != 7 {
			t.Fatalf("unexpected cards: %+v", cards)
		}
	})

	t.Run("omits zero-value params", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			q := r.URL.Query()
			if q.Get("board_id") != "5" {
				t.Fatalf("expected board_id=5, got %s", q.Get("board_id"))
			}
			if q.Get("condition") != "" {
				t.Fatalf("condition should be omitted, got %s", q.Get("condition"))
			}
			if q.Get("limit") != "" {
				t.Fatalf("limit should be omitted, got %s", q.Get("limit"))
			}
			if q.Get("offset") != "" {
				t.Fatalf("offset should be omitted, got %s", q.Get("offset"))
			}
			_, _ = w.Write([]byte(`[]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		_, err := client.ListCardsWithOptions(ListCardsOptions{BoardID: 5})
		if err != nil {
			t.Fatalf("ListCardsWithOptions returned error: %v", err)
		}
	})
}

func TestCardExpandedFields(t *testing.T) {
	t.Run("unmarshal expanded fields", func(t *testing.T) {
		jsonData := `[
  {
    "id": 100,
    "title": "Expanded Card",
    "board_id": 3,
    "column_id": 5,
    "owner_id": 10,
    "updater_id": 20,
    "completed_at": "2024-06-01T12:00:00Z",
    "column_changed_at": "2024-06-01T11:00:00Z",
    "comments_total": 7,
    "tag_ids": [1, 2, 3],
    "sprint_id": 15,
    "sort_order": 500
  }
]`
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(jsonData))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		cards, err := client.ListCardsWithOptions(ListCardsOptions{BoardID: 3})
		if err != nil {
			t.Fatalf("ListCardsWithOptions returned error: %v", err)
		}
		if len(cards) != 1 {
			t.Fatalf("expected 1 card, got %d", len(cards))
		}
		c := cards[0]
		if c.ID != 100 || c.Title != "Expanded Card" {
			t.Fatalf("basic fields mismatch: %+v", c)
		}
		if c.OwnerID == nil || *c.OwnerID != 10 {
			t.Fatalf("OwnerID expected 10, got %+v", c.OwnerID)
		}
		if c.UpdaterID == nil || *c.UpdaterID != 20 {
			t.Fatalf("UpdaterID expected 20, got %+v", c.UpdaterID)
		}
		if c.CompletedAt == nil || *c.CompletedAt != "2024-06-01T12:00:00Z" {
			t.Fatalf("CompletedAt mismatch: %+v", c.CompletedAt)
		}
		if c.ColumnChangedAt == nil || *c.ColumnChangedAt != "2024-06-01T11:00:00Z" {
			t.Fatalf("ColumnChangedAt mismatch: %+v", c.ColumnChangedAt)
		}
		if c.CommentsTotal == nil || *c.CommentsTotal != 7 {
			t.Fatalf("CommentsTotal expected 7, got %+v", c.CommentsTotal)
		}
		if len(c.TagIDs) != 3 || c.TagIDs[0] != 1 || c.TagIDs[1] != 2 || c.TagIDs[2] != 3 {
			t.Fatalf("TagIDs expected [1 2 3], got %v", c.TagIDs)
		}
		if c.SprintID == nil || *c.SprintID != 15 {
			t.Fatalf("SprintID expected 15, got %+v", c.SprintID)
		}
		if c.SortOrder == nil || *c.SortOrder != 500 {
			t.Fatalf("SortOrder expected 500, got %+v", c.SortOrder)
		}
	})

	t.Run("raw json populated", func(t *testing.T) {
		jsonData := `[{"id":1,"title":"Raw","board_id":2,"column_id":3}]`
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(jsonData))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		cards, err := client.ListCardsWithOptions(ListCardsOptions{BoardID: 2})
		if err != nil {
			t.Fatalf("ListCardsWithOptions returned error: %v", err)
		}
		if len(cards) != 1 {
			t.Fatalf("expected 1 card, got %d", len(cards))
		}
		if cards[0].RawJSON != jsonData[1:len(jsonData)-1] {
			t.Fatalf("RawJSON mismatch:\nexpected: %s\ngot: %s", jsonData, cards[0].RawJSON)
		}
	})

	t.Run("raw json via GetCard", func(t *testing.T) {
		jsonData := `{"id":1,"title":"Single","board_id":2,"column_id":3}`
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(jsonData))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		card, err := client.GetCard(1)
		if err != nil {
			t.Fatalf("GetCard returned error: %v", err)
		}
		if card.RawJSON != jsonData {
			t.Fatalf("RawJSON mismatch:\nexpected: %s\ngot: %s", jsonData, card.RawJSON)
		}
	})
}

func TestListCardsAllPages(t *testing.T) {
	t.Run("fetches single page when results < limit", func(t *testing.T) {
		callCount := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":1,"title":"A","board_id":2,"column_id":3}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		cards, err := client.ListCardsAllPages(ListCardsOptions{BoardID: 2, Limit: 100})
		if err != nil {
			t.Fatalf("ListCardsAllPages returned error: %v", err)
		}
		if len(cards) != 1 {
			t.Fatalf("expected 1 card, got %d", len(cards))
		}
		if callCount != 1 {
			t.Fatalf("expected 1 call, got %d", callCount)
		}
	})

	t.Run("fetches multiple pages", func(t *testing.T) {
		callCount := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			q := r.URL.Query()
			offset, _ := strconv.Atoi(q.Get("offset"))
			w.Header().Set("Content-Type", "application/json")
			switch callCount {
			case 1:
				if offset != 0 {
					t.Fatalf("expected offset=0 on first call, got %d", offset)
				}
				_, _ = w.Write([]byte(`[{"id":1},{"id":2},{"id":3}]`))
			case 2:
				if offset != 3 {
					t.Fatalf("expected offset=3 on second call, got %d", offset)
				}
				_, _ = w.Write([]byte(`[{"id":4}]`))
			default:
				t.Fatalf("unexpected call #%d", callCount)
			}
		}
		client, server := newTestClient(handler)
		defer server.Close()

		cards, err := client.ListCardsAllPages(ListCardsOptions{BoardID: 2, Limit: 3})
		if err != nil {
			t.Fatalf("ListCardsAllPages returned error: %v", err)
		}
		if len(cards) != 4 {
			t.Fatalf("expected 4 cards, got %d", len(cards))
		}
		if callCount != 2 {
			t.Fatalf("expected 2 calls, got %d", callCount)
		}
	})

	t.Run("default limit is 100", func(t *testing.T) {
		callCount := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			q := r.URL.Query()
			if q.Get("limit") != "100" {
				t.Fatalf("expected limit=100 by default, got %s", q.Get("limit"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":1}]`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		cards, err := client.ListCardsAllPages(ListCardsOptions{BoardID: 2})
		if err != nil {
			t.Fatalf("ListCardsAllPages returned error: %v", err)
		}
		if len(cards) != 1 {
			t.Fatalf("expected 1 card, got %d", len(cards))
		}
		if callCount != 1 {
			t.Fatalf("expected 1 call, got %d", callCount)
		}
	})
}

func TestChecklists(t *testing.T) {
	t.Run("CreateChecklist", func(t *testing.T) {
		reqBody := &CreateChecklistRequest{Name: "Todo"}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/checklists" {
				t.Fatalf("expected path /cards/42/checklists, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"id":1,"name":"Todo","items":[],"created":"now","updated":"now"}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		cl, err := client.CreateChecklist(42, reqBody)
		if err != nil {
			t.Fatalf("CreateChecklist returned error: %v", err)
		}
		if cl.ID != 1 || cl.Name != "Todo" {
			t.Fatalf("unexpected checklist: %+v", cl)
		}
	})

	t.Run("GetChecklist", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/checklists/1" {
				t.Fatalf("expected path /cards/42/checklists/1, got %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"id":1,"name":"Todo","items":[{"id":10,"text":"Item 1","checked":false}],"created":"now","updated":"now"}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		cl, err := client.GetChecklist(42, 1)
		if err != nil {
			t.Fatalf("GetChecklist returned error: %v", err)
		}
		if cl.ID != 1 || len(cl.Items) != 1 || cl.Items[0].Text != "Item 1" {
			t.Fatalf("unexpected checklist: %+v", cl)
		}
	})

	t.Run("DeleteChecklist", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Fatalf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/checklists/1" {
				t.Fatalf("expected path /cards/42/checklists/1, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}
		client, server := newTestClient(handler)
		defer server.Close()

		if err := client.DeleteChecklist(42, 1); err != nil {
			t.Fatalf("DeleteChecklist returned error: %v", err)
		}
	})

	t.Run("AddChecklistItem", func(t *testing.T) {
		reqBody := &AddChecklistItemRequest{Text: "New item"}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/checklists/1/items" {
				t.Fatalf("expected path /cards/42/checklists/1/items, got %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}
			if payload["text"] != "New item" {
				t.Fatalf("unexpected payload: %v", payload)
			}
			_, _ = w.Write([]byte(`{"id":10,"text":"New item","checked":false,"sort_order":1,"created":"now","updated":"now"}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		item, err := client.AddChecklistItem(42, 1, reqBody)
		if err != nil {
			t.Fatalf("AddChecklistItem returned error: %v", err)
		}
		if item.ID != 10 || item.Text != "New item" || item.Checked {
			t.Fatalf("unexpected item: %+v", item)
		}
	})

	t.Run("UpdateChecklistItem", func(t *testing.T) {
		checked := true
		reqBody := &UpdateChecklistItemRequest{Checked: &checked}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				t.Fatalf("expected PATCH, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/checklists/1/items/10" {
				t.Fatalf("expected path /cards/42/checklists/1/items/10, got %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}
			if payload["checked"] != true {
				t.Fatalf("unexpected payload: %v", payload)
			}
			_, _ = w.Write([]byte(`{"id":10,"text":"New item","checked":true,"sort_order":1,"created":"now","updated":"now"}`))
		}
		client, server := newTestClient(handler)
		defer server.Close()

		item, err := client.UpdateChecklistItem(42, 1, 10, reqBody)
		if err != nil {
			t.Fatalf("UpdateChecklistItem returned error: %v", err)
		}
		if !item.Checked {
			t.Fatalf("expected checked=true, got %+v", item)
		}
	})

	t.Run("DeleteChecklistItem", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Fatalf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/cards/42/checklists/1/items/10" {
				t.Fatalf("expected path /cards/42/checklists/1/items/10, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}
		client, server := newTestClient(handler)
		defer server.Close()

		if err := client.DeleteChecklistItem(42, 1, 10); err != nil {
			t.Fatalf("DeleteChecklistItem returned error: %v", err)
		}
	})
}

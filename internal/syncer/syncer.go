// Package syncer orchestrates fetching Kaiten data via API and storing it in SQLite.
package syncer

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.life-pay.ru/ai/kaiten-cli/internal/api"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/db"
)

// SyncResult captures the outcome of a sync operation.
type SyncResult struct {
	ScopeType       string `json:"scope_type"`
	ScopeID         int    `json:"scope_id"`
	Boards          int    `json:"boards"`
	CardsUpserted   int    `json:"cards_upserted"`
	MembersUpserted int    `json:"members_upserted"`
	TagsUpserted    int    `json:"tags_upserted"`
	StartedAt       string `json:"started_at"`
	FinishedAt      string `json:"finished_at"`
	Status          string `json:"status"`
	Error           string `json:"error,omitempty"`
}

// Syncer pulls data from the Kaiten API and persists it via the DB layer.
type Syncer struct {
	api *api.Client
	db  *db.DB
}

// New creates a new Syncer.
func New(apiClient *api.Client, database *db.DB) *Syncer {
	return &Syncer{
		api: apiClient,
		db:  database,
	}
}

// --------------------------------------------------------------------------
//  Public sync methods
// --------------------------------------------------------------------------

// SyncBoard fetches all data for a single board and writes it to the DB.
func (s *Syncer) SyncBoard(boardID int) (SyncResult, error) {
	result := SyncResult{
		ScopeType: "board",
		ScopeID:   boardID,
		StartedAt: nowISO(),
	}

	// Mark in-progress upfront.
	_ = s.db.SetSyncState(db.SyncState{
		ScopeType: result.ScopeType,
		ScopeID:   result.ScopeID,
		StartedAt: result.StartedAt,
		Status:    "in_progress",
	})

	if err := s.syncBoard(&result, boardID); err != nil {
		result.Status = "error"
		result.Error = err.Error()
		result.FinishedAt = nowISO()
		_ = s.db.SetSyncState(db.SyncState{
			ScopeType:  result.ScopeType,
			ScopeID:    result.ScopeID,
			StartedAt:  result.StartedAt,
			FinishedAt: result.FinishedAt,
			Status:     result.Status,
			Error:      result.Error,
		})
		return result, err
	}

	result.Boards = 1
	result.Status = "completed"
	result.FinishedAt = nowISO()
	_ = s.db.SetSyncState(db.SyncState{
		ScopeType:   result.ScopeType,
		ScopeID:     result.ScopeID,
		StartedAt:   result.StartedAt,
		FinishedAt:  result.FinishedAt,
		Status:      result.Status,
		CardsSynced: result.CardsUpserted,
	})
	return result, nil
}

// SyncSpace fetches all data for a space (boards + cards) and writes it to the DB.
func (s *Syncer) SyncSpace(spaceID int) (SyncResult, error) {
	result := SyncResult{
		ScopeType: "space",
		ScopeID:   spaceID,
		StartedAt: nowISO(),
	}

	_ = s.db.SetSyncState(db.SyncState{
		ScopeType: result.ScopeType,
		ScopeID:   result.ScopeID,
		StartedAt: result.StartedAt,
		Status:    "in_progress",
	})

	// Upsert space metadata.
	apiSpace, err := s.api.GetSpace(spaceID)
	if err != nil {
		s.failSync(&result, fmt.Errorf("get space %d: %w", spaceID, err))
		return result, err
	}
	spaceRaw, _ := json.Marshal(apiSpace)
	if err := s.db.UpsertSpace(db.Space{
		ID:      apiSpace.ID,
		Title:   apiSpace.Title,
		RawJSON: string(spaceRaw),
	}); err != nil {
		s.failSync(&result, fmt.Errorf("upsert space %d: %w", spaceID, err))
		return result, err
	}

	// List boards in the space.
	apiBoards, err := s.api.ListBoards(spaceID)
	if err != nil {
		s.failSync(&result, fmt.Errorf("list boards for space %d: %w", spaceID, err))
		return result, err
	}

	for _, apiBoard := range apiBoards {
		boardRaw, _ := json.Marshal(apiBoard)
		if err := s.db.UpsertBoard(db.Board{
			ID:      apiBoard.ID,
			SpaceID: spaceID,
			Title:   apiBoard.Title,
			RawJSON: string(boardRaw),
		}); err != nil {
			s.failSync(&result, fmt.Errorf("upsert board %d: %w", apiBoard.ID, err))
			return result, err
		}
		result.Boards++

		// Sync cards for this board, accumulating stats.
		if err := s.syncBoard(&result, apiBoard.ID); err != nil {
			s.failSync(&result, fmt.Errorf("sync board %d: %w", apiBoard.ID, err))
			return result, err
		}
	}

	result.Status = "completed"
	result.FinishedAt = nowISO()
	_ = s.db.SetSyncState(db.SyncState{
		ScopeType:   result.ScopeType,
		ScopeID:     result.ScopeID,
		StartedAt:   result.StartedAt,
		FinishedAt:  result.FinishedAt,
		Status:      result.Status,
		CardsSynced: result.CardsUpserted,
	})
	return result, nil
}

// SyncAll fetches all spaces, boards, and cards from the API and persists them.
func (s *Syncer) SyncAll() (SyncResult, error) {
	result := SyncResult{
		ScopeType: "all",
		ScopeID:   0,
		StartedAt: nowISO(),
	}

	_ = s.db.SetSyncState(db.SyncState{
		ScopeType: result.ScopeType,
		ScopeID:   result.ScopeID,
		StartedAt: result.StartedAt,
		Status:    "in_progress",
	})

	// List all spaces.
	apiSpaces, err := s.api.ListSpaces()
	if err != nil {
		s.failSync(&result, fmt.Errorf("list spaces: %w", err))
		return result, err
	}

	for _, apiSpace := range apiSpaces {
		spaceRaw, _ := json.Marshal(apiSpace)
		if err := s.db.UpsertSpace(db.Space{
			ID:      apiSpace.ID,
			Title:   apiSpace.Title,
			RawJSON: string(spaceRaw),
		}); err != nil {
			s.failSync(&result, fmt.Errorf("upsert space %d: %w", apiSpace.ID, err))
			return result, err
		}

		// List boards for this space.
		apiBoards, err := s.api.ListBoards(apiSpace.ID)
		if err != nil {
			s.failSync(&result, fmt.Errorf("list boards for space %d: %w", apiSpace.ID, err))
			return result, err
		}

		for _, apiBoard := range apiBoards {
			boardRaw, _ := json.Marshal(apiBoard)
			if err := s.db.UpsertBoard(db.Board{
				ID:      apiBoard.ID,
				SpaceID: apiSpace.ID,
				Title:   apiBoard.Title,
				RawJSON: string(boardRaw),
			}); err != nil {
				s.failSync(&result, fmt.Errorf("upsert board %d: %w", apiBoard.ID, err))
				return result, err
			}
			result.Boards++

			// Sync cards for this board.
			if err := s.syncBoard(&result, apiBoard.ID); err != nil {
				s.failSync(&result, fmt.Errorf("sync board %d: %w", apiBoard.ID, err))
				return result, err
			}
		}
	}

	result.Status = "completed"
	result.FinishedAt = nowISO()
	_ = s.db.SetSyncState(db.SyncState{
		ScopeType:   result.ScopeType,
		ScopeID:     result.ScopeID,
		StartedAt:   result.StartedAt,
		FinishedAt:  result.FinishedAt,
		Status:      result.Status,
		CardsSynced: result.CardsUpserted,
	})
	return result, nil
}

// --------------------------------------------------------------------------
//  Internal board sync
// --------------------------------------------------------------------------

// syncBoard fetches cards, members, and tags for a board and persists them.
// It writes board-level sync_state and accumulates counters into result.
// The boardID parameter specifies which board to sync; result carries the
// parent scope (used for error reporting) and accumulates counters.
func (s *Syncer) syncBoard(result *SyncResult, boardID int) error {
	apiBoard, err := s.api.GetBoard(boardID)
	if err != nil {
		return fmt.Errorf("get board %d: %w", boardID, err)
	}
	boardSpaceID := 0
	if apiBoard != nil {
		boardSpaceID = apiBoard.SpaceID
		boardRaw, _ := json.Marshal(apiBoard)
		if err := s.db.UpsertBoard(db.Board{
			ID:      apiBoard.ID,
			SpaceID: apiBoard.SpaceID,
			Title:   apiBoard.Title,
			RawJSON: string(boardRaw),
		}); err != nil {
			return fmt.Errorf("upsert board %d: %w", apiBoard.ID, err)
		}
	}

	// Fetch all cards for the board using pagination.
	cards, err := s.api.ListCardsAllPages(api.ListCardsOptions{
		BoardID: boardID,
		Limit:   100,
	})
	if err != nil {
		return fmt.Errorf("list cards for board %d: %w", boardID, err)
	}

	// Convert API cards to DB cards.
	dbCards := make([]db.Card, 0, len(cards))
	for _, c := range cards {
		dbCards = append(dbCards, apiCardToDBCard(c, boardSpaceID))
	}

	if err := s.db.BulkUpsertCards(dbCards); err != nil {
		return fmt.Errorf("bulk upsert cards for board %d: %w", boardID, err)
	}
	result.CardsUpserted += len(dbCards)

	// Process members and tags for each card.
	memberCount := 0
	tagCount := 0
	for _, c := range cards {
		n, err := s.syncCardMembers(c.ID)
		if err != nil {
			return fmt.Errorf("sync members for card %d: %w", c.ID, err)
		}
		memberCount += n

		n, err = s.syncCardTags(c.ID)
		if err != nil {
			return fmt.Errorf("sync tags for card %d: %w", c.ID, err)
		}
		tagCount += n
	}
	result.MembersUpserted += memberCount
	result.TagsUpserted += tagCount

	// Write board-level sync_state.
	_ = s.db.SetSyncState(db.SyncState{
		ScopeType:   "board",
		ScopeID:     boardID,
		StartedAt:   nowISO(),
		FinishedAt:  nowISO(),
		Status:      "completed",
		CardsSynced: len(dbCards),
	})

	return nil
}

// syncCardMembers fetches members for a card and persists users + card_members.
// Returns the number of users upserted.
func (s *Syncer) syncCardMembers(cardID int) (int, error) {
	members, err := s.api.ListCardMembers(cardID)
	if err != nil {
		return 0, fmt.Errorf("list members for card %d: %w", cardID, err)
	}

	userIDs := make([]int, 0, len(members))
	for _, m := range members {
		memberRaw, _ := json.Marshal(m)
		if err := s.db.UpsertUser(db.User{
			ID:       m.ID,
			FullName: m.FullName,
			Name:     m.Username,
			Email:    m.Email,
			RawJSON:  string(memberRaw),
		}); err != nil {
			return 0, fmt.Errorf("upsert user %d: %w", m.ID, err)
		}
		userIDs = append(userIDs, m.ID)
	}

	if err := s.db.SetCardMembers(cardID, userIDs); err != nil {
		return 0, fmt.Errorf("set card members for card %d: %w", cardID, err)
	}

	return len(members), nil
}

// syncCardTags fetches tags for a card and persists tags + card_tags.
// Returns the number of tags upserted.
func (s *Syncer) syncCardTags(cardID int) (int, error) {
	cardTags, err := s.api.ListCardTags(cardID)
	if err != nil {
		return 0, fmt.Errorf("list tags for card %d: %w", cardID, err)
	}

	tagIDs := make([]int, 0, len(cardTags))
	for _, ct := range cardTags {
		tagRaw, _ := json.Marshal(ct)
		if err := s.db.UpsertTag(db.Tag{
			ID:      ct.TagID,
			Name:    ct.Name,
			Color:   ct.Color,
			RawJSON: string(tagRaw),
		}); err != nil {
			return 0, fmt.Errorf("upsert tag %d: %w", ct.TagID, err)
		}
		tagIDs = append(tagIDs, ct.TagID)
	}

	if err := s.db.SetCardTags(cardID, tagIDs); err != nil {
		return 0, fmt.Errorf("set card tags for card %d: %w", cardID, err)
	}

	return len(cardTags), nil
}

// failSync writes an error sync_state and updates result fields.
func (s *Syncer) failSync(result *SyncResult, syncErr error) {
	result.Status = "error"
	result.Error = syncErr.Error()
	result.FinishedAt = nowISO()
	_ = s.db.SetSyncState(db.SyncState{
		ScopeType:  result.ScopeType,
		ScopeID:    result.ScopeID,
		StartedAt:  result.StartedAt,
		FinishedAt: result.FinishedAt,
		Status:     result.Status,
		Error:      result.Error,
	})
}

// --------------------------------------------------------------------------
//  Helpers
// --------------------------------------------------------------------------

// apiCardToDBCard converts an api.Card to a db.Card, mapping pointer fields.
func apiCardToDBCard(c api.Card, spaceID int) db.Card {
	return db.Card{
		ID:              c.ID,
		Title:           c.Title,
		Description:     c.Description,
		BoardID:         c.BoardID,
		SpaceID:         spaceID,
		ColumnID:        c.ColumnID,
		LaneID:          c.LaneID,
		TypeID:          c.TypeID,
		OwnerID:         intPtrVal(c.OwnerID),
		UpdaterID:       intPtrVal(c.UpdaterID),
		Condition:       c.Condition,
		SizeText:        c.SizeText,
		DueDate:         c.DueDate,
		Blocked:         c.Blocked,
		BlockReason:     c.BlockReason,
		Created:         c.CreatedAt,
		Updated:         c.UpdatedAt,
		CompletedAt:     strPtrVal(c.CompletedAt),
		ColumnChangedAt: strPtrVal(c.ColumnChangedAt),
		CommentsTotal:   intPtrVal(c.CommentsTotal),
		SprintID:        intPtrVal(c.SprintID),
		SortOrder:       intPtrToFloat64(c.SortOrder),
		RawJSON:         c.RawJSON,
	}
}

func intPtrVal(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func strPtrVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func intPtrToFloat64(p *int) float64 {
	if p == nil {
		return 0
	}
	return float64(*p)
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

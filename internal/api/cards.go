package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// Card represents a Kaiten card.
type Card struct {
	ID              int     `json:"id"`
	Title           string  `json:"title"`
	Description     string  `json:"description,omitempty"`
	BoardID         int     `json:"board_id"`
	ColumnID        int     `json:"column_id"`
	LaneID          int     `json:"lane_id,omitempty"`
	TypeID          *int    `json:"type_id,omitempty"`
	SizeText        string  `json:"size_text,omitempty"`
	Condition       int     `json:"condition,omitempty"`
	DueDate         string  `json:"due_date,omitempty"`
	Blocked         bool    `json:"blocked,omitempty"`
	BlockReason     string  `json:"block_reason,omitempty"`
	CreatedAt       string  `json:"created"`
	UpdatedAt       string  `json:"updated"`
	OwnerID         *int    `json:"owner_id,omitempty"`
	UpdaterID       *int    `json:"updater_id,omitempty"`
	CompletedAt     *string `json:"completed_at,omitempty"`
	ColumnChangedAt *string `json:"column_changed_at,omitempty"`
	CommentsTotal   *int    `json:"comments_total,omitempty"`
	TagIDs          []int   `json:"tag_ids,omitempty"`
	SprintID        *int    `json:"sprint_id,omitempty"`
	SortOrder       *int    `json:"sort_order,omitempty"`
	RawJSON         string  `json:"-" yaml:"-"`
}

// UnmarshalJSON captures RawJSON while decoding the rest of the fields.
func (c *Card) UnmarshalJSON(data []byte) error {
	c.RawJSON = string(data)
	type cardAlias Card // avoid infinite recursion
	var alias cardAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*c = Card(alias)
	c.RawJSON = string(data)
	return nil
}

// CreateCardRequest is the payload for creating a card.
type CreateCardRequest struct {
	Title       string `json:"title"`
	BoardID     int    `json:"board_id"`
	ColumnID    int    `json:"column_id"`
	LaneID      int    `json:"lane_id,omitempty"`
	TypeID      int    `json:"type_id,omitempty"`
	Position    int    `json:"position,omitempty"` // 1=top, 2=bottom
	Description string `json:"description,omitempty"`
	SizeText    string `json:"size_text,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
}

// UpdateCardRequest is the payload for updating a card.
type UpdateCardRequest struct {
	Title       *string `json:"title,omitempty"`
	ColumnID    *int    `json:"column_id,omitempty"`
	LaneID      *int    `json:"lane_id,omitempty"`
	TypeID      *int    `json:"type_id,omitempty"`
	Description *string `json:"description,omitempty"`
	SizeText    *string `json:"size_text,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	Condition   *int    `json:"condition,omitempty"`
}

// ListCardsOptions defines optional filters/pagination for listing cards.
type ListCardsOptions struct {
	BoardID   int
	Condition int // 0=all, 1=active, 2=archived (only sent when > 0)
	Limit     int
	Offset    int
	OwnerID   int
	MemberID  int
}

// buildQuery builds URL query values from non-zero ListCardsOptions fields.
func (opts ListCardsOptions) buildQuery() url.Values {
	q := url.Values{}
	if opts.BoardID > 0 {
		q.Set("board_id", strconv.Itoa(opts.BoardID))
	}
	if opts.Condition > 0 {
		q.Set("condition", strconv.Itoa(opts.Condition))
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.OwnerID > 0 {
		q.Set("owner_id", strconv.Itoa(opts.OwnerID))
	}
	if opts.MemberID > 0 {
		q.Set("member_id", strconv.Itoa(opts.MemberID))
	}
	return q
}

// ListCards returns cards for a board with a simple boardID+condition interface.
// Deprecated: Use ListCardsWithOptions for richer filtering.
func (c *Client) ListCards(boardID int, condition int) ([]Card, error) {
	return c.ListCardsWithOptions(ListCardsOptions{BoardID: boardID, Condition: condition})
}

// ListCardsWithOptions returns cards with full filtering and pagination.
// GET /cards?board_id=... &condition=... &limit=... &offset=... &owner_id=... &member_id=...
func (c *Client) ListCardsWithOptions(opts ListCardsOptions) ([]Card, error) {
	q := opts.buildQuery()
	path := "/cards"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	data, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	var cards []Card
	if err := json.Unmarshal(data, &cards); err != nil {
		return nil, fmt.Errorf("parsing cards: %w", err)
	}
	return cards, nil
}

// ListCardsAllPages fetches all cards matching opts using automatic pagination.
// Default page size is 100 when opts.Limit is not set.
func (c *Client) ListCardsAllPages(opts ListCardsOptions) ([]Card, error) {
	if opts.Limit <= 0 {
		opts.Limit = 100
	}
	var all []Card
	for {
		cards, err := c.ListCardsWithOptions(opts)
		if err != nil {
			return nil, err
		}
		all = append(all, cards...)
		if len(cards) < opts.Limit {
			break
		}
		opts.Offset += opts.Limit
	}
	return all, nil
}

// GetCard returns a single card by ID.
// GET /cards/{card_id}
func (c *Client) GetCard(cardID int) (*Card, error) {
	data, err := c.Get(fmt.Sprintf("/cards/%d", cardID))
	if err != nil {
		return nil, err
	}
	var card Card
	if err := json.Unmarshal(data, &card); err != nil {
		return nil, fmt.Errorf("parsing card: %w", err)
	}
	return &card, nil
}

// CreateCard creates a new card.
// POST /cards
func (c *Client) CreateCard(req *CreateCardRequest) (*Card, error) {
	data, err := c.Post("/cards", req)
	if err != nil {
		return nil, err
	}
	var card Card
	if err := json.Unmarshal(data, &card); err != nil {
		return nil, fmt.Errorf("parsing card: %w", err)
	}
	return &card, nil
}

// UpdateCard updates an existing card.
// PATCH /cards/{card_id}
func (c *Client) UpdateCard(cardID int, req *UpdateCardRequest) (*Card, error) {
	data, err := c.Patch(fmt.Sprintf("/cards/%d", cardID), req)
	if err != nil {
		return nil, err
	}
	var card Card
	if err := json.Unmarshal(data, &card); err != nil {
		return nil, fmt.Errorf("parsing card: %w", err)
	}
	return &card, nil
}

// DeleteCard deletes a card.
// DELETE /cards/{card_id}
func (c *Client) DeleteCard(cardID int) error {
	_, err := c.Delete(fmt.Sprintf("/cards/%d", cardID))
	return err
}

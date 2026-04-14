package api

import (
	"encoding/json"
	"fmt"
)

// Card represents a Kaiten card.
type Card struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	BoardID     int    `json:"board_id"`
	ColumnID    int    `json:"column_id"`
	LaneID      int    `json:"lane_id,omitempty"`
	TypeID      *int   `json:"type_id,omitempty"`
	SizeText    string `json:"size_text,omitempty"`
	Condition   int    `json:"condition,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Blocked     bool   `json:"blocked,omitempty"`
	BlockReason string `json:"block_reason,omitempty"`
	CreatedAt   string `json:"created"`
	UpdatedAt   string `json:"updated"`
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

// ListCards returns cards for a board.
// GET /cards?board_id={board_id}
func (c *Client) ListCards(boardID int, condition int) ([]Card, error) {
	url := fmt.Sprintf("/cards?board_id=%d", boardID)
	if condition > 0 {
		url += fmt.Sprintf("&condition=%d", condition)
	}
	data, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	var cards []Card
	if err := json.Unmarshal(data, &cards); err != nil {
		return nil, fmt.Errorf("parsing cards: %w", err)
	}
	return cards, nil
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

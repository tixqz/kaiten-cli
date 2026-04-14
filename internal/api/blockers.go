package api

import (
	"encoding/json"
	"fmt"
)

// Blocker represents a card blocker.
type Blocker struct {
	ID            int    `json:"id"`
	Reason        string `json:"reason,omitempty"`
	CardID        int    `json:"card_id"`
	BlockerID     int    `json:"blocker_id"`
	BlockerCardID *int   `json:"blocker_card_id,omitempty"`
	Released      bool   `json:"released"`
	ReleasedByID  *int   `json:"released_by_id,omitempty"`
	DueDate       string `json:"due_date,omitempty"`
	CreatedAt     string `json:"created"`
	UpdatedAt     string `json:"updated"`
}

// BlockCardRequest is the payload for blocking a card.
type BlockCardRequest struct {
	Reason string `json:"reason,omitempty"`
}

// UpdateBlockerRequest is the payload for updating a blocker.
// reason and blocker_card_id are required by Kaiten API even on update.
type UpdateBlockerRequest struct {
	Reason        string `json:"reason"`
	BlockerCardID *int   `json:"blocker_card_id"`
	Released      *bool  `json:"released,omitempty"`
}

// BlockCard blocks a card.
// POST /cards/{card_id}/blockers
func (c *Client) BlockCard(cardID int, req *BlockCardRequest) (*Blocker, error) {
	data, err := c.Post(fmt.Sprintf("/cards/%d/blockers", cardID), req)
	if err != nil {
		return nil, err
	}
	var blocker Blocker
	if err := json.Unmarshal(data, &blocker); err != nil {
		return nil, fmt.Errorf("parsing blocker: %w", err)
	}
	return &blocker, nil
}

// UpdateBlocker updates a blocker (e.g. to release/unblock).
// PATCH /cards/{card_id}/blockers/{id}
func (c *Client) UpdateBlocker(cardID, blockerID int, req *UpdateBlockerRequest) (*Blocker, error) {
	data, err := c.Patch(fmt.Sprintf("/cards/%d/blockers/%d", cardID, blockerID), req)
	if err != nil {
		return nil, err
	}
	var blocker Blocker
	if err := json.Unmarshal(data, &blocker); err != nil {
		return nil, fmt.Errorf("parsing blocker: %w", err)
	}
	return &blocker, nil
}

// DeleteBlocker deletes a blocker.
// DELETE /cards/{card_id}/blockers/{id}
func (c *Client) DeleteBlocker(cardID, blockerID int) error {
	_, err := c.Delete(fmt.Sprintf("/cards/%d/blockers/%d", cardID, blockerID))
	return err
}

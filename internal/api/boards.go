package api

import (
	"encoding/json"
	"fmt"
)

// Board represents a Kaiten board.
type Board struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	SpaceID     int    `json:"space_id"`
}

// ListBoards returns all boards in a space.
// GET /spaces/{space_id}/boards
func (c *Client) ListBoards(spaceID int) ([]Board, error) {
	data, err := c.Get(fmt.Sprintf("/spaces/%d/boards", spaceID))
	if err != nil {
		return nil, err
	}
	var boards []Board
	if err := json.Unmarshal(data, &boards); err != nil {
		return nil, fmt.Errorf("parsing boards: %w", err)
	}
	return boards, nil
}

// GetBoard returns a single board by ID.
// GET /boards/{id}
func (c *Client) GetBoard(boardID int) (*Board, error) {
	data, err := c.Get(fmt.Sprintf("/boards/%d", boardID))
	if err != nil {
		return nil, err
	}
	var board Board
	if err := json.Unmarshal(data, &board); err != nil {
		return nil, fmt.Errorf("parsing board: %w", err)
	}
	return &board, nil
}

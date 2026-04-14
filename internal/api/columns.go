package api

import (
	"encoding/json"
	"fmt"
)

// Column represents a Kaiten board column.
type Column struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	BoardID int    `json:"board_id"`
	Sort    int    `json:"sort"`
}

// ListColumns returns all columns on a board.
// GET /boards/{board_id}/columns
func (c *Client) ListColumns(boardID int) ([]Column, error) {
	data, err := c.Get(fmt.Sprintf("/boards/%d/columns", boardID))
	if err != nil {
		return nil, err
	}
	var columns []Column
	if err := json.Unmarshal(data, &columns); err != nil {
		return nil, fmt.Errorf("parsing columns: %w", err)
	}
	return columns, nil
}

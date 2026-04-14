package api

import (
	"encoding/json"
	"fmt"
)

// Lane represents a Kaiten board lane (swimlane).
type Lane struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	BoardID int    `json:"board_id"`
	Sort    int    `json:"sort"`
}

// ListLanes returns all lanes on a board.
// GET /boards/{board_id}/lanes
func (c *Client) ListLanes(boardID int) ([]Lane, error) {
	data, err := c.Get(fmt.Sprintf("/boards/%d/lanes", boardID))
	if err != nil {
		return nil, err
	}
	var lanes []Lane
	if err := json.Unmarshal(data, &lanes); err != nil {
		return nil, fmt.Errorf("parsing lanes: %w", err)
	}
	return lanes, nil
}

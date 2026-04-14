package api

import (
	"encoding/json"
	"fmt"
)

// Member represents a card member.
type Member struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Type     int    `json:"type"`
}

// ListCardMembers returns members of a card.
// GET /cards/{card_id}/members
func (c *Client) ListCardMembers(cardID int) ([]Member, error) {
	data, err := c.Get(fmt.Sprintf("/cards/%d/members", cardID))
	if err != nil {
		return nil, err
	}
	var members []Member
	if err := json.Unmarshal(data, &members); err != nil {
		return nil, fmt.Errorf("parsing members: %w", err)
	}
	return members, nil
}

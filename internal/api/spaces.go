package api

import (
	"encoding/json"
	"fmt"
)

// Space represents a Kaiten space.
type Space struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// ListSpaces returns all spaces.
// GET /spaces
func (c *Client) ListSpaces() ([]Space, error) {
	data, err := c.Get("/spaces")
	if err != nil {
		return nil, err
	}
	var spaces []Space
	if err := json.Unmarshal(data, &spaces); err != nil {
		return nil, fmt.Errorf("parsing spaces: %w", err)
	}
	return spaces, nil
}

// GetSpace returns a single space by ID.
// GET /spaces/{space_id}
func (c *Client) GetSpace(spaceID int) (*Space, error) {
	data, err := c.Get(fmt.Sprintf("/spaces/%d", spaceID))
	if err != nil {
		return nil, err
	}
	var space Space
	if err := json.Unmarshal(data, &space); err != nil {
		return nil, fmt.Errorf("parsing space: %w", err)
	}
	return &space, nil
}
